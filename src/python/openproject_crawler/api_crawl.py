from .http_request import SendAPIRequest

class CrawlProject(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, path_uri="/projects"):
    super().__init__(api_url, base64_token=base64_token, path_uri=path_uri)
    self.data = self.get()

  def total(self) -> int:
    return self.data['total']
  
  def get_project_id(self) -> dict:
    projects = self.data['_embedded']['elements']
    return {project['identifier']:project['id'] for project in projects}
  
class CrawlWorkPackages(SendAPIRequest):
  def __init__(self, api_url=None, base64_token=None, project_name=None):
    super().__init__(api_url, base64_token=base64_token)
    self._project_name = project_name
    
    if self._project_name is None: self.path_uri = "/work_packages"
    else: self.path_uri = f"/projects/{self._project_name}/work_packages"
    
    self.data = self.get()['_embedded']['elements']

  @property
  def project_name(self):
    return self._project_name
  
  @project_name.setter
  def project_name(self, value):
    self.path_uri = f"/projects/{self._project_name}/work_packages" if value else "/work_packages"

  def get_tasks_id(self):
    return [ val['id'] for val in self.data ]
  
  def get_tasks_subject(self):
    return [ val['subject'] for val in self.data ]

  def get_tasks_subject_id(self):
    return {val['subject']:val['id'] for val in self.data}

  def get_tasks_attributes(self):
    result = {}

    for val in self.data:
      _child = val['_links']
      result.update({
        val['subject'] : {
          'id' : val['id'],
          'type' : _child['type']['title'],
          'priority' : _child['priority']['title'],
          'status' : _child['status']['title']
        }
      })

    return result

  def sum_tasks_type(self):
    type_list = [ val['_links']['type']['title'] for val in self.data ]
    result = { val:0 for val in type_list }
    for val in type_list: result[val] += 1
    return result
  
  def sum_tasks_status(self):
    status = [ val['_links']['status']['title'] for val in self.data ]
    result = { val:0 for val in status }
    for val in status: result[val] += 1
    return result