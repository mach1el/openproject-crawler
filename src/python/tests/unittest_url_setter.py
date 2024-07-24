import os
import sys
import unittest

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

from openproject_crawler.http_request import SetURL

class TestSetURL(unittest.TestCase):
  def setUp(self):
    self.url_instance = SetURL()

  def test_api_url(self):
    self.url_instance.api_url = "https://api.example.com"
    self.assertEqual(self.url_instance.api_url, "https://api.example.com")

  def test_portal_url(self):
    self.url_instance.portal_url = "https://portal.example.com"
    self.assertEqual(self.url_instance.portal_url, "https://portal.example.com")

  def test_path_uri(self):
    self.url_instance.path_uri = "path/to/resource"
    self.assertEqual(self.url_instance.path_uri, "path/to/resource")

  def test_selected_url_with_api_url(self):
    self.url_instance.api_url = "https://api.example.com"
    self.assertEqual(self.url_instance.selected_url, "https://api.example.com")

  def test_selected_url_with_portal_url_and_path_uri(self):
    self.url_instance.portal_url = "https://portal.example.com"
    self.url_instance.path_uri = "path/to/resource"
    self.assertEqual(self.url_instance.selected_url, "https://portal.example.com/path/to/resource")

  def test_selected_url_no_base_url_set(self):
    with self.assertRaises(ValueError): self.url_instance.selected_url

if __name__ == '__main__': unittest.main()