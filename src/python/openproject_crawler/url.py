from typing import Optional

class UrlSetter(object):
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

    if self._api_url is not None:
      base_url = self._api_url
    elif self._portal_url is not None:
      base_url = self._portal_url
    else:
      raise ValueError("Need to set API url or Web portal URL!")

    if self._path_uri:
      return f"{base_url.rstrip('/')}/{self._path_uri.lstrip('/')}"

    return base_url