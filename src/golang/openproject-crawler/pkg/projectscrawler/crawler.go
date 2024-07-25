package projectcrawler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"openproject-crawler/internal/httpclient"
)

type ProjectCrawler struct {
	*httpclient.APIClient
	data map[string]interface{}
	mu   sync.Mutex
}

func NewProjectCrawler(apiURL, authToken string) (*ProjectCrawler, error) {
	apiClient, err := httpclient.NewAPIClient(apiURL, authToken)
	if err != nil {
		return nil, err
	}
	apiClient.SetURIPath("/projects")
	return &ProjectCrawler{
		APIClient: apiClient,
		data:      make(map[string]interface{}),
	}, nil
}

func (pc *ProjectCrawler) FetchData() error {
	var wg sync.WaitGroup
	ch := make(chan error, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := pc.GetRequest("", nil)
			if err == nil {
				var responseData map[string]interface{}
				err = json.Unmarshal([]byte(response), &responseData)
				if err != nil {
					ch <- err
					return
				}
				pc.mu.Lock()
				defer pc.mu.Unlock()
				embedded, ok := responseData["_embedded"].(map[string]interface{})
				if !ok {
					ch <- fmt.Errorf("_embedded key not found or not a map")
					return
				}
				elements, ok := embedded["elements"].([]interface{})
				if !ok {
					ch <- fmt.Errorf("elements key not found or not a slice")
					return
				}
				pc.data = map[string]interface{}{
					"elements": elements,
				}
				ch <- nil
				return
			}
			time.Sleep(1 * time.Second)
			ch <- err
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for err := range ch {
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to fetch data after several retries")
}

func (pc *ProjectCrawler) GetTotalProjects() (int, error) {
	err := pc.FetchData()
	if err != nil {
		return 0, err
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	total, ok := pc.data["total"].(float64)
	if !ok {
		return 0, fmt.Errorf("total key not found in data")
	}
	return int(total), nil
}

func (pc *ProjectCrawler) GetProjectIDs() (map[int]string, error) {
	err := pc.FetchData()
	if err != nil {
		return nil, err
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	projectMap := make(map[int]string)
	elements, ok := pc.data["elements"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("elements key not found or not a slice")
	}
	for _, element := range elements {
		project, ok := element.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("element is not a map")
		}
		id, ok := project["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("id not found or not a float64")
		}
		identifier, ok := project["identifier"].(string)
		if !ok {
			return nil, fmt.Errorf("identifier not found or not a string")
		}
		projectMap[int(id)] = identifier
	}
	return projectMap, nil
}
