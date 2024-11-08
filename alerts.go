package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/nma-io/alerter/gchat"
	"github.com/nma-io/alerter/teams"
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
		if "RESOLVED" == alert.Node.Status {
			continue // Skip resolved alerts
		}
		// Print detailed information
		rstring += fmt.Sprintf(
			"%s\nProduct: %s\nName: %s\nSeverity: %s\nStatus: %s\nDetectedAt: %s\nClassification: %s\nAnalystVerdict: %s\nExternalID: %s\nAsset: %s\nCmdLine: %s\n\n",
			separator, alert.Node.DetectionSource.Product, alert.Node.Name, alert.Node.Severity, alert.Node.Status,
			detectedAt.Format("2006-01-02 15:04:05"), // Format detectedAt for human readable output
			alert.Node.Classification, alert.Node.AnalystVerdict, alert.Node.ExternalID,
			alert.Node.Asset.Name, alert.Node.Process.CmdLine,
		)
	}
	return rstring
}

func listAlertsWithoutComments(apiKey string, lookbackMinutes int) {
	startTimestamp := calculateStartTimestamp(lookbackMinutes)

	// Define the product filter, setting it to all products if 'ALL' is selected
	var productFilter string
	if "ALL" == strings.ToUpper(product) {
		productFilter = `["EDR", "STAR", "Identity"]`
	} else {
		productFilter = fmt.Sprintf(`["%s"]`, product)
	}

	queryTemplate := `{
        alerts(filters: [
                {fieldId: "detectionProduct", stringIn: { values: %s }},
                {fieldId: "detectedAt", dateTimeRange: { start: %d, startInclusive: true, end: null }}
            ],
            sort: { by: "detectedAt", order: DESC },
            first: 1000,
            %s) {
            edges { node { id name result status detectedAt noteExists severity externalId analystVerdict classification asset { name } process { cmdLine } detectionSource { product } }}
            pageInfo { endCursor hasNextPage }
        }
    }`

	// Initialize results array and pagination variables
	var results []AlertData
	hasNextPage := true
	endCursor := ""

	for hasNextPage {
		afterClause := ""
		if endCursor != "" {
			afterClause = fmt.Sprintf(`after: "%s"`, endCursor)
		}

		// Construct the query
		query := fmt.Sprintf(queryTemplate, productFilter, startTimestamp, afterClause)

		// Send the GraphQL request
		resp, err := sendGraphQLRequest(apiEndpoint, apiKey, query)
		if err != nil {
			log.Fatalf("Failed to fetch alerts: %v", err)
		}

		// Append results and update pagination variables
		results = append(results, resp.Data.Alerts.Edges...)
		endCursor, hasNextPage = resp.Data.Alerts.PageInfo.EndCursor, resp.Data.Alerts.PageInfo.HasNextPage

		// Debug output
		// fmt.Printf("Fetched %d alerts so far. Next endCursor: %s, hasNextPage: %v", len(results), endCursor, hasNextPage)

		// Check to prevent infinite loop
		if endCursor == "" && hasNextPage {
			log.Fatal("Error: Pagination is in infinite loop.")
		}
	}

	// Process and display the results
	fmt.Printf("Fetched %d alerts.\n", len(results))
	data := processResults(results)
	if len(data) == 0 {
		fmt.Println("No alerts found without comments.")
	} else {
		counts := strings.Count(data, separator)
		data = fmt.Sprintf("Warning: %d Unresolved SentinelOne Alerts\n\n%s\n\n", counts, data)
		// Handle Webhook notifications
		switch {
		case strings.Contains(webHookUrl, "office.com"):
			teams.Send(webHookUrl, data)
		case strings.Contains(webHookUrl, "chat.googleapis.com"):
			gchat.Send(webHookUrl, data)
		default:
			fmt.Println(data)
		}
	}
}
