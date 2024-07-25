package main

import (
	"fmt"
	"log"
	"openproject-crawler/internal/credential"
	crawlprojects "openproject-crawler/pkg/projectscrawler"
)

func main() {
	apiURL := "https://project.jayeson.com.sg/api/v3"
	cred, err := credential.SetCredential("apikey", "a6a2081fd40e89612e0d362753d3cf843974cfde7c67821c03b9c851dfc138d1")
	if err != nil {
		fmt.Println(err)
	}
	base64Token := cred.GenerateToken()

	crawlProjects, err := crawlprojects.NewProjectCrawler(apiURL, base64Token)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	projectIDs, err := crawlProjects.GetProjectIDs()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println("Project IDs and Identifiers:", projectIDs)
}
