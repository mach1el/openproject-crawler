from openproject_crawler.config.logging import logger
from openproject_crawler.http.http_request import SendAPIRequest

import asyncio
import aiohttp
import functools
from typing import Any, Dict, List, Optional, Callable

def auto_process(*method_names: str):
  def decorator(func: Callable):
    @functools.wraps(func)
    async def async_wrapper(self, *args, **kwargs):
      tasks = [getattr(self, method_name)() for method_name in method_names if callable(getattr(self, method_name, None))]
      await asyncio.gather(*tasks)
      return await func(self, *args, **kwargs)
    return async_wrapper
  return decorator

class CrawlWorkPackages(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, project_name=None, params: Optional[Dict[str, Any]] = None):
    super().__init__(api_url, base64_token=base64_token)
    self._project_name = project_name
    self.params = params
    self.path_uri = self._get_path_uri()
    self.data = []

  async def get_data(self):
    for _ in range(3):
      try:
        response = await self.send_get_request(params=self.params)
        self.data = response['_embedded']['elements']
        return
      except aiohttp.ClientError:
        await asyncio.sleep(1)
    raise Exception("Failed to fetch data after several retries")

  async def get_data_if_needed(self):
    if not self.data:
      await self.get_data()

  def _get_path_uri(self):
    if self._project_name:
      return f"/projects/{self._project_name}/work_packages"
    return "/work_packages"

  @property
  def project_name(self):
    return self._project_name

  @project_name.setter
  def project_name(self, value):
    self._project_name = value
    self.path_uri = self._get_path_uri()

  @auto_process('get_data')
  async def get_tasks_id(self) -> List[int]:
    try:
      await self.get_data_if_needed()
      return [val['id'] for val in self.data]
    except Exception as e:
      logger.error(f"Error in get_tasks_id: {e}")
      return []

  @auto_process('get_data')
  async def get_tasks_subject(self) -> List[str]:
    try:
      await self.get_data_if_needed()
      return [val['subject'] for val in self.data]
    except Exception as e:
      logger.error(f"Error in get_tasks_subject: {e}")
      return []

  @auto_process('get_data')
  async def get_tasks_subject_id(self) -> Dict[int, str]:
    try:
      await self.get_data_if_needed()
      return {val['id']: val['subject'] for val in self.data}
    except Exception as e:
      logger.error(f"Error in get_tasks_subject_id: {e}")
      return {}

  @auto_process('get_data')
  async def get_tasks_attributes(self) -> Dict[str, Dict[str, Any]]:
    try:
      await self.get_data_if_needed()
      result = {}
      for val in self.data:
        _child = val['_links']
        result[val['subject']] = {
          'id': val['id'],
          'type': _child['type']['title'],
          'priority': _child['priority']['title'],
          'status': _child['status']['title']
        }
      return result
    except Exception as e:
      logger.error(f"Error in get_tasks_attributes: {e}")
      return {}

  @auto_process('get_data')
  async def sum_tasks_type(self) -> Dict[str, int]:
    try:
      await self.get_data_if_needed()
      type_list = [val['_links']['type']['title'] for val in self.data]
      result = {val: 0 for val in type_list}
      for val in type_list:
        result[val] += 1
      return result
    except Exception as e:
      logger.error(f"Error in sum_tasks_type: {e}")
      return {}

  @auto_process('get_data')
  async def sum_tasks_status(self) -> Dict[str, int]:
    try:
      await self.get_data_if_needed()
      status_list = [val['_links']['status']['title'] for val in self.data]
      result = {val: 0 for val in status_list}
      for val in status_list:
        result[val] += 1
      return result
    except Exception as e:
      logger.error(f"Error in sum_tasks_status: {e}")
      return {}

  async def __aexit__(self, exc_type, exc_val, exc_tb):
    await self.close_session()