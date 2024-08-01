// seoBusinessInsights: SEO insights broadsheet
// Written by Jason Vicinanza

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/ini.v1"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Version
var version = "v0.3"

// changelog v0.3
// UI updates & refinements (seoBusinessInsights)
// UI updates & refinements (index.html)
// Added env. variable "envInsightsHostingMode". Set to "local" or "docker"
// Bug fix. Error when running the broadsheet on the last day of the month

// changelog v0.2
// Added tooltips to login page (org and project name)
// Set API time out to 30 seconds
// Version displayed in broadsheet header
// Port & protocol not required keys in .ini file when hosted on Botify infra
// Wordclouds now display correctly (fixed 404 errors)
// Fixed "division by zero" error when EA is configured but not data availanle
// Removed experimental news section

// Token, log folder and cache folder acquired from environment variables
var envBotifyAPIToken string
var envInsightsLogFolder string
var envInsightsFolder string
var envInsightsHostingMode string

// Declare the mutex
var mutex sync.Mutex

// Colours, symbols etc
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var yellow = "\033[0;33m"
var white = "\033[37m"
var bold = "\033[1m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"
var lineSeparator = "█" + strings.Repeat("█", 130)

// KPI Specific colours
var kpiColourRevenue = "Coral"
var kpiColourVisits = "Green"
var kpiColourVisitsPerOrder = "DarkGoldenRod"
var kpiColourRevenueForecast = "Orange"
var kpiColourOrganicVisitValue = "CornflowerBlue"
var kpiColourNoOfOrders = "IndianRed"
var kpiColourOrderValue = "MediumSlateBlue"

// Slice used to store the month names. These are used in the chart X axis
var startMonthNames []string

// Slice used to store projected revenue values
var forecastRevenue []int

// Used for the branded/non branded keyword title in the wordcloud
var wordcloudTitle string

// Slices used to store the startMonthDate and endMonthDate
var startMonthDates = make([]string, 0)
var endMonthDates = make([]string, 0)

// Slices used to store the SEO metrics
var seoRevenue []int
var seoVisits []int
var seoOrders []int
var seoOrderValue []int
var seoVisitValue []float64
var seoVisitsPerOrder []int

// Slices used to store branded Keywords KPIs
var kwKeywords []string
var kwCountClicks []int
var kwMetricsCTR []float64
var kwMetricsAvgPosition []float64

// Slices used to store non-branded Keywords KPIsd
var kwKeywordsNonBranded []string
var kwCountClicksNonBranded []int
var kwCTRNonBranded []float64
var kwAvgPositionNonBranded []float64

// Variables used to store the CMGR values
var cmgrRevenue float64
var cmgrVisits float64
var cmgrVisitValue float64
var cmgrOrderValue float64
var cmgrOrderValueValue float64

// Variables used to store the total values
var totalVisits int
var totalRevenue int
var totalOrders int
var totalAverageOrderValue int

// Bools used to flag if some data is missing
var revenueDataIssue bool
var visitsDataIssue bool
var ordersDataIssue bool

// Project URL. Used to provide a link to the Botify project
var projectURL string

// Strings used to store the project credentials for API access
var organization string
var project string

// Variables used to store the min and max visits per order
var minVisitsPerOrder int
var maxVisitsPerOrder int

// No. of months processed
var noOfMonths int

// Average visits per order
var totalAverageVisitsPerOrder int

// Average visit value
var totalAverageVisitValue float64

// The number of keywords to include in the wordcloud
var noKeywordsInCloud = 50

// No. of keywords returned by the API
var noKeywordsFound int

// No of executions & generated session ID
var sessionIDCounter int
var sessionID string

// The number of top keywords to include in the keywords detail table
var noTopKeywords = 50

// Used to set the default size for all chart types
var chartDefaultWidth = "85vw"
var chartDefaultHeight = "90vh"

var wordcloudDefaultWidth = "95vw"
var wordcloudDefaultHeight = "90vh"

var badgeDefaultWidth = "95vw"
var badgeDefaultHeight = "90vh"

var gaugeDefaultWidth = "95vw"
var gaugeDefaultHeight = "90vh"

// Define the increment and the maximum value
var forecastIncrement = 500000
var forecastMaxVisits = 10000000

// Slices used to store the visit increment values
var forecastVisitIncrements []int
var forecastVisitIncrementsString []string

// Project currency
var currencyCode string
var currencySymbol string

// Name of the seoBusinessInsights folder used to store the generated HTML
var insightsCacheFolder string

// Host name and port the web server runs on
var protocol string
var hostname string
var port string
var fullHost string

// Dashboard permalink
var dashboardPermaLink string
var insightsCacheFolderTrimmed string

type botifyResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Owner struct {
			Login          string      `json:"login"`
			Email          string      `json:"email"`
			IsOrganisation bool        `json:"is_organisation"`
			URL            string      `json:"url"`
			DateJoined     string      `json:"date_joined"`
			Status         interface{} `json:"status"`
			FirstName      string      `json:"first_name"`
			LastName       string      `json:"last_name"`
			CompanyName    interface{} `json:"company_name"`
		} `json:"owner"`
		Features struct {
			SemanticMetadata struct {
				StructuredData struct {
					Currencies struct {
						Offer []string `json:"offer"`
					} `json:"currencies"`
				} `json:"structured_data"`
			} `json:"semantic_metadata"`
		} `json:"features"`
	} `json:"results"`
}

// KeywordsData struct used to store Keywords dimensions and metrics
type KeywordsData struct {
	Results []struct {
		Dimensions []interface{} `json:"dimensions"`
		Metrics    []*float64    `json:"metrics,omitempty"`
	} `json:"results"`
}

// AnalyticsID is used to identify which analytics tool is in use
type AnalyticsID struct {
	ID                 string `json:"id"`
	AnalyticsDateStart string `json:"date_start"`
}

// The Result struct is used to store the revenue, orders and visits
type Result struct {
	Dimensions []interface{} `json:"dimensions"`
	Metrics    []float64     `json:"metrics"`
}
type Response struct {
	Results []Result `json:"results"`
}

var company string

func main() {

	// Display the welcome banner
	startup()

	// Serve static files from the current folder
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Define a handler function for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {

		// Lock the function until it's complete to prevent race conditions
		mutex.Lock()
		defer mutex.Unlock()

		// Retrieve the form data from the request (org and username)
		err := r.ParseForm()
		if err != nil {
			fmt.Println(red+"Error. Cannot parse form:"+reset, err)
			return
		}
		organization = r.Form.Get("organization")
		project = r.Form.Get("project")

		// Generate a session ID used for grouping log entries
		sessionID, err = generateSessionID(8)
		if err != nil {
			fmt.Println(red+"Error. writeLog. Failed generating a session ID: %s"+reset, err)
			os.Exit(0)
		}

		// Acquire the business insights
		dataStatus := getBusinessInsights(sessionID)

		// Evaluate the results of getBusinessInsights before generating the broadsheet

		// All good! Generate the broadsheet
		if dataStatus == "success" {
			// Define the projectURL
			projectURL = "https://app.botify.com/" + organization + "/" + project
			writeLog(sessionID, organization, project, "-", "SEO Insights acquired")
			// Generate the broadsheet components and container
			businessInsightsDashboard(sessionID)
			writeLog(sessionID, organization, project, company, "Broadsheet generated")
			// Respond to the client with a success message or redirect to another page
			http.Redirect(w, r, insightsCacheFolder+"/go_seo_BusinessInsights.html", http.StatusFound)
		}

		// Manage errors
		// An invalid org/project name has been specified
		if dataStatus == "errorNoProjectFound" {
			writeLog(sessionID, organization, project, "-", "No project found")
			generateErrorPage("No project found. Try another organisation and project. (" + organization + "/" + project + ")")
			http.Redirect(w, r, insightsCacheFolder+"/"+"go_seo_BusinessInsights_error.html", http.StatusFound)
			return
		}

		// No analytics tool has been integrated
		if dataStatus == "errorNoAnalyticsIntegrated" {
			writeLog(sessionID, organization, project, "-", "No analytics found")
			generateErrorPage("No analytics tool has been integrated into the specified project (" + organization + "/" + project + ")")
			http.Redirect(w, r, insightsCacheFolder+"/"+"go_seo_BusinessInsights_error.html", http.StatusFound)
			return
		}

		// Engagement analytics has not been configured
		if dataStatus == "errorNoEAFound" {
			writeLog(sessionID, organization, project, "-", "No revenue data found")
			generateErrorPage("Engagement analytics with visits, revenue & transactions has not been configured for the specified project (" + organization + "/" + project + ")")
			http.Redirect(w, r, insightsCacheFolder+"/"+"go_seo_BusinessInsights_error.html", http.StatusFound)
			return
		}

		// Engagement analytics has not been configured
		if dataStatus == "errorNoKWFound" {
			writeLog(sessionID, organization, project, "-", "No keywords data found")
			generateErrorPage("RealKeywords has not been configured for the specified project (" + organization + "/" + project + ")")
			http.Redirect(w, r, insightsCacheFolder+"/"+"go_seo_BusinessInsights_error.html", http.StatusFound)
			return
		}
	})

	// Start the HTTP server
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf(red+"Error. Main. Failed to start HTTP server: %v\n"+reset, err)
	}
}

// Generate the broadsheet
func businessInsightsDashboard(sessionID string) {

	// Broadsheet header
	headerNotes()

	// Visits, orders & revenue totals
	tableVisitsOrdersRevenue()

	// Badges for CMGR KPIs
	badgeCMGR()

	// Visits per order gauge chart
	gaugeVisitsPerOrder()

	// Revenue & visits bar chart
	barRevenueVisits()

	// Visits per order line chart
	lineVisitsPerOrder()

	// Organic visit value
	barVisitValue()

	// Order volume bar chart
	barOrders()

	// Order value bar chart
	barOrderValue()

	// Revenue and visits river chart
	riverRevenueVisits()

	// Wordclouds
	// Branded
	wordcloudBrandedNonBranded(true)
	// Non branded
	wordcloudBrandedNonBranded(false)

	// Winning branded keyword narrative
	textWinningKeywords(true, sessionID)
	// Winning non branded keyword
	textWinningKeywords(false, sessionID)

	// Detailed keyword insights table - Branded
	textDetailedKeywordsInsights(true)
	// Detailed keyword insights table - Non-branded
	textDetailedKeywordsInsights(false)

	// KPI details table
	textTableDataDetail()

	// Revenue forecast line chart
	lineRevenueForecast()

	// Forecast narrative
	textForecastNarrative()

	// Footer notes
	footerNotes()

	// Generate the container to present the previously generated components
	generateDashboardContainer(company)

	fmt.Println()
	fmt.Println(lineSeparator)

	// We're done
	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println("\nSession ID: " + yellow + sessionID + reset)
	fmt.Println("\nseoBusinessInsights: Done at " + formattedTime)
	fmt.Printf("\nOrganization: %s, Project: %s\n"+reset, organization, project)

	fmt.Println()
	fmt.Println(lineSeparator)

	// Wait for the next request
	return
}

func getBusinessInsights(sessionID string) string {

	fmt.Println()
	fmt.Println(yellow + sessionID + purple + " Getting SEO insights" + reset)
	fmt.Printf("\n%s%s%s Organization: %s, Project: %s\n", yellow, sessionID, reset, organization, project)
	fmt.Println()

	// Create the seoBusinessInsights folder for the generated HTML if it does not exist
	insightsCacheFolder = envInsightsFolder + "/" + sessionID + organization
	createInsightsCacheFolder(insightsCacheFolder)

	// Get the currency used
	getCurrencyStatus := getCurrencyCompany()
	if getCurrencyStatus == "errorNoProjectFound" {
		fmt.Println(red+"Error. getBusinessInsights. No project found for", organization+"/"+project+reset)
		return getCurrencyStatus
	}

	// Identify the analytics tool in use
	analyticsID, analyticsDateStart := getAnalyticsID()
	fmt.Printf("%s%s%s Analytics identified: %s\n", yellow, sessionID, reset, analyticsID)
	fmt.Printf("%s%s%s Data available from: %s\n", yellow, sessionID, reset, analyticsDateStart)

	fmt.Println()

	// Error checking
	// Exit if no project has been found
	if analyticsID == "errorNoProjectFound" {
		fmt.Println(red+"Error. getBusinessInsights. No project found for", organization+"/"+project+reset)
		return analyticsID
	}

	// Exit if no analytics tool has been detected
	if analyticsID == "errorNoAnalyticsIntegrated" {
		fmt.Println(red+"Error. getBusinessInsights. No analytics tool integrated for", organization+"/"+project+reset)
		return analyticsID
	}

	// Reset the KPIs to their default values
	resetMetrics()

	// Get the date ranges
	dateRanges := calculateDateRanges(analyticsDateStart)

	// Populate the slice with string versions of the dates for use in the BQL
	for _, dateRange := range dateRanges.MonthlyRanges {
		startMonthDate := dateRange[0].Format("20060102")
		endMonthDate := dateRange[1].Format("20060102")
		startMonthDates = append(startMonthDates, startMonthDate)
		endMonthDates = append(endMonthDates, endMonthDate)

		// Get the month name
		startDate, _ := time.Parse("20060102", startMonthDate)
		startMthName := startDate.Format("January 2006")
		startMonthNames = append(startMonthNames, startMthName)
	}

	// Invert the date slices in order to display the oldest date first in the charts
	invertStringSlice(startMonthDates)
	invertStringSlice(endMonthDates)
	invertStringSlice(startMonthNames)

	// Get the revenue data
	getRevenueDataStatus := getRevenueData(analyticsID, startMonthDates, endMonthDates, sessionID)

	// Error checking
	// Exit if Engagement Analytics has not been configured
	if getRevenueDataStatus == "errorNoEAFound" {
		writeLog(sessionID, organization, project, analyticsID, "EngagementAnalytics not configured")
		return getRevenueDataStatus
	}

	writeLog(sessionID, organization, project, analyticsID, "Revenue data acquired")

	// Get the keywords data
	// Get last months' date range
	kwStartDate := startMonthDates[len(startMonthDates)-1]
	kwEndDate := endMonthDates[len(endMonthDates)-1]

	// Get the keywords data
	getKeywordsDataStatus := getKeywordsCloudData(kwStartDate, kwEndDate)

	// Error checking
	// Exit if Real Keywords has not been configured
	if getKeywordsDataStatus == "errorNoKWFound" {
		writeLog(sessionID, organization, project, analyticsID, "RealKeywords not configured")
		return getKeywordsDataStatus
	}

	writeLog(sessionID, organization, project, analyticsID, "Keyword data acquired")

	seoRevenue, seoVisits, seoOrders, seoOrderValue, seoVisitValue, seoVisitsPerOrder, startMonthDates, endMonthDates, startMonthNames = cleanInsights(seoRevenue, seoVisits, seoOrders, seoOrderValue, seoVisitValue, seoVisitsPerOrder, startMonthDates, endMonthDates, startMonthNames)

	// Calculate the CMGR values
	calculateCMGR(sessionID)

	// Calculate the forecast
	forecastDataCompute()

	println()
	println(green+"No. of months: "+reset, noOfMonths)

	return "success"
}

func resetMetrics() {

	// Reset slices
	startMonthDates = nil
	endMonthDates = nil
	startMonthNames = nil
	seoRevenue = nil
	seoVisits = nil
	seoOrders = nil
	seoOrderValue = nil
	totalAverageOrderValue = 0
	seoVisitValue = nil
	seoVisitsPerOrder = nil
	kwKeywords = nil
	kwCountClicks = nil
	kwMetricsCTR = nil
	kwMetricsAvgPosition = nil
	kwKeywordsNonBranded = nil
	kwCountClicksNonBranded = nil
	kwCTRNonBranded = nil
	kwAvgPositionNonBranded = nil

	// Reset integers and floats
	totalVisits = 0
	totalRevenue = 0
	totalOrders = 0
	cmgrRevenue = 0.00
	cmgrVisits = 0.00
	cmgrVisitValue = 0.00
	cmgrOrderValue = 0.00
	cmgrOrderValueValue = 0.00
}

// Get the revenue, orders and visits data
func getRevenueData(analyticsID string, startMonthDates []string, endMonthDates []string, sessionID string) string {

	var metricsOrders = 0
	var metricsRevenue = 0
	var metricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	revenueDataIssue = false
	visitsDataIssue = false
	ordersDataIssue = false

	// Get monthly insights
	for i := range startMonthDates {

		getRevenueDataStatus := ""
		metricsOrders, metricsRevenue, metricsVisits, avgOrderValue, avgVisitValue, getRevenueDataStatus = generateRevenueBQL(analyticsID, startMonthDates[i], endMonthDates[i])

		// Error checking
		if getRevenueDataStatus == "errorNoEAFound" {
			return getRevenueDataStatus
		}

		// Check revenue, visits or orders values are missing
		if metricsRevenue == 0 {
			revenueDataIssue = true
			println("revenue issue")
		}
		if metricsVisits == 0 {
			println("visits issue")
			visitsDataIssue = true
		}
		if metricsOrders == 0 {
			ordersDataIssue = true
		}
		// Append the metrics to the slices
		seoOrders = append(seoOrders, metricsOrders)
		seoRevenue = append(seoRevenue, metricsRevenue)
		seoOrderValue = append(seoOrderValue, avgOrderValue)
		seoVisits = append(seoVisits, metricsVisits)

		// Round avgVisitValue to 2 decimal places
		avgVisitValueRounded := math.Round(avgVisitValue*100) / 100
		seoVisitValue = append(seoVisitValue, avgVisitValueRounded)

		// Calculate the visits per order (for the month)
		// Check division by zero
		visitsPerOrderDisplay := 0
		if metricsOrders != 0 {
			seoVisitsPerOrder = append(seoVisitsPerOrder, metricsVisits/metricsOrders)
			visitsPerOrderDisplay = metricsVisits / metricsOrders
		} else {
			seoVisitsPerOrder = append(seoVisitsPerOrder, 0)
		}

		// Calculate the grand total for revenue visits & orders
		totalRevenue += metricsRevenue
		totalVisits += metricsVisits
		totalOrders += metricsOrders

		formatInteger := message.NewPrinter(language.English)

		// Display the KPIs
		fmt.Printf(yellow+sessionID+white+" Date Start: %s End: %s\n"+reset, startMonthDates[i], endMonthDates[i])
		formattedOrders := formatInteger.Sprintf("%d", metricsOrders)
		formattedRevenue := formatInteger.Sprintf("%d", metricsRevenue)
		formattedVisits := formatInteger.Sprintf("%d", metricsVisits)
		fmt.Println("No. Orders:", formattedOrders)
		fmt.Println("Total revenue:", formattedRevenue)
		fmt.Println("Average order value:", avgOrderValue)
		fmt.Println("No. of visits:", formattedVisits)
		fmt.Println("Average visit value:", avgVisitValue)
		fmt.Println("Average visits per order:", visitsPerOrderDisplay)
	}

	// Calculate the average visits per order
	totalVisitsPerOrder := 0
	// Sum the total visits per order over the period
	for _, value := range seoVisitsPerOrder {
		totalVisitsPerOrder += value
	}
	// Divide the total by the number of periods
	if len(seoVisitsPerOrder) > 0 {
		totalAverageVisitsPerOrder = totalVisitsPerOrder / len(seoVisitsPerOrder)
	}

	// Calculate the minimum and maximum visits per order
	minVisitsPerOrder = -1
	maxVisitsPerOrder = seoVisitsPerOrder[0]

	// Iterate through the slice to find the min and max values
	for _, value := range seoVisitsPerOrder {
		if value >= 2 {
			if minVisitsPerOrder == -1 || value < minVisitsPerOrder {
				minVisitsPerOrder = value
			}
		}
		if value > maxVisitsPerOrder {
			maxVisitsPerOrder = value
		}
	}

	// Calculate the average visit value
	totalVisitsValue := 0.00
	for _, value := range seoVisitValue {
		totalVisitsValue += value
	}
	if len(seoVisitValue) > 0 {
		totalAverageVisitValue = totalVisitsValue / float64(len(seoVisitValue))
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

	fmt.Println("\n" + yellow + sessionID + reset + " Totals" + reset)
	fmt.Println("Total visits:", totalVisits)
	fmt.Println("Total revenue:", totalRevenue)
	fmt.Println("Total orders:", totalOrders)
	fmt.Println("Total average order value:", totalAverageOrderValue)
	fmt.Println("Total average visits per order:", totalAverageVisitsPerOrder)
	fmt.Println("Total average visit value:", totalAverageVisitValue)

	return "success"
}

// Get the keywords data
func getKeywordsCloudData(startMonthDates string, endMonthDates string) string {

	var getKeywordsDataStatus = ""

	// Branded keywords
	noKeywordsFound = generateKeywordsCloudBQL(startMonthDates, endMonthDates, "true")
	if noKeywordsFound == 0 {
		getKeywordsDataStatus = "errorNoKWFound"
	}

	// Non-branded keywords
	noKeywordsFound = generateKeywordsCloudBQL(startMonthDates, endMonthDates, "false")
	if noKeywordsFound == 0 {
		getKeywordsDataStatus = "errorNoKWFound"
	}

	return getKeywordsDataStatus
}

// Execute the BQL to acquire the keywords data
func generateKeywordsCloudBQL(startDate string, endDate string, brandedFlag string) int {

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

	// Get the keyword data
	responseData := executeBQL(noKeywordsInCloud, bqlCloudKeywords)

	// Unmarshal JSON data into KeywordsData struct
	var response KeywordsData
	err := json.Unmarshal(responseData, &response)
	if err != nil {
		fmt.Printf(red+"Error. generateKeywordsCloudBQL. Cannot unmarshal the JSON: %v"+reset, err)
	}

	noKeywordsFound := len(response.Results)

	// Load the response into the slices - branded keywords
	if brandedFlag == "true" {
		for _, result := range response.Results {
			if len(result.Dimensions) >= 1 && len(result.Metrics) >= 3 {
				kwKeywords = append(kwKeywords, result.Dimensions[0].(string))
				kwCountClicks = append(kwCountClicks, int(*result.Metrics[0]))
				kwMetricsAvgPosition = append(kwMetricsAvgPosition, *result.Metrics[1])
				kwMetricsCTR = append(kwMetricsCTR, *result.Metrics[2])
			}
		}
	}

	// Load the response into the slices - non-branded keywords
	if brandedFlag == "false" {
		for _, result := range response.Results {
			if len(result.Dimensions) >= 1 && len(result.Metrics) >= 3 {
				kwKeywordsNonBranded = append(kwKeywordsNonBranded, result.Dimensions[0].(string))
				kwCountClicksNonBranded = append(kwCountClicksNonBranded, int(*result.Metrics[0]))
				kwAvgPositionNonBranded = append(kwAvgPositionNonBranded, *result.Metrics[1])
				kwCTRNonBranded = append(kwCTRNonBranded, *result.Metrics[2])
			}
		}
	}
	return noKeywordsFound
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
		fmt.Printf(red+"Error. generateRevenueBQL. Cannot unmarshal the JSON: %v"+reset, err)
	}

	var metricsOrders = 0
	var metricsRevenue = 0
	var metricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Check if any data has been returned from the API. Count the number of elements in the response.Results slice
	responseCount := len(response.Results)

	if responseCount == 0 {
		fmt.Println(red+"Error. generateRevenueBQL. Engagement analytics with visits, revenue & transactions (orders) has not been configured for the specified project ", organization+"/"+project+reset)
		fmt.Println(startDate)
		fmt.Println(endDate)

		getRevenueDataStatus := "errorNoEAFound"
		return 0, 0, 0, 0, 0.0, getRevenueDataStatus
	} else {
		metricsOrders = int(response.Results[0].Metrics[0])
		metricsRevenue = int(response.Results[0].Metrics[1])
		metricsVisits = int(response.Results[0].Metrics[2])
		// Compute the average Order value
		// Check division by zero
		if metricsOrders != 0 {
			avgOrderValue = metricsRevenue / metricsOrders
		}

		// Calculate avgVisitValue only if metricsVisits is not zero
		if metricsVisits != 0 {
			avgVisitValue = float64(metricsRevenue) / float64(metricsVisits)
		}
	}
	getRevenueDataStatus := "success"
	return metricsOrders, metricsRevenue, metricsVisits, avgOrderValue, avgVisitValue, getRevenueDataStatus
}

// Header for the broadsheet
func headerNotes() {

	currentTime := time.Now()
	currentDate := currentTime.Format("02 January 2006")
	currentTimeFormatted := currentTime.Format("15:04")

	htmlDataIssue := ""
	// If any issues have been found in the data (i.e. mlissing data) generate the HTML for inclusion in the header
	if revenueDataIssue || visitsDataIssue || ordersDataIssue {
		htmlDataIssue = generateDataIssueHTML(revenueDataIssue, visitsDataIssue, ordersDataIssue)
	}

	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
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
		.right-justify {
            text-align: right;
            display: block; 
        }
		.deepskyblue {
            color: DeepSkyBlue;
        }
        .darkgrey {
            color: #00796b;
        }
    </style>
</head>
<body>
    <div class="content">
	<span class="header-font right-justify">
        <span class="deepskyblue">Version:</span>
        <span class="darkgrey">` + fmt.Sprintf("%s", version) + `</span>
    </span> 
	<span class="header-font right-justify">
        <span class="deepskyblue">Session:</span>
        <span class="darkgrey">` + fmt.Sprintf("%s", sessionID) + `</span>
    </span>
	<span class="header-font">The following insights are based on the previous ` + fmt.Sprintf("%d", noOfMonths) + ` months.</span>
		<span class="header-font">Access the Botify project <a href="` + organization + `" target="_blank">here</a></span> (` + organization + "/" + project + `)
        <br>
        <br>
        <span class="header-font">Click the chart title to view the chart in a new window.</span>
        <br>
		<br>
			<span class="header-font">This broadsheet for <strong style="color: DeepSkyBlue;">` + organization + "/" + project + `</strong> was generated on ` + currentDate + ` at ` + currentTimeFormatted + `</span>
		<br>
		` + htmlDataIssue + `
    </div>
</body>
</html>
`
	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_HeaderNotes.html")
}

// If data issue have been detected generate the HTML to include in the header
func generateDataIssueHTML(revenueDataIssue bool, visitsDataIssue bool, ordersDataIssue bool) string {

	htmlDataIssue := "<br>"

	htmlDataIssue += "<span style=\"color: red;\">Warning: Less than 12 months valid data has been found for "

	// Check which variables are true and include them in the HTML content
	if revenueDataIssue {
		htmlDataIssue += "revenue, "
	}
	if visitsDataIssue {
		htmlDataIssue += "visits, "
	}
	if ordersDataIssue {
		htmlDataIssue += "orders, "
	}

	htmlDataIssue += "</br>"

	// Trim the trailing comma and space
	htmlDataIssue = htmlDataIssue[:len(htmlDataIssue)-2] + ". </span>"

	return htmlDataIssue
}

// CMGR Badges
func badgeCMGR() {

	cmgrRevenue32 := float32(cmgrRevenue)
	cmgrVisits32 := float32(cmgrVisits)
	cmgrVisitValue32 := float32(cmgrVisitValue)
	cmgrOrderValue32 := float32(cmgrOrderValue)
	cmgrOrderValueValue32 := float32(cmgrOrderValueValue)

	// Generate the badges
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")

	// URL to full screen badge display
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_CMGRRevenue.html"
	generateLiquidBadge("Revenue", cmgrRevenue32, clickURL, "Revenue growth over the period")

	clickURL = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_CMGRVisits.html"
	generateLiquidBadge("Visits", cmgrVisits32, clickURL, "Visits growth")

	clickURL = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_CMGRVisitValue.html"
	generateLiquidBadge("Visit Value", cmgrVisitValue32, clickURL, "Visit value (RPV)")

	clickURL = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_CMGROrders.html"
	generateLiquidBadge("Orders", cmgrOrderValue32, clickURL, "Order volume")

	clickURL = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_CMGROrderValue.html"
	generateLiquidBadge("Order Value", cmgrOrderValueValue32, clickURL, "Order value (AOV)")
}

// Total Visits, Orders & Revenue
func tableVisitsOrdersRevenue() {

	formatInteger := message.NewPrinter(language.English)

	totalVisitsFormatted := formatInteger.Sprintf("%d", totalVisits)
	totalOrdersFormatted := formatInteger.Sprintf("%d", totalOrders)
	totalRevenueFormatted := formatInteger.Sprintf("%d", totalRevenue)

	totalAverageOrderValueFormatted := formatInteger.Sprintf("%d", totalAverageOrderValue)
	totalAverageVisitsPerOrderFormatted := formatInteger.Sprintf("%d", totalAverageVisitsPerOrder)
	totalAverageVisitValueFormatted := fmt.Sprintf("%.2f", totalAverageVisitValue)

	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            color: #333;
            display: flex;
            justify-content: center;
            align-items: center;
            margin: 0;
            height: 100vh;
        }
        .container {
            display: flex;
            justify-content: center;
            align-items: center;
            width: 100%;
            height: 100%;
        }
        .wrapper {
            display: flex;
            justify-content: space-between;
            width: 80%;
            max-width: 1200px;
            padding: 20px;
            border-radius: 8px;
        }
        .column {
            flex: 1;
            text-align: center;
            margin: 0 20px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            font-size: 35px;
            padding: 10px;
        }
        th {
            color: #555;
            font-weight: 600;
            text-transform: uppercase;
        }
        td {
            color: #00796b;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header" style="font-size: 30px; font-weight: bold; color: Grey; margin-bottom: 20px; text-align: center;">Key Performance Metrics</div>
        <div class="wrapper">
            <div class="column">
                <table>
            	<div class="row">
                    <tr>
					<th style="color: deepskyblue;">VISITS</th>
                    </tr>
                    <tr>
                        <td>` + fmt.Sprintf("%s", totalVisitsFormatted) + `</td>
                    </tr>
                    <tr>
					<th style="color: deepskyblue;">VISIT VALUE (RPV)</th>
                    </tr>
                    <tr>
						<td>` + fmt.Sprintf("%s%s", currencySymbol, totalAverageVisitValueFormatted) + `</td>
                    </tr>
                </table>
            </div>
            <div class="column">
                <table>
                    <tr>
					<th style="color: deepskyblue;">ORDERS</th>
                    </tr>
                    <tr>
                        <td>` + fmt.Sprintf("%s", totalOrdersFormatted) + `</td>
                    </tr>
                    <tr>
					<th style="color: deepskyblue;">ORDER VALUE (AOV)</th>
                    </tr>
                    <tr>
						<td>` + fmt.Sprintf("%s%s", currencySymbol, totalAverageOrderValueFormatted) + `</td>
                    </tr>
                </table>
            </div>
            <div class="column">
                <table>
                    <tr>
					<th style="color: deepskyblue;">REVENUE</th>
                    </tr>
                    <tr>
						<td>` + fmt.Sprintf("%s%s", currencySymbol, totalRevenueFormatted) + `</td>
                    </tr>
                    <tr>
					<th style="color: deepskyblue;">VISITS PER ORDER</th>
                    </tr>
                    <tr>
                        <td>` + fmt.Sprintf("%s", totalAverageVisitsPerOrderFormatted) + `</td>
                    </tr>
                </table>
            </div>
        </div>
    </div>
</body>
</html>
`

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_TotalsVisitsOrdersRevenue.html")
}

// Bar chart. Revenue and Visits

func barRevenueVisits() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_RevenueVisitsBar.html"

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Revenue & visits",
			Subtitle: "How many organic visits does the site attract and what is the generated revenue from those visits?",
			Link:     clickURL,
		}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Revenue & visits",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}),
	)

	barDataRevenue := generateBarItems(seoRevenue)
	barDataVisits := generateBarItems(seoVisits)

	var seriesWithCurrency = "Revenue (" + currencySymbol + ")"

	bar.SetXAxis(startMonthNames).
		AddSeries(seriesWithCurrency, barDataRevenue).
		AddSeries("Visits", barDataVisits).
		SetSeriesOptions(
			charts.WithMarkPointNameTypeItemOpts(
				opts.MarkPointNameTypeItem{Name: "Highest", Type: "max", ItemStyle: &opts.ItemStyle{Color: "rgb(144, 238, 144)"}},
				opts.MarkPointNameTypeItem{Name: "Lowest", Type: "min", ItemStyle: &opts.ItemStyle{Color: "rgb(255, 55, 55)"}},
			),
			charts.WithMarkPointStyleOpts(
				opts.MarkPointStyle{
					SymbolSize: 90,
				}),
		)

	var f *os.File
	var err error

	// Assign 'f' here
	f, err = os.Create(insightsCacheFolder + "/go_seo_RevenueVisitsBar.html")
	if err != nil {
		fmt.Printf(red+"Error. barRevenueVisits. Cannot create go_seo_RevenueVisitsBar.html: %v\n"+reset, err)
		return
	}

	// Render the chart to the file
	_ = bar.Render(f)
}

// Visits per order line chart
func lineVisitsPerOrder() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_VisitsPerOrderLine.html"

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Visits per order",
			Subtitle: "High number of visits per order indicates poor quality traffic or conversion inefficiency.",
			Link:     clickURL,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Average visits per order",
		}),

		charts.WithColorsOpts(opts.Colors{kpiColourVisitsPerOrder}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	lineVisitsPerOrderValue := generateLineItems(seoVisitsPerOrder)

	line.SetXAxis(startMonthNames).AddSeries("Visits per order", lineVisitsPerOrderValue).SetSeriesOptions(
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Color: "lightSkyBlue",
		}),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithMarkPointNameTypeItemOpts(
			opts.MarkPointNameTypeItem{Name: "Highest No. of visits", Type: "max", ItemStyle: &opts.ItemStyle{Color: "rgb(255, 55, 55)"}},
			opts.MarkPointNameTypeItem{Name: "Lowest  No. of visits", Type: "min", ItemStyle: &opts.ItemStyle{Color: "rgb(144, 238, 144)"}},
		),
		charts.WithMarkPointStyleOpts(
			opts.MarkPointStyle{
				// the size of the markLine symbol
				SymbolSize: 90,
			}),
	)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_VisitsPerOrderLine.html")

	_ = line.Render(f)
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
func barVisitValue() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_VisitValueBar.html"

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Revenue per visit (RPV)",
		Subtitle: "A high organic visit value is a strong indicator of the effectiveness and profitability of the site's organic traffic.",
		Link:     clickURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Organic visit value",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourOrganicVisitValue}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	barDataVisitValue := generateBarItemsFloat(seoVisitValue)

	bar.SetXAxis(startMonthNames).
		AddSeries("Organic visit value", barDataVisitValue).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Lowest visit value", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Highest visit value", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average visit value", Type: "average"},
		),
			charts.WithMarkLineStyleOpts(
				opts.MarkLineStyle{
					Label:     &opts.Label{FontSize: 15},
					LineStyle: &opts.LineStyle{Color: "rgb(255, 128, 0)", Width: 3, Opacity: .7, Type: "dotted"},
				},
			),
		)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_VisitValueBar.html")

	_ = bar.Render(f)
}

// Order volume bar chart
func barOrders() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_OrdersBar.html"

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Order volume",
		Subtitle: "Order volume placed during an organic visit.",
		Link:     clickURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Order volume",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourNoOfOrders}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	barDataOrders := generateBarItems(seoOrders)

	bar.SetXAxis(startMonthNames).
		AddSeries("Orders", barDataOrders).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Lowest No. orders", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Highest No. orders", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average No. orders", Type: "average"},
		),
			charts.WithMarkLineStyleOpts(
				opts.MarkLineStyle{
					Label:     &opts.Label{FontSize: 15},
					LineStyle: &opts.LineStyle{Color: "rgb(255, 128, 0)", Width: 3, Opacity: .7, Type: "dotted"},
				},
			),
		)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_OrdersBar.html")

	_ = bar.Render(f)
}

// Order value bar chart
func barOrderValue() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_OrderValueBar.html"

	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Order value (AOV)",
		Subtitle: "The average value of an order placed during a visit from an organic source. A higher value reflects effective SEO strategies driving quality traffic.",
		Link:     clickURL,
	}),
		charts.WithLegendOpts(opts.Legend{Right: "80px"}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 1,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Average order value",
		}),
		charts.WithColorsOpts(opts.Colors{kpiColourOrderValue}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	barDataOrderValue := generateBarItems(seoOrderValue)

	bar.SetXAxis(startMonthNames).
		AddSeries("Order value", barDataOrderValue).
		SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(
			opts.MarkLineNameTypeItem{Name: "Lowest order value", Type: "min"},
			opts.MarkLineNameTypeItem{Name: "Highest order value", Type: "max"},
			opts.MarkLineNameTypeItem{Name: "Average order value", Type: "average"},
		),
			charts.WithMarkLineStyleOpts(
				opts.MarkLineStyle{
					Label:     &opts.Label{FontSize: 15},
					LineStyle: &opts.LineStyle{Color: "rgb(255, 128, 0)", Width: 3, Opacity: .7, Type: "dotted"},
				},
			),
		)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_OrderValueBar.html")

	_ = bar.Render(f)
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

// CMGR badges
func generateLiquidBadge(badgeKPI string, badgeKPIValue float32, clickURL string, title string) {

	badgeKPIValueCalc := badgeKPIValue * 100

	subTitle := fmt.Sprintf("Compound growth CMGR. Rounded from %.2f%%", badgeKPIValueCalc)

	liquid := charts.NewLiquid()

	liquid.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: badgeKPI + " growth (CMGR)",
			Width:     badgeDefaultWidth,
			Height:    badgeDefaultHeight,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subTitle,
			Link:     clickURL,
		}),
	)

	liquid.AddSeries("", genLiquidItems([]float32{badgeKPIValue})).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show: opts.Bool(true),
			}),

			charts.WithLiquidChartOpts(opts.LiquidChart{
				IsWaveAnimation: opts.Bool(true),
				Shape:           "roundRect",
			}),
		)

	// Removing spaces from badgeKPI to ensure a clean URL for the HTML is generated.
	badgeKPI = strings.ReplaceAll(badgeKPI, " ", "")
	badgeFileName := fmt.Sprintf("/go_seo_CMGR%s.html", badgeKPI)
	f, _ := os.Create(insightsCacheFolder + badgeFileName)

	_ = liquid.Render(f)
}

// Get data for the liquid badge
func genLiquidItems(data []float32) []opts.LiquidData {

	items := make([]opts.LiquidData, 0)
	for i := 0; i < len(data); i++ {
		items = append(items, opts.LiquidData{Value: data[i]})
	}
	return items
}

// Branded and non-branded wordclouds
func wordcloudBrandedNonBranded(brandedMode bool) {

	var clickURL string
	var subtitle string
	var pageTitle string

	if brandedMode {
		wordcloudTitle = fmt.Sprintf("Top %d branded keywords generating clicks", noKeywordsInCloud)
		// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
		insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
		clickURL = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_WordcloudBranded.html"
		pageTitle = "Branded wordcloud"
	}
	if !brandedMode {
		wordcloudTitle = fmt.Sprintf("Top %d non branded keywords generating clicks", noKeywordsInCloud)
		// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
		insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
		clickURL = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_WordcloudNonBranded.html"
		pageTitle = "Non Branded wordcloud"
	}

	wordcloud := charts.NewWordCloud()

	wordcloud.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:     wordcloudDefaultWidth,
			Height:    wordcloudDefaultHeight,
			PageTitle: pageTitle,
		}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithTitleOpts(opts.Title{
			Title:    wordcloudTitle,
			Link:     clickURL,
			Subtitle: subtitle,
		}))

	// Generate the branded wordcloud
	if brandedMode {
		wordcloud.AddSeries("Clicks", generateWCData(kwKeywords, kwCountClicks)).
			SetSeriesOptions(
				charts.WithWorldCloudChartOpts(
					opts.WordCloudChart{
						SizeRange: []float32{10, 90},
						Shape:     "roundRect",
					}),
			)
	}

	// Generate the non-branded wordcloud
	if !brandedMode {
		wordcloud.AddSeries("Clicks", generateWCDataNonBranded(kwKeywordsNonBranded, kwCountClicksNonBranded)).
			SetSeriesOptions(
				charts.WithWorldCloudChartOpts(
					opts.WordCloudChart{
						SizeRange: []float32{10, 90},
						Shape:     "basic",
					}),
			)
	}

	if brandedMode {
		f, _ := os.Create(insightsCacheFolder + "/go_seo_WordcloudBranded.html")
		_ = wordcloud.Render(f)
	}

	if !brandedMode {
		f, _ := os.Create(insightsCacheFolder + "/go_seo_WordcloudNonBranded.html")
		_ = wordcloud.Render(f)
	}

}

// Generate the data for the wordcloud - Branded
func generateWCData(kwKeywords []string, kwCountClicks []int) (items []opts.WordCloudData) {

	items = make([]opts.WordCloudData, 0)
	// Iterate over kwKeywords and kwCountClicks slices starting from index 1
	// We start at index 1 because the top keyword is generally significantly higher performing than the following keywords and will distort the wordcloud if included
	for i := 1; i < len(kwKeywords); i++ {
		// Check if index is valid for kwCountClicks slice
		if i < len(kwCountClicks) {
			// Append WordCloudData struct with keyword and corresponding count
			items = append(items, opts.WordCloudData{Name: kwKeywords[i], Value: kwCountClicks[i]})
		}
	}
	return items
}

// Generate the data for the wordcloud - Non-branded
func generateWCDataNonBranded(kwKeywordsNonBranded []string, kwCountClicksNonBranded []int) (items []opts.WordCloudData) {

	items = make([]opts.WordCloudData, 0)
	// Iterate over kwKeywords and kwCountClicks slices starting from index 1
	// We start at index 1 because the top keyword is generally significantly higher performing than the following keywords and will distort the wordcloud if included
	for i := 1; i < len(kwKeywordsNonBranded); i++ {
		// Check if index is valid for kwCountClicks slice
		if i < len(kwCountClicksNonBranded) {
			// Append WordCloudData struct with keyword and corresponding count
			items = append(items, opts.WordCloudData{Name: kwKeywordsNonBranded[i], Value: kwCountClicksNonBranded[i]})
		}
	}
	return items
}

// River chart
func riverRevenueVisits() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_VisitsRevenueRiver.html"

	river := charts.NewThemeRiver()

	river.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Revenue & visits",
			Subtitle: "Insights into the fluctuations in organic visitors to a site and the corresponding revenue generation.",
			Link:     clickURL}),
		charts.WithSingleAxisOpts(opts.SingleAxis{
			Type:   "time",
			Bottom: "10%",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "time",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
		}),
		// Increase the canvas size
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Revenue & visits",
		}),
		//charts.WithDataZoomOpts(opts.DataZoom{
		//	Type:  "slider",
		//	Start: 1,
		//	End:   100,
		//}),
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Populate ThemeRiverData slice
	var themeRiverData []opts.ThemeRiverData

	// Add the Revenue data
	// The date is formatted from YYYYMMDD to YYYY/MM/DD
	for i, date := range startMonthDates {
		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			fmt.Printf(red+"Error. riverRevenueVisits. Error parsing date: %v\n"+reset, err)
			break
		}
		formattedDate := parsedDate.Format("2006/01/02")
		themeRiverData = append(themeRiverData, opts.ThemeRiverData{
			Date:  formattedDate,
			Value: float64(seoRevenue[i]),
			Name:  "Revenue",
		})
	}

	// Add the Visits data
	// The date is formatted from YYYYMMDD to YYYY/MM/DD
	for i, date := range startMonthDates {
		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			fmt.Printf(red+"Error. riverRevenueVisits. Error parsing date: %v\n"+reset, err)
			break
		}
		formattedDate := parsedDate.Format("2006/01/02")
		themeRiverData = append(themeRiverData, opts.ThemeRiverData{
			Date:  formattedDate,
			Value: float64(seoVisits[i]),
			Name:  "Visits",
		})
	}

	river.AddSeries("themeRiver", themeRiverData)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_VisitsRevenueRiver.html")

	_ = river.Render(f)
}

func gaugeVisitsPerOrder() {

	gauge := charts.NewGauge()

	insightsCacheFolderTrimmed = strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_Gauge.html"

	setMinMax := charts.WithSeriesOpts(func(s *charts.SingleSeries) {
		s.Min = minVisitsPerOrder
		s.Max = maxVisitsPerOrder
	})

	gauge.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Lowest, highest and average organic visits per order",
			Link:  clickURL}),

		charts.WithInitializationOpts(opts.Initialization{
			Width:     gaugeDefaultWidth,
			Height:    gaugeDefaultHeight,
			PageTitle: "Visits per order",
		}),
	)
	gauge.AddSeries("",
		[]opts.GaugeData{{Value: totalAverageVisitsPerOrder}}, setMinMax)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_Gauge.html")

	_ = gauge.Render(f)
}

// Table containing the detailed KPI insights
func textTableDataDetail() {

	var detailedKPITableData [][]string

	formatInteger := message.NewPrinter(language.English)

	for i := 0; i < noOfMonths; i++ {
		formattedDate := formatDate(startMonthDates[i])
		orders := formatInteger.Sprintf("%d", seoOrders[i])
		revenue := formatInteger.Sprintf("%d", seoRevenue[i])
		orderValue := formatInteger.Sprintf("%d", seoOrderValue[i])
		visits := formatInteger.Sprintf("%d", seoVisits[i])
		visitValue := formatInteger.Sprintf("%.2f", seoVisitValue[i])
		visitsPerOrderValue := formatInteger.Sprintf("%d", seoVisitsPerOrder[i])

		row := []string{
			formattedDate,
			orders,
			currencySymbol + revenue,
			currencySymbol + orderValue,
			visits,
			currencySymbol + visitValue,
			visitsPerOrderValue,
		}
		detailedKPITableData = append(detailedKPITableData, row)
	}

	// Generate the table
	htmlContent := generateHTMLDetailedKPIInsightsTable(detailedKPITableData)

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_DataInsightDetailKPIs.html")
}

// Winning keywords, branded & non-branded
func textWinningKeywords(brandedMode bool, sessionID string) {

	var htmlFileName = ""

	// Define the display values based on branded or non-branded mode
	var htmlKeyword = ""
	var htmlClicks = ""
	var htmlClickGap = 0
	var htmlSecondPlaceKW = ""
	var htmlCTR float64
	var htmlAvgPosition float64

	formatInteger := message.NewPrinter(language.English)

	if brandedMode {
		htmlKeyword = kwKeywords[0]
		htmlClicks = formatInteger.Sprintf("%d", kwCountClicks[0])
		htmlClickGap = int(((float64(kwCountClicks[0]) - float64(kwCountClicks[1])) / float64(kwCountClicks[1])) * 100)
		htmlSecondPlaceKW = kwKeywords[1]
		htmlCTR = kwMetricsCTR[0]
		htmlAvgPosition = kwMetricsAvgPosition[0]
		fmt.Println("\n" + yellow + sessionID + reset + " Branded keywords\n")
		for i := 0; i < len(kwKeywords); i++ {
			fmt.Printf(green+"Keyword:"+reset+bold+" %s"+reset+","+green+" Clicks:"+reset+" %d,"+green+" CTR:"+reset+" %.2f,"+green+" Avg. Position:"+reset+" %.2f\n",
				kwKeywords[i], kwCountClicks[i], kwMetricsCTR[i], kwMetricsAvgPosition[i])
		}
	}

	if !brandedMode {
		htmlKeyword = kwKeywordsNonBranded[0]
		htmlClicks = formatInteger.Sprintf("%d", kwCountClicksNonBranded[0])
		htmlClickGap = int(((float64(kwCountClicksNonBranded[0]) - float64(kwCountClicksNonBranded[1])) / float64(kwCountClicksNonBranded[1])) * 100)
		htmlSecondPlaceKW = kwKeywordsNonBranded[1]
		htmlCTR = kwCTRNonBranded[0]
		htmlAvgPosition = kwAvgPositionNonBranded[0]
		fmt.Println("\n" + yellow + sessionID + reset + " Non branded keywords\n")
		for i := 0; i < len(kwKeywords); i++ {
			fmt.Printf(green+"Keyword:"+reset+bold+" %s"+reset+","+green+" Clicks:"+reset+" %d,"+green+" CTR:"+reset+" %.2f,"+green+" Avg. Position:"+reset+" %.2f\n",
				kwKeywordsNonBranded[i], kwCountClicksNonBranded[i], kwCTRNonBranded[i], kwAvgPositionNonBranded[i])
		}
	}

	// Get the last month name
	htmlLastMonthName := ""
	if len(startMonthNames) > 0 {
		htmlLastMonthName = startMonthNames[len(startMonthNames)-1]
	}

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
        }
        .content {
  			border: 3px solid lightSkyBlue;  
            border-radius: 25px; 
            padding: 15px;
            text-align: center;  
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
    <span class="keyword-font">
        The winning keyword during <b>%s</b> was <span class="blueText">%s</span>. 
        This keyword generated <b>%s</b> clicks which is <b>%d%%</b> more clicks than the second placed keyword  <b>%s</b>. The click-through rate for the winning keyword was <b>%.2f%%</b> 
        from an average position of <b>%.2f</b>.
    </span>
	</div>
</body>
</html>
`, htmlLastMonthName, htmlKeyword, htmlClicks, htmlClickGap, htmlSecondPlaceKW, htmlCTR, htmlAvgPosition)

	if brandedMode {
		htmlFileName = "/go_seo_WinningKeywordBranded.html"
	} else {
		htmlFileName = "/go_seo_WinningKeywordNonBranded.html"
	}

	// Save the HTML to a file
	saveHTML(htmlContent, htmlFileName)
}

// Generate the HTML for the table
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
				<th class="title" style="color: DeepSkyBlue;">Order volume</th>
                <th class="title" style="color: DeepSkyBlue;">Revenue</th>
                <th class="title" style="color: DeepSkyBlue;">Order Value</th>
                <th class="title" style="color: DeepSkyBlue;">No. of Visits</th>
                <th class="title" style="color: DeepSkyBlue;">Revenue per visit</th>
                <th class="title" style="color: DeepSkyBlue;">Visits per Order</th>
            </tr>
        </thead>
        <tbody>`

	// Title
	htmlContent += fmt.Sprintf("<h2>\n\nSEO Business insights for the previous %d months</h2>", noOfMonths)

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

// Generate the HTML for the keywords insights
func textDetailedKeywordsInsights(brandedMode bool) {

	formatInteger := message.NewPrinter(language.English)

	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
        }
        .table-container {
            overflow: auto;
            height: 500px; 
            border: 2px solid transparent; 
            border-radius: 16px; 
            margin: 20px; 
        }
        table {
            width: 100%;
            border-collapse: collapse; 
            color: DimGray;
            font-size: 17px;
            text-align: left;
            border-radius: 16px; 
        }
        th, td {
            padding: 20px;
        }
        th {
            background-color: White;
            color: deepskyblue;
            position: sticky; /* Sticky position */
            top: 0; /* Stick to the top */
            z-index: 10; /* Ensure header stays on top */
        }
        td {
            border-bottom: 1px solid transparent; 
        }
        tr:nth-child(odd) {
            background-color: #f9f9f9;
        }
        tr:hover {
            background-color: DeepSkyBlue;
            color: white; 
        }
        /* Apply rounded corners to header and footer rows */
        thead th:first-child {
            border-top-left-radius: 16px;
        }
        thead th:last-child {
            border-top-right-radius: 16px;
        }
        tbody tr:last-child td:first-child {
            border-bottom-left-radius: 16px;
        }
        tbody tr:last-child td:last-child {
            border-bottom-right-radius: 16px;
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
		for i := 0; i < noTopKeywords; i++ {
			kwCountClicksFormatted := formatInteger.Sprintf("%d", kwCountClicks[i])
			htmlContent += fmt.Sprintf("<tr>\n"+
				"    <td>%s</td>\n"+
				"    <td>%s</td>\n"+
				"    <td>%.2f%%</td>\n"+
				"    <td>%.2f</td>\n"+
				"</tr>\n",
				kwKeywords[i],
				kwCountClicksFormatted,
				kwMetricsCTR[i],
				kwMetricsAvgPosition[i])
		}
	}

	// Non branded keywords details
	if !brandedMode {
		for i := 0; i < noTopKeywords; i++ {
			kwCountClicksFormattedNonBranded := formatInteger.Sprintf("%d", kwCountClicksNonBranded[i])
			htmlContent += fmt.Sprintf("<tr>\n"+
				"    <td>%s</td>\n"+
				"    <td>%s</td>\n"+
				"    <td>%.2f%%</td>\n"+
				"    <td>%.2f</td>\n"+
				"</tr>\n",
				kwKeywordsNonBranded[i],
				kwCountClicksFormattedNonBranded,
				kwCTRNonBranded[i],
				kwAvgPositionNonBranded[i])
		}
	}

	htmlContent += fmt.Sprintf("</table>\n")
	htmlContent += fmt.Sprintf("</body>\n")
	htmlContent += fmt.Sprintf("</html>\n")

	// Save the HTML to a file
	// Branded keywords details
	if brandedMode {
		saveHTML(htmlContent, "/go_seo_BrandedInsights.html")
	}
	// Branded keywords details
	if !brandedMode {
		saveHTML(htmlContent, "/go_seo_NonBrandedInsights.html")
	}
}

// generate the slice containing the projected revenue data
func forecastDataCompute() {

	// First create a slice containing the visit ranges
	numElements := forecastMaxVisits/forecastIncrement + 1
	forecastVisitIncrements = make([]int, numElements)
	forecastVisitIncrementsString = make([]string, numElements)

	// Populate the slice with the visit ranges
	formatInteger := message.NewPrinter(language.English)

	for i := 0; i < numElements; i++ {
		forecastVisitIncrements[i] = i * forecastIncrement
		// Create a formatted String version for use in the chart XAxis
		forecastVisitIncrementsString[i] = formatInteger.Sprintf("%d", forecastVisitIncrements[i])
	}

	// Create a slice to hold the forecast revenue values
	forecastRevenue = make([]int, numElements)
	for i := 0; i < numElements; i++ {
		if totalAverageVisitsPerOrder != 0 {
			forecastRevenue[i] = forecastVisitIncrements[i] / totalAverageVisitsPerOrder * totalAverageOrderValue
		} else {
			forecastRevenue[i] = 0
		}
	}
}

// Revenue forecast line chart
func lineRevenueForecast() {

	// Generate the URL to the chart. Used to display the chart full screen when the header is clicked
	insightsCacheFolderTrimmed := strings.TrimPrefix(insightsCacheFolder, ".")
	clickURL := protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_Forecast.html"

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Revenue forecast",
			Subtitle: "Use the slider to forecast potential revenue growth with increased visits",
			Link:     clickURL,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 5,
			End:   20,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:     chartDefaultWidth,
			Height:    chartDefaultHeight,
			PageTitle: "Organic revenue forecast",
		}),

		charts.WithColorsOpts(opts.Colors{kpiColourRevenueForecast}),
		// disable show the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Pass visitsPerOrder directly to generaLineItems
	lineVisitsPerOrderValue := generateLineItemsRevenueForecast(forecastRevenue)

	line.SetXAxis(forecastVisitIncrementsString).AddSeries("Revenue forecast", lineVisitsPerOrderValue).SetSeriesOptions(
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Color: "lightSkyBlue",
		}),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithMarkPointNameTypeItemOpts(
			opts.MarkPointNameTypeItem{Name: "High End", Type: "max", ItemStyle: &opts.ItemStyle{Color: "rgb(144, 238, 144)"}},
			opts.MarkPointNameTypeItem{Name: "Mid Range", Type: "average", ItemStyle: &opts.ItemStyle{Color: "rgb(255, 165, 0)"}},
			opts.MarkPointNameTypeItem{Name: "Low End", Type: "min", ItemStyle: &opts.ItemStyle{Color: "rgb(255, 55, 55)"}},
		),
		charts.WithMarkPointStyleOpts(
			opts.MarkPointStyle{
				SymbolSize: 90,
			}),
	)

	f, _ := os.Create(insightsCacheFolder + "/go_seo_Forecast.html")

	_ = line.Render(f)
}

// Populate the chart with the revenue forecast data
func generateLineItemsRevenueForecast(forecastRevenue []int) []opts.LineData {

	items := make([]opts.LineData, len(forecastRevenue))
	for i, val := range forecastRevenue {
		items[i] = opts.LineData{Value: val}
	}
	return items
}

func textForecastNarrative() {

	var htmlFileName = ""

	var noOfOrderVisits = 0
	if totalAverageVisitsPerOrder != 0 {
		noOfOrderVisits = forecastIncrement / totalAverageVisitsPerOrder
	} else {
		noOfOrderVisits = 0
	}

	var projectedRevenue = noOfOrderVisits * totalAverageOrderValue

	// Format the integers with commas
	formatInteger := message.NewPrinter(language.English)
	formattedForecastIncrement := formatInteger.Sprintf("%d", forecastIncrement)
	formattedProjectedRevenue := formatInteger.Sprintf("%d", projectedRevenue)

	// HTML content for the revenue forecast narrative
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
			padding-left: 30px;
			padding-right: 30px;
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
			<b>Example scenario:</b>
			On average an order is placed every
			<span class="blueText">%d</span> visits from an organic source. For each additional 
			<span class="blueText">%s</span> organic visits, the current forecast is 
			<span class="blueText">%d</span> orders will be placed. With an average 
			order value of <span class="blueText">%s%d</span> the projected 
			incremental revenue from <span class="blueText">%s</span> additional visits will be 
			<span class="blueText">%s%s</span>
		</p>
	</div>
</body>
</html>
`, totalAverageVisitsPerOrder, formattedForecastIncrement, noOfOrderVisits, currencySymbol, totalAverageOrderValue,
		formattedForecastIncrement, currencySymbol, formattedProjectedRevenue,
	)

	// Define the HTML filename
	htmlFileName = "/go_seo_ForecastNarrative.html"

	// Save the HTML to a file
	saveHTML(htmlContent, htmlFileName)
}

// Footer
func footerNotes() {

	dashboardPermaLink = protocol + "://" + fullHost + insightsCacheFolderTrimmed + "/go_seo_BusinessInsights.html"

	// Text content for the footer
	var footerNotesStrings = []string{
		"The current month is not included in the analysis, only full months are reported on.",
		"Compound Growth (CMGR) refers to the Compound Monthly Growth Rate of the KPI. CMGR is a financial term used to measure the growth rate of a metric over a monthly basis taking into account the compounding effect. CMGR provides a clear and standardised method to measure growth over time.",
		"The permalink for this broadsheet is <a href=\"" + dashboardPermaLink + "\" target=\"_blank\">" + dashboardPermaLink + "</a>",
	}

	// Generate HTML content
	htmlContent := `<html>
<head>
</head>
	<style>
		body {
		font-family: Arial, sans-serif;
		font-size: 18px;
		color: LightSlateGray; 
		}
	</style>
<body>
    <br>
    <ul>
`
	for _, note := range footerNotesStrings {
		htmlContent += fmt.Sprintf("<li>%s</li>\n", note)
	}

	htmlContent += "</ul>\n</body>\n</html>"

	htmlContent += "<br>"

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_FooterNotes.html")
}

// formatDate converts date from YYYYMMDD to Month-Year format
func formatDate(dateStr string) string {

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		fmt.Println(red+"Error. formatDate. Cannot parse date:"+reset, err)
		return dateStr
	}
	return date.Format("January 2006")
}

// Function used to generate and save the HTML content to a file
func saveHTML(genHTML string, genFilename string) {

	file, err := os.Create(insightsCacheFolder + genFilename)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Cannot create:"+reset, insightsCacheFolder, genFilename, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. saveHTML. Failed to close file: %v\n"+reset, err)
			return
		}
	}()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fullFolder := insightsCacheFolder + genFilename
		fmt.Printf(red+"Error. saveHTML. Cannot write HTML file: %s"+reset, fullFolder)
		fmt.Printf(red+"Error. saveHTML. Error %s:"+reset, err)
		return
	}
}

// Define the HTML for the container. Used to consolidate the generated charts into a single page.
// Container start
func generateDashboardContainer(company string) {

	// Using these variables to replace width values in the HTML below because string interpolation confuses the percent signs as variables
	width90 := "90%"
	width100 := "100%"
	width0 := "0%"
	percent := "%"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>seoBusinessInsights</title>
    <style>
        body {
   			margin: 0;
            font-family: 'Helvetica Neue', Arial, sans-serif;
            background-color: #f4f7f6;
            color: #333;
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
            width: %s;
        }
        .container {
            display: flex;
            flex-wrap: wrap;
            align-items: center;
            gap: 20px;
            margin: 5px auto;

            width: %s;
        }
        .row {
            flex-wrap: nowrap;
        }
        iframe {
            flex: 1 1 auto;
            min-width: 200px;
            width: %s; 
            border: 2px solid lightSkyBlue;
            border-radius: 10px;
			margin: 10px 0;
        }
        .no-border iframe {
            border: none;
        }
        .tall-iframe {
            height: 750px;
        }
        .medium-iframe {
            height: 600px;
        }
        .short-iframe {
            height: 450px;
        }
        .back-button {
            padding: 12px 24px;
            font-size: 18px;
            color: white;
            background-color: Green;
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
            background-color: DeepSkyBlue;
            box-shadow: 0 6px 8px rgba(0, 0, 0, 0.15);
        }
        .section-padding-top {
            padding-top: 35px;
        }
        .section-padding-bottom {
            padding-bottom: 35px;
        }
        /* Scroll Indicator Styles */
        #progressContainer {
            position: fixed;
            width: %s;
            height: 8px;
            top: 0;
            left: 0;
            background: #f3f3f3;
            z-index: 9999;
        }
        #progressBar {
            height: %s;
            background: DeepSkyBlue;
            width: %s;
        }
        /* Navigation Links */
		nav {
            position: sticky;
            top: 0;
            background-color: #f2f2f2;
            border-bottom: 2px solid #ddd;
            z-index: 1000; 
        }
        nav ul {
            list-style-type: none;
            display: flex;
            justify-content: center;
            margin: 0;
            padding: 15px 0;
            background-color: #f4f7f6;
            border-bottom: 2px solid #ddd;
        }
        nav li {
            padding: 0 20px;
            text-align: center;
        }
		nav a {
            text-decoration: none;
            color: #00796b;
            font-weight: bold;
            font-size: 16px;
            transition: color 0.3s;
        }
        nav a:hover {
            color: #00aaff;
        }
    </style>
</head>
<body>

<!-- Scroll Indicator -->
<div id="progressContainer">
    <div id="progressBar"></div>
</div>

<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">Business Insights Broadsheet for %s</span>
</header>

<!-- Navigation Links -->
<nav>
    <ul>	
        <li><a href="#revenue_visits">Revenue & visits</a></li>
        <li><a href="#visits_per_order">Visits per order</a></li>
        <li><a href="#orders">Order volume</a></li>
        <li><a href="#order_value">Order value (AOV)</a></li>
        <li><a href="#visit_value">Revenue per visit (RPV)</a></li>
        <li><a href="#detailed_insights">Detailed insights</a></li>
        <li><a href="#revenue_forecast">Revenue forecast</a></li>
        <li><a href="#wordcloud_branded">Top branded keywords</a></li>
        <li><a href="#wordcloud_non_branded">Top non branded keywords</a></li>
    </ul>
</nav>

<button class="back-button" onclick="goHome()">New broadsheet</button>

<script>
    function goHome() {
        window.open('%s://%s/', '_blank');
    }
    // Scroll Indicator Script
    window.onscroll = function() { updateProgressBar(); };

    function updateProgressBar() {
        const scrollTop = document.documentElement.scrollTop || document.body.scrollTop;
        const scrollHeight = document.documentElement.scrollHeight || document.body.scrollHeight;
        const clientHeight = document.documentElement.clientHeight;
        const scrolled = (scrollTop / (scrollHeight - clientHeight)) * 100;
        document.getElementById('progressBar').style.width = scrolled + "%s";
    }
</script>

<div class="content">
	<section class="container row no-border">
		<iframe src="go_seo_HeaderNotes.html" title="Header" style="height: 180px;"></iframe>
	</section>
	
	<section class="container row no-border">
		<iframe src="go_seo_TotalsVisitsOrdersRevenue.html" title="Your SEO KPI totals" style="height: 340px;"></iframe>
	</section>
	
	<section id="revenue_visits" class="container row">
		<iframe src="go_seo_RevenueVisitsBar.html" title="Revenue & visits" class="tall-iframe"></iframe>
	</section>

	<section class="container row">
		<iframe src="go_seo_VisitsRevenueRiver.html" title="Revenue & visits" class="medium-iframe"></iframe>
	</section>

	<section id="visits_per_order" class="container row">
		<iframe src="go_seo_VisitsPerOrderLine.html" title="Visits per order" class="tall-iframe"></iframe>
	</section>

	<section class="container row">
		<iframe src="go_seo_CMGRRevenue.html" title="CMGR Revenue" class="short-iframe"></iframe>
		<iframe src="go_seo_CMGRVisits.html" title="CMGR Visits" class="short-iframe"></iframe>
	</section>
	
	<section id="orders" class="container row">
		<iframe src="go_seo_OrdersBar.html" title="Order volume" class="medium-iframe"></iframe>
	</section>
	
	<section id="order_value" class="container row">
		<iframe src="go_seo_OrderValueBar.html" title="Order value" class="medium-iframe"></iframe>
	</section>
	
	<section class="container row">
		<iframe src="go_seo_CMGROrders.html" title="CMGR Orders" class="short-iframe"></iframe>
		<iframe src="go_seo_CMGROrderValue.html" title="CMGR Order Value" class="short-iframe"></iframe>
	</section>

	<section id="visit_value" class="container row">
		<iframe src="go_seo_VisitValueBar.html" title="Organic visit value" class="tall-iframe"></iframe>
	</section>
	
	<section class="container row">
		<iframe src="go_seo_CMGRVisitValue.html" title="CMGR Visit Value" class="short-iframe"></iframe>
		<iframe src="go_seo_Gauge.html" title="Visits per order gauge" class="short-iframe"></iframe>
	</section>
	
	<section id="detailed_insights" class="container row no-border">
		<iframe src="go_seo_DataInsightDetailKPIs.html" title="KPIs" class="tall-iframe" style="height: 690px;"></iframe>
	</section>

	<section id="revenue_forecast" class="container row">
		<iframe src="go_seo_Forecast.html" title="Revenue forecast" class="tall-iframe"></iframe>
		<iframe src="go_seo_ForecastNarrative.html" title="Visits per order" class="tall-iframe"></iframe>
	</section>

	<section id="wordcloud_branded" class="container row no-border">
 	   <iframe src="go_seo_WordcloudBranded.html" title="Branded Keyword wordcloud" class="tall-iframe" style="height: 650px; font-size: 10px;"></iframe>
 	   <iframe src="go_seo_BrandedInsights.html" title="Branded Keyword Insights" class="tall-iframe" style="height: 700px; font-size: 10px;"></iframe>
	</section>

	<section class="container row no-border section-padding-top section-padding-bottom">
 	   <iframe src="go_seo_WinningKeywordBranded.html" title="Winning branded keyword"s class="tall-iframe" style="height: 150px; font-size: 10px;"></iframe>
	</section>

	<section id="wordcloud_non_branded" class="container row no-border">
  	  <iframe src="go_seo_WordcloudNonBranded.html" title="Non Branded Keyword wordcloud" class="tall-iframe" style="height: 650px; font-size: 10px;"></iframe>
  	  <iframe src="go_seo_NonBrandedInsights.html" title="Non Branded Keyword Insights" class="tall-iframe" style="height: 700px; font-size: 10px;"></iframe>
	</section>

	<section class="container row no-border section-padding-top section-padding-bottom">
 	   <iframe src="go_seo_WinningKeywordNonBranded.html" title="Winning non-branded keywords" class="tall-iframe" style="height: 150px; font-size: 10px;"></iframe>
	</section>

	<section id="footer" class="container row no-border">
    	<iframe src="go_seo_FooterNotes.html" title="Footer" class="tall-iframe"></iframe>
	</section>
</div>

<!-- Create a new broadsheet -->
<button class="back-button" onclick="goHome()">New broadsheet</button>

</body>
</html>
`, width90, width90, width100, width100, width100, width0, company, protocol, fullHost, percent)
	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_BusinessInsights.html")
}

// Execute the BQL
func executeBQL(returnSize int, bqlToExecute string) []byte {

	// If a size needs to be added to the URL, define it here
	var returnSizeAppend string
	if returnSize > 0 {
		returnSizeAppend = "?size=" + fmt.Sprint(returnSize)
	}

	// Define the URL
	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query%s", organization, project, returnSizeAppend)

	// Define the body
	httpBody := []byte(bqlToExecute)

	// Create the POST request
	req, errorCheck := http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
	if errorCheck != nil {
		fmt.Println(red+"Error. executeBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
	}

	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)
	req.Header.Add("Content-Type", "application/json")

	// Create HTTP client and execute the request
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(red+"Error. executeBQL.  Cannot create the HTTP client:", errorCheck)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(red+"Error. executeBQL. Failed to close response body: %v\n"+reset, err)
		}
	}()

	// Read the response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(red+"Error. executeBQL. Cannot read response body:", err)
	}

	// Return the response body as a byte slice
	return responseData
}

// Compute the CMGR
func calculateCMGR(sessionID string) {

	// Revenue
	// Convert slice of integers to slice of floats for CMGR compute
	var seoRevenueFloat []float64
	for _, v := range seoRevenue {
		seoRevenueFloat = append(seoRevenueFloat, float64(v))
	}

	cmgrRevenue = computeCMGR(seoRevenueFloat, "Revenue")

	// Visits
	var seoVisitsFloat []float64
	for _, v := range seoVisits {
		seoVisitsFloat = append(seoVisitsFloat, float64(v))
	}
	cmgrVisits = computeCMGR(seoVisitsFloat, "Visits")

	// Visit value
	var seoMetricsVisitValueFloat []float64
	for _, v := range seoVisitValue {
		seoMetricsVisitValueFloat = append(seoMetricsVisitValueFloat, v)
	}
	cmgrVisitValue = computeCMGR(seoMetricsVisitValueFloat, "Visit Value")

	// Order volume
	var seoOrdersFloat []float64
	for _, v := range seoOrders {
		seoOrdersFloat = append(seoOrdersFloat, float64(v))
	}
	cmgrOrderValue = computeCMGR(seoOrdersFloat, "Orders")

	// Order value
	var seoOrdersValueFloat []float64
	for _, v := range seoOrderValue {
		seoOrdersValueFloat = append(seoOrdersValueFloat, float64(v))
	}
	cmgrOrderValueValue := computeCMGR(seoOrdersValueFloat, "Order value")

	fmt.Printf("\n" + yellow + sessionID + reset + " Compound Monthly Growth Rate\n" + reset)
	fmt.Printf("Revenue: %.2f\n", cmgrRevenue)
	fmt.Printf("Visits: %.2f\n", cmgrVisits)
	fmt.Printf("Visit value: %.2f\n", cmgrVisitValue)
	fmt.Printf("Order volume: %.2f\n", cmgrOrderValue)
	fmt.Printf("Order value: %.2f\n", cmgrOrderValueValue)
}

func computeCMGR(values []float64, calculatedKPIName string) float64 {

	if len(values) < 2 {
		return 0.0 // Cannot calculate CMGR with less than 2 values
	}

	initialValue := values[0]

	// The final period value is not included as it is not a full month
	finalValue := values[len(values)-1]
	numberOfPeriods := float64(noOfMonths)

	// CMGR formula: (finalValue / initialValue) ^ (1 / numberOfPeriods) - 1
	cmgr := math.Pow(finalValue/initialValue, 1/numberOfPeriods) - 1

	println()
	println(green + "CMGR Inputs" + reset)
	println(yellow + calculatedKPIName + reset)
	fmt.Printf("initialValue: %f\n", initialValue)
	fmt.Printf("finalValue: %f\n", finalValue)
	fmt.Printf("numberOfPeriods: %f\n", numberOfPeriods)

	return cmgr
}

// Get the analytics ID
func getAnalyticsID() (string, string) {

	// First identify which analytics tool is integrated
	urlAPIAnalyticsID := "https://api.botify.com/v1/projects/" + organization + "/" + project + "/collections"
	req, errorCheck := http.NewRequest("GET", urlAPIAnalyticsID, nil)

	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)
	req.Header.Add("Content-Type", "application/json")

	if errorCheck != nil {
		fmt.Println(red+"\nError. getAnalyticsID. Cannot create request:"+reset, errorCheck)
	}
	// Create HTTP client and execute the request
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, errorCheck := client.Do(req)
	if errorCheck != nil {
		fmt.Println(red+"Error. getAnalyticsID. Error: "+reset, errorCheck)
	}

	//defer resp.Body.Close()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(red+"Error. getAnalysisID. Failed to close response body: %v\n"+reset, err)
			return
		}
	}()

	// Read the response body
	responseData, errorCheck := io.ReadAll(resp.Body)
	if errorCheck != nil {
		fmt.Println(red+"Error. getAnalyticsID. Cannot read response body:"+reset, errorCheck)
	}

	// Unmarshal the JSON data into the struct
	var analyticsIDs []AnalyticsID
	if err := json.Unmarshal(responseData, &analyticsIDs); err != nil {
		fmt.Println(red+"Error. getAnalyticsID. The organisation and/or project name are probably incorrect. Cannot unmarshall the JSON:"+reset, err)
		return "errorNoProjectFound", ""
	}

	// Find and print the name value when the ID contains the word "visit"
	// Assume the first instance of "visit" contains the analytics ID
	for _, analyticsID := range analyticsIDs {
		if strings.Contains(analyticsID.ID, "visit") {
			return analyticsID.ID, analyticsID.AnalyticsDateStart
		}
	}

	return "errorNoAnalyticsIntegrated", ""
}

// Get the date ranges for the revenue and visits
func calculateDateRanges(analyticsStartDate string) DateRanges {

	// Is it the last day of the month?
	date := time.Now()
	isLastDayoftheMonth := isLastDayOfMonth(date)
	println(isLastDayoftheMonth)

	startTime, err := time.Parse("2006-01-02", analyticsStartDate)
	if err != nil {
		fmt.Println("Error parsing start date:", err)
		return DateRanges{}
	}

	// Get the current year and month
	currentYear, currentMonth, _ := time.Now().Date()

	// Calculate the last day of the previous month of the current year
	previousMonth := time.Date(currentYear, currentMonth-1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfPreviousMonth := previousMonth.AddDate(0, 1, -1)
	// Check if analyticsStartDate is more than 12 months older than lastDayOfPreviousMonth
	isMoreThan12MonthsDataAvailable := startTime.Before(lastDayOfPreviousMonth.AddDate(-1, 0, 0))

	var dateRanges [][2]time.Time

	// Full year data available
	if isMoreThan12MonthsDataAvailable {
		noOfMonths = 11
		currentTime := time.Now()
		// Calculate the date ranges for the last 12 months
		for i := 0; i < 12; i++ {
			prevMonth := currentTime.AddDate(0, -1, 0)
			startDate := time.Date(prevMonth.Year(), prevMonth.Month(), 1, 0, 0, 0, 0, currentTime.Location())

			var endDate time.Time

			if i == 0 {
				firstDayOfCurrentMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
				endDate = firstDayOfCurrentMonth.AddDate(0, 0, -1)
				fmt.Println("firstDayOfCurrentMonth")
				fmt.Println(firstDayOfCurrentMonth)
			} else {
				endDate = startDate.AddDate(0, 1, -1)
			}

			dateRanges = append(dateRanges, [2]time.Time{startDate, endDate}) //bloo
			fmt.Println("start end")
			fmt.Println(startDate)
			fmt.Println(endDate)
			fmt.Println("_______________")
			currentTime = startDate.AddDate(0, 0, 0)
		}
		// Less than a full year data available
	} else {
		currentTime := time.Now()
		monthsBetween := monthsBetween(startTime, currentTime)
		noOfMonths = monthsBetween - 1

		for i := monthsBetween - 1; i >= 0; i-- {
			currentStartDate := startTime.AddDate(0, i, 0)
			endDate := time.Date(currentStartDate.Year(), currentStartDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
			dateRanges = append(dateRanges, [2]time.Time{currentStartDate, endDate})
		}
	}

	// Return the date range slice
	return DateRanges{MonthlyRanges: dateRanges}
}

func isLastDayOfMonth(date time.Time) bool {
	nextDay := date.AddDate(0, 0, 1)
	return nextDay.Day() == 1
}

// DateRanges struct is used to store the date ranges for use in the BQL when the SEO KPIs are acquired
type DateRanges struct {
	MonthlyRanges [][2]time.Time
}

// Function to calculate the number of months between two dates
func monthsBetween(startDate, endDate time.Time) int {
	// Calculate the difference in months
	year1, month1, _ := startDate.Date()
	year2, month2, _ := endDate.Date()

	months := (year2-year1)*12 + int(month2-month1)

	// Adjust if endDate is before startDate
	if endDate.Before(startDate) {
		months = -months
	}

	return months
}

// Define the error page
func generateErrorPage(displayMessage string) {

	// If displayMessage is empty or nil display a default error message.
	if displayMessage == "" {
		displayMessage = "An Unknown error has occurred"
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go_Seo Broadsheet</title>
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
            background-color: Green;
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
            background-color: DeepSkyBlue;
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

<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">Business insights broadsheet</span>
</header>

<!-- Back Button -->
<button class="back-button" onclick="goHome()">Try again</button>

<!-- Error message -->
<div class="error-message" id="error-message">
    %s
</div>

<script>
    function goHome() {
        window.open('http://%s/', '_blank');
    }
</script>

</body>
</html>`, displayMessage, fullHost)

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_BusinessInsights_error.html")

}

func writeLog(sessionID, organization, project, analyticsID, statusDescription string) {

	// Define log file name
	fileName := envInsightsLogFolder + "/_seoBusinessInsights.log"

	// Check if the log file exists
	fileExists := true
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fileExists = false
	}

	// Open or create the log file
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf(red+"Error. writeLog. Cannot open log file: %s"+reset, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf(red+"Error. writeLog. Failed to close log file: %v\n"+reset, err)
		}
	}()

	// Get current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Construct log record
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
		sessionID, currentTime, organization, project, analyticsID, statusDescription)

	// If the file doesn't exist, write header first
	if !fileExists {
		header := "SessionID,Date,Organisation,Project,AnalyticsID,Status\n"
		if _, err := file.WriteString(header); err != nil {
			fmt.Printf(red+"Error. writeLog. Failed to write log header: %s"+reset, err)
		}
	}

	if _, err := file.WriteString(logRecord); err != nil {
		fmt.Printf(red+"Error. writeLog. Cannot write to log file: %s"+reset, err)
	}
}

func generateSessionID(length int) (string, error) {
	sessionID := make([]byte, length)
	if _, err := rand.Read(sessionID); err != nil {
		return "", err
	}

	// Add to the execution increment
	sessionIDCounter++

	var builder strings.Builder
	builder.WriteString(strconv.Itoa(sessionIDCounter))
	builder.WriteString("-")
	builder.WriteString(base64.URLEncoding.EncodeToString(sessionID))

	// Convert the builder to a string and return
	return builder.String(), nil
}

// Get the currency used
func getCurrencyCompany() string {

	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", organization, project)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(red+"\nError. getCurrencyCompany. Cannot create request:"+reset, err)
	}
	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(red+"\nError. getCurrencyCompany. Cannot sent request:"+reset, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(red+"Error. getCurrencyCompany. Failed to close response body: %v\n"+reset, err)
			return
		}
	}()

	responseData, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(red+"\nError. getCurrencyCompany. Cannot read response body:"+reset, err)
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		fmt.Println(red+"\nError. getCurrencyCompany. Cannot unmarshall JSON:"+reset, err)
	}

	// Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError. getCurrencyCompany. Invalid crawl or no crawls found in the project" + reset)
		return "errorNoProjectFound"
	}

	// If one currency has been found assume that's the base currency. If multiple currencies are found assume a default of $
	if len(responseObject.Results[0].Features.SemanticMetadata.StructuredData.Currencies.Offer) == 1 {
		currencyCode = responseObject.Results[0].Features.SemanticMetadata.StructuredData.Currencies.Offer[0]
	} else {
		currencyCode = "USD"
	}

	switch currencyCode {
	case "USD":
		currencySymbol = "$" // US Dollar
	case "EUR":
		currencySymbol = "€" // Euro
	case "GBP":
		currencySymbol = "£" // British Pound
	case "JPY":
		currencySymbol = "¥" // Japanese Yen
	case "AUD":
		currencySymbol = "A$" // Australian Dollar
	case "CAD":
		currencySymbol = "C$" // Canadian Dollar
	case "CHF":
		currencySymbol = "CHF" // Swiss Franc
	case "CNY":
		currencySymbol = "CN¥" // Chinese Yuan
	case "INR":
		currencySymbol = "₹" // Indian Rupee
	case "SGD":
		currencySymbol = "S$" // Singapore Dollar
	case "ZAR":
		currencySymbol = "R" // South African Rand
	case "AED":
		currencySymbol = "د.إ" // UAE Dirham
	default:
		currencySymbol = currencyCode // Unknown currency defaults to the code
	}

	// To determine the customer name first check the CompanyName. if it is empty use the first word of the FirstName, if a CompanyName is present use it.
	if responseObject.Results[0].Owner.CompanyName == nil {
		fullFirstName := strings.Fields(responseObject.Results[0].Owner.FirstName)
		company = fullFirstName[0]
	} else {
		companyName := responseObject.Results[0].Owner.CompanyName
		company = companyName.(string)
	}

	return "success"
}

func createInsightsCacheFolder(cacheFolder string) {

	insightsDir := cacheFolder
	// Check if the folder already exists
	if _, err := os.Stat(insightsDir); os.IsNotExist(err) {
		// Create the folder and any necessary parents
		err := os.MkdirAll(insightsDir, 0755)
		if err != nil {
			fmt.Printf(red+"Error. Failed to create the insights cache folder: %v"+insightsDir+reset, err)
			fmt.Println()
		}
	}
}

func getHostnamePort() {

	// Load the INI file
	cfg, err := ini.Load("seoBusinessInsights.ini")
	if err != nil {
		fmt.Printf(red+"Error. getHostnamePort. Failed to read seoBusinessInsights.ini file: %v"+reset, err)
	}

	// Get values from the .ini file
	if !cfg.Section("").HasKey("protocol") {
		fmt.Println(yellow + "Warning: 'protocol' not found in configuration file. Will default to HTTPS." + reset)
		// Default when no protocol key is found in the .ini file
		protocol = "https"
	} else {
		protocol = cfg.Section("").Key("protocol").String()
	}

	if !cfg.Section("").HasKey("hostname") {
		fmt.Println(yellow + "Warning: 'hostname' not found in configuration file. Will default to localhost." + reset)
	} else {
		hostname = cfg.Section("").Key("hostname").String()
	}

	if !cfg.Section("").HasKey("port") {
		fmt.Println(yellow + "Warning: 'port' not found in configuration file. By default no port number will be used." + reset)
		port = ""
	} else {
		port = cfg.Section("").Key("port").String()
		port = ":" + port
	}

	// Add port to the hostname if running locally.
	if envInsightsHostingMode == "local" {
		fullHost = hostname + port
	} else {
		fullHost = hostname
	}

	var serverHostname, serverPort string
	serverHostname = hostname
	serverPort = port

	fmt.Printf(green+"\nHostname: %s\n"+reset, serverHostname)
	fmt.Printf(green+"Port: %s\n"+reset, serverPort)
}

// Function used to inverse the dates in the date slice. Used to ensure the latest data is display to the right side of the chart
func invertStringSlice(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// Get environment variables for token and cache folders
func getEnvVariables() (envBotifyAPIToken string, envInsightsLogFolder string, envInsightsFolder string, envInsightsHostingMode string) {

	// Botify API token from the env. variable getbotifyAPIToken
	envBotifyAPIToken = os.Getenv("envBotifyAPIToken")
	if envBotifyAPIToken == "" {
		fmt.Println(red + "Error. getEnvVariables. envBotifyAPIToken environment variable not set." + reset)
		fmt.Println(red + "Cannot start seoBusinessInsights server." + reset)
		os.Exit(0)
	}

	// Storage folder for the log file
	envInsightsLogFolder = os.Getenv("envInsightsLogFolder")
	if envInsightsLogFolder == "" {
		fmt.Println(red + "Error. getEnvVariables. envInsightsLogFolder environment variable not set." + reset)
		fmt.Println(red + "Cannot start seoBusinessInsights server." + reset)
		os.Exit(0)
	} else {
		fmt.Println()
		fmt.Println(green + "Log folder: " + envInsightsLogFolder + reset)
	}

	// Storage folder for the cached insights
	envInsightsFolder = os.Getenv("envInsightsFolder")
	if envInsightsFolder == "" {
		fmt.Println(red + "Error. getEnvVariables. envInsightsFolder environment variable not set." + reset)
		fmt.Println(red + "Cannot start seoBusinessInsights server." + reset)
		os.Exit(0)
	} else {
		fmt.Println(green + "seoBusinessInsights cache folder: " + envInsightsFolder + reset)
	}

	// Hosting mode. This will be either "local" or "docker"
	envInsightsHostingMode = os.Getenv("envInsightsHostingMode")
	if envInsightsHostingMode == "" {
		fmt.Println(red + "Error. getEnvVariables. envInsightsHostingMode environment variable not set." + reset)
		fmt.Println(red + "Cannot start seoBusinessInsights server." + reset)
		os.Exit(0)
	} else {
		fmt.Println(green + "seoBusinessInsights hosting mode: " + envInsightsHostingMode + reset)
	}

	return envBotifyAPIToken, envInsightsLogFolder, envInsightsFolder, envInsightsHostingMode
}

// Display the welcome banner, get the hostname and environment variables
func startup() {

	// Clear the screen
	fmt.Print(clearScreen)

	fmt.Print(green + `
 ██████╗  ██████╗         ███████╗███████╗ ██████╗ 
██╔════╝ ██╔═══██╗        ██╔════╝██╔════╝██╔═══██╗
██║  ███╗██║   ██║        ███████╗█████╗  ██║   ██║
██║   ██║██║   ██║        ╚════██║██╔══╝  ██║   ██║
╚██████╔╝╚██████╔╝███████╗███████║███████╗╚██████╔╝
 ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚══════╝ ╚═════╝`)

	fmt.Print(purple + `
██████╗ ██╗   ██╗███████╗██╗███╗   ██╗███████╗███████╗███████╗██╗███╗   ██╗███████╗██╗ ██████╗ ██╗  ██╗████████╗███████╗
██╔══██╗██║   ██║██╔════╝██║████╗  ██║██╔════╝██╔════╝██╔════╝██║████╗  ██║██╔════╝██║██╔════╝ ██║  ██║╚══██╔══╝██╔════╝
██████╔╝██║   ██║███████╗██║██╔██╗ ██║█████╗  ███████╗███████╗██║██╔██╗ ██║███████╗██║██║  ███╗███████║   ██║   ███████╗
██╔══██╗██║   ██║╚════██║██║██║╚██╗██║██╔══╝  ╚════██║╚════██║██║██║╚██╗██║╚════██║██║██║   ██║██╔══██║   ██║   ╚════██║
██████╔╝╚██████╔╝███████║██║██║ ╚████║███████╗███████║███████║██║██║ ╚████║███████║██║╚██████╔╝██║  ██║   ██║   ███████║
╚═════╝  ╚═════╝ ╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝ 
`)

	fmt.Println()
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(green + "\nseoBusinessInsights server is ON\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Server started at " + formattedTime + reset)

	// Get the environment variables for token, log folder & cache folder
	envBotifyAPIToken, envInsightsLogFolder, envInsightsFolder, envInsightsHostingMode = getEnvVariables()

	// Get the hostname and port
	getHostnamePort()

	fmt.Println(green + "\n... waiting for requests\n" + reset)
}

// CleanInsights is used to remove all slices where there are zero values in the revenue and / or visits data
func cleanInsights(seoRevenue []int, seoVisits []int, seoOrders []int, seoOrderValue []int, seoVisitValue []float64, visitsPerOrder []int, startMonthDates, endMonthDates, startMonthNames []string) ([]int, []int, []int, []int, []float64, []int, []string, []string, []string) {
	var filteredSEORevenue []int
	var filteredSEOVisits []int
	var filteredSEOOrders []int
	var filteredSEOOrderValue []int
	var filteredSEOVisitValue []float64
	var filteredVisitsPerOrder []int
	var filteredStartMonthDates []string
	var filteredEndMonthDates []string
	var filteredStartMonthNames []string

	for i, value := range seoRevenue {
		if value != 0 {
			filteredSEORevenue = append(filteredSEORevenue, value)
			filteredSEOVisits = append(filteredSEOVisits, seoVisits[i])
			filteredSEOOrders = append(filteredSEOOrders, seoOrders[i])
			filteredSEOOrderValue = append(filteredSEOOrderValue, seoOrderValue[i])
			filteredSEOVisitValue = append(filteredSEOVisitValue, seoVisitValue[i])
			filteredVisitsPerOrder = append(filteredVisitsPerOrder, visitsPerOrder[i])
			filteredStartMonthDates = append(filteredStartMonthDates, startMonthDates[i])
			filteredEndMonthDates = append(filteredEndMonthDates, endMonthDates[i])
			filteredStartMonthNames = append(filteredStartMonthNames, startMonthNames[i])
		}
	}

	// Update the number of months based on the reduced slice size
	noOfMonths = len(filteredStartMonthDates)

	return filteredSEORevenue, filteredSEOVisits, filteredSEOOrders, filteredSEOOrderValue, filteredSEOVisitValue, filteredVisitsPerOrder, filteredStartMonthDates, filteredEndMonthDates, filteredStartMonthNames
}
