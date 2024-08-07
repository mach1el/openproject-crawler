from openproject_crawler.config.logging import logger
from openproject_crawler.core.url import UrlSetter

import time
import asyncio
import aiohttp
from aiohttp import ClientTimeout
from typing import Any, Dict, Optional

class SendAPIRequest(UrlSetter):
  param = {'pageSize': 1000}

  def __init__(self, api_url=None, base64_token=None, rate_limit_per_second=5):
    super().__init__(api_url)
    self._base64_token = base64_token
    self.rate_limit_per_second = rate_limit_per_second
    self.semaphore = asyncio.Semaphore(rate_limit_per_second)
    self.headers = {
      'Authorization': f'Basic {self._base64_token}' if self._base64_token else None,
      'Content-Type': 'application/json'
    }
    self.session = None
    self._last_request_time = 0

  @property
  def base64_token(self):
    return self._base64_token

  @base64_token.setter
  def base64_token(self, value: str):
    self._base64_token = value
    self.headers['Authorization'] = f'Basic {value}' if value else None

  async def _rate_limit(self):
   async with self.semaphore:
      current_time = time.monotonic()
      elapsed = current_time - self._last_request_time
      wait_time = max(0, (1 / self.rate_limit_per_second) - elapsed)
      if wait_time > 0:
          await asyncio.sleep(wait_time)
      self._last_request_time = time.monotonic()

  async def send_get_request(self, custom_uri: Optional[str] = None, params: Optional[Dict[str, Any]] = None) -> Any:
    if params is None:
      params = SendAPIRequest.param
    if custom_uri:
      self.path_uri = custom_uri

    url = self.selected_url

    if self.session is None:
      self.session = aiohttp.ClientSession(headers=self.headers)

    await self._rate_limit()

    for attempt in range(3):
      try:
        async with self.session.get(url, params=params, timeout=ClientTimeout(total=30)) as response:
          response.raise_for_status()
          return await response.json()
      except aiohttp.ClientError as e:
        logger.error(f"Error fetching {url}: {e} - Params: {params}")
        if attempt < 2:
          await asyncio.sleep(2 ** attempt)
        else:
          return None

  async def close_session(self):
    if self.session is not None:
      await self.session.close()
      self.session = None

  async def __aenter__(self):
    if self.session is None:
      self.session = aiohttp.ClientSession(headers=self.headers)
    return self

  async def __aexit__(self, exc_type, exc_val, exc_tb):
    await self.close_session()
