from .http_request import SendAPIRequest
import asyncio
import aiohttp
import logging
import functools
from typing import Any, Dict, List, Union, Optional, Callable

logger = logging.getLogger(__name__)

def auto_process(*method_names: str):
  def decorator(func: Callable):
    @functools.wraps(func)
    async def async_wrapper(self, *args, **kwargs):
      tasks = [getattr(self, method_name)() for method_name in method_names if callable(getattr(self, method_name, None))]
      await asyncio.gather(*tasks)
      return await func(self, *args, **kwargs)
    return async_wrapper
  return decorator

class CrawlStatuses(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, path_uri="/statuses", params=None):
    super().__init__(api_url, base64_token=base64_token)
    self.path_uri = path_uri
    self.params = params

  async def fetch_data(self):
    for _ in range(3):
      try:
        response = await self.get(params=self.params)
        self.data = response['_embedded']['elements']
        return
      except aiohttp.ClientError as e:
        await asyncio.sleep(1)
    raise Exception("Failed to fetch data after several retries")

  async def initialize(self):
      await self.fetch_data()

  def __repr__(self) -> str:
      return str({val['id']: val['name'] for val in self.data})

class CrawlProjects(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, path_uri="/projects"):
    super().__init__(api_url, base64_token=base64_token)
    self.path_uri = path_uri
    self.data = {}

  async def get_data(self):
    for _ in range(3):
      try:
        response = await self.get()
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
        response = await self.get(params=self.params)
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
    await self.get_data_if_needed()
    return [val['id'] for val in self.data]

  @auto_process('get_data')
  async def get_tasks_subject(self) -> List[str]:
    await self.get_data_if_needed()
    return [val['subject'] for val in self.data]

  @auto_process('get_data')
  async def get_tasks_subject_id(self) -> Dict[int, str]:
    await self.get_data_if_needed()
    return {val['id']: val['subject'] for val in self.data}

  @auto_process('get_data')
  async def get_tasks_attributes(self) -> Dict[str, Dict[str, Any]]:
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

  @auto_process('get_data')
  async def sum_tasks_type(self) -> Dict[str, int]:
    await self.get_data_if_needed()
    type_list = [val['_links']['type']['title'] for val in self.data]
    result = {val: 0 for val in type_list}
    for val in type_list:
      result[val] += 1
    return result

  @auto_process('get_data')
  async def sum_tasks_status(self) -> Dict[str, int]:
    await self.get_data_if_needed()
    status_list = [val['_links']['status']['title'] for val in self.data]
    result = {val: 0 for val in status_list}
    for val in status_list:
      result[val] += 1
    return result

  async def __aexit__(self, exc_type, exc_val, exc_tb):
    await self.close_session()

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
    else:
        raise ValueError("Tasks id must be a single int, str, or a list of ids!")

  async def fetch_data(self, task_id: Union[int, str]) -> Any:
    return await self.get(f"/work_packages/{task_id}/activities", self.params)

  @auto_process('check_var')
  async def get_full_attributes(self) -> Any:
    tasks = [self.fetch_data(task_id) for task_id in self._tasks_id]
    responses = await asyncio.gather(*tasks)

    for response in responses:
      if response:
        self.tasks_data.append(response['_embedded']['elements'])
    return self.tasks_data

  async def __aexit__(self, exc_type, exc_val, exc_tb):
    await self.close_session()
