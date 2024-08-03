package main

import (
	"encoding/json"
	"fmt"
	"log"
	"openproject-crawler/internal/credential"
	"openproject-crawler/pkg/crawlact"
	"openproject-crawler/pkg/crawlprojects"
	"openproject-crawler/pkg/crawlwp"
)

type Crawler struct {
	*crawlprojects.CrawlProjects
	*crawlwp.CrawlWorkPackages
	*crawlact.CrawlActivities
	authToken string
}

func (c *Crawler) NewCrawler(apiURL, username, password string) (*Crawler, error) {
	cred, err := credential.SetCredential(username, password)
	if err != nil {
		return nil, err
	}
	c.authToken = cred.GenerateToken()

	crawlProject, err := crawlprojects.NewCrawlProjects(apiURL, c.authToken)
	if err != nil {
		return nil, err
	}

	crawlWorkPackages, err := crawlwp.NewCrawlWorkPackages(apiURL, c.authToken, "")
	if err != nil {
		return nil, err
	}

	crawlAct, err := crawlact.NewCrawlActivities(apiURL, c.authToken)
	if err != nil {
		return nil, err
	}

	return &Crawler{
		CrawlProjects:     crawlProject,
		CrawlWorkPackages: crawlWorkPackages,
		CrawlActivities:   crawlAct,
		authToken:         c.authToken,
	}, nil
}

func (c *Crawler) crawlProjectsID() (map[int]string, error) {
	IDs, err := c.GetProjectsID()
	if err != nil {
		return nil, err
	}
	return IDs, nil
}

func (c *Crawler) crawlTasksID(projectName string) ([]int, error) {
	projectsID, err := c.crawlProjectsID()
	if err != nil {
		return nil, err
	}

	validID, found := func(mappedID map[int]string) (int, bool) {
		for k, v := range mappedID {
			if v == projectName {
				return k, true
			}
		}
		return 0, false
	}(projectsID)
	if !found {
		return nil, fmt.Errorf("found no valid project name")
	}

	params := make(map[string]interface{})
	filters := []map[string]interface{}{
		{
			"project": map[string]interface{}{
				"operator": "=",
				"values":   []int{validID},
			},
		},
	}
	filtersJSON, _ := json.Marshal(filters)
	params["pageSize"] = "1000"
	params["filters"] = string(filtersJSON)
	c.SetParams(params)

	tasksID, err := c.GetTasksID()
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}
	return tasksID, nil
}

func main() {
	apiURL := "https://myopenproject.com/api/v3"
	username := "apikey"
	password := "a6a2081fd40e89612e0d362753d3cf843974cfde7c67821c03b9c851dfc1"
	projectName := "viclass"

	crawler, err := (&Crawler{}).NewCrawler(apiURL, username, password)
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}

	projectsID, err := crawler.crawlProjectsID()
	if err != nil {
		log.Fatalf("Failed to crawl project IDs: %v", err)
	}

	fmt.Println("Project IDs:", projectsID)

	tasksID, err := crawler.crawlTasksID(projectName)
	if err != nil {
		log.Fatalf("Failed to crawl task IDs for project '%s': %v", projectName, err)
	}

	crawler.SettasksID(tasksID)
	tasksActivities, err := crawler.GetTasksActivities()
	if err != nil {
		log.Fatalf("Failed to crawl tasks activities: %v", err)
	}

	jsonData, err := json.MarshalIndent(tasksActivities, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal task activities to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
