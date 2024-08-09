from openproject_crawler.config import logger
from openproject_crawler.http import SendAPIRequest

import asyncio
import aiohttp
import functools
from typing import Dict, Callable

def auto_process(*method_names: str):
  def decorator(func: Callable):
    @functools.wraps(func)
    async def async_wrapper(self, *args, **kwargs):
      tasks = [getattr(self, method_name)() for method_name in method_names if callable(getattr(self, method_name, None))]
      await asyncio.gather(*tasks)
      return await func(self, *args, **kwargs)
    return async_wrapper
  return decorator

class CrawlProjects(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, path_uri="/projects"):
    super().__init__(api_url, base64_token=base64_token)
    self.path_uri = path_uri
    self.data = {}

  async def get_data(self):
    for _ in range(3):
      try:
        response = await self.send_get_request()
        self.data = response['_embedded']['elements']
        return
      except aiohttp.ClientError:
        await asyncio.sleep(1)
    raise Exception("Failed to fetch data after several retries")

  @auto_process('get_data')
  async def total(self) -> int:
    return self.data['total']

  @auto_process('get_data')
  async def get_projects_id(self) -> Dict[int, str]:
    return {project['id']: project['identifier'] for project in self.data}