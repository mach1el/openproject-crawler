package crawlact

import (
	"encoding/json"
	"fmt"
	"log"
	"openproject-crawler/internal/core"
	"openproject-crawler/internal/httpclient"
	"sync"
)

type CrawlActivities struct {
	*httpclient.APIClient
	*core.DataParser
	tasksID   []int
	params    map[string]interface{}
	tasksData [][]map[string]interface{}
	mu        sync.Mutex
	once      sync.Once
}

func NewCrawlActivities(apiURL, authToken string) (*CrawlActivities, error) {
	apiClient, err := httpclient.NewAPIClient(apiURL, authToken)
	if err != nil {
		return nil, err
	}
	parser := core.NewDataParser(nil)
	return &CrawlActivities{
		APIClient:  apiClient,
		DataParser: parser,
		tasksID:    make([]int, 0),
		params:     make(map[string]interface{}),
		tasksData:  [][]map[string]interface{}{},
	}, nil
}

func (c *CrawlActivities) GettasksID() []int {
	return c.tasksID
}

func (c *CrawlActivities) SettasksID(value []int) {
	c.tasksID = value
}

func (c *CrawlActivities) fetchData(taskID interface{}, ch chan<- []map[string]interface{}, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	customURI := fmt.Sprintf("/work_packages/%v/activities", taskID)
	response, err := c.GetRequest(customURI, c.params)
	if err != nil {
		errCh <- fmt.Errorf("failed to fetch data for task %v: %v", taskID, err)
		ch <- nil
		return
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		errCh <- fmt.Errorf("failed to parse JSON response for task %v: %v", taskID, err)
		ch <- nil
		return
	}

	if embedded, ok := jsonResponse["_embedded"].(map[string]interface{}); ok {
		if elements, ok := embedded["elements"].([]interface{}); ok {
			elementsMap := make([]map[string]interface{}, len(elements))
			for i, elem := range elements {
				elementsMap[i] = elem.(map[string]interface{})
			}
			ch <- elementsMap
			return
		} else {
			errCh <- fmt.Errorf("invalid 'elements' array for task %v", taskID)
		}
	} else {
		errCh <- fmt.Errorf("missing or invalid '_embedded' key for task %v", taskID)
	}
	ch <- nil
}

func (c *CrawlActivities) GetTasksActivities() ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	ch := make(chan []map[string]interface{}, len(c.tasksID))
	errCh := make(chan error, len(c.tasksID))

	for _, taskID := range c.tasksID {
		wg.Add(1)
		go c.fetchData(taskID, ch, errCh, &wg)
	}

	wg.Wait()
	close(ch)
	close(errCh)

	batchResults := make([][]map[string]interface{}, 0, len(c.tasksID))
	for elementsMap := range ch {
		if elementsMap != nil {
			batchResults = append(batchResults, elementsMap)
		}
	}

	c.mu.Lock()
	c.tasksData = append(c.tasksData, batchResults...)
	c.mu.Unlock()

	for err := range errCh {
		if err != nil {
			log.Printf("Error: %v\n", err)
		}
	}

	var mergedData []map[string]interface{}
	var mergeErr error
	c.once.Do(func() {
		c.SetDataInput(c.tasksData)
		mergedData, mergeErr = c.MergeData()
	})

	if mergeErr != nil {
		return nil, fmt.Errorf("error merging data: %v", mergeErr)
	}

	return mergedData, nil
}
