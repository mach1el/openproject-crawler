# openproject-crawler

![OpenProject](https://img.shields.io/badge/OpenProject-2D8CFF?style=for-the-badge&logo=openproject&logoColor=white)
![Python](https://img.shields.io/badge/python-3670A0?style=for-the-badge&logo=python&logoColor=ffdd54)
![Selenium](https://img.shields.io/badge/-selenium-%43B02A?style=for-the-badge&logo=selenium&logoColor=white)

This tool supports collecting data from OpenProject, forcing users to use the available API of OpenProject and additional Web Selenium for scraping more data, which the API doesn't support.

## Installation

To install the required dependencies, use:
```
pip install -r requirements.txt
```

## Important variables

- *`username`*: This variable could be change when collect data from API or from web portal, for the API value should be `apikey`, [check this](https://www.openproject.org/docs/api/introduction/#api-key-through-basic-auth) for more information. For the portal value should be the username you use to access the web portal
- *`password`*: Also like the username, for the API it must be access token, check [this note](https://www.openproject.org/docs/api/introduction/#api-key-through-basic-auth).
- *`api_url`*: The value should be `https://myopenproject.example/api/v3` (endswith `/api/v3`)
- *`portal_url`*: The value should be `https://myopenproject.example` (no need any uri path)

## Example to use

```
from openproject_crawler.api_crawl import CrawlProject
from openproject_crawler.credential import SetCredential

credential = SetCredential("apikey", "a6a2081fd40e89612e0d362753d3cf843974cfde7c67821c03b9c158dfc138d1")
project_crawler = CrawlProject(api_url="https://openproject.mich43l.io/api/v3", base64_token=credential.base64_token)
print(project_crawler.total()) #Get total projects available
print(project_crawler.get_project_id()) #Get dict of project's identifier and its ID
```
1. Set the credential for the API
2. The API request use `Basic Authorization` so we set the credential as base64 token

## License
![GitHub License](https://img.shields.io/github/license/mach1el/openproject-crawler?style=flat-square&color=%23FF5E0E)