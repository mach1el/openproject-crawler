package main

import (
	"encoding/json"
	"fmt"
	"log"
	"openproject-crawler/internal/credential"

	// pjcrawler "openproject-crawler/pkg/projectscrawler"
	"openproject-crawler/pkg/wpcrawler"
)

func main() {
	apiURL := "https://project.jayeson.com.sg/api/v3"
	cred, err := credential.SetCredential("apikey", "a6a2081fd40e89612e0d362753d3cf843974cfde7c67821c03b9c851dfc138d1")
	if err != nil {
		fmt.Println(err)
	}
	base64Token := cred.GenerateToken()

	crawler, err := wpcrawler.NewWPCrawler(apiURL, base64Token, "")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// crawler.SetProjectName("vi")
	params := make(map[string]interface{})
	filter := []map[string]interface{}{
		{
			"project": map[string]interface{}{
				"operator": "=",
				"values":   []string{"3"},
			},
		},
	}
	filterJSON, _ := json.Marshal(filter)
	params["pageSize"] = "1000"
	params["filters"] = string(filterJSON)

	// crawler.SetParams(params)
	tasksAttr, err := crawler.GetTasksAttrAsync()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(tasksAttr)

	// pjcrawler, err := pjcrawler.NewProjectCrawler(apiURL, base64Token)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// projectIDs, err := pjcrawler.GetProjectIDs()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println("Project IDs and Identifiers:", projectIDs)
}
