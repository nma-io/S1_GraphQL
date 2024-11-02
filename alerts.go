package main

import (
	"fmt"
	"log"
)

// Process results to print out relevant alert details
func processResults(results []AlertData) (rstring string) {

	for _, alert := range results {
		// Parse DetectedAt
		detectedAt, err := parseDetectedAt(alert.Node.DetectedAt)
		if err != nil {
			log.Printf("Failed to parse detectedAt for alert ID %s: %v", alert.Node.ID, err)
			continue
		}
		if alert.Node.NoteExists {
			continue // Skip alerts with notes
		}
		// Print detailed information
		rstring += fmt.Sprintf(
			"%s\nName: %s\nSeverity: %s\nStatus: %s\nDetectedAt: %s\nClassification: %s\nAnalystVerdict: %s\nExternalID: %s\nAsset: %s\nCmdLine: %s\n\n",
			separator, alert.Node.Name, alert.Node.Severity, alert.Node.Status,
			detectedAt.Format("2006-01-02 15:04:05"), // Format detectedAt for human readable output
			alert.Node.Classification, alert.Node.AnalystVerdict, alert.Node.ExternalID,
			alert.Node.Asset.Name, alert.Node.Process.CmdLine,
		)
	}
	return rstring
}

// Function to list alerts without comments by product, filtering on date and detection product only
func listAlertsWithoutComments(apiKey string, lookbackDays int) {
	// Calculate the start timestamp based on lookbackDays
	startTimestamp := calculateStartTimestamp(lookbackDays)

	// Build the query with the dynamically calculated start timestamp
	// Todo - we should probably add an end timestamp as well
	query := fmt.Sprintf(`{
		alerts(filters: [
				{fieldId: "detectionProduct", stringIn: { values: ["%s"] }},
				{fieldId: "detectedAt", dateTimeRange: { start: %d, startInclusive: true, end: null }}
			],
			sort: { by: "detectedAt", order: DESC },
			first: 100) {
			edges { node { id name result status detectedAt noteExists severity externalId analystVerdict classification asset { name } process { cmdLine } }}
			pageInfo { endCursor hasNextPage }
		}
	}`, product, startTimestamp)

	var results []AlertData
	hasNextPage, endCursor := true, ""

	for hasNextPage {
		resp, err := sendGraphQLRequest(apiEndpoint, apiKey, query, endCursor)
		if err != nil {
			log.Fatalf("Failed to fetch alerts: %v", err)
		}
		results = append(results, resp.Data.Alerts.Edges...)
		endCursor, hasNextPage = resp.Data.Alerts.PageInfo.EndCursor, resp.Data.Alerts.PageInfo.HasNextPage
	}

	fmt.Printf("Fetched %d alerts.\n", len(results))
	data := processResults(results)
	if len(data) == 0 {
		fmt.Println("No alerts found without comments.")
	} else {
		fmt.Println(data)

	}
}
