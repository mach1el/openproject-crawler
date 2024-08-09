import os
import sys
import base64
import unittest

from openproject_crawler.config import SetCredential

class TestSetCredential(unittest.TestCase):
  def test_initialization(self):
    cred = SetCredential("user", "pass")
    self.assertEqual(cred.username, "user")
    self.assertEqual(cred.password, "pass")
  
  def test_username_setter(self):
    cred = SetCredential()
    cred.username = "new_user"
    self.assertEqual(cred.username, "new_user")
  
  def test_password_setter(self):
    cred = SetCredential()
    cred.password = "new_pass"
    self.assertEqual(cred.password, "new_pass")
  
  def test_base64_token(self):
    cred = SetCredential("user", "pass")
    expected_token = base64.b64encode("user:pass".encode()).decode('utf-8')
    self.assertEqual(cred.base64_token, expected_token)

  def test_base64_token_without_username(self):
    cred = SetCredential(password="pass")
    with self.assertRaises(ValueError): _ = cred.base64_token

  def test_base64_token_without_password(self):
    cred = SetCredential(username="user")
    with self.assertRaises(ValueError): _ = cred.base64_token

if __name__ == '__main__': unittest.main()