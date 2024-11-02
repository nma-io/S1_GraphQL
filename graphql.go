package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// There is much more data available here - see "schema.query/schema.response" if you are interested in getting more detail.
type AlertData struct {
	Node struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Result         string `json:"result"`
		Status         string `json:"status"`
		Severity       string `json:"severity"`
		NoteExists     bool   `json:"noteExists"`
		DetectedAt     string `json:"detectedAt"`
		ExternalID     string `json:"externalId"`
		AnalystVerdict string `json:"analystVerdict"`
		Classification string `json:"classification"`
		Asset          struct {
			Name string `json:"name"`
		} `json:"asset"`
		Process struct {
			CmdLine string `json:"cmdLine"`
		} `json:"process"`
	} `json:"node"`
}

type GraphQLResponse struct {
	Data struct {
		Alerts struct {
			Edges    []AlertData `json:"edges"`
			PageInfo PageInfo    `json:"pageInfo"`
		} `json:"alerts"`
	} `json:"data"`
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

// Function to send GraphQL requests and parse the response - this is a generic function that can be used for any GraphQL query
func sendGraphQLRequest(endpoint, apiKey, query, cursor string) (*GraphQLResponse, error) {
	payload := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"after": cursor,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// DEBUG: Print the response body
	// log.Printf("Response body: %s\n", body)

	// Check if the response status is not OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, body)
	}

	var result GraphQLResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v, body: %s", err, body)
	}

	// Todo - This needs to be moved to alerts, Just quick pre-check here.
	if len(result.Data.Alerts.Edges) == 0 {
		log.Println("No alerts found.")
	}

	return &result, nil
}
