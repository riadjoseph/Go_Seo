// Go_seo_dashboard: SEO insights dashboard
// Written by Jason Vicinanza

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// Version
var version = "v0.1"

// Botify API token
var botifyApiToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Colours, symbols etc
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"
var lineSeparator = "█" + strings.Repeat("█", 129)

// KPI Specific colours
var kpiColourRevenue = "Coral"
var kpiColourVisits = "Green"
var kpiColourVisitsPerOrder = "DarkGoldenRod"
var kpiColourRevenueProjection = "Orange"
var kpiColourOrganicVisitValue = "CornflowerBlue"
var kpiColourNoOfOrders = "IndianRed"
var kpiColourOrderValue = "MediumSlateBlue"

// Anonymous mode. When set to true the URL to the project defaults to 'https://www.botify.com'
// If set to false a link is provided to the analysis project
var anonymousMode = false

// DateRanges struct used to hold the monthly date ranges
// Used for revenue and visits data
type DateRanges struct {
	MonthlyRanges [][2]time.Time
}

// Slice used to store the month names
var startMthNames []string

// Slice used to store projected revenue values
var projectionRevenue []int

// KeywordsData struct used to store Keywords dimensions and metrics
type KeywordsData struct {
	Results []struct {
		Dimensions []interface{} `json:"dimensions"`
		Metrics    []*float64    `json:"metrics,omitempty"`
	} `json:"results"`
}

// Used for the branded/non branded title in the wordcloud
var wordcloudTitle = ""

// Slices used to store the startMthDate and endMthDate
var startMthDates = make([]string, 0)
var endMthDates = make([]string, 0)

// Slices used to store the SEO metrics
var seoMetricsRevenue []int
var seoMetricsVisits []int
var seoMetricsOrders []int
var seoOrderValue []int
var seoVisitValue []float64

// Slices used to store branded Keywords KPIs
var kwKeywords []string
var kwMetricsCountUrls []int
var kwMetricsCountClicks []int
var kwMetricsCTR []float64
var kwMetricsAvgPosition []float64

// Slices used to store non-branded Keywords KPIsd
var kwKeywordsNB []string
var kwMetricsCountClicksNB []int
var kwMetricsCTRNB []float64
var kwMetricsAvgPositionNB []float64

// Variables used to store the CMGR values
var cmgrRevenue float64
var cmgrVisits float64
var cmgrVisitValue float64
var cmgrOrdersValue float64
var cmgrOrdersValueValue float64

// Variables used to store the total values
var totalVisits int
var totalRevenue int
var totalOrders int
var totalAverageOrderValue int

// Slices used  to store the visits per order for each month
var visitsPerOrder []int

// AnalyticsID is used to identify which analytics tool is in use
type AnalyticsID struct {
	ID string `json:"id"`
}

// The Result struct is used to store the revenue, orders and visits
type Result struct {
	Dimensions []interface{} `json:"dimensions"`
	Metrics    []float64     `json:"metrics"`
}
type Response struct {
	Results []Result `json:"results"`
}

// Project URL. Used to provide a link to the Botify project
var projectURL = ""

// Organization name used for display purposes
var displayOrgName = ""

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Variables used to store the min and max visits per order
var minVisitsPerOrder = 0
var maxVisitsPerOrder = 0

// No. of months processed
var noOfMonths = 0

// Average visits per order
var totalAverageVisitsPerOrder = 0

// The number of keywords to include in the wordcloud
var noOfKWInCloud = 50

// The number of top keywords to include in the keywords detail table
var noOfTopKeywords = 10

// Used to set the default size for all chart types
var chartDefaultWidth = "85vw"
var chartDefaultHeight = "90vh"

var wcDefaultWidth = "95vw"
var wcDefaultHeight = "95vh"

var badgeDefaultWidth = "95vw"
var badgeDefaultHeight = "95vh"

var gaugeDefaultWidth = "96vw"
var gaugeDefaultHeight = "96vh"

// Define the increment and the maximum value
var projectionIncrement = 10000
var projectionMaxVisits = 1000000

// Slices used to store the visit increment values
var projectionVisitIncrements []int
var projectionVisitIncrementsString []string

func main() {

	// Display the welcome banner
	displayBanner()

	// Serve static files from the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Define a handler function for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the form data from the request
		r.ParseForm()
		organization := r.Form.Get("organization")
		project := r.Form.Get("project")

		fmt.Printf("\nOrganization: %s, Project: %s\n", organization, project)

		// Set the organization and project name
		orgName = organization
		projectName = project

		// Generate a session ID used for grouping log entries
		sessionID, err := generateLogSessionID(8)
		if err != nil {
			log.Fatalf(red+"Error. writeLog. Failed generating session ID: %s"+reset, err)
		}

		// Get revenue, visits, orders and keyword data
		dataStatus := getSeoInsights(sessionID)

		// An invalid org/project name has been specified
		if dataStatus == "errorNoProjectFound" {
			// Write to the log
			writeLog(sessionID, orgName, projectName, "-", "No project found")
			generateErrorPage("No project found. Try another organisation and project name. (" + orgName + "/" + projectName + ")")
			http.Redirect(w, r, "go_seo_errorPage.html", http.StatusFound)
			return
		}

		// No analytics tool has been integrated
		if dataStatus == "errorNoAnalyticsIntegrated" {
			// Write to the log
			writeLog(sessionID, orgName, projectName, "-", "No analytics found")
			generateErrorPage("No analytics tool has been integrated into the specified project (" + orgName + "/" + projectName + ")")
			http.Redirect(w, r, "go_seo_errorPage.html", http.StatusFound)
			return
		}

		// Engagement analytics has not been configured
		if dataStatus == "errorNoEAFound" {
			// Write to the log
			writeLog(sessionID, orgName, projectName, "-", "No revenue data found")
			generateErrorPage("Engagement analytics with visits, revenue & transactions has not been configured for the specified project (" + orgName + "/" + projectName + ")")
			http.Redirect(w, r, "go_seo_errorPage.html", http.StatusFound)
			return
		}

		// Generate the dashboard HTML
		goSeoDashboard()

		// Write to the log
		writeLog(sessionID, orgName, projectName, "-", "Dashboard generated")

		// Respond to the client with a success message or redirect to another page
		http.Redirect(w, r, "go_seo_dashboard.html", http.StatusFound)
	})

	// Start the HTTP server
	http.ListenAndServe(":8080", nil)
}

func goSeoDashboard() {

	// Start of charts

	// Generate the header
	dashboardHeader()

	// Total vales
	tableTotalsVisitsOrdersRevenue()

	// Badges for CMGR KPIs
	badgesForKPIs()

	// Visits per order gauge
	gaugeVisitsPerOrder()

	// Revenue & visits bar chart
	barChartRevenueVisits()

	// Visits per order line chart
	lineChartVisitsPerOrder()

	// Organic visit value
	barChartVisitValue()

	// No. of Orders bar chart
	barChartOrders()

	// Order value barchart
	barChartOrderValue()

	// Revenue and visits river chart
	riverCharRevenueVisits()

	// Wordclouds
	// Branded keywords
	wordCloudBrandedUnbranded(true)
	// Non branded keywords
	wordCloudBrandedUnbranded(false)

	// Winning branded keyword
	winningKeywords(true)
	// Winning non branded keyword
	winningKeywords(false)

	// Detailed keyword insights - Branded
	generateHTMLDetailedKeywordsInsights(true)
	// Detailed keyword insights - Non branded
	generateHTMLDetailedKeywordsInsights(false)

	// KPI details table
	tableDataDetail()

	// Revenue projection
	// Revenue projection line chart
	lineChartRevenueProjection()

	// Projection narrative
	projectionNarrative()

	// Footer notes
	footerNotes()

	// Generate go_seo_dashboard.html container
	generateDashboard()

	// Make a tidy display
	fmt.Println()
	fmt.Println(lineSeparator)

	// We're done
	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(purple + "\ngo_seo_dashboard: Done at " + formattedTime)
	fmt.Printf("\nOrganization: %s, Project: %s\n"+reset, orgName, projectName)

	// Make a tidy display
	fmt.Println()
	fmt.Println(lineSeparator)

	// Wait for the next request
	return

}

