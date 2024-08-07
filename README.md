# openproject-crawler

![OpenProject](https://img.shields.io/badge/OpenProject-2D8CFF?style=for-the-badge&logo=openproject&logoColor=white)
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Python](https://img.shields.io/badge/python-3670A0?style=for-the-badge&logo=python&logoColor=ffdd54)
![Selenium](https://img.shields.io/badge/-selenium-%43B02A?style=for-the-badge&logo=selenium&logoColor=white)

This tool supports collecting data from OpenProject, forcing users to use the available API of OpenProject and additional Web Selenium for scraping more data, which the API doesn't support. Scraping processes are using asynchronous programming to make it faster and stable.

## Installation

To install the required dependencies, use:
```bash
pip install -r requirements.txt
```

## Important variables

- *`username`*: This variable could be change when collect data from API or from web portal, for the API value should be `apikey`, [check this](https://www.openproject.org/docs/api/introduction/#api-key-through-basic-auth) for more information. For the portal value should be the username you use to access the web portal
- *`password`*: Also like the username, for the API it must be access token, check [this note](https://www.openproject.org/docs/api/introduction/#api-key-through-basic-auth).
- *`api_url`*: The value should be `https://myopenproject.example/api/v3` (endswith `/api/v3`)
- *`portal_url`*: The value should be `https://myopenproject.example` (no need any uri path)

## Example to use (Python)

For example, to use this module, I provide a script named [utils.py](src/python/sample/utils.py) to scrape data from a specific project. This will use the asynchronous method, execpt `DataParser`; it will use [ThreadPool](https://docs.python.org/3/library/concurrent.futures.html) instead. Hence, you need to setup it in an asynchronous way with *`async/await`* syntax. Give some explanation.

- *`Crawler`* class where to init crawler and get data such project's ID, project's tasks ID, tasks's activities
  - function `get_projects_id` -> Get all projects available and its ID
  - function `get_tasks_id` -> Get all tasks that belong to project `"my_project"` with filters parameters in HTTP request
  - function `get_tasks_activities_data` -> Scrape data from `work_packages/{id}`

### Setup Python venv to use the tool

* Navigate to the project source: 

```bash
cd /path/to/openproject-crawler/src/python/openproject_crawler
```

* Create a virutal environment:

```bash
python -m venv venv
```

* Active environment

  + `On Windows`

      ```
      .\venv\Scripts\activate
      ```
  + `Unix or MacOS`

      ```bash
      source venv/bin/activate
      ```
* Install the required dependencies:

```bash
pip install -e .
```

## Example to use (GoLang)

Given detail usage on [main.go](src/golang/openproject-crawler/cmd/main.go) as same as Python process, the flow is

`Get projects ID` -> `Get tasks ID of specific project` -> `Get tasks activities of specific project`

```bash
go run main.go
```

# Data structure

* Projects ID:
```json
{
  "1" : "mainproject",
  "2" : "demoproject"
}
```

* Tasks ID:

  * Golang data:

    ```go
    [45 278 13 225]
    ```

  * Python data:

    ```python
    [45, 278, 13, 225]
    ```

* Tasks activities:

```json
{
  "Task name": "Scraping data from openproject",
  "Task info": {
    "Project": "Data collection",
    "ID": "2",
    "Type": "Task",
    "Priority": "Normal",
    "Create date": "2024-06-09 15:12:26",
    "End Date": "2024-06-19 16:44:31",
    "Duration": "10 days"
  },
  "Task activities": [
    {
      "Datetime": "2024-06-19 16:44:31",
      "Action": [
        "Status changed from In progress to Closed"
      ]
    }
  ]
}
```

## License
![GitHub License](https://img.shields.io/github/license/mach1el/openproject-crawler?style=flat-square&color=%23FF5E0E)