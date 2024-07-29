package wpcrawler

import (
	"encoding/json"
	"fmt"
	"openproject-crawler/internal/httpclient"
	"sync"
)

type WPCrawler struct {
	*httpclient.APIClient
	projectName string
	params      map[string]interface{}
	data        []map[string]interface{}
	mu          sync.Mutex
	wg          sync.WaitGroup
}

func NewWPCrawler(apiURL, authToken, projectName string) (*WPCrawler, error) {
	apiClient, err := httpclient.NewAPIClient(apiURL, authToken)
	if err != nil {
		return nil, err
	}
	wpc := &WPCrawler{
		APIClient:   apiClient,
		projectName: projectName,
		params:      make(map[string]interface{}),
		data:        []map[string]interface{}{},
	}
	apiClient.SetURIPath(wpc.getUriPath())
	return wpc, nil
}

func (wpc *WPCrawler) getUriPath() string {
	if wpc.projectName != "" {
		return fmt.Sprintf("/projects/%s/work_packages", wpc.projectName)
	}
	return "/work_packages"
}

func (wpc *WPCrawler) SetProjectName(value string) {
	wpc.projectName = value
	wpc.SetURIPath(wpc.getUriPath())
}

func (wpc *WPCrawler) GetProjectName() string {
	return wpc.projectName
}

func (wpc *WPCrawler) SetParams(value map[string]interface{}) {
	wpc.params = value
	wpc.SetURIPath(wpc.getUriPath())
}

func (wpc *WPCrawler) GetParams() map[string]interface{} {
	return wpc.params
}

func (wpc *WPCrawler) fetchData(ch chan<- []byte) {
	params := wpc.params
	response, err := wpc.GetRequest("", params)
	if err != nil {
		ch <- nil
		fmt.Printf("failed to get data: %v\n", err)
		return
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		ch <- nil
		fmt.Printf("failed to parse JSON response: %v\n", err)
		return
	}
	elements, ok := result["_embedded"].(map[string]interface{})["elements"].([]interface{})
	if !ok {
		ch <- nil
		fmt.Printf("unexpected JSON structure\n")
		return
	}
	for _, element := range elements {
		if data, ok := element.(map[string]interface{}); ok {
			marshaledData, err := json.Marshal(data)
			if err != nil {
				ch <- nil
				fmt.Printf("failed to marshal data: %v\n", err)
				return
			}
			ch <- marshaledData
		}
	}
	ch <- nil
}

func (wpc *WPCrawler) FetchDataAsync() error {
	ch := make(chan []byte)
	wpc.wg.Add(1)
	go func() {
		defer wpc.wg.Done()
		wpc.fetchData(ch)
	}()

	go func() {
		wpc.wg.Wait()
		close(ch)
	}()

	var data []map[string]interface{}
	for marshaledData := range ch {
		if marshaledData == nil {
			continue
		}
		var unmarshaledData map[string]interface{}
		if err := json.Unmarshal(marshaledData, &unmarshaledData); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
		data = append(data, unmarshaledData)
	}

	wpc.mu.Lock()
	wpc.data = data
	wpc.mu.Unlock()

	return nil
}

func (wpc *WPCrawler) GetTasksIDAsync() ([]int, error) {
	err := wpc.FetchDataAsync()
	if err != nil {
		return nil, err
	}
	var ids []int
	wpc.mu.Lock()
	for _, val := range wpc.data {
		if id, ok := val["id"].(float64); ok {
			ids = append(ids, int(id))
		}
	}
	wpc.mu.Unlock()
	return ids, nil
}

func (wpc *WPCrawler) GetTasksAttrAsync() ([]map[string]interface{}, error) {
	err := wpc.FetchDataAsync()
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	tasksMap := make(map[string]interface{})
	wpc.mu.Lock()
	for _, val := range wpc.data {
		valChild := val["_links"].(map[string]interface{})
		tasksMap["taskName"] = val["subject"]
		tasksMap["taskAttr"] = map[string]interface{}{
			"id":       val["id"],
			"type":     valChild["type"].(map[string]interface{})["title"].(string),
			"priority": valChild["priority"].(map[string]interface{})["title"].(string),
			"status":   valChild["status"].(map[string]interface{})["title"].(string),
		}
		result = append(result, tasksMap)
	}
	wpc.mu.Unlock()
	return result, nil
}