func getSeoInsights(sessionID string) string {

	fmt.Println(purple + bold + "\nGetting SEO insights" + reset)
	fmt.Println("Session ID:", sessionID)

	// Get the date ranges
	dateRanges := calculateDateRanges()
	// Iterate over the MonthlyRanges and print each range

	// Identify which analytics tool is used
	analyticsID := getAnalyticsID()
	fmt.Println("Analytics identified:", analyticsID)

	// Error checking
	// Exit if no project has been found
	if analyticsID == "errorNoProjectFound" {
		fmt.Println(red+"Error. getSeoInsights. No project found for", orgName+"/"+projectName+reset)
		return analyticsID
	}
	// Exit if no analytics tool has been detected
	if analyticsID == "errorNoAnalyticsIntegrated" {
		fmt.Println(red+"Error. getSeoInsights. No analytics tool integrated for", orgName+"/"+projectName+reset)
		return analyticsID
	}

	// Reset the dates slice
	resetMetrics()

	// Populate the slice with string versions of the dates ready for use in the BQL
	for _, dateRange := range dateRanges.MonthlyRanges {
		startMthDate := dateRange[0].Format("20060102")
		endMthDate := dateRange[1].Format("20060102")
		startMthDates = append(startMthDates, startMthDate)
		endMthDates = append(endMthDates, endMthDate)

		// Get the month name
		startDate, _ := time.Parse("20060102", startMthDate)
		startMthName := startDate.Format("January 2006")
		startMthNames = append(startMthNames, startMthName)
	}

	// Get the revenue data
	getRevenueDataStatus := getRevenueData(analyticsID, startMthDates, endMthDates)

	// Error checking
	if getRevenueDataStatus == "errorNoEAFound" {
		return getRevenueDataStatus
	}

	// Write to the log. Data acquired
	writeLog(sessionID, orgName, projectName, analyticsID, "Revenue data acquired")

	// Get the keywords data
	// Get last months' date range
	kwStartDate := startMthDates[len(startMthDates)-1]
	kwEndDate := endMthDates[len(endMthDates)-1]

	getKeywordsCloudData(kwStartDate, kwEndDate)

	// Write to the log. Data acquired
	writeLog(sessionID, orgName, projectName, analyticsID, "Keyword data acquired")

	// Calculate the CMGR values
	calculateCMGR()

	// Calculate the projections
	projectionDataCompute()

	return "success"
}

func resetMetrics() {
	// Reset slices
	startMthDates = nil
	endMthDates = nil
	startMthNames = nil
	seoMetricsRevenue = nil
	seoMetricsVisits = nil
	seoMetricsOrders = nil
	seoOrderValue = nil
	totalAverageOrderValue = 0
	seoVisitValue = nil
	visitsPerOrder = nil
	kwKeywords = nil
	kwMetricsCountUrls = nil
	kwMetricsCountClicks = nil
	kwMetricsCTR = nil
	kwMetricsAvgPosition = nil
	kwKeywordsNB = nil
	kwMetricsCountClicksNB = nil
	kwMetricsCTRNB = nil
	kwMetricsAvgPositionNB = nil

	// Reset integers and floats
	totalVisits = 0
	totalRevenue = 0
	totalOrders = 0
	cmgrRevenue = 0.00
	cmgrVisits = 0.00
	cmgrVisitValue = 0.00
	cmgrOrdersValue = 0.00
	cmgrOrdersValueValue = 0.00
}

// Get the revenue, orders and visits data
func getRevenueData(analyticsID string, startMthDates []string, endMthDates []string) string {

	var metricsOrders = 0
	var metricsRevenue = 0
	var metricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Get monthly insights
	for i := range startMthDates {

		getRevenueDataStatus := ""
		metricsOrders, metricsRevenue, metricsVisits, avgOrderValue, avgVisitValue, getRevenueDataStatus = generateRevenueBQL(analyticsID, startMthDates[i], endMthDates[i])

		// Error checking
		if getRevenueDataStatus == "errorNoEAFound" {
			return getRevenueDataStatus
		}

		// Append the metrics to the slices
		seoMetricsOrders = append(seoMetricsOrders, metricsOrders)
		seoMetricsRevenue = append(seoMetricsRevenue, metricsRevenue)
		seoOrderValue = append(seoOrderValue, avgOrderValue)
		seoMetricsVisits = append(seoMetricsVisits, metricsVisits)

		// Round avgVisitValue to 2 decimal places
		avgVisitValueRounded := math.Round(avgVisitValue*100) / 100
		seoVisitValue = append(seoVisitValue, avgVisitValueRounded)

		// Calculate the visits per order (for the month)
		visitsPerOrder = append(visitsPerOrder, metricsVisits/metricsOrders)

		// Calculate the grand total for revenue visits & orders
		totalRevenue += metricsRevenue
		totalVisits += metricsVisits
		totalOrders += metricsOrders

		// Use the printer to format an integer
		formatInteger := message.NewPrinter(language.English)

		// Display the KPIs
		fmt.Printf(green+"\nDate Start: %s End: %s\n"+reset, startMthDates[i], endMthDates[i])
		formattedOrders := formatInteger.Sprintf("%d", metricsOrders)
		formattedRevenue := formatInteger.Sprintf("%d", metricsRevenue)
		formattedVisits := formatInteger.Sprintf("%d", metricsVisits)
		fmt.Println("No. Orders:", formattedOrders)
		fmt.Println("Total revenue:", formattedRevenue)
		fmt.Println("Average order value:", avgOrderValue)
		fmt.Println("No. of visits:", formattedVisits)
		fmt.Println("Average visit value:", avgVisitValue)
	}

	// Calculate the average visits per order
	totalVisitsPerOrder := 0
	for _, value := range visitsPerOrder {
		totalVisitsPerOrder += value
	}

	if len(visitsPerOrder) > 0 {
		totalAverageVisitsPerOrder = totalVisitsPerOrder / len(visitsPerOrder)
	}

	// Get the min and max visits per order
	minVisitsPerOrder = -1
	maxVisitsPerOrder = visitsPerOrder[0]

	// Iterate through the slice to find the min and max values
	for _, value := range visitsPerOrder {
		if value >= 2 {
			if minVisitsPerOrder == -1 || value < minVisitsPerOrder {
				minVisitsPerOrder = value
			}
		}
		if value > maxVisitsPerOrder {
			maxVisitsPerOrder = value
		}
	}

	// Calculate the average order value for all months
	var totalOrderValue = 0
	var mthAverageOrderValue = 0
	// Sum the total of averages
	for _, mthAverageOrderValue = range seoOrderValue {
		totalOrderValue += mthAverageOrderValue
	}
	// Calculate the average of averages. Ensure there is no division by zero
	totalAverageOrderValue = 0
	if len(seoOrderValue) > 0 {
		totalAverageOrderValue = totalOrderValue / len(seoOrderValue)
	}

	fmt.Println(green + "\nTotals" + reset)
	fmt.Println("Total visits:", totalVisits)
	fmt.Println("Total revenue:", totalRevenue)
	fmt.Println("Total orders:", totalOrders)
	fmt.Println("Total average order value:", totalAverageOrderValue)
	fmt.Println("Average visits per order:", totalAverageVisitsPerOrder)

	return "success"
}

// Get the keywords data
func getKeywordsCloudData(startMthDates string, endMthDates string) {

	// Branded keywords
	generateKeywordsCloudBQL(startMthDates, endMthDates, "true")

	// Non-branded keywords
	generateKeywordsCloudBQL(startMthDates, endMthDates, "false")

}

// Execute the BQL to acquire keywords data
func generateKeywordsCloudBQL(startDate string, endDate string, brandedFlag string) ([]string, []int, []int, []float64, []float64) {

	// Get the keyword data. Define the BQL
	bqlCloudKeywords := fmt.Sprintf(`{
		"collections": [
						"search_console_by_property"
		],
		"periods": [
			[
						%s,
						%s
			]
		],
		"query": {
			"dimensions": [
				"keyword"
			],
			"metrics": [
					"search_console_by_property.period_0.count_clicks",
					"search_console_by_property.period_0.avg_position",
					"search_console_by_property.period_0.ctr"
			],
			"sort": [{"index": 0, "type": "metrics", "order": "desc"}],
	
			"filters": {
				"and": [
					{
						"field": "keyword_meta.branded",
						"predicate": "eq",
						"value": %s
					}
				]
			}
		}
	}`, startDate, endDate, brandedFlag)

	// get the keyword's data. Receiving top 50 keys here
	responseData := executeBQL(noOfKWInCloud, bqlCloudKeywords)

	// Unmarshal JSON data into KeywordsData struct
	var response KeywordsData
	err := json.Unmarshal(responseData, &response)
	if err != nil {
		log.Fatalf(red+"Error. generateKeywordsCloudBQL. Cannot unmarshal the JSON: %v"+reset, err)
	}

	// Load the response into the slices - branded keywords
	if brandedFlag == "true" {
		for _, result := range response.Results {
			if len(result.Dimensions) >= 1 && len(result.Metrics) >= 3 {
				kwKeywords = append(kwKeywords, result.Dimensions[0].(string))
				kwMetricsCountClicks = append(kwMetricsCountClicks, int(*result.Metrics[0]))
				kwMetricsAvgPosition = append(kwMetricsAvgPosition, *result.Metrics[1])
				kwMetricsCTR = append(kwMetricsCTR, *result.Metrics[2])
			}
		}
	}

	// Load the response into the slices - non-branded keywords
	if brandedFlag == "false" {
		for _, result := range response.Results {
			if len(result.Dimensions) >= 1 && len(result.Metrics) >= 3 {
				kwKeywordsNB = append(kwKeywordsNB, result.Dimensions[0].(string))
				kwMetricsCountClicksNB = append(kwMetricsCountClicksNB, int(*result.Metrics[0]))
				kwMetricsAvgPositionNB = append(kwMetricsAvgPositionNB, *result.Metrics[1])
				kwMetricsCTRNB = append(kwMetricsCTRNB, *result.Metrics[2])
			}
		}
	}
	return kwKeywords, kwMetricsCountUrls, kwMetricsCountClicks, kwMetricsCTR, kwMetricsAvgPosition
}

// Execute the BQL for the specified date range
func generateRevenueBQL(analyticsID string, startDate string, endDate string) (int, int, int, int, float64, string) {

	conversionCollection := ""
	conversionTransactionField := ""
	// GA4
	conversionCollection = "conversion.dip"
	conversionTransactionField = "transactions"
	// Support for Adobe
	if analyticsID == "visits.adobe" {
		conversionCollection = "conversion"
		conversionTransactionField = "orders"
	}

	// Get the revenue, no. Orders and visits
	bqlRevTrans := fmt.Sprintf(`
	{
    "collections": [
                    "%s",
                    "%s"
    ],
    "periods": [
        [
                    "%s",
                    "%s"
        ]
    ],
    "query": {
        "dimensions": [],
        "metrics": [
                    "%s.period_0.%s",
                    "%s.period_0.revenue",    
                    "%s.period_0.nb"
        ],
        "filters": {
            "and": [
                {
                    "field": "%s.period_0.medium",
                    "predicate": "eq",
                    "value": "organic"
                },
                {
                    "field": "%s.period_0.medium",
                    "predicate": "eq",
                    "value": "organic"
           	     }
      	      ]
    	    }
 	   }
	}`, conversionCollection, analyticsID, startDate, endDate, conversionCollection, conversionTransactionField, conversionCollection, analyticsID, conversionCollection, analyticsID)

	// get the revenue and transaction
	responseData := executeBQL(0, bqlRevTrans)

	// Unmarshal the JSON data into the struct
	var response Response
	err := json.Unmarshal(responseData, &response)
	if err != nil {
		log.Fatalf(red+"Error. generateRevenueBQL. Cannot unmarshal the JSON: %v"+reset, err)
	}

	var metricsOrders = 0
	var metricsRevenue = 0
	var metricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Check if any data has been returned from the API. Count the number of elements in the response.Results slice
	responseCount := len(response.Results)

	if responseCount == 0 {
		fmt.Println(red+"Error. generateRevenueBQL. No engagement analytics (revenue, transactions & visits) has been configured for", orgName+"/"+projectName+reset)
		getRevenueDataStatus := "errorNoEAFound"
		return 0, 0, 0, 0, 0.0, getRevenueDataStatus
	} else {
		metricsOrders = int(response.Results[0].Metrics[0])
		metricsRevenue = int(response.Results[0].Metrics[1])
		metricsVisits = int(response.Results[0].Metrics[2])
		// Compute the average Order value
		avgOrderValue = metricsRevenue / metricsOrders
		avgVisitValue = float64(metricsRevenue) / float64(metricsVisits)
	}
	getRevenueDataStatus := "success"
	return metricsOrders, metricsRevenue, metricsVisits, avgOrderValue, avgVisitValue, getRevenueDataStatus
}

// Header for the dashboard
func dashboardHeader() {

	// 	Anonymous mode
	if anonymousMode {
		displayOrgName = "anonymized"
		projectURL = "https://www.botify.com"
	}

	// 	Not anonymous mode
	if !anonymousMode {
		displayOrgName = orgName
		projectURL = "https://app.botify.com/" + orgName + "/" + projectName
	}

	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            margin: 0;
            height: 100pv;
            overflow: hidden;
        }
        .content {
            color: gray;
            font-size: 15px;
            padding: 5px;
        }
	.header-font {
            font-size: 15px;  
        }
    </style>
</head>
<body>
    <div class="content">
        <span class="header-font">The following insights are based on the previous ` + fmt.Sprintf("%d", noOfMonths) + ` months.</span>
		<span class="header-font">Access the Botify project <a href="` + projectURL + `" target="_blank">here</a></span> (` + displayOrgName + `)

    </div>
</body>
</html>
`

	// Save the HTML to a file
	saveHTML(htmlContent, "./seoDashboardHeader.html")

}

// Badges
func badgesForKPIs() {

	cmgrRevenue32 := float32(cmgrRevenue)
	cmgrVisits32 := float32(cmgrVisits)
	cmgrVisitValue32 := float32(cmgrVisitValue)
	cmgrOrdersValue32 := float32(cmgrOrdersValue)
	cmgrOrdersValueValue32 := float32(cmgrOrdersValueValue)

	// Generate the badges
	liquidBadge("Revenue", cmgrRevenue32)
	liquidBadge("Visits", cmgrVisits32)
	liquidBadge("Visit Value", cmgrVisitValue32)
	liquidBadge("Orders", cmgrOrdersValue32)
	liquidBadge("Order Value", cmgrOrdersValueValue32)
}

// Table for total Visits, Orders & Revenue
func tableTotalsVisitsOrdersRevenue() {

	// Use the printer to format an integer
	formatInteger := message.NewPrinter(language.English)

	totalVisitsFormatted := formatInteger.Sprintf("%d", totalVisits)
	totalOrdersFormatted := formatInteger.Sprintf("%d", totalOrders)
	totalRevenueFormatted := formatInteger.Sprintf("%d", totalRevenue)

	htmlContent := `
<!DOCTYPE html>
<html>
<head>
        <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            margin: 0;
            height: 100vh; /* Ensure the body takes full viewport height */
        }
        .wrapper {
            display: flex;
            justify-content: space-between;
            width: 70%;
        }
        .column {
            flex: 1;
            text-align: center;
            margin: 0 50px;
        }
        th, td {
            font-size: 25px;
            padding: 10px; /* Add padding to the table cells */
        }
        th {
            color: DimGray;
        }
        td {
            color: LightSeaGreen;
        }
    </style>
</head>
<body>
    <div class="container">
    <div class="wrapper">
        <div class="column">
            <table>
				<tr>
                    <th>Visits</th>
                </tr>
                 <tr>
                    <td>` + fmt.Sprintf("%s", totalVisitsFormatted) + `</td>
                </tr>
            </table>
        </div>
        <div class="column">
            <table>
                <tr>
                    <th>Orders</th>
                </tr>
                <tr>
                    <td>` + fmt.Sprintf("%s", totalOrdersFormatted) + `</td>
                </tr>
            </table>
        </div>
        <div class="column">
            <table>
                <tr>
                    <th>Revenue</th>
                </tr>
                <tr>
                    <td>` + fmt.Sprintf("%s", totalRevenueFormatted) + `</td>
                </tr>
            </table>
        </div>
    </div>
    </div>
</body>
</html>`

	// Save the HTML to a file
	saveHTML(htmlContent, "./seoTableTotalsVisitsOrdersRevenue.html")

}

// Bar chart. Revenue and Visits
func barChartRevenueVisits() {

	// create a new bar instance
	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Revenue & visits",
		Subtitle: "Understand your organic visit performance and how much revenue those visits are generating.",
		Link:     projectURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}),
	)

	barDataRevenue := generateBarItems(seoMetricsRevenue)
	barDataVisits := generateBarItems(seoMetricsVisits)

	bar.SetXAxis(startMthNames).
		AddSeries("Revenue", barDataRevenue).
		AddSeries("Visits", barDataVisits).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Minimum", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average", Type: "average"},
		))

	f, _ := os.Create("./seoRevenueVisitsBar.html")
	bar.Render(f)
}

func lineChartVisitsPerOrder() {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Average visits per order",
			Subtitle: "On average, how many organic visits are needed to generate one order?",
			Link:     projectURL,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),

		charts.WithColorsOpts(opts.Colors{kpiColourVisitsPerOrder}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Pass visitsPerOrder directly to generaLineItems
	lineVisitsPerOrderValue := generateLineItems(visitsPerOrder)

	line.SetXAxis(startMthNames).AddSeries("Month", lineVisitsPerOrderValue).SetSeriesOptions(
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: 0.2,
		}),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithMarkPointNameTypeItemOpts(
			opts.MarkPointNameTypeItem{Name: "Maximum visits per order", Type: "max"},
			opts.MarkPointNameTypeItem{Name: "Average visits per order", Type: "average"},
			opts.MarkPointNameTypeItem{Name: "Minimum  visits per order", Type: "min"},
		),
		charts.WithMarkPointStyleOpts(
			opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(true)}},
		),
	)

	f, _ := os.Create("./seoVisitsPerOrderLine.html")
	line.Render(f)
}

// Function to generate line chart items from an array of float64 values
func generateLineItems(visitsPerOrder []int) []opts.LineData {
	items := make([]opts.LineData, len(visitsPerOrder))
	for i, val := range visitsPerOrder {
		items[i] = opts.LineData{Value: val}
	}
	return items
}

// Visit value bar chart
func barChartVisitValue() {
	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Organic visit value",
		Subtitle: "What is the value of a single organic visit?",
		Link:     projectURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		// Increase the canvas size
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourOrganicVisitValue}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	barDataVisitValue := generateBarItemsFloat(seoVisitValue)

	bar.SetXAxis(startMthNames).
		AddSeries("Organic visit value", barDataVisitValue).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Minimum", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average", Type: "average"},
		))

	f, _ := os.Create("./seoVisitValueBar.html")
	bar.Render(f)
}

// Bar chart. No. of Orders
func barChartOrders() {

	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Number of orders",
		Subtitle: "How many orders are placed by organic visitors?",
		Link:     projectURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourNoOfOrders}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	barDataOrders := generateBarItems(seoMetricsOrders)

	bar.SetXAxis(startMthNames).
		AddSeries("Number of orders", barDataOrders).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Minimum", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average", Type: "average"},
		))

	f, _ := os.Create("./seoOrdersBar.html")
	bar.Render(f)
}

// Bar chart. No. of Orders
func barChartOrderValue() {
	// create a new bar instance
	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Average order value",
		Subtitle: "What is the average value of an order placed by an organic visitor?",
		Link:     projectURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourOrderValue}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	barDataOrderValue := generateBarItems(seoOrderValue)

	bar.SetXAxis(startMthNames).
		AddSeries("Order value", barDataOrderValue).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Minimum", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average", Type: "average"},
		))

	f, _ := os.Create("./seoOrderValueBar.html")
	bar.Render(f)
}

// Function to generate BarData items from an array of integers
func generateBarItems(revenue []int) []opts.BarData {

	items := make([]opts.BarData, len(revenue))
	for i, val := range revenue {
		items[i] = opts.BarData{Value: val}
	}
	return items
}

// Function to generate BarData items from an array of float64 values
func generateBarItemsFloat(revenue []float64) []opts.BarData {

	items := make([]opts.BarData, len(revenue))
	for i, val := range revenue {
		items[i] = opts.BarData{Value: val}
	}
	return items
}

func liquidBadge(badgeKPI string, badgeKPIValue float32) {

	page := components.NewPage()
	page.AddCharts(
		generateLiquidBadge(badgeKPI, badgeKPIValue),
	)

	// Removing spaces from badgeKPI to ensure a clean URL for the HTML is generated.
	badgeKPI = strings.ReplaceAll(badgeKPI, " ", "")
	badgeFileName := fmt.Sprintf("./seoCMGR%s.html", badgeKPI)
	f, err := os.Create(badgeFileName)
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func generateLiquidBadge(badgeKPI string, badgeKPIValue float32) *charts.Liquid {

	liquid := charts.NewLiquid()
	liquid.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  badgeDefaultWidth,
			Height: badgeDefaultHeight,
		}),
	)

	liquid.AddSeries(badgeKPI, genLiquidItems([]float32{badgeKPIValue})).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show: opts.Bool(true),
			}),

			charts.WithLiquidChartOpts(opts.LiquidChart{
				IsWaveAnimation: opts.Bool(true),
				Shape:           "circle",
			}),
		)
	return liquid
}

// Get data for the liquid badge
func genLiquidItems(data []float32) []opts.LiquidData {

	items := make([]opts.LiquidData, 0)
	for i := 0; i < len(data); i++ {
		items = append(items, opts.LiquidData{Value: data[i]})
	}
	return items
}

// Top keywords for branded and non-branded
func wordCloudBrandedUnbranded(brandedMode bool) {

	// Generate the HTML for branded keywords
	if brandedMode {
		page := components.NewPage()
		page.AddCharts(
			generateWordCloud(true),
		)
		f, err := os.Create("./seoWordCloudBranded.html")
		if err != nil {
			panic(err)
		}
		page.Render(io.MultiWriter(f))
	}

	// Generate the HTML for non-branded keywords
	if !brandedMode {
		page := components.NewPage()
		page.AddCharts(
			generateWordCloud(false),
		)
		f, err := os.Create("./seoWordCloudNonBranded.html")
		if err != nil {
			panic(err)
		}
		page.Render(io.MultiWriter(f))
	}
}

func generateWordCloud(brandedMode bool) *charts.WordCloud {

	if brandedMode {
		wordcloudTitle = fmt.Sprintf("Top %d branded keywords generating clicks", noOfKWInCloud)
	}
	if !brandedMode {
		wordcloudTitle = fmt.Sprintf("Top %d non branded keywords generating clicks", noOfKWInCloud)
	}

	wc := charts.NewWordCloud()
	wc.SetGlobalOptions(
		//  No options defined
		charts.WithInitializationOpts(opts.Initialization{
			Width:  wcDefaultWidth,
			Height: wcDefaultHeight,
		}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithTitleOpts(opts.Title{
			Title: wordcloudTitle,
			Link:  projectURL,
		}))

	// Generate the branded wordcloud
	if brandedMode {
		wc.AddSeries("Keywords", generateWCData(kwKeywords, kwMetricsCountClicks)).
			SetSeriesOptions(
				charts.WithWorldCloudChartOpts(
					opts.WordCloudChart{
						SizeRange: []float32{10, 90},
						Shape:     "basic",
					}),
			)
	}

	// Generate the non-branded wordcloud
	if !brandedMode {
		wc.AddSeries("Keywords", generateWCDataNB(kwKeywordsNB, kwMetricsCountClicksNB)).
			SetSeriesOptions(
				charts.WithWorldCloudChartOpts(
					opts.WordCloudChart{
						SizeRange: []float32{10, 90},
						Shape:     "basic",
					}),
			)
	}

	return wc
}

func generateWCData(kwKeywords []string, kwMetricsCountClicks []int) (items []opts.WordCloudData) {

	items = make([]opts.WordCloudData, 0)
	// Iterate over kwKeywords and kwMetricsCountClicks slices starting from index 1
	// We start at index 1 because the top keyword is generally significantly higher performing than the following keywords and will distort the wordcloud if included
	for i := 1; i < len(kwKeywords); i++ {
		// Check if index is valid for kwMetricsCountClicks slice
		if i < len(kwMetricsCountClicks) {
			// Append WordCloudData struct with keyword and corresponding count
			items = append(items, opts.WordCloudData{Name: kwKeywords[i], Value: kwMetricsCountClicks[i]})
		}
	}
	return items
}

func generateWCDataNB(kwKeywordsNB []string, kwMetricsCountClicksNB []int) (items []opts.WordCloudData) {

	items = make([]opts.WordCloudData, 0)
	// Iterate over kwKeywords and kwMetricsCountClicks slices starting from index 1
	// We start at index 1 because the top keyword is generally significantly higher performing than the following keywords and will distort the wordcloud if included
	for i := 1; i < len(kwKeywordsNB); i++ {
		// Check if index is valid for kwMetricsCountClicks slice
		if i < len(kwMetricsCountClicksNB) {
			// Append WordCloudData struct with keyword and corresponding count
			items = append(items, opts.WordCloudData{Name: kwKeywordsNB[i], Value: kwMetricsCountClicksNB[i]})
		}
	}
	return items
}

// River chart
func riverCharRevenueVisits() {

	page := components.NewPage()
	page.AddCharts(
		generateRiverTime(),
	)
	f, err := os.Create("./seoVisitsRevenueRiver.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

// Theme river chart
func generateRiverTime() *charts.ThemeRiver {

	tr := charts.NewThemeRiver()
	tr.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Revenue & visits",
			Subtitle: "Gain an insight into the fluctuations in organic visitors to a site and the corresponding revenue generation.",
			Link:     projectURL}),
		charts.WithSingleAxisOpts(opts.SingleAxis{
			Type:   "time",
			Bottom: "10%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
		}),
		// Increase the canvas size
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Populate ThemeRiverData slice
	var themeRiverData []opts.ThemeRiverData

	// Add the Revenue data
	// The date is formatted from YYYYMMDD to YYYY/MM/DD
	for i, date := range startMthDates {
		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			fmt.Printf(red+"Error. generateRiverTime. Error parsing date: %v\n"+reset, err)
			break
		}
		formattedDate := parsedDate.Format("2006/01/02")
		themeRiverData = append(themeRiverData, opts.ThemeRiverData{
			Date:  formattedDate,
			Value: float64(seoMetricsRevenue[i]),
			Name:  "Revenue",
		})
	}

	// Add the Visits data
	// The date is formatted from YYYYMMDD to YYYY/MM/DD
	for i, date := range startMthDates {
		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			fmt.Printf(red+"Error. generateRiverTime. Error parsing date: %v\n"+reset, err)
			break
		}
		formattedDate := parsedDate.Format("2006/01/02")
		themeRiverData = append(themeRiverData, opts.ThemeRiverData{
			Date:  formattedDate,
			Value: float64(seoMetricsVisits[i]),
			Name:  "Visits",
		})
	}

	tr.AddSeries("themeRiver", themeRiverData)

	return tr
}

// Gauge chart
func gaugeVisitsPerOrder() {

	page := components.NewPage()
	page.AddCharts(
		gaugeBase(),
	)

	f, err := os.Create("./seoGauge.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func gaugeBase() *charts.Gauge {

	gauge := charts.NewGauge()

	setMinMax := charts.WithSeriesOpts(func(s *charts.SingleSeries) {
		s.Min = minVisitsPerOrder
		s.Max = maxVisitsPerOrder
	})

	gauge.SetGlobalOptions(
		//  No options defined
		charts.WithInitializationOpts(opts.Initialization{
			Width:  gaugeDefaultWidth,
			Height: gaugeDefaultHeight,
		}),
	)

	//gauge.AddSeries("Visits Per Order", []opts.GaugeData{{Name: "Visits / order", Value: totalAverageVisitsPerOrder}}, setMinMax)
	gauge.AddSeries("Visits Per Order", []opts.GaugeData{{Value: totalAverageVisitsPerOrder}}, setMinMax)

	return gauge
}

// Generate an HTML table containing the detailed KPI insights
func tableDataDetail() {
	var detailedKPITableData [][]string

	// Use the printer to format an integer (or a float)
	formatInteger := message.NewPrinter(language.English)

	for i := 0; i < len(startMthDates); i++ {
		formattedDate := formatDate(startMthDates[i])
		orders := formatInteger.Sprintf("%d", seoMetricsOrders[i])
		revenue := formatInteger.Sprintf("%d", seoMetricsRevenue[i])
		orderValue := formatInteger.Sprintf("%d", seoOrderValue[i])
		visits := formatInteger.Sprintf("%d", seoMetricsVisits[i])
		visitValue := formatInteger.Sprintf("%.2f", seoVisitValue[i])
		visitsPerOrderValue := formatInteger.Sprintf("%d", visitsPerOrder[i])

		row := []string{
			formattedDate,
			orders,
			revenue,
			orderValue,
			visits,
			visitValue,
			visitsPerOrderValue,
		}
		detailedKPITableData = append(detailedKPITableData, row)
	}

	// Generate the table
	htmlContent := generateHTMLDetailedKPIInsightsTable(detailedKPITableData)

	// Save the HTML to a file
	saveHTML(htmlContent, "./seoDataInsightDetailKPIs.html")

}

func winningKeywords(brandedMode bool) {

	var htmlFileName = ""

	// Define the display values based on branded or non-branded mode
	var htmlKeyword = ""
	var htmlClicks = ""
	var htmlClickGap = 0
	var htmlSecondPlaceKW = ""
	var htmlCTR float64
	var htmlAvgPosition float64

	// Use the printer to format an integer (clicks)
	formatInteger := message.NewPrinter(language.English)

	if brandedMode {
		htmlKeyword = kwKeywords[0]
		htmlClicks = formatInteger.Sprintf("%d", kwMetricsCountClicks[0])
		htmlClickGap = int(((float64(kwMetricsCountClicks[0]) - float64(kwMetricsCountClicks[1])) / float64(kwMetricsCountClicks[1])) * 100)
		htmlSecondPlaceKW = kwKeywords[1]
		htmlCTR = kwMetricsCTR[0]
		htmlAvgPosition = kwMetricsAvgPosition[0]
		fmt.Printf(green + "\nBranded keywords\n" + reset)
		for i := 0; i < len(kwKeywords); i++ {
			fmt.Printf(green+"Keyword:"+reset+bold+" %s"+reset+","+green+" Clicks:"+reset+" %d,"+green+" CTR:"+reset+" %.2f,"+green+" Avg. Position:"+reset+" %.2f\n",
				kwKeywords[i], kwMetricsCountClicks[i], kwMetricsCTR[i], kwMetricsAvgPosition[i])
		}
	}

	if !brandedMode {
		htmlKeyword = kwKeywordsNB[0]
		htmlClicks = formatInteger.Sprintf("%d", kwMetricsCountClicksNB[0])
		htmlClickGap = int(((float64(kwMetricsCountClicksNB[0]) - float64(kwMetricsCountClicksNB[1])) / float64(kwMetricsCountClicksNB[1])) * 100)
		htmlSecondPlaceKW = kwKeywordsNB[1]
		htmlCTR = kwMetricsCTRNB[0]
		htmlAvgPosition = kwMetricsAvgPositionNB[0]
		fmt.Printf(green + "\nNon Branded keywords\n" + reset)
		for i := 0; i < len(kwKeywords); i++ {
			fmt.Printf(green+"Keyword:"+reset+bold+" %s"+reset+","+green+" Clicks:"+reset+" %d,"+green+" CTR:"+reset+" %.2f,"+green+" Avg. Position:"+reset+" %.2f\n",
				kwKeywordsNB[i], kwMetricsCountClicksNB[i], kwMetricsCTRNB[i], kwMetricsAvgPositionNB[i])
		}
	}

	// Get the last month name
	htmlLastMonthName := startMthNames[len(startMthNames)-1]

	// HTML content for the winning keyword
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<style>
        body {
            font-family: 'Arial', sans-serif;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            background-color: Cornsilk;
        }
        .content {
            max-width: 600px;
            text-align: center;  
            padding-bottom: 40px; 
        }
        .blueText {
            color: DeepSkyBlue;
            font-size: 25px;
            font-weight: bold;  
        }
        .keyword-font {
            font-family: 'Arial', sans-serif;
            font-size: 18px;
            color: LightSlateGray;
            line-height: 1.6; 
        }
        b {
            color: #333;
        }
    </style>
</head>

<body>
	<p>
    <div class="content">
    <span class="keyword-font">
        The winning keyword was <span class="blueText">%s</span> during <b>%s</b>. 
        This keyword generated <b>%s</b> clicks which is <b>%d%%</b> more clicks than the second placed keyword (<b>%s</b>). The click-through rate for the winning keyword was <b>%.2f%%</b> 
        from an average position of <b>%.2f</b>.
    </span>
	</p>
	</div>
</body>
</html>
`, htmlKeyword, htmlLastMonthName, htmlClicks, htmlClickGap, htmlSecondPlaceKW, htmlCTR, htmlAvgPosition)

	// Define the HTML filename
	if brandedMode {
		htmlFileName = "./seoWinningKeywordBranded.html"
	} else {
		htmlFileName = "./seoWinningKeywordNonBranded.html"
	}

	// Save the HTML to a file
	saveHTML(htmlContent, htmlFileName)
}

// generateHTML generates the HTML content for the table
func generateHTMLDetailedKPIInsightsTable(data [][]string) string {
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
<style>
    body {
        font-family: Arial, sans-serif;
    }
    table {
        width: 100%;
        border-collapse: collapse;
        margin: 10px 0;
        font-size: 18px;
        text-align: left;
    }
    th, td {
        padding: 12px;
        border-bottom: 1px solid #ddd;
    }
    th {
        background-color: #f2f2f2;
    }
    th.title {
        color: gray;
        font-weight: bold;
    }
    td {
        color: dimgray;
    }
    tr:nth-child(even) {
        background-color: #f9f9f9;
    }
    tr:hover {
        background-color: deepskyblue;
    }
    h2 {
        color: dimgray;
        margin-bottom: 20px;
    }
</style>
</head>
<body style="min-height: 10vh;">
    <table>
        <thead>
            <tr>
                <th class="title" style="color: DeepSkyBlue;">Date</th>
				<th class="title" style="color: DeepSkyBlue;">No. of Orders</th>
                <th class="title" style="color: DeepSkyBlue;">Revenue</th>
                <th class="title" style="color: DeepSkyBlue;">Order Value</th>
                <th class="title" style="color: DeepSkyBlue;">No. of Visits</th>
                <th class="title" style="color: DeepSkyBlue;">Visit Value</th>
                <th class="title" style="color: DeepSkyBlue;">Visits per Order</th>
            </tr>
        </thead>
        <tbody>`

	// Title
	htmlContent += fmt.Sprintf("<h2>\n\nMonthly summary for the previous %d months</h2>", noOfMonths)

	// Insert tke KPI details
	for _, row := range data {
		htmlContent += "<tr>"
		for _, cell := range row {
			htmlContent += "<td>" + cell + "</td>"
		}
		htmlContent += "</tr>"
	}
	htmlContent += `
        </tbody>
    </table>
</body>
</html>`

	return htmlContent
}

func generateHTMLDetailedKeywordsInsights(brandedMode bool) {

	// Use the printer to format an integer (clicks)
	formatInteger := message.NewPrinter(language.English)

	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            color: DimGray;
            margin: 5px 0;
            font-size: 18px;
            text-align: left;
        }
        th, td {
            padding: 12px;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: White;
            color: deepskyblue;
        }
        tr:nth-child(odd) {
            background-color: #f9f9f9;
        }
        tr:hover {
            background-color: DeepSkyBlue;
        }
  		h2 {
            color: DimGray;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <table>
        <tr>
            <th>Keyword</th>
            <th>No. of Clicks</th>
            <th>CTR</th>
            <th>Average Position</th>
        </tr>`

	// Branded keywords details
	if brandedMode {
		htmlContent += fmt.Sprintf("<h2>\n\nTop %d branded keywords generating clicks</h2>", noOfTopKeywords)
		for i := 0; i < noOfTopKeywords; i++ {
			kwMetricsCountClicksFormatted := formatInteger.Sprintf("%d", kwMetricsCountClicks[i])
			htmlContent += fmt.Sprintf("<tr>\n"+
				"    <td>%s</td>\n"+
				"    <td>%s</td>\n"+
				"    <td>%.2f%%</td>\n"+
				"    <td>%.2f</td>\n"+
				"</tr>\n",
				kwKeywords[i],
				kwMetricsCountClicksFormatted,
				kwMetricsCTR[i],
				kwMetricsAvgPosition[i])
		}
	}

	// Non branded keywords details
	if !brandedMode {
		htmlContent += fmt.Sprintf("<h2>\n\nTop %d non branded keywords generating clicks</h2>", noOfTopKeywords)
		for i := 0; i < noOfTopKeywords; i++ {
			kwMetricsCountClicksFormatted := formatInteger.Sprintf("%d", kwMetricsCountClicksNB[i])
			htmlContent += fmt.Sprintf("<tr>\n"+
				"    <td>%s</td>\n"+
				"    <td>%s</td>\n"+
				"    <td>%.2f%%</td>\n"+
				"    <td>%.2f</td>\n"+
				"</tr>\n",
				kwKeywordsNB[i],
				kwMetricsCountClicksFormatted,
				kwMetricsCTRNB[i],
				kwMetricsAvgPositionNB[i])
		}
	}

	htmlContent += fmt.Sprintf("</table>\n")
	htmlContent += fmt.Sprintf("</body>\n")
	htmlContent += fmt.Sprintf("</html>\n")

	// Save the HTML to a file
	// Branded keywords details
	if brandedMode {
		saveHTML(htmlContent, "./seoDataInsightKeywordsKPIsBranded.html")
	}
	// Branded keywords details
	if !brandedMode {
		saveHTML(htmlContent, "./seoDataInsightKeywordsKPIsNonBranded.html")
	}
}

// generate the slice containing the projected revenue data
func projectionDataCompute() {

	// First create a slice containing the visit ranges
	numElements := projectionMaxVisits/projectionIncrement + 1
	projectionVisitIncrements = make([]int, numElements)
	projectionVisitIncrementsString = make([]string, numElements)

	// Populate the slice with the visit ranges
	formatInteger := message.NewPrinter(language.English)

	for i := 0; i < numElements; i++ {
		projectionVisitIncrements[i] = i * projectionIncrement
		// Create a formatted String version for use in the chart XAxis
		projectionVisitIncrementsString[i] = formatInteger.Sprintf("%d", projectionVisitIncrements[i])
	}

	// Create a slice to hold the projection revenue values
	projectionRevenue = make([]int, numElements)
	// Populate the projection revenue slice
	for i := 0; i < numElements; i++ {
		projectionRevenue[i] = projectionVisitIncrements[i] / totalAverageVisitsPerOrder * totalAverageOrderValue
	}
}

// Revenue projection line chart
func lineChartRevenueProjection() {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Revenue projection",
			Subtitle: "What is the revenue potential of increasing the number of organic visits?",
			Link:     projectURL,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  chartDefaultWidth,
			Height: chartDefaultHeight,
		}),

		charts.WithColorsOpts(opts.Colors{kpiColourRevenueProjection}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Pass visitsPerOrder directly to generaLineItems
	lineVisitsPerOrderValue := generateLineItemsRevenueProjection(projectionRevenue)

	line.SetXAxis(projectionVisitIncrementsString).AddSeries("Revenue projection", lineVisitsPerOrderValue).SetSeriesOptions(

		charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: 0.2,
		}),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithMarkPointNameTypeItemOpts(
			opts.MarkPointNameTypeItem{Name: "Maximum revenue", Type: "max"},
			opts.MarkPointNameTypeItem{Name: "Average revenue", Type: "average"},
			opts.MarkPointNameTypeItem{Name: "Minimum revenue", Type: "min"},
		),
		charts.WithMarkPointStyleOpts(
			opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(true)}},
		),
	)

	f, _ := os.Create("./seoVisitsPerOrderLineRevenueProjection.html")
	line.Render(f)
}

// Populate the chart with the revenue projection data
func generateLineItemsRevenueProjection(projectionRevenue []int) []opts.LineData {
	items := make([]opts.LineData, len(projectionRevenue))
	for i, val := range projectionRevenue {
		items[i] = opts.LineData{Value: val}
	}
	return items
}

func projectionNarrative() {

	var htmlFileName = ""

	var noOfOrderVisits = projectionIncrement / totalAverageVisitsPerOrder
	var projectedRevenue = noOfOrderVisits * totalAverageOrderValue

	// HTML content for the revenue projection narrative
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<style>
        body {
            font-family: 'Arial', sans-serif;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            background-color: Cornsilk;
        }
        .content {
            max-width: 600px;
            text-align: center;  
            padding-bottom: 40px; 
        }
        .blueText {
            color: DeepSkyBlue;
            font-size: 25px;
            font-weight: bold;  
        }
        .keyword-font {
            font-family: 'Arial', sans-serif;
            font-size: 18px;
            color: LightSlateGray;
            line-height: 1.6; 
        }
        b {
            color: #333;
        }
    </style>
</head>
<body>
	<div class="content">
		<p class="keyword-font">
			On average over the period the number of visits required in order to generate one order is 
			<span class="blueText">%d</span>. For each additional 
			<span class="blueText">%d</span> visits, we can project 
			<span class="blueText">%d</span> orders will be placed. With an average 
			order value of <span class="blueText">%d</span>, the projected 
			incremental revenue for <span class="blueText">%d</span> visits will be 
			<span class="blueText">%d</span>.
		</p>
	</div>
</body>
</html>
`, totalAverageVisitsPerOrder, projectionIncrement, noOfOrderVisits, totalAverageOrderValue,
		projectionIncrement, projectedRevenue,
	)

	// Define the HTML filename
	htmlFileName = "./seoVisitsPerOrderLineRevenueProjectionNarrative.html"

	// Save the HTML to a file
	saveHTML(htmlContent, htmlFileName)

}

// Footer notes
func footerNotes() {

	// Text content for the footer
	var footerNotesStrings = []string{
		"These current month is not included in the analysis, only full months are reported on",
		"Compound growth rate refers to CMGR. CMGR is a financial term used to measure the growth rate of a metric over a monthly basis taking into account the compounding effect",
	}

	// Generate HTML content
	htmlContent := `<html>
<head>
</head>
	<style>
		body {
		font-family: Arial, sans-serif;
		font-size: 15px;
		color: LightSlateGray;
		}
	</style>
<body>
    <ul>
    <hr>
`
	for _, note := range footerNotesStrings {
		htmlContent += fmt.Sprintf("<li>%s</li>\n", note)
	}

	htmlContent += "</ul>\n</body>\n</html>"

	// Save the HTML to a file
	saveHTML(htmlContent, "./seoFooterNotes.html")

}

// formatDate converts date from YYYYMMDD to Month-Year format
func formatDate(dateStr string) string {
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		fmt.Println(red+"Error. formatDate. Cannot parse date:"+reset, err)
		return dateStr // return the original string in case of error
	}
	return date.Format("January 2006")
}

// Function used to generate and save the HTML content to a file
func saveHTML(genHTML string, genFilename string) {

	file, err := os.Create(genFilename)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Can create %s:"+reset, genFilename, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Can write %s:"+reset, genFilename, err)
		return
	}
}

// Define the HTML for the go_seo_dashboard.html container. Used to consolidate the generated charts into a single page.
func generateDashboard() {

	htmlContent := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go_Seo Dashboard</title>
    <style>
        body {
            margin: 0;
            font-family: Arial, sans-serif;
            background-color: Cornsilk;
        }
        .banner {
            background-color: DeepSkyBlue;
            color: white;
            text-align: center;
            padding: 15px 0;
        }
        .banner.top {
            font-size: 24px;
        }
        .banner.bottom {
            font-size: 12px;
        }
        .title {
            background-color: LightCyan;
            color: LightSlateGray;
            text-align: center;
            padding: 5px;
            margin: 5px auto;
            font-size: 22px;
            border-radius: 10px;
            width: 90%;
        }
        .iframe-container {
            display: flex;
            flex-wrap: wrap;
            align-items: center;
            gap: 20px;
            margin: 5px auto;
            width: 90%;
        }
        .iframe-container.row {
            flex-wrap: nowrap;
        }
        iframe {
            flex: 1 1 auto;
            min-width: 200px;
            width: 100%;
            border: 2px solid LightGray;
            border-radius: 10px;
        }
        .iframe-container.row iframe {
            height: 530px;
        }
        .iframe-container.no-border iframe {
            border: none;
        }
        .column iframe {
            height: 500px;
        }
        a {
            color: white;
            text-decoration: none;
        }
.badge-container {
            display: flex;
            flex-wrap: wrap;
            align-items: center;
            gap: 20px;
            margin: 5px auto;
            width: 90%;
        }
        .badge-container.row {
            flex-wrap: nowrap;
        }
        iframe {
            flex: 1 1 auto;
            min-width: 200px;
            width: 100%;
            border: 2px solid LightGray;
            border-radius: 10px;
        }
        .badge-container.row iframe {
            height: 400px;
        }
.back-button {
            padding: 12px 24px;
            font-size: 18px;
            color: white;
            background-color: #007BFF;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            position: fixed;
            bottom: 20px;
            right: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            transition: background-color 0.3s, box-shadow 0.3s;
        }
        .back-button:hover {
            background-color: #0056b3;
            box-shadow: 0 6px 8px rgba(0, 0, 0, 0.15);
        }
    </style>
</head>
<body>

<!-- Top Banner -->
<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">Business insights dashboard</span>
</header>

<!-- Back Button -->
<button class="back-button" onclick="goHome()">New dashboard</button>

<script>
    function goHome() {
        window.open('http://localhost:8080/', '_blank');
    }
</script>


<!-- Header for the totals display -->
<section class="iframe-container row no-border">
    <iframe src="seoDashboardHeader.html" title="Totals" style="height: 30px;"></iframe>
</section>

<!-- Totals of visits, orders & revenue -->
<section class="iframe-container row no-border">
    <iframe src="seoTableTotalsVisitsOrdersRevenue.html" title="Totals" style="height: 120px;"></iframe>
</section>

<!-- Bar chart for revenue & visits and line chart for visits per order -->
<section class="iframe-container row">
    <iframe src="seoRevenueVisitsBar.html" title="Revenue & visits"></iframe>
</section>

<!-- Bar chart for revenue & visits and line chart for visits per order -->
<section class="badge-container row">
    <iframe src="seoCMGRRevenue.html" title="CMGR Revenue"></iframe>
    <iframe src="seoCMGRVisits.html" title="CMGR Visits"></iframe>
</section>

<section class="iframe-container row">
    <iframe src="seoVisitsPerOrderLine.html" title="Visits per order"></iframe>
</section>

<!-- Orders and order value bar charts -->
<section class="iframe-container row">
    <iframe src="seoOrdersBar.html" title="No. of orders"></iframe>
</section>

<section class="iframe-container row">
    <iframe src="seoVisitsPerOrderLineRevenueProjection.html" title="Revenue projection"></iframe>
    <iframe src="seoVisitsPerOrderLineRevenueProjectionNarrative.html" title="Visits per order"></iframe>
</section>

<section class="badge-container row">
    <iframe src="seoGauge.html" title="Visits per order gauge"></iframe>
    <iframe src="seoCMGROrders.html" title="CMGR Orders"></iframe>
</section>

<!-- Orders and order value bar charts -->
<section class="iframe-container row">
    <iframe src="seoOrderValueBar.html" title="Order value"></iframe>
</section>

<!-- Revenue/Visits river chart and visit value bar chart -->
<section class="iframe-container row">
    <iframe src="seoVisitsRevenueRiver.html" title="Revenue & visits"></iframe>
</section>

<!-- Revenue/Visits river chart and visit value bar chart -->
<section class="iframe-container row">
    <iframe src="seoVisitValueBar.html" title="Organic visit value"></iframe>
</section>

<!-- Revenue/Visits river chart and visit value bar chart -->
<section class="badge-container row">
    <iframe src="seoCMGRVisitValue.html" title="CMGR Visit Value"></iframe>
    <iframe src="seoCMGROrderValue.html" title="CMGR Order Value"></iframe>
</section>

<!-- KPI details table-->
<section class="iframe-container row no-border">
    <iframe src="seoDataInsightDetailKPIs.html" style="height: 630px; title="KPIs"></iframe>
</section>

<!-- Wordclouds for top keywords -->
<section class="iframe-container row no-border">
    <iframe src="seoWordCloudBranded.html" title="Branded Keyword wordcloud" style="height: 700px; font-size: 10px;"></iframe>
</section>

<!-- Wordclouds for top keywords -->
<section class="iframe-container row no-border">
    <iframe src="seoWinningKeywordBranded.html" title="Winning branded keyword" style="height: 190px; font-size: 10px;"></iframe>
</section>

<!-- Wordclouds for top keywords -->
<section class="iframe-container row no-border">
    <iframe src="seoWordCloudNonBranded.html" title="Non Branded Keyword wordcloud" style="height: 700px; font-size: 10px;"></iframe>
</section>

<!-- Wordclouds for top keywords -->
<section class="iframe-container row no-border">
    <iframe src="seoWinningKeywordNonBranded.html" title="Winning non Branded keyword" style="height: 190px; font-size: 10px;"></iframe>
</section>

<!-- Keywords insights for top keywords - Branded-->
<section class="iframe-container row no-border">
    <iframe src="seoDataInsightKeywordsKPIsBranded.html" title="Branded keyword insights"; style="height: 600px; font-size: 10px;"></iframe>
</section>

<!-- Keywords insights for top keywords - Non branded -->
<section class="iframe-container row no-border">
    <iframe src="seoDataInsightKeywordsKPIsNonBranded.html" title="Branded keyword insights"; style="height: 600px; font-size: 10px;"></iframe>
</section>

<!-- Footer notes -->
<footer class="iframe-container row">
    <iframe src="seoFooterNotes.html" title="Footer notes" style="height: 200px; font-size: 10px; border: none;"></iframe>
</footer>

<!-- Bottom Banner -->
<footer class="banner bottom">
    Go_Seo. Jason Vicinanza. Github: <a href="https://github.com/flaneur7508/Go_SEO">https://github.com/flaneur7508/Go_SEO</a>
</footer>

</body>
</html>`

	// Save the HTML to a file
	saveHTML(htmlContent, "./go_seo_dashboard.html")

}

func executeBQL(returnSize int, bqlToExecute string) []byte {
	// If a size needs to be added to the URL, define it here
	var returnSizeAppend string
	if returnSize > 0 {
		returnSizeAppend = "?size=" + fmt.Sprint(returnSize)
	}

	// Define the URL
	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query%s", orgName, projectName, returnSizeAppend)

	// Define the body
	httpBody := []byte(bqlToExecute)

	// Create the POST request
	req, errorCheck := http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
	if errorCheck != nil {
		log.Fatal("Error. executeBQL. Cannot create request. Perhaps the provided credentials are invalid: ", errorCheck)
	}

	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botifyApiToken)
	req.Header.Add("Content-Type", "application/json")

	// Create HTTP client and execute the request
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, errorCheck := client.Do(req)
	if errorCheck != nil {
		log.Fatal("Error. executeBQL.  Cannot create the HTTP client: ", errorCheck)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, errorCheck := io.ReadAll(resp.Body)
	if errorCheck != nil {
		log.Fatal("Error. executeBQL. Cannot read response body: ", errorCheck)
	}

	// Return the response body as a byte slice
	return responseData
}

// CMGR
func calculateCMGR() {

	// Revenue
	// Convert slice of integers to slice of floats for CMGR compute
	var seoMetricsRevenueFloat []float64
	for _, v := range seoMetricsRevenue {
		seoMetricsRevenueFloat = append(seoMetricsRevenueFloat, float64(v))
	}
	cmgrRevenue = computeCMGR(seoMetricsRevenueFloat)

	// Visits
	var seoMetricsVisitsFloat []float64
	for _, v := range seoMetricsVisits {
		seoMetricsVisitsFloat = append(seoMetricsVisitsFloat, float64(v))
	}
	cmgrVisits = computeCMGR(seoMetricsVisitsFloat)

	// Visit value
	var seoMetricsVisitValueFloat []float64
	for _, v := range seoVisitValue {
		seoMetricsVisitValueFloat = append(seoMetricsVisitValueFloat, v)
	}
	cmgrVisitValue = computeCMGR(seoMetricsVisitValueFloat)

	// No. of Orders
	var seoMetricsOrdersFloat []float64
	for _, v := range seoMetricsOrders {
		seoMetricsOrdersFloat = append(seoMetricsOrdersFloat, float64(v))
	}
	cmgrOrdersValue = computeCMGR(seoMetricsOrdersFloat)

	// Order value
	var seoMetricsOrdersValueFloat []float64
	for _, v := range seoOrderValue {
		seoMetricsOrdersValueFloat = append(seoMetricsOrdersValueFloat, float64(v))
	}
	cmgrOrdersValueValue := computeCMGR(seoMetricsOrdersValueFloat)

	fmt.Printf(green + "\nCompound Monthly Growth Rate\n" + reset)
	fmt.Printf("Revenue: %.2f\n", cmgrRevenue)
	fmt.Printf("Visits: %.2f\n", cmgrVisits)
	fmt.Printf("Visit value: %.2f\n", cmgrVisitValue)
	fmt.Printf("No. of Orders: %.2f\n", cmgrOrdersValue)
	fmt.Printf("Order value: %.2f\n", cmgrOrdersValueValue)
}

// Calculate the Compound Monthly Growth Rate
func computeCMGR(values []float64) float64 {

	if len(values) < 2 {
		return 0.0 // Cannot calculate CMGR with less than 2 values
	}

	initialValue := values[0]

	// The final period value is not included as it is not a full month
	finalValue := values[len(values)-2]
	numberOfPeriods := float64(len(values))

	// CMGR formula: (finalValue / initialValue) ^ (1 / numberOfPeriods) - 1
	cmgr := math.Pow(finalValue/initialValue, 1/numberOfPeriods) - 1

	// Round CMGR to 2 decimal places
	cmgr = math.Round(cmgr*100) / 100

	return cmgr
}

// Get the analytics ID
func getAnalyticsID() string {
	// First identify which analytics tool is integrated
	urlAPIAnalyticsID := "https://api.botify.com/v1/projects/" + orgName + "/" + projectName + "/collections"
	req, errorCheck := http.NewRequest("GET", urlAPIAnalyticsID, nil)

	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botifyApiToken)
	req.Header.Add("Content-Type", "application/json")

	if errorCheck != nil {
		log.Fatal(red+"\nError. getAnalyticsID. Cannot create request: "+reset, errorCheck)
	}
	// Create HTTP client and execute the request
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, errorCheck := client.Do(req)
	if errorCheck != nil {
		log.Fatal(red+"Error. getAnalyticsID. Error: "+reset, errorCheck)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, errorCheck := io.ReadAll(resp.Body)
	if errorCheck != nil {
		log.Fatal(red+"Error. getAnalyticsID. Cannot read response body: "+reset, errorCheck)
	}

	// Unmarshal the JSON data into the struct
	var analyticsIDs []AnalyticsID
	if err := json.Unmarshal(responseData, &analyticsIDs); err != nil {
		fmt.Println(red+"Error. getAnalyticsID. The organisation and/or project name are probably incorrect. Cannot unmarshall the JSON:: "+reset, err)
		return "errorNoProjectFound"
	}

	// Find and print the name value when the ID contains the word "visit"
	// Assume the first instance of "visit" contains the analytics ID
	for _, analyticsID := range analyticsIDs {
		if strings.Contains(analyticsID.ID, "visit") {
			return analyticsID.ID
		}
	}
	return "errorNoAnalyticsIntegrated"
}

// Get the date ranges for the revenue and visits
func calculateDateRanges() DateRanges {
	currentTime := time.Now()
	dateRanges := make([][2]time.Time, 12)

	// Calculate the date ranges for the last 12 months
	for i := 0; i < 12; i++ {
		// Calculate the start and end dates for the current range
		// Adjust to the previous month. We don't count the current month.
		prevMonth := currentTime.AddDate(0, -1, 0)

		// Start of the previous month range
		startDate := time.Date(prevMonth.Year(), prevMonth.Month(), 1, 0, 0, 0, 0, currentTime.Location())

		var endDate time.Time
		if i == 0 {
			// The end date is the End of the previous month. We don't use the current month for the analysis.
			firstDayOfCurrentMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
			endDate = firstDayOfCurrentMonth.AddDate(0, 0, -1)
		} else {
			// End of the previous month range
			endDate = startDate.AddDate(0, 1, -1)
		}

		// Store the range
		dateRanges[11-i] = [2]time.Time{startDate, endDate}

		// Move to the previous month
		currentTime = startDate.AddDate(0, 0, 0)
	}

	// Subtract 1 day from the end date in the last element of the array
	dateRanges[0][1] = dateRanges[0][1].AddDate(0, 0, -1)

	// Save the number of months
	noOfMonths = len(dateRanges)

	return DateRanges{MonthlyRanges: dateRanges}
}

// Define the error page
func generateErrorPage(displayMessage string) {

	// If displayMessage is empty or nil display a default error message.
	if displayMessage == "" {
		displayMessage = "An Unknown error has occurred" // Provide a default message if needed
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go_Seo Dashboard</title>
    <style>
        body {
            margin: 0;
            font-family: Arial, sans-serif;
            background-color: Cornsilk;
        }
        .banner {
            background-color: DeepSkyBlue;
            color: white;
            text-align: center;
            padding: 15px 0;
        }
        .banner.top {
            font-size: 24px;
        }
        .back-button {
            padding: 12px 24px;
            font-size: 18px;
            color: white;
            background-color: #007BFF;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            position: fixed;
            bottom: 20px;
            right: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            transition: background-color 0.3s, box-shadow 0.3s;
        }
        .back-button:hover {
            background-color: #0056b3;
            box-shadow: 0 6px 8px rgba(0, 0, 0, 0.15);
        }
		.error-container {
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        .error-message {
            color: red;
            font-weight: bold;
            text-align: center;
            padding: 100px;
        }
    </style>
</head>
<body>

<!-- Top Banner -->
<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">Business insights dashboard</span>
</header>

<!-- Back Button -->
<button class="back-button" onclick="goHome()">Try again</button>

<!-- Error message -->
<div class="error-message" id="error-message">
    %s
</div>

<script>
    function goHome() {
        window.open('http://localhost:8080/', '_blank');
    }
</script>

</body>
</html>`, displayMessage)

	// Save the HTML to a file
	saveHTML(htmlContent, "./go_seo_errorPage.html")

}

func writeLog(sessionID, orgName, projectName, analyticsID, statusDescription string) {
	// Define log file name
	fileName := "_seoDashboardlogfile.log"

	// Check if the log file exists
	fileExists := true
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fileExists = false
	}

	// Open or create the log file
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf(red+"Error. writeLog. Cannot oprn log file: %s"+reset, err)
	}
	defer file.Close()

	// Get current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Construct log record
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
		sessionID, currentTime, orgName, projectName, analyticsID, statusDescription)

	// If the file doesn't exist, write header first
	if !fileExists {
		header := "SessionID,Date,Organisation,Project,AnalyticsID,Status\n"
		if _, err := file.WriteString(header); err != nil {
			log.Fatalf(red+"Error. writeLog. Failed to write log header: %s"+reset, err)
		}
	}

	// Write log record to file
	if _, err := file.WriteString(logRecord); err != nil {
		log.Fatalf(red+"Error. writeLog. Cannot write to log file: %s", err)
	}
}

func generateLogSessionID(length int) (string, error) {
	// Generate random bytes
	sessionIDLength := make([]byte, length)
	if _, err := rand.Read(sessionIDLength); err != nil {
		return "", err
	}
	// Encode bytes to base64 string
	return base64.URLEncoding.EncodeToString(sessionIDLength), nil
}

// Display the welcome banner
func displayBanner() {

	// Clear the screen
	fmt.Print(clearScreen)

	fmt.Println(green + `
 ██████╗  ██████╗         ███████╗███████╗ ██████╗ 
██╔════╝ ██╔═══██╗        ██╔════╝██╔════╝██╔═══██╗
██║  ███╗██║   ██║        ███████╗█████╗  ██║   ██║
██║   ██║██║   ██║        ╚════██║██╔══╝  ██║   ██║
╚██████╔╝╚██████╔╝███████╗███████║███████╗╚██████╔╝
 ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚══════╝ ╚═════╝
`)

	fmt.Println(purple + `
██████╗  █████╗ ███████╗██╗  ██╗██████╗  ██████╗  █████╗ ██████╗ ██████╗     ███████╗███████╗██████╗ ██╗   ██╗███████╗██████╗ 
██╔══██╗██╔══██╗██╔════╝██║  ██║██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██╔══██╗    ██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
██║  ██║███████║███████╗███████║██████╔╝██║   ██║███████║██████╔╝██║  ██║    ███████╗█████╗  ██████╔╝██║   ██║█████╗  ██████╔╝
██║  ██║██╔══██║╚════██║██╔══██║██╔══██╗██║   ██║██╔══██║██╔══██╗██║  ██║    ╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗
██████╔╝██║  ██║███████║██║  ██║██████╔╝╚██████╔╝██║  ██║██║  ██║██████╔╝    ███████║███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║
  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝     ╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝ 
`)
	//Display welcome message
	fmt.Println(purple+"Version:"+reset, version)

	fmt.Println(green + "\nThe Go_Seo dashboard server is ON.\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Server started at " + formattedTime + reset)

	fmt.Println(green + "\n... waiting for requests\n" + reset)

}
