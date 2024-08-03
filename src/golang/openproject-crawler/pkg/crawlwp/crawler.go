package crawlwp

import (
	"encoding/json"
	"fmt"
	"openproject-crawler/internal/httpclient"
	"sync"
)

type CrawlWorkPackages struct {
	*httpclient.APIClient
	projectName string
	params      map[string]interface{}
	data        []map[string]interface{}
	mu          sync.Mutex
	wg          sync.WaitGroup
}

func NewCrawlWorkPackages(apiURL, authToken, projectName string) (*CrawlWorkPackages, error) {
	apiClient, err := httpclient.NewAPIClient(apiURL, authToken)
	if err != nil {
		return nil, err
	}
	c := &CrawlWorkPackages{
		APIClient:   apiClient,
		projectName: projectName,
		params:      make(map[string]interface{}),
		data:        []map[string]interface{}{},
	}
	apiClient.SetURIPath(c.getUriPath())
	return c, nil
}

func (c *CrawlWorkPackages) getUriPath() string {
	if c.projectName != "" {
		return fmt.Sprintf("/projects/%s/work_packages", c.projectName)
	}
	return "/work_packages"
}

func (c *CrawlWorkPackages) SetProjectName(value string) {
	c.projectName = value
	c.SetURIPath(c.getUriPath())
}

func (c *CrawlWorkPackages) GetProjectName() string {
	return c.projectName
}

func (c *CrawlWorkPackages) SetParams(value map[string]interface{}) {
	c.params = value
	c.SetURIPath(c.getUriPath())
}

func (c *CrawlWorkPackages) GetParams() map[string]interface{} {
	return c.params
}

func (c *CrawlWorkPackages) fetchData(ch chan<- map[string]interface{}) {
	defer c.wg.Done()
	params := c.params
	response, err := c.GetRequest("", params)
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
			ch <- data
		}
	}
}

func (c *CrawlWorkPackages) FetchDataAsync() error {
	ch := make(chan map[string]interface{})
	c.wg.Add(1)
	go c.fetchData(ch)

	go func() {
		c.wg.Wait()
		close(ch)
	}()

	var data []map[string]interface{}
	for marshaledData := range ch {
		if marshaledData == nil {
			continue
		}
		data = append(data, marshaledData)
	}

	c.mu.Lock()
	c.data = data
	c.mu.Unlock()

	return nil
}

func (c *CrawlWorkPackages) GetTasksID() ([]int, error) {
	if err := c.FetchDataAsync(); err != nil {
		return nil, err
	}

	var ids []int
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, val := range c.data {
		if id, ok := val["id"].(float64); ok {
			ids = append(ids, int(id))
		}
	}
	return ids, nil
}

func (c *CrawlWorkPackages) GetTasksAttr() ([]map[string]interface{}, error) {
	if err := c.FetchDataAsync(); err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, val := range c.data {
		valChild := val["_links"].(map[string]interface{})
		result = append(result, map[string]interface{}{
			"taskName": val["subject"],
			"taskAttr": map[string]interface{}{
				"id":       val["id"],
				"type":     valChild["type"].(map[string]interface{})["title"].(string),
				"priority": valChild["priority"].(map[string]interface{})["title"].(string),
				"status":   valChild["status"].(map[string]interface{})["title"].(string),
			},
		})
	}
	return result, nil
}

func (c *CrawlWorkPackages) sumTasks(attribute string) (map[string]int, error) {
	if err := c.FetchDataAsync(); err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, val := range c.data {
		valChild := val["_links"].(map[string]interface{})
		attrValue := valChild[attribute].(map[string]interface{})["title"].(string)
		counts[attrValue]++
	}
	return counts, nil
}

func (c *CrawlWorkPackages) SumTasksType() (map[string]int, error) {
	return c.sumTasks("type")
}

func (c *CrawlWorkPackages) SumTasksPriority() (map[string]int, error) {
	return c.sumTasks("priority")
}

func (c *CrawlWorkPackages) SumTasksStatus() (map[string]int, error) {
	return c.sumTasks("status")
}
