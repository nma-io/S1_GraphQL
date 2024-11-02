package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

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

func sendGraphQLRequest(endpoint, apiKey, query, cursor string) (*GraphQLResponse, error) {
	payload := map[string]interface{}{
		"query": query,
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Debug
	// log.Printf("Response: %s", body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, body)
	}

	var result GraphQLResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v, body: %s", err, body)
	}

	if len(result.Data.Alerts.Edges) == 0 {
		log.Println("No alerts found.")
	}

	return &result, nil
}
