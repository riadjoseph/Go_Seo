// seoCharts: SEO insights dashboard
// Written by Jason Vicinanza

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Anonymous mode. When set to true the URL to the project defaults to 'https://www.botify.com'
// If set to false a link is provided to the Botify project
var anonymousMode = false

// DateRanges struct used to hold the monthly date ranges
// Used for revenue and visits data
type DateRanges struct {
	MonthlyRanges [][2]time.Time
}

// Slice used to store the month names
var startMthNames []string

// Struct used to store Keywords dimensions and metrics
type KeywordsData struct {
	Results []struct {
		Dimensions []interface{} `json:"dimensions"`
		Metrics    []*float64    `json:"metrics,omitempty"`
	} `json:"results"`
}

// Used for the branded/non branded title in the wordcloud
var wordcloudTitle = ""
var wordcloudSubTitle = ""

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

// Slices used to store non branded Keywords KPIsd
var kwKeywordsNB []string
var kwMetricsCountUrlsNB []int
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
var totalVisitsPerOrder float64

// Slice to store the visits per order for each month
var visitsPerOrder []int

// AnalyticsID is used to identify which analytics tool is in use and is used for some API cals
type AnalyticsID struct {
	ID string `json:"id"`
}

// The Result struct is used to store the revenue, Orders and visits
type Result struct {
	Dimensions []interface{} `json:"dimensions"`
	Metrics    []float64     `json:"metrics"`
}
type Response struct {
	Results []Result `json:"results"`
}

// Project URL. Used to provide a link to the Botify project
var projectURL = ""

// Organisation name used for display purposes
var displayOrgName = ""

// Version
var version = "v0.1"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var yellow = "\033[33m"
var bold = "\033[1m"
var reset = "\033[0m"
var checkmark = "\u2713"

// KPI Specific colours
var kpiColourRevenue = "Coral"
var kpiColourVisits = "Green"
var kpiColourVisitsPerOrder = "DarkGoldenRod"
var kpiColourOrganicVisitValue = "CornflowerBlue"
var kpiColourNoOfOrders = "IndianRed"
var kpiColourOrderValue = "MediumSlateBlue"

// Botify API token
var botifyApiToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

// Variables used to store the min and max visits per order
var minVisitsPerOrder = 0
var maxVisitsPerOrder = 0

// No. of months processed
var noOfMonths = 0

// Average visits per order
var averageVisitsPerOrder = 0

func main() {

	clearScreen()

	displayBanner()

	// Get the credentials if they have not been specified on the command line
	checkCredentials()

	// If the credentials have been provided on the command line use them
	if !credentialsInput {
		orgName = os.Args[1]
		projectName = os.Args[2]
	} else {
		orgName = orgNameInput
		projectName = projectNameInput
	}

	fmt.Println(bold+"\nOrganisation name:", orgName)
	fmt.Println(bold+"Project name:", projectName+reset)
	fmt.Println()

	// Generate the link to the project
	if anonymousMode {
		projectURL = "http://www.botify.com/"
		displayOrgName = "Anonymised"
	} else {
		projectURL = "https://app.botify.com/" + orgName + "/" + projectName
		displayOrgName = orgName
	}

	displaySeparator()

	// Get revenue, visits and orders
	getSeoInsights()

	// Start of charts

	// Insert the header
	dashboardHeader()

	// Total vales
	tableTotalsVisitsOrdersRevenue()

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

	// Badges
	// Values for the badges
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

	// Revenue and visits river chart
	riverCharRevenueVisits()

	// Visits per order Gauge
	gaugeVisitsPerOrder(totalVisitsPerOrder)

	// Wordclouds
	// Branded keywords
	wordCloudBrandedUnbranded(true)
	// Non branded keywords
	wordCloudBrandedUnbranded(false)

	// Winning branded keyword
	winningKeywords(true)
	// Winning non branded keyword
	winningKeywords(false)

	// KPI details table
	tableDataDetail()

	// Footer notes
	footerNotes()

	displaySeparator()

	seoChartsDone()
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true
		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit seoCharts" +
			"\n")
		fmt.Print(purple + "\nEnter organisation name: " + reset)
		_, _ = fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using seoCharts. Goodbye!")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		_, _ = fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using seoCharts. Goodbye!")
			os.Exit(0)
		}
	}
}

func getSeoInsights() {

	fmt.Println(purple + bold + "\nGetting SEO insights" + reset)

	// Get the date ranges
	dateRanges := calculateDateRanges()

	// Identify which analytics tool is used
	analyticsID := getAnalyticsID()
	fmt.Println("Analytics identified:", analyticsID)

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
	getRevenueData(analyticsID, startMthDates, endMthDates)

	// Get the keywords data
	// Get last months' date range
	kwStartDate := startMthDates[len(startMthDates)-1]
	kwEndDate := endMthDates[len(endMthDates)-1]

	getKeywordsData(kwStartDate, kwEndDate)

	// Calculate the CMGR values
	calculateCMGR()

}

// Get the revenue, orders and visits data
func getRevenueData(analyticsID string, startMthDates []string, endMthDates []string) {

	var metricsOrders = 0
	var metricsRevenue = 0
	var metricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Get monthly insights
	fmt.Println(purple + "\nMonthly organic insights\n" + reset)
	for i := range startMthDates {

		metricsOrders, metricsRevenue, metricsVisits, avgOrderValue, avgVisitValue = executeRevenueBQL(analyticsID, startMthDates[i], endMthDates[i])

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

		// Display the KPIs
		fmt.Printf(green+"\nDate Start: %s End: %s\n"+reset, startMthDates[i], endMthDates[i])
		formattedOrders := formatWithCommas(metricsOrders)
		fmt.Println("No. Orders:", formattedOrders)
		formattedRevenue := formatWithCommas(metricsRevenue)
		fmt.Println("Total revenue:", formattedRevenue)
		fmt.Println("Average order value:", avgOrderValue)
		formattedVisits := formatWithCommas(metricsVisits)
		fmt.Println("No. of visits:", formattedVisits)
		fmt.Println("Average visit value:", avgVisitValue)
	}

	// Calculate the average visits per order
	totalVisitsPerOrder := 0
	for _, value := range visitsPerOrder {
		totalVisitsPerOrder += value
	}

	if len(visitsPerOrder) > 0 {
		averageVisitsPerOrder = totalVisitsPerOrder / len(visitsPerOrder)
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

	fmt.Println(green + "\nTotals" + reset)
	fmt.Println("Total visits:", totalVisits)
	fmt.Println("Total revenue:", totalRevenue)
	fmt.Println("Total orders:", totalOrders)
	fmt.Println("Average visits per order:", averageVisitsPerOrder)
}

// Get the keywords data
func getKeywordsData(startMthDates string, endMthDates string) {

	//bloo
	// Branded keywords
	executeKeywordsBQL(startMthDates, endMthDates, "true")

	// Non-branded keywords
	executeKeywordsBQL(startMthDates, endMthDates, "false")

}

// bloo continue from here
// Execute the BQL to acquire keywords data
func executeKeywordsBQL(startDate string, endDate string, brandedFlag string) ([]string, []int, []int, []float64, []float64) {

	// Get the keywords data
	bqlKeywords := fmt.Sprintf(`{
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

	// Define the URL
	// We get 21 keywords because the first keyword is not included in the wordcloud
	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query?size=21", orgName, projectName)

	// GET the HTTP request
	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal(red+"\nError. executeKeywordBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
	}

	// Define the body
	httpBody := []byte(bqlKeywords)

	// Create the POST request
	req, errorCheck = http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
	if errorCheck != nil {
		log.Fatal("Error. executeKeywordsBQL. Cannot create request. Perhaps the provided credentials are invalid: ", errorCheck)
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
		log.Fatal(red+"Error. executeKeywordsBQL.  Cannot create the HTTP client: "+reset, errorCheck)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, errorCheck := ioutil.ReadAll(resp.Body)
	if errorCheck != nil {
		log.Fatal(red+"Error. executeKeywordsBQL. Cannot read response body: "+reset, errorCheck)
	}

	// Unmarshal JSON data into KeywordsData struct
	var response KeywordsData
	err := json.Unmarshal(responseData, &response)
	if err != nil {
		log.Fatalf("Error. executeKeywordsBQL. Cannot unmarshal the JSON: %v", err)
		os.Exit(1)
	}

	// Load the response into the slices - branded keywords
	if brandedFlag == "true" {
		for _, result := range response.Results {
			if len(result.Dimensions) >= 1 && len(result.Metrics) >= 3 {
				kwKeywords = append(kwKeywords, result.Dimensions[0].(string))
				kwMetricsCountClicks = append(kwMetricsCountClicks, int(*result.Metrics[0]))
				kwMetricsAvgPosition = append(kwMetricsAvgPosition, float64(*result.Metrics[1])) //bloo error
				kwMetricsCTR = append(kwMetricsCTR, float64(*result.Metrics[2]))
			}
		}
	}

	// Load the response into the slices - non-branded keywords
	if brandedFlag == "false" {
		for _, result := range response.Results {
			if len(result.Dimensions) >= 1 && len(result.Metrics) >= 3 {
				kwKeywordsNB = append(kwKeywordsNB, result.Dimensions[0].(string))
				kwMetricsCountClicksNB = append(kwMetricsCountClicksNB, int(*result.Metrics[0]))
				kwMetricsAvgPositionNB = append(kwMetricsAvgPositionNB, float64(*result.Metrics[1]))
				kwMetricsCTRNB = append(kwMetricsCTRNB, float64(*result.Metrics[2]))
			}
		}
	}

	return kwKeywords, kwMetricsCountUrls, kwMetricsCountClicks, kwMetricsCTR, kwMetricsAvgPosition
}

// Get the analytics ID
func getAnalyticsID() string {
	// First identify which analytics tool is integrated
	urlAPIAnalyticsID := "https://api.botify.com/v1/projects/" + orgName + "/" + projectName + "/collections"
	//fmt.Println(bold+"\nAnalytics ID end point:"+reset, urlAPIAnalyticsID)
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
		log.Fatal("Error. getAnalyticsID. Error: ", errorCheck)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, errorCheck := ioutil.ReadAll(resp.Body)
	if errorCheck != nil {
		log.Fatal(red+"Error. getAnalyticsID. Cannot read response body: "+reset, errorCheck)
	}

	// Unmarshal the JSON data into the struct
	var analyticsIDs []AnalyticsID
	if err := json.Unmarshal(responseData, &analyticsIDs); err != nil {
		log.Fatal(red+"Error. getAnalyticsID. Cannot unmarshall the JSON: "+reset, err)
	}

	// Find and print the name value when the ID contains the word "visit"
	// Assume the first instance of "visit" contains the analytics ID
	for _, analyticsID := range analyticsIDs {
		if strings.Contains(analyticsID.ID, "visit") {
			return analyticsID.ID
		}
	}
	return "noAnalyticsFound"
}

// Get the date ranges for the revenue and visits
func calculateDateRanges() DateRanges {
	currentTime := time.Now()
	dateRanges := make([][2]time.Time, 12)

	// Calculate the date ranges for the last 12 months
	for i := 0; i < 12; i++ {
		// Calculate the start and end dates for the current range
		// Adjust to the previous month. We doint caount the current month.
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

// Execute the BQL for the specified date range
func executeRevenueBQL(analyticsID string, startDate string, endDate string) (int, int, int, int, float64) {

	// Get the revenue, no. Orders and visits
	bqlRevTrans := fmt.Sprintf(`
	{
    "collections": [
                    "conversion.dip",
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
                    "conversion.dip.period_0.transactions",
                    "conversion.dip.period_0.revenue",    
                    "visits.dip.period_0.nb"
        ],
        "filters": {
            "and": [
                {
                    "field": "conversion.dip.period_0.medium",
                    "predicate": "eq",
                    "value": "organic"
                },
                {
                    "field": "visits.dip.period_0.medium",
                    "predicate": "eq",
                    "value": "organic"
           	     }
      	      ]
    	    }
 	   }
	}`, analyticsID, startDate, endDate)

	// Define the URL
	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query", orgName, projectName)

	// GET the HTTP request
	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal(red+"\nError. executeRevenueBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
	}

	// Define the body
	httpBody := []byte(bqlRevTrans)

	// Create the POST request
	req, errorCheck = http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
	if errorCheck != nil {
		log.Fatal(red+"Error. executeRevenueBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
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
		log.Fatal("Error. executeRevenueBQL. Error: ", errorCheck)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, errorCheck := ioutil.ReadAll(resp.Body)
	if errorCheck != nil {
		log.Fatal(red+"Error. executeRevenueBQL. Cannot read response body: "+reset, errorCheck)
	}

	// Unmarshal the JSON data into the struct
	var response Response
	err := json.Unmarshal(responseData, &response)
	if err != nil {
		log.Fatalf(red+"Error. executeRevenueBQL. Cannot unmarshal the JSON: %v"+reset, err)
	}

	var metricsOrders = 0
	var metricsRevenue = 0
	var metricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Check if any data has been returned from the API. Count the number of elements in the response.Results slice
	responseCount := len(response.Results)

	if responseCount == 0 {
		fmt.Println(yellow + "Warning. executeRevenueBQL. Some data may default to 1 if it's the first day of the month." + reset)
	} else {
		metricsOrders = int(response.Results[0].Metrics[0])
		metricsRevenue = int(response.Results[0].Metrics[1])
		metricsVisits = int(response.Results[0].Metrics[2])
		// Compute the average Order value
		avgOrderValue = metricsRevenue / metricsOrders
		avgVisitValue = float64(metricsRevenue) / float64(metricsVisits)
	}
	return metricsOrders, metricsRevenue, metricsVisits, avgOrderValue, avgVisitValue
}

// Header for the dashboard
func dashboardHeader() {

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
	saveHTML(htmlContent, "./Utilities/seoChartsWeb/seoDashboardHeader.html")

}

// Table for total Visits, Orders & Revenue
func tableTotalsVisitsOrdersRevenue() {

	totalVisitsFormatted := formatInt(totalVisits)
	totalOrdersFormatted := formatInt(totalOrders)
	totalRevenueFormatted := formatInt(totalRevenue)

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
        }
     .wrapper {
            display: flex;
            justify-content: space-between; /* Distribute space between columns */
            width: 80%; /* Adjust the width as needed */
        }
        .column {
            flex: 1;
            text-align: center;
            margin: 0 130px; /* Increase margin for more space between columns */
        }
        th, td {
            padding: 4px; /* Adjust padding here */
            font-size: 30px; /* Double the font size */
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
                    <th>Visits</th>
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
	saveHTML(htmlContent, "./Utilities/seoChartsWeb/seoTableTotalsVisitsOrdersRevenue.html")

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
			Width:  "600px",
			Height: "480px",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}),
	)

	barDataRevenue := generateBarItems(seoMetricsRevenue)
	barDataVisits := generateBarItems(seoMetricsVisits)

	bar.SetXAxis(startMthNames).
		AddSeries("Revenue", barDataRevenue).
		AddSeries("Visits", barDataVisits).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Avg", Type: "average"},
		))

	f, _ := os.Create("./Utilities/seoChartsWeb/seoRevenueVisitsBar.html")
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
			Width:  "600px",
			Height: "480px",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisitsPerOrder}),
	)

	// Pass visitsPerOrder directly to generaLineItems
	lineVisitsPerOrderValue := generateLineItems(visitsPerOrder)

	line.SetXAxis(startMthNames).AddSeries("Month", lineVisitsPerOrderValue).SetSeriesOptions(
		charts.WithMarkPointNameTypeItemOpts(
			opts.MarkPointNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkPointNameTypeItem{Name: "Average", Type: "average"},
			opts.MarkPointNameTypeItem{Name: "Minimum", Type: "min"},
		),
		charts.WithMarkPointStyleOpts(
			opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(true)}},
		),
	)

	f, _ := os.Create("./Utilities/seoChartsWeb/seoVisitsPerOrderLine.html")
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
		Subtitle: "What is the value of a single organic visit to the site?",
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
			Width:  "600px",
			Height: "480px",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourOrganicVisitValue}),
	)

	barDataVisitValue := generateBarItemsFloat(seoVisitValue)

	bar.SetXAxis(startMthNames).
		AddSeries("Organic visit value", barDataVisitValue).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Avg", Type: "average"},
		))

	f, _ := os.Create("./Utilities/seoChartsWeb/seoVisitValueBar.html")
	bar.Render(f)
}

// Bar chart. No. of Orders
func barChartOrders() {

	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "No. of orders",
		Subtitle: "How many orders are placed by organic visitors to your site?",
		Link:     projectURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "600px",
			Height: "480px",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourNoOfOrders}),
	)

	barDataOrders := generateBarItems(seoMetricsOrders)

	bar.SetXAxis(startMthNames).
		AddSeries("No. of Orders", barDataOrders).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Avg", Type: "average"},
		))

	f, _ := os.Create("./Utilities/seoChartsWeb/seoOrdersBar.html")
	bar.Render(f)
}

// Bar chart. No. of Orders
func barChartOrderValue() {
	// create a new bar instance
	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Average order value",
		Subtitle: "What is the average value of an order placed by an organic visitor to your site?",
		Link:     projectURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "600px",
			Height: "480px",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourOrderValue}),
	)

	barDataOrderValue := generateBarItems(seoOrderValue)

	bar.SetXAxis(startMthNames).
		AddSeries("Order value", barDataOrderValue).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Avg", Type: "average"},
		))

	f, _ := os.Create("./Utilities/seoChartsWeb/seoOrderValueBar.html")
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
	badgeFileName := fmt.Sprintf("./Utilities/seoChartsWeb/seoCMGR%s.html", badgeKPI)
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
			Width:  "400px",
			Height: "480px",
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

// Top 20 keywords for branded and non branded
func wordCloudBrandedUnbranded(brandedMode bool) {

	// Generate the HTML for branded keywords
	if brandedMode {
		page := components.NewPage()
		page.AddCharts(
			generateWordCloud(true),
		)
		f, err := os.Create("./Utilities/seoChartsWeb/seoWordCloudBranded.html")
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
		f, err := os.Create("./Utilities/seoChartsWeb/seoWordCloudNonBranded.html")
		if err != nil {
			panic(err)
		}
		page.Render(io.MultiWriter(f))
	}
}

func generateWordCloud(brandedMode bool) *charts.WordCloud {

	if brandedMode {
		wordcloudTitle = "Top branded keywords"
		wordcloudSubTitle = "Branded keywords driving clicks to the site last month."
	}
	if !brandedMode {
		wordcloudTitle = "Top non branded keywords"
		wordcloudSubTitle = "Non branded keywords driving clicks to the site last month."
	}

	wc := charts.NewWordCloud()
	wc.SetGlobalOptions(
		// Remove the legend from the wordcloud
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithTitleOpts(opts.Title{
			Title:    wordcloudTitle,
			Subtitle: wordcloudSubTitle,
			Link:     projectURL,
		}))

	// Generate the branded wordcloud
	if brandedMode {
		wc.AddSeries("Keywords", generateWCData(kwKeywords, kwMetricsCountClicks)).
			SetSeriesOptions(
				charts.WithWorldCloudChartOpts(
					opts.WordCloudChart{
						SizeRange: []float32{10, 90},
						Shape:     "cardioid",
					}),
			)
	}

	// Generate the non-branded wordcloud
	if !brandedMode {
		wc.AddSeries("Keywords", generateWCData(kwKeywordsNB, kwMetricsCountClicks)).
			SetSeriesOptions(
				charts.WithWorldCloudChartOpts(
					opts.WordCloudChart{
						SizeRange: []float32{10, 90},
						Shape:     "cardioid",
					}),
			)
	}

	return wc
}

func generateWCData(kwKeywords []string, kwMetricsCountClicks []int) (items []opts.WordCloudData) {
	items = make([]opts.WordCloudData, 0)
	// Iterate over kwKeywords and kwMetricsCountClicks slices starting from index 1
	// We start at index 1 because the top keyword is generally significantly higher performing than the following keywords
	for i := 1; i < len(kwKeywords); i++ {
		// Check if index is valid for kwMetricsCountClicks slice
		if i < len(kwMetricsCountClicks) {
			// Append WordCloudData struct with keyword and corresponding count
			items = append(items, opts.WordCloudData{Name: kwKeywords[i], Value: kwMetricsCountClicks[i]})
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
	f, err := os.Create("./Utilities/seoChartsWeb/seoVisitsRevenueRiver.html")
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
			Width:  "600px",
			Height: "480px",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}),
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
func gaugeVisitsPerOrder(visitsPerOrder float64) {

	page := components.NewPage()
	page.AddCharts(
		gaugeBase(visitsPerOrder),
	)

	f, err := os.Create("./Utilities/seoChartsWeb/seoGauge.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func gaugeBase(visitsPerOrder float64) *charts.Gauge {

	gauge := charts.NewGauge()

	setMinMax := charts.WithSeriesOpts(func(s *charts.SingleSeries) {
		s.Min = minVisitsPerOrder
		s.Max = maxVisitsPerOrder
	})

	gauge.SetGlobalOptions(
		//  No options defined
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "500px",
			Height: "500px",
		}),
	)

	gauge.AddSeries("Visits Per Order", []opts.GaugeData{{Name: "Average visits per order", Value: averageVisitsPerOrder}}, setMinMax)

	return gauge
}

// CMGR
func calculateCMGR() {

	// Revenue
	// Convert slice of ints to slice of floats for CMGR compute
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
		seoMetricsVisitValueFloat = append(seoMetricsVisitValueFloat, float64(v))
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

func seoChartsDone() {

	// We're done
	fmt.Println(purple + "\nseoCharts: Done!")
	fmt.Println(bold + green + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)
}

// Generate an HTML table containing the detailed KPI insights
func tableDataDetail() {
	var detailedKPIstableData [][]string

	for i := 0; i < len(startMthDates); i++ {
		formattedDate := formatDate(startMthDates[i])
		orders := formatInt(seoMetricsOrders[i])
		revenue := formatInt(seoMetricsRevenue[i])
		orderValue := formatInt(seoOrderValue[i])
		visits := formatInt(seoMetricsVisits[i])
		visitValue := formatFloat(seoVisitValue[i])
		visitsPerOrder := formatInt(visitsPerOrder[i])

		row := []string{
			formattedDate,
			orders,
			revenue,
			orderValue,
			visits,
			visitValue,
			visitsPerOrder,
		}
		detailedKPIstableData = append(detailedKPIstableData, row)
	}

	// Generate the table
	htmlContent := generateHTMLDetailedKPIInsightsTable(detailedKPIstableData)

	// Save the HTML to a file
	saveHTML(htmlContent, "./Utilities/seoChartsWeb/seoDataInsightDetailKPIs.html")

}

func winningKeywords(brandedMode bool) {

	var file *os.File
	var err error

	if brandedMode {
		file, err = os.Create("./Utilities/seoChartsWeb/seoWinningKeywordBranded.html")
		if err != nil {
			fmt.Println(red+"Error. winningKeywords. Cannot create branded winning keyword HTML:"+reset, err)
			return
		}
		defer file.Close()
	}

	if !brandedMode {
		file, err = os.Create("./Utilities/seoChartsWeb/seoWinningKeywordNonBranded.html")
		if err != nil {
			fmt.Println(red+"Error. winningKeywords. Cannot create non branded winning keyword HTML:"+reset, err)
			return
		}
		defer file.Close()
	}

	// Define the display values based on branded or non branded mode
	var htmlKeyword = ""
	var htmlClicks = ""
	var htmlCTR float64
	var htmlAvgPosition float64

	if brandedMode {
		htmlKeyword = kwKeywords[0]
		htmlClicks = formatInt(kwMetricsCountClicks[0])
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
		htmlClicks = formatInt(kwMetricsCountClicksNB[0])
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
		.blueText {
			color: DeepSkyBlue;
			font-size: 40px;
		}
	.keyword-font {
		font-family: Arial, sans-serif;
		font-size: 20px;
		color: LightSlateGray;
	}
	</style>
</head>
<body>
	<p><span class="keyword-font">The winning keyword was <span class="blueText">%s<span class="keyword-font"> during <b>%s</b></span></span></p>
	<p><span class="keyword-font">This keyword generated <b>%s</b> clicks. The click through rate was <b>%.2f%%</b> from an average position of <b>%.2f</b></span></p>
</body>
</html>
`, htmlKeyword, htmlLastMonthName, htmlClicks, htmlCTR, htmlAvgPosition)

	_, err = file.WriteString(htmlContent)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

// Display the line break
func displaySeparator() {
	block := "█"
	fmt.Println()

	for i := 0; i < 130; i++ {
		fmt.Print(block)
	}

	fmt.Println()
}

// formatInt formats integer values with comma separator
func formatInt(num int) string {
	return formatIntWithCommas(int64(num))
}

// formatIntWithCommas formats an integer with commas as thousand separators
func formatIntWithCommas(num int64) string {
	in := strconv.FormatInt(num, 10)
	n := len(in)
	if n <= 3 {
		return in
	}

	var sb strings.Builder
	pre := n % 3
	if pre > 0 {
		sb.WriteString(in[:pre])
		if n > pre {
			sb.WriteString(",")
		}
	}

	for i := pre; i < n; i += 3 {
		sb.WriteString(in[i : i+3])
		if i+3 < n {
			sb.WriteString(",")
		}
	}

	return sb.String()
}

// formatFloat formats float values with 2 decimal places
func formatFloat(num float64) string {
	return strconv.FormatFloat(num, 'f', 2, 64)
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

// generateHTML generates the HTML content for the table
func generateHTMLDetailedKPIInsightsTable(data [][]string) string {
	html := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        table { width: 100%; border-collapse: collapse; margin: 25px 0; font-size: 18px; text-align: left; }
        th, td { padding: 12px; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        th.title { color: Gray; font-weight: bold; }
        td { color: DimGray; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        tr:hover { background-color: HoneyDew; }
    </style>
</head>
<body style="min-height: 10vh;">
    <table>
        <thead>
            <tr>
                <th class="title">Date</th>
                <th class="title">No. of Orders</th>
                <th class="title">Revenue</th>
                <th class="title">Order Value</th>
                <th class="title">No. of Visits</th>
                <th class="title">Visit Value</th>
                <th class="title">Visits per Order</th>

            </tr>
        </thead>
        <tbody>`
	for _, row := range data {
		html += "<tr>"
		for _, cell := range row {
			html += "<td>" + cell + "</td>"
		}
		html += "</tr>"
	}
	html += `
        </tbody>
    </table>
</body>
</html>`
	return html
}

// Function used to write HTML content to a file
func saveHTML(genHTML string, genFilename string) {
	file, err := os.Create(genFilename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Cannot save HTML:"+reset, err)
		return
	}
}

func footerNotes() {

	// Slice of strings used for the footer text
	var footerNotesStrings = []string{
		"1. Only complete months are included in the analysis.",
		"2. Compound growth rate refers to CMGR. CMGR is a financial term used to measure the growth rate of a business metric over a monthly basis taking into account the compounding effect. ",
	}

	// Create or open the HTML file
	file, err := os.Create("./Utilities/seoChartsWeb/seoFooterNotes.html")
	if err != nil {
		fmt.Println(red+"Error. footerNotes. Cannot create footer notes:"+reset, err)
		return
	}
	defer file.Close()

	// Write the HTML header
	_, err = file.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<body>\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	// Write each string as a paragraph in the HTML body
	for _, str := range footerNotesStrings {
		_, err := file.WriteString("<p>" + str + "</p>\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}

	// Write the HTML footer
	_, err = file.WriteString("</body>\n</html>")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

// Function to format an integer with comma separation
func formatWithCommas(n int) string {
	s := strconv.Itoa(n)
	nLen := len(s)
	if nLen <= 3 {
		return s
	}

	var result strings.Builder
	commaOffset := nLen % 3
	if commaOffset > 0 {
		result.WriteString(s[:commaOffset])
		if nLen > commaOffset {
			result.WriteString(",")
		}
	}

	for i := commaOffset; i < nLen; i += 3 {
		result.WriteString(s[i : i+3])
		if i+3 < nLen {
			result.WriteString(",")
		}
	}

	return result.String()
}

// Display the welcome banner
func displayBanner() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Println(green + `
███████╗███████╗ ██████╗  ██████╗██╗  ██╗ █████╗ ██████╗ ████████╗███████╗
██╔════╝██╔════╝██╔═══██╗██╔════╝██║  ██║██╔══██╗██╔══██╗╚══██╔══╝██╔════╝
███████╗█████╗  ██║   ██║██║     ███████║███████║██████╔╝   ██║   ███████╗
╚════██║██╔══╝  ██║   ██║██║     ██╔══██║██╔══██║██╔══██╗   ██║   ╚════██║
███████║███████╗╚██████╔╝╚██████╗██║  ██║██║  ██║██║  ██║   ██║   ███████║
╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚══════╝
`)
	//Display welcome message
	fmt.Println(purple+"Version:"+reset, version)

	fmt.Println(purple + "seoCharts: Test Botify BQL.\n" + reset)
	fmt.Println(purple + "Use it as a template for your Botify integration needs.\n" + reset)
	fmt.Println(purple + "BQL tests performed in this version.\n" + reset)
	fmt.Println(checkmark + green + bold + " Revenue (monthly)" + reset)
	fmt.Println(checkmark + green + bold + " Visits (monthly)" + reset)
	fmt.Println(checkmark + green + bold + " Orders (monthly)" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) Average order value" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) Average visit value" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) CMGR for Revenue, Visits, Orders, Order value, Visit value" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) Visits per order" + reset)
	fmt.Println(checkmark + green + bold + " Top branded keywords" + reset)
	fmt.Println(checkmark + green + bold + " Top non branded keywords" + reset)
}

// Function to clear the screen
func clearScreen() {

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
