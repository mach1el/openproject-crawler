from openproject_crawler.config import logger

import os
import traceback

from datetime import datetime
from concurrent.futures import ThreadPoolExecutor, as_completed

class DataParser(object):
  def __init__(self, data_input=None):
    self._data_input = data_input or []
    self.text_filtering = {
      "type" : "Type set to ",
      "project" : "Project set to ",
      "priority" : "Priority set to "
    }

  @property
  def data_input(self):
    return self._data_input
  
  @data_input.setter
  def data_input(self, value):
    if isinstance(value, dict):
      self._data_input = [value]
    elif isinstance(value, list):
      self._data_input = value
    else:
      raise ValueError("Only accept list or dict format!")

  def convert_time(self, timestamp) -> str:
    datetime_obj = datetime.strptime(timestamp, '%Y-%m-%dT%H:%M:%S.%fZ')
    formatted_date_str = datetime_obj.strftime('%Y-%m-%d %H:%M:%S')
    return formatted_date_str
  
  def calculate_duration(self, start, end) -> str:
    start_date = datetime.strptime(start, '%Y-%m-%d %H:%M:%S')
    end_date = datetime.strptime(end, '%Y-%m-%d %H:%M:%S')
    duration = end_date - start_date
    return duration.days
  
  def parse_activity(self, element):
    activity = {
      'Datetime': self.convert_time(element['createdAt']),
      'Action': [detail['raw'] for detail in element['details']]
    }
    closed_date = 'null'
    for detail in element['details']:
      if detail['raw'] == 'Status changed from In progress to Closed':
        closed_date = self.convert_time(element['createdAt'])
    return activity, closed_date

  def parse_task_details(self, details):
    task_info = {}
    for detail in details:
      if self.text_filtering['type'] in detail['raw']:
        task_info['Type'] = detail['raw'].replace(self.text_filtering['type'], '')
      if self.text_filtering['project'] in detail['raw']:
        task_info['Project'] = detail['raw'].replace(self.text_filtering['project'], '')
      if self.text_filtering['priority'] in detail['raw']:
        task_info['Priority'] = detail['raw'].replace(self.text_filtering['priority'], '')
    return task_info

  def parse_item(self, item):
    try:
      activities = []
      closed_date = 'null'
      task_id = item[0]['_links']['workPackage']['href'].split('/')[4]
      subject = item[0]['_links']['workPackage']['title']
      created_date = self.convert_time(item[0]['createdAt'])

      for index, element in enumerate(item):
        if index == 0:
          task_info = self.parse_task_details(item[0]['details'])
        else: 
          if element['_type'] == 'Activity':
            activity, closed = self.parse_activity(element)
            activities.append(activity)
            if closed != 'null':
              closed_date = closed

      duration = "null" if closed_date == "null" else f"{self.calculate_duration(created_date, closed_date)} days"

      return {
        "Task name": subject,
        "Task info": {
          "Project": task_info.get('Project', ''),
          "ID": task_id,
          "Type": task_info.get('Type', ''),
          "Priority": task_info.get('Priority', ''),
          "Create date": created_date,
          "End Date": closed_date,
          "Duration": duration
        },
        "Task activities": activities
      }
    except Exception as exc:
      logger.error(f"Error parsing item: {item} - Exception: {exc}")
      logger.error(traceback.format_exc())
      return None

  def merge_data(self) -> list:
    result = []
    cpu_cores = os.cpu_count()
    max_workers = cpu_cores * 2
    
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
      future_to_item = {executor.submit(self.parse_item, item): item for item in self._data_input}
      for future in as_completed(future_to_item):
        try:
          result.append(future.result())
        except Exception as exc:
          logger.error(f"Error parsing item: {exc}")

    return result
