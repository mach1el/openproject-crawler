package projectcrawler

import (
	"encoding/json"
	"fmt"
	"sync"

	"openproject-crawler/internal/httpclient"
)

const (
	projectsPath  = "/projects"
	elementsKey   = "elements"
	embeddedKey   = "_embedded"
	totalKey      = "total"
	idKey         = "id"
	identifierKey = "identifier"
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
	apiClient.SetURIPath(projectsPath)
	return &ProjectCrawler{
		APIClient: apiClient,
		data:      make(map[string]interface{}),
	}, nil
}

func (pc *ProjectCrawler) fetchDataOnce(ch chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			ch <- fmt.Errorf("panic occurred: %v", r)
		}
	}()
	response, err := pc.GetRequest("", nil)
	if err != nil {
		ch <- err
		return
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &responseData); err != nil {
		ch <- err
		return
	}

	embedded, ok := responseData[embeddedKey].(map[string]interface{})
	if !ok {
		ch <- fmt.Errorf("%s key not found or not a map", embeddedKey)
		return
	}
	elements, ok := embedded[elementsKey].([]interface{})
	if !ok {
		ch <- fmt.Errorf("%s key not found or not a slice", elementsKey)
		return
	}

	pc.mu.Lock()
	pc.data = map[string]interface{}{
		elementsKey: elements,
		totalKey:    responseData[totalKey],
	}
	pc.mu.Unlock()
	ch <- nil
}

func (pc *ProjectCrawler) FetchData() error {
	var wg sync.WaitGroup
	ch := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		pc.fetchDataOnce(ch)
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	for err := range ch {
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to fetch data")
}

func (pc *ProjectCrawler) GetTotalProjects() (int, error) {
	if err := pc.FetchData(); err != nil {
		return 0, err
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	total, ok := pc.data[totalKey].(float64)
	if !ok {
		return 0, fmt.Errorf("%s key not found in data", totalKey)
	}

	return int(total), nil
}

func (pc *ProjectCrawler) GetProjectIDs() (map[int]string, error) {
	if err := pc.FetchData(); err != nil {
		return nil, err
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	projectMap := make(map[int]string)
	elements, ok := pc.data[elementsKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s key not found or not a slice", elementsKey)
	}

	for _, element := range elements {
		project, ok := element.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("element is not a map")
		}

		id, ok := project[idKey].(float64)
		if !ok {
			return nil, fmt.Errorf("%s not found or not a float64", idKey)
		}

		identifier, ok := project[identifierKey].(string)
		if !ok {
			return nil, fmt.Errorf("%s not found or not a string", identifierKey)
		}

		projectMap[int(id)] = identifier
	}

	return projectMap, nil
}
