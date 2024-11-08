package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func closeAlertsWithFilters(apiKey string, lookbackDays int, message string, falsePositive, truePositive bool, startDate, endDate string) {
	alertIds := fetchAlertIds(apiKey, lookbackDays, startDate, endDate)

	for i := 0; i < len(alertIds); i += 100 {
		end := i + 100
		if end > len(alertIds) {
			end = len(alertIds)
		}
		processAlertBatch(apiKey, alertIds[i:end], message, falsePositive, truePositive)
		time.Sleep(3 * time.Second)
	}
}

// Fetch alert IDs based on filters
func fetchAlertIds(apiKey string, lookbackDays int, startDate, endDate string) []string {
	var startEpoch, endEpoch int64

	if startDate != "" {
		startEpoch = parseDateToEpoch(startDate)
	} else {
		startEpoch = getEpochDaysAgo(lookbackDays)
	}

	if endDate != "" {
		endEpoch = parseDateToEpoch(endDate)
	} else {
		endEpoch = time.Now().UnixMilli()
	}

	query := `{
		alerts(filters: [
			{fieldId: "detectionProduct", stringEqual: { value: "` + product + `" }},
			{fieldId: "status", stringEqual: { value: "NEW" }},
			{fieldId: "detectedAt", dateTimeRange: { start: ` + strconv.FormatInt(startEpoch, 10) + `, end: ` + strconv.FormatInt(endEpoch, 10) + ` }}
		],
		first: 1000) {
			edges { node { id }}
		}
	}`

	resp, err := sendGraphQLRequest(apiEndpoint, apiKey, query)
	if err != nil {
		log.Fatalf("Failed to fetch alert IDs: %v", err)
	}

	alertIds := make([]string, len(resp.Data.Alerts.Edges))
	for i, edge := range resp.Data.Alerts.Edges {
		alertIds[i] = edge.Node.ID
	}

	fmt.Printf("Fetched %d alert IDs\n", len(alertIds))
	return alertIds
}

// This function handles adding a note, setting verdict, and closing alerts in batches.
// I have lots of identity alerts that are from our kickoff. I want to close them all.
func processAlertBatch(apiKey string, alertIds []string, message string, falsePositive, truePositive bool) {
	fmt.Printf("Processing %d alerts with message: %s\n", len(alertIds), message)
	// To Properly close these we have to follow a specific flow:

	// Step 1: Add note to alerts
	notePayload := generatePayload(alertIds, "S1/alert/addNote", map[string]interface{}{
		"note": map[string]string{"value": message},
	})
	if err := sendMutation(apiKey, notePayload); err != nil {
		log.Printf("Error adding note to alerts: %v\n", err)
	}
	time.Sleep(2 * time.Second) // Give it a bit to catchup, otherwise we get "All Shards Down" errors.

	// Step 2: Set analyst verdict
	// NOTE - This is currently hard coded, if we add more functionality to the main.go file we need to fix this.
	verdictValue := "FALSE_POSITIVE_UNDEFINED"
	if truePositive {
		verdictValue = "TRUE_POSITIVE_UNDEFINED"
	}

	verdictPayload := generatePayload(alertIds, "S1/alert/analystVerdictUpdate", map[string]interface{}{
		"analystVerdict": map[string]string{"value": verdictValue},
	})
	if err := sendMutation(apiKey, verdictPayload); err != nil {
		log.Printf("Error setting analyst verdict on alerts: %v\n", err)
	}
	time.Sleep(1 * time.Second) // Another sleep for the shard problem - this is stupid.

	// Step 3: Update status to RESOLVED
	resolvePayload := generatePayload(alertIds, "S1/alert/statusUpdate", map[string]interface{}{
		"status": map[string]string{"value": "RESOLVED"},
	})
	if err := sendMutation(apiKey, resolvePayload); err != nil {
		log.Printf("Error resolving alerts: %v\n", err)
	}
	time.Sleep(2 * time.Second) // Last sleep. It takes longer to do this than verdict update for some reason.
}

// generatePayload creates the payload for each action in the mutation
func generatePayload(ids []string, actionID string, payloadContent map[string]interface{}) map[string]interface{} {
	orConditions := make([]map[string]interface{}, len(ids))
	for i, alertID := range ids {
		orConditions[i] = map[string]interface{}{
			"and": []map[string]interface{}{
				{"fieldId": "id", "stringIn": map[string]interface{}{"values": []string{alertID}}},
				{"fieldId": "status", "stringIn": map[string]interface{}{"values": []string{"NEW"}}},
				// product is a global, careful!
				{"fieldId": "detectionProduct", "stringIn": map[string]interface{}{"values": []string{product}}},
			},
		}
	}

	return map[string]interface{}{
		"operationName": "AlertTriggerActions",
		"variables": map[string]interface{}{
			// scopeId is a global - careful with it!
			"scope":   map[string]interface{}{"scopeIds": []string{scopeId}, "scopeType": "ACCOUNT"},
			"filter":  map[string]interface{}{"or": orConditions},
			"actions": []map[string]interface{}{{"id": actionID, "payload": payloadContent}},
		},
		"query": `
			fragment ActionsTriggered on ActionsTriggered {
				actions {
					actionId
					alertCount
					skip { id __typename }
					failure { id __typename }
					success { id __typename }
					__typename
				}
				__typename
			}
			mutation AlertTriggerActions($filter: OrFilterSelectionInput, $scope: ScopeSelectorInput, $actions: [TriggerActionInput!]!) {
				alertTriggerActions(filter: $filter, scope: $scope, actions: $actions) {
					...ActionsTriggered
					__typename
				}
			}
		`,
	}
}

// sendMutation sends the mutation request to the GraphQL API
func sendMutation(apiKey string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check for any errors in the response JSON
	if errors, found := response["errors"]; found {
		return fmt.Errorf("API returned errors: %v", errors)
	}

	return nil
}
