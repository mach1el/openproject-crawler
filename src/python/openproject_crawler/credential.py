import base64
from typing import Optional

class SetCredential(object):
  def __init__(self, username=None, password=None):
    self._username = username
    self._password = password

  @property
  def username(self) -> Optional[str]:
    return self._username
  
  @username.setter
  def username(self, value: str) -> None:
    self._username = value

  @property
  def password(self) -> Optional[str]:
    return self._password

  @password.setter
  def password(self, value) -> None:
    self._password = value

  @property
  def base64_token(self) -> str:
    if self._username is None or self._password is None: raise ValueError("Username and password must be set")
    base64_token = base64.b64encode(f'{self._username}:{self._password}'.encode()).decode('utf-8')
    return base64_token