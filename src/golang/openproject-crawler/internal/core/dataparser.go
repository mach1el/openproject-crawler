package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type DataParser struct {
	dataInput     [][]map[string]interface{}
	TextFiltering map[string]string
}

func NewDataParser(dataInput [][]map[string]interface{}) *DataParser {
	if dataInput == nil {
		dataInput = [][]map[string]interface{}{}
	}
	return &DataParser{
		dataInput: dataInput,
		TextFiltering: map[string]string{
			"type":     "Type set to ",
			"project":  "Project set to ",
			"subject":  "Subject set to ",
			"priority": "Priority set to ",
		},
	}
}

func (dp *DataParser) GetDataInput() [][]map[string]interface{} {
	return dp.dataInput
}

func (dp *DataParser) SetDataInput(value [][]map[string]interface{}) {
	dp.dataInput = value
}

func convertTime(timestamp string) (string, error) {
	datetimeObj, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return "", fmt.Errorf("failed to parse timestamp: %v", err)
	}
	localTimezone := time.Now().Location()
	localDatetimeObj := datetimeObj.In(localTimezone)
	return localDatetimeObj.Format("2006-01-02 15:04:05"), nil
}

func calculateDuration(start, end string) (int, error) {
	const layout = "2006-01-02 15:04:05"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		return 0, fmt.Errorf("failed to parse start date: %v", err)
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		return 0, fmt.Errorf("failed to parse end date: %v", err)
	}
	duration := endDate.Sub(startDate)
	return int(duration.Hours() / 24), nil
}

func parseActivity(element map[string]interface{}) (map[string]interface{}, string, error) {
	var closedDate string
	activity := map[string]interface{}{
		"dateTime": "",
		"action":   []string{},
	}

	createdAt, ok := element["createdAt"].(string)
	if !ok {
		return nil, "", fmt.Errorf("missing or invalid 'createdAt' field")
	}

	dateTime, err := convertTime(createdAt)
	if err != nil {
		return nil, "", err
	}
	activity["dateTime"] = dateTime

	details, ok := element["details"].([]interface{})
	if !ok {
		return nil, "", fmt.Errorf("missing or invalid 'details' field")
	}

	for _, detail := range details {
		detailMap, ok := detail.(map[string]interface{})
		if !ok {
			continue
		}
		raw, rawOk := detailMap["raw"].(string)
		if rawOk {
			activity["action"] = append(activity["action"].([]string), raw)
			if raw == "Status changed from In progress to Closed" {
				closedDate = dateTime
			}
		}
	}
	return activity, closedDate, nil
}

func (dp *DataParser) parseTasksDetails(item interface{}) (map[string]string, error) {
	taskInfo := make(map[string]string)
	details, ok := item.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid 'details' field")
	}
	for _, detail := range details {
		raw, ok := detail.(map[string]interface{})["raw"].(string)
		if !ok {
			continue
		}
		for key, prefix := range dp.TextFiltering {
			if strings.Contains(raw, prefix) {
				taskInfo[key] = strings.Replace(raw, prefix, "", 1)
			}
		}
	}
	return taskInfo, nil
}

func (dp *DataParser) parseItem(item []map[string]interface{}) (map[string]interface{}, error) {
	var createdDate string
	var duration int
	mappedData := map[string]interface{}{
		"taskName": nil,
		"taskInfo": map[string]interface{}{
			"id":          nil,
			"project":     nil,
			"type":        nil,
			"priority":    nil,
			"createdDate": nil,
			"closedDate":  nil,
			"duration":    nil,
		},
		"taskActivities": []map[string]interface{}{},
	}

	for index, val := range item {
		if index == 0 {
			href, ok := val["_links"].(map[string]interface{})["workPackage"].(map[string]interface{})["href"].(string)
			if !ok {
				return nil, fmt.Errorf("missing or invalid 'href' field")
			}
			taskID := strings.Split(href, "/")[4]

			createdAt, ok := val["createdAt"].(string)
			if !ok {
				return nil, fmt.Errorf("missing or invalid 'createdAt' field")
			}

			var err error
			createdDate, err = convertTime(createdAt)
			if err != nil {
				return nil, err
			}

			tasksInfo, err := dp.parseTasksDetails(val["details"])
			if err != nil {
				return nil, err
			}

			mappedData["taskName"] = tasksInfo["subject"]
			mappedData["taskInfo"].(map[string]interface{})["id"] = taskID
			mappedData["taskInfo"].(map[string]interface{})["project"] = tasksInfo["project"]
			mappedData["taskInfo"].(map[string]interface{})["type"] = tasksInfo["type"]
			mappedData["taskInfo"].(map[string]interface{})["priority"] = tasksInfo["priority"]
			mappedData["taskInfo"].(map[string]interface{})["createdDate"] = createdDate
		} else {
			if val["_type"].(string) == "Activity" {
				activities, closedDate, err := parseActivity(val)
				if err != nil {
					return nil, err
				}
				if closedDate != "" {
					duration, err = calculateDuration(createdDate, closedDate)
					if err != nil {
						return nil, err
					}
					mappedData["taskInfo"].(map[string]interface{})["duration"] = fmt.Sprintf("%v days", duration)
					mappedData["taskInfo"].(map[string]interface{})["closedDate"] = closedDate
				}
				taskActivities := mappedData["taskActivities"].([]map[string]interface{})
				mappedData["taskActivities"] = append(taskActivities, activities)
			}
		}
	}
	return mappedData, nil
}

func (dp *DataParser) processItem(item []map[string]interface{}, wg *sync.WaitGroup, resultChan chan<- map[string]interface{}, errChan chan<- error) {
	defer wg.Done()
	parsedItem, err := dp.parseItem(item)
	if err != nil {
		errChan <- err
		return
	}
	resultChan <- parsedItem
}

func (dp *DataParser) MergeData() ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	result := []map[string]interface{}{}
	resultChan := make(chan map[string]interface{}, len(dp.dataInput))
	errChan := make(chan error, len(dp.dataInput))

	for _, item := range dp.dataInput {
		wg.Add(1)
		go dp.processItem(item, &wg, resultChan, errChan)
	}

	wg.Wait()
	close(resultChan)
	close(errChan)

	for res := range resultChan {
		result = append(result, res)
	}

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
