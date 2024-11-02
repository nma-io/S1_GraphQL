package main

/* SentinelOne GraphQL API Client v2024.0.1 - Nicholas Albright (@nma-io)

TODO:
	- Add different options for verdict flags, there are many more to choose from.
	- Add a capability to get details of a specific event - by ExternalID
*/

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// Globals:
var scopeId string
var product string
var apiKey string

func main() {

	fmt.Printf("SentinelOne GraphQL API Client v%s - %s\n", version, author)
	productFlag := flag.String("product", "EDR", "Specify detection product: EDR, Identity, STAR")
	// Fix the lookback flag to use start/end date instead. This is just temporay.
	lookbackFlag := flag.Int("t", 2, "Look back in days - this is for alerts with missing comments only")
	closeAlerts := flag.Bool("c", false, "Close alerts")
	missingComments := flag.Bool("missing-comments", false, "Show alerts without comments")
	message := flag.String("m", "Closed by S1_GraphQL", "Closure message")
	falsePositive := flag.Bool("fp", false, "Set verdict to FALSE_POSITIVE_UNDEFINED")
	truePositive := flag.Bool("tp", false, "Set verdict to TRUE_POSITIVE_UNDEFINED")
	scopeFlag := flag.String("scope", "", "Scope ID for the account")
	startDate := flag.String("start", "", "Start date for alert detection range (YYYY-MM-DD)")
	endDate := flag.String("end", "", "End date for alert detection range (YYYY-MM-DD)")

	flag.Parse()
	product = *productFlag
	scopeId = *scopeFlag

	apiKey = os.Getenv("S1_TOKEN")
	if apiKey == "" {
		log.Fatal("API Key not found in environment variables. Please set S1_TOKEN with the S1 API Token provided by your administrator.")
	}
	if scopeId == "" {
		log.Fatal("Scope ID not provided. Please use --scope flag to specify the scope ID for the SentinelOne account.")
	}

	switch {
	case *missingComments:
		listAlertsWithoutComments(apiKey, *lookbackFlag)
	case *closeAlerts:
		closeAlertsWithFilters(apiKey, *lookbackFlag, *message, *falsePositive, *truePositive, *startDate, *endDate)
	default:
		fmt.Println("Please specify a valid action: --missing-comments or -c for closing alerts")
	}
}
