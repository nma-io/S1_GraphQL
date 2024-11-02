package main

import (
	"fmt"
	"log"
	"strings"
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

// Function to list alerts without comments by product - allowing for ALL as a product filter.
func listAlertsWithoutComments(apiKey string, lookbackDays int) {
	startTimestamp := calculateStartTimestamp(lookbackDays)

	// Define the product filter, setting it to all products if 'ALL' is selected
	var productFilter string
	if "ALL" == strings.ToUpper(product) {
		productFilter = `["EDR", "STAR", "IDENTITY"]`
	} else {
		productFilter = fmt.Sprintf(`["%s"]`, product)
	}

	// GraphQL query without 'after' but with 'endCursor' for pagination
	queryTemplate := `{
		alerts(filters: [
				{fieldId: "detectionProduct", stringIn: { values: %s }},
				{fieldId: "detectedAt", dateTimeRange: { start: %d, startInclusive: true, end: null }}
			],
			sort: { by: "detectedAt", order: DESC },
			first: 1000) {
			edges { node { id name result status detectedAt noteExists severity externalId analystVerdict classification asset { name } process { cmdLine } }}
			pageInfo { endCursor hasNextPage }
		}
	}`

	// Initialize results array and pagination variables
	var results []AlertData
	hasNextPage, endCursor := true, ""

	for hasNextPage {
		query := fmt.Sprintf(queryTemplate, productFilter, startTimestamp)

		resp, err := sendGraphQLRequest(apiEndpoint, apiKey, query, endCursor)
		if err != nil {
			log.Fatalf("Failed to fetch alerts: %v", err)
		}

		results = append(results, resp.Data.Alerts.Edges...)
		endCursor, hasNextPage = resp.Data.Alerts.PageInfo.EndCursor, resp.Data.Alerts.PageInfo.HasNextPage

		// Check to prevent infinite loop
		if endCursor == "" && hasNextPage {
			log.Fatal("Error: Pagination endCursor is empty, but hasNextPage is still true. Exiting to avoid infinite loop.")
		}
	}

	// Process and display the results
	fmt.Printf("Fetched %d alerts.\n", len(results))
	data := processResults(results)
	if len(data) == 0 {
		fmt.Println("No alerts found without comments.")
	} else {
		fmt.Println(data)
	}
}
