from openproject_crawler.config.logging import logger
from openproject_crawler.http.http_request import SendAPIRequest
from openproject_crawler.data.parser import DataParser

import asyncio
import functools
from typing import Any, List, Union, Callable

def auto_process(*method_names: str):
  def decorator(func: Callable):
    @functools.wraps(func)
    async def async_wrapper(self, *args, **kwargs):
      tasks = [getattr(self, method_name)() for method_name in method_names if callable(getattr(self, method_name, None))]
      await asyncio.gather(*tasks)
      return await func(self, *args, **kwargs)
    return async_wrapper
  return decorator

class CrawlActivities(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, tasks_id=None, params=None):
    super().__init__(api_url, base64_token=base64_token)
    self._tasks_id = tasks_id
    self.params = params
    self.tasks_data = []

  async def check_var(self):
    if self._tasks_id is None:
      raise ValueError("No tasks ID found, please set it")

  @property
  def tasks_id(self):
    return self._tasks_id

  @tasks_id.setter
  def tasks_id(self, value: Union[int, str, List[Union[int, str]]]):
    if isinstance(value, (str, int)):
        self._tasks_id = [value]
    elif isinstance(value, list):
        self._tasks_id = value

  async def fetch_data(self, task_id: Union[int, str]) -> Any:
    return await self.send_get_request(f"/work_packages/{task_id}/activities", self.params)

  @auto_process('check_var')
  async def get_tasks_activities(self) -> Any:
    tasks = [self.fetch_data(task_id) for task_id in self._tasks_id]
    try:
      responses = await asyncio.gather(*tasks, return_exceptions=True)
    except Exception as e:
      logger.error(f"Error while gathering tasks: {e}")
      return None

    for response in responses:
      if response:
        self.tasks_data.append(response['_embedded']['elements'])
    
    data_parser = DataParser(data_input=self.tasks_data)
    parsed_data = data_parser.merge_data()
    return parsed_data

  async def __aexit__(self, exc_type, exc_val, exc_tb):
    await self.close_session()