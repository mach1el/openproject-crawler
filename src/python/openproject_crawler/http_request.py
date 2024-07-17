import requests
from typing import Optional, Dict, Any
   
class SetURL(object):
  def __init__(self, api_url: Optional[str] = None, portal_url: Optional[str] = None, path_uri: Optional[str] = None):
    self._api_url = api_url
    self._portal_url = portal_url
    self._path_uri = path_uri

  @property
  def api_url(self) -> Optional[str]:
    return self._api_url

  @api_url.setter
  def api_url(self, value: str) -> None:
    self._api_url = value

  @property
  def portal_url(self) -> Optional[str]:
    return self._portal_url

  @portal_url.setter
  def portal_url(self, value: str) -> None:
    self._portal_url = value

  @property
  def path_uri(self) -> Optional[str]:
    return self._path_uri

  @path_uri.setter
  def path_uri(self, value: str) -> None:
    self._path_uri = value

  @property
  def selected_url(self) -> str:
    base_url = None

    if self._api_url is not None: base_url = self._api_url
    elif self._portal_url is not None: base_url = self._portal_url
    else: raise ValueError("Need to set API url or Web portal URL!")

    if self._path_uri: return f"{base_url.rstrip('/')}/{self._path_uri.lstrip('/')}"

    return base_url
  
class SendAPIRequest(SetURL):

  param = { 'pageSize' : 100 }

  def __init__(self, api_url=None, path_uri=None, base64_token=None):
    super().__init__(api_url, path_uri)
    self.path_uri = path_uri
    self._base64_token = base64_token
    
  @property
  def base64_token(self):
    return self._base64_token
  
  @base64_token.setter
  def base64_token(self, value):
    self._base64_token = value

  def get(self, param: Optional[Dict[str, Any]] = None) -> Any:
    if param is None: param = SendAPIRequest.param
    if self._base64_token is not None:
      headers = {
        'Authorization': f'Basic {self._base64_token}',
        'Content-Type': 'application/json'
      }
    else: raise ValueError("Need to set username:password as base64 format!")

    response = requests.get(self.selected_url, params=param, headers=headers)
    response.raise_for_status()
    return response.json()