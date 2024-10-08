import json
import asyncio
from typing import Any, Dict
from openproject_crawler.config import SetCredential
from openproject_crawler.models import (
  CrawlProjects,
  CrawlWorkPackages,
  CrawlActivities
)

class Crawler(object):
  def __init__(self):
    self.url = "https://openproject.example.com/api/v3"
    self.credential = SetCredential("apikey", "a6a2081fd40e89612e0d362753d3cf843974cfde7c67821c03b9c851dfc1")
    self.base64_token = self.credential.base64_token

  async def get_projects_id(self) -> Dict[int, str]:
    crawler = CrawlProjects(api_url=self.url, base64_token=self.base64_token)
    return await crawler.get_projects_id()

  async def get_tasks_id(self, project_name="viclass", closed=False) -> Any:
    
    if closed:
      params = {
        'pageSize': 1000,
        'filters': json.dumps([{ "status_id": { "operator": "=", "values": ["12"] }}])
      }
    
    if project_name:
      projects_id = await self.get_projects_id()
      project_id = next( ( pid for pid, value in projects_id.items() if value == project_name ), None )
      
      params = {
      'pageSize': 1000,
      'filters': json.dumps([{ "project": { "operator": "=", "values": [project_id] }}])
    }

    crawler = CrawlWorkPackages(api_url=self.url, base64_token=self.base64_token, params=params)
    
    return await crawler.get_tasks_id()
  
  async def get_tasks_activities_data(self):
    crawler = CrawlActivities(api_url=self.url, base64_token=self.base64_token)
    crawler.tasks_id = await self.get_tasks_id(project_name="viclass")
    data = await crawler.get_tasks_activities()
    return data

def main():
  crawler = Crawler()
  loop = asyncio.get_event_loop()
  if loop.is_running():
    future = asyncio.ensure_future(crawler.get_tasks_activities_data())
    loop.run_until_complete(future)
    data = future.result()
  else:
    data = loop.run_until_complete(crawler.get_tasks_activities_data())
  print(data)

if __name__ == "__main__":
  main()