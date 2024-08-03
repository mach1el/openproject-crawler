package crawlprojects

import (
	"encoding/json"
	"fmt"
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

type CrawlProjects struct {
	*httpclient.APIClient
	data map[string]interface{}
}

func NewCrawlProjects(apiURL, authToken string) (*CrawlProjects, error) {
	apiClient, err := httpclient.NewAPIClient(apiURL, authToken)
	if err != nil {
		return nil, err
	}
	apiClient.SetURIPath(projectsPath)
	return &CrawlProjects{
		APIClient: apiClient,
		data:      make(map[string]interface{}),
	}, nil
}

func (c *CrawlProjects) fetchData() error {
	response, err := c.GetRequest("", nil)
	if err != nil {
		return err
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &responseData); err != nil {
		return err
	}

	embedded, ok := responseData[embeddedKey].(map[string]interface{})
	if !ok || embedded[elementsKey] == nil {
		return fmt.Errorf("%s or %s key not found", embeddedKey, elementsKey)
	}

	c.data = map[string]interface{}{
		elementsKey: embedded[elementsKey],
		totalKey:    responseData[totalKey],
	}
	return nil
}

func (c *CrawlProjects) GetTotalProjects() (int, error) {
	if err := c.fetchData(); err != nil {
		return 0, err
	}

	total, ok := c.data[totalKey].(float64)
	if !ok {
		return 0, fmt.Errorf("%s key not found in data", totalKey)
	}

	return int(total), nil
}

func (c *CrawlProjects) GetProjectsID() (map[int]string, error) {
	if err := c.fetchData(); err != nil {
		return nil, err
	}

	projectMap := make(map[int]string)
	elements, ok := c.data[elementsKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s key not found or not a slice", elementsKey)
	}

	for _, element := range elements {
		project, ok := element.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("element is not a map")
		}

		id, ok := project[idKey].(float64)
		identifier, ok2 := project[identifierKey].(string)
		if !ok || !ok2 {
			return nil, fmt.Errorf("%s or %s not found or invalid", idKey, identifierKey)
		}

		projectMap[int(id)] = identifier
	}

	return projectMap, nil
}
