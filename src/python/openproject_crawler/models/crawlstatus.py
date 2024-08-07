from openproject_crawler.config.logging import logger
from openproject_crawler.http.http_request import SendAPIRequest

import asyncio
import aiohttp

class CrawlStatuses(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, path_uri="/statuses", params=None):
    super().__init__(api_url, base64_token=base64_token)
    self.path_uri = path_uri
    self.params = params

  async def fetch_data(self):
    for _ in range(3):
      try:
        response = await self.send_get_request(params=self.params)
        self.data = response['_embedded']['elements']
        return
      except aiohttp.ClientError as e:
        await asyncio.sleep(1)
    raise Exception("Failed to fetch data after several retries")

  async def initialize(self):
    await self.fetch_data()

  def __repr__(self) -> str:
    return str({val['id']: val['name'] for val in self.data})