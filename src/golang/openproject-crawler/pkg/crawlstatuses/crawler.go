package crawlstatuses

import (
	"encoding/json"
	"fmt"
	"openproject-crawler/internal/httpclient"
)

const (
	statusPath  = "/statuses"
	elementsKey = "elements"
	embeddedKey = "_embedded"
	idKey       = "id"
	nameKey     = "name"
)

type CrawlStatuses struct {
	*httpclient.APIClient
	data map[string]interface{}
}

func NewCrawlStatuses(apiURL, authToken string) (*CrawlStatuses, error) {
	apiClient, err := httpclient.NewAPIClient(apiURL, authToken)
	if err != nil {
		return nil, err
	}
	apiClient.SetURIPath(statusPath)
	return &CrawlStatuses{
		APIClient: apiClient,
		data:      make(map[string]interface{}),
	}, nil
}

func (c *CrawlStatuses) FetchData() error {
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
	}
	return nil
}

func (c *CrawlStatuses) Initialize() error {
	return c.FetchData()
}

func (c *CrawlStatuses) String() string {
	elements, ok := c.data[elementsKey].([]interface{})
	if !ok {
		return "Invalid data format"
	}

	statusMap := make(map[int]string)
	for _, element := range elements {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			continue
		}
		id, idOk := elementMap["id"].(float64)
		name, nameOk := elementMap["name"].(string)
		if idOk && nameOk {
			statusMap[int(id)] = name
		}
	}

	result, _ := json.Marshal(statusMap)
	return string(result)
}
