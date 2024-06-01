// seoCharts: Charting SEO insights
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

// anonymous mode. When set to true the URL to the project defaults to 'http://go-seo.rf.gd/'
var anonymousMode = true

// DateRanges struct used to hold the monthly date ranges and the YTD date range
// Used for revenue and visits data
type DateRanges struct {
	MonthlyRanges [][2]time.Time
	YTDRange      [2]time.Time
}

// Slice used to store the name of the month
var startMthNames []string

// Define the slice to store startMthDate and endMthDate separately
var startMthDates = make([]string, 0)
var endMthDates = make([]string, 0)

// Slices used to store the SEO metrics
var seoMetricsRevenue []int
var seoMetricsVisits []int
var seoMetricsOrders []int
var seoOrderValue []int
var seoVisitValue []float64

// Variables used to store the CMGR values
var cmgrRevenue float64
var cmgrVisits float64
var cmgrVisitValue float64
var cmgrOrdersValue float64
var cmgrOrdersValueValue float64

// Totals
var totalVisits int
var totalRevenue int
var totalOrders int
var totalVisitsPerOrder float64

// Slice to store the visits per order for each month
var visitsPerOrder []int

// AnalyticsID is used to identify which analytics tool is in use
type AnalyticsID struct {
	ID string `json:"id"`
}

// Result is used to store the revenue, Orders and visits
type Result struct {
	Dimensions []interface{} `json:"dimensions"`
	Metrics    []float64     `json:"metrics"`
}

type Response struct {
	Results []Result `json:"results"`
}

// Project URL
var projectURL = ""

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

// Botify API token here
var botifyApiToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

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
		projectURL = "http://go-seo.rf.gd/"
	} else {
		projectURL = "https://app.botify.com/" + orgName + "/" + projectName
	}

	displaySeparator()

	// Get revenue, visit and Order for the last 12 months
	getSeoInsights()

	// Start of charts

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

	// Generate the charts for the insights detail
	dataInsightsDetail()

	// Badges
	cmgrRevenue32 := float32(cmgrRevenue)                   // Cast to float32
	cmgrVisits32 := float32(cmgrVisits)                     // Cast to float32
	cmgrVisitValue32 := float32(cmgrVisitValue)             // Cast to float32
	cmgrOrdersValue32 := float32(cmgrOrdersValue)           // Cast to float32
	cmgrOrdersValueValue32 := float32(cmgrOrdersValueValue) // Cast to float32

	liquidBadges("Revenue", cmgrRevenue32)
	liquidBadges("Visits", cmgrVisits32)
	liquidBadges("Visit Value", cmgrVisitValue32)
	liquidBadges("Orders", cmgrOrdersValue32)
	liquidBadges("Order Value", cmgrOrdersValueValue32)

	// River chart
	riverCharRevenueVisits() // Revenue & visits

	// Gauge chart
	gaugeChartVisitsPerOrder(totalVisitsPerOrder)

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

	// Populate the slice with string versions of the date ready for use in the BQL
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

	// Format the YTD range ready for use in the BQL
	startYTDDate := dateRanges.YTDRange[0].Format("20060102")
	endYTDDate := dateRanges.YTDRange[1].Format("20060102")

	// Get the revenue data
	getRevenueData(analyticsID, startYTDDate, endYTDDate, startMthDates, endMthDates)

	// Calculate the CMGR for the metrics
	getCMGR()

}

// Get the revenue, Orders and visits data
func getRevenueData(analyticsID string, startYTDDate string, endYTDDate string, startMthDates []string, endMthDates []string) {

	var ytdMetricsOrders = 0
	var ytdMetricsRevenue = 0
	var ytdMetricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Get monthly insights
	fmt.Println(bold + "\nMonthly organic insights" + reset)
	for i := range startMthDates {

		ytdMetricsOrders, ytdMetricsRevenue, ytdMetricsVisits, avgOrderValue, avgVisitValue = executeRevenueBQL(analyticsID, startMthDates[i], endMthDates[i])

		// A little hack
		// If any of the returned values is zero default them to 1 in order to avoid divide by zero issues later.
		if ytdMetricsVisits == 0 {
			ytdMetricsVisits = 1
		}
		if avgOrderValue == 0 {
			avgOrderValue = 1
		}
		if avgVisitValue == 0 {
			avgVisitValue = 1
		}
		if ytdMetricsOrders == 0 {
			ytdMetricsOrders = 1
		}
		if ytdMetricsRevenue == 0 {
			ytdMetricsRevenue = 1
		}

		// Display the metrics (formatted)
		fmt.Printf(green+"Start: %s End: %s\n"+reset, startMthDates[i], endMthDates[i])
		formattedOrders := formatWithCommas(ytdMetricsOrders)
		fmt.Println("No. Orders:", formattedOrders)
		formattedRevenue := formatWithCommas(ytdMetricsRevenue)
		fmt.Println("Total revenue:", formattedRevenue)
		fmt.Println("Average order value:", avgOrderValue)
		formattedVisits := formatWithCommas(ytdMetricsVisits)
		fmt.Println("No. of visits:", formattedVisits)
		fmt.Println("Average visit value:", avgVisitValue)
		fmt.Println()

		// Append the metrics to the slices
		seoMetricsOrders = append(seoMetricsOrders, ytdMetricsOrders)
		seoMetricsRevenue = append(seoMetricsRevenue, ytdMetricsRevenue)
		seoOrderValue = append(seoOrderValue, avgOrderValue)
		seoMetricsVisits = append(seoMetricsVisits, ytdMetricsVisits)
		// Round avgVisitValue to 3 decimal places
		avgVisitValueRounded := math.Round(avgVisitValue*100) / 100
		seoVisitValue = append(seoVisitValue, avgVisitValueRounded)

		// Calculate the visits per order (for the month)
		visitsPerOrder = append(visitsPerOrder, ytdMetricsVisits/ytdMetricsOrders)

		// Calculate the total revenue
		totalRevenue += ytdMetricsRevenue
		// Calculate the visits per order (average across all data)
		totalVisits += ytdMetricsVisits
		totalOrders += ytdMetricsOrders
		totalVisitsPerOrder = float64(totalVisits / totalOrders)
	}

	fmt.Println(green + "Totals" + reset)
	fmt.Println("Total visits:", totalVisits)
	fmt.Println("Total revenue:", totalRevenue)
	fmt.Println("Total orders:", totalOrders)
	fmt.Println("Visits per order:", totalVisitsPerOrder)
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

	// Calculate the YTD date range
	year, _, _ := currentTime.Date()
	loc := currentTime.Location()
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
	endOfYTD := currentTime
	yearToDateRange := [2]time.Time{startOfYear, endOfYTD}

	// Calculate the date ranges for the last 12 months
	for i := 0; i < 12; i++ {
		// Calculate the start and end dates for the current range
		year, month, _ := currentTime.Date()
		loc := currentTime.Location()

		// Start of the current month range
		startDate := time.Date(year, month, 1, 0, 0, 0, 0, loc)

		var endDate time.Time
		if i == 0 {
			// End of the current month range (up to the current date)
			endDate = currentTime
		} else {
			// End of the previous month range
			endDate = startDate.AddDate(0, 1, -1)
		}

		// Store the range
		dateRanges[11-i] = [2]time.Time{startDate, endDate}

		// Move to the previous month
		currentTime = startDate.AddDate(0, -1, 0)
	}

	return DateRanges{MonthlyRanges: dateRanges, YTDRange: yearToDateRange}
}

// Execute the BQL for the specified date range
func executeRevenueBQL(analyticsID string, startDate string, endDate string) (int, int, int, int, float64) {

	// Get the revenue, no. Orders and visits - YTD
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
		log.Fatal("Error. executeRevenueBQL. Cannot create request. Perhaps the provided credentials are invalid: ", errorCheck)
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
		log.Fatalf("Error. executeRevenueBQL. Cannot unmarshal the JSON: %v", err)
	}

	var ytdMetricsOrders = 0
	var ytdMetricsRevenue = 0
	var ytdMetricsVisits = 0
	var avgOrderValue = 0
	var avgVisitValue = 0.00

	// Check if any data has been returned from the API. Count the number of elements in the response.Results slice
	responseCount := len(response.Results)

	if responseCount == 0 {
		fmt.Println(yellow + "Warning. executeRevenueBQL. Some data may default to 1 if it's the first day of the month." + reset)
	} else {
		ytdMetricsOrders = int(response.Results[0].Metrics[0])
		ytdMetricsRevenue = int(response.Results[0].Metrics[1])
		ytdMetricsVisits = int(response.Results[0].Metrics[2])
		// Compute the average Order value
		avgOrderValue = ytdMetricsRevenue / ytdMetricsOrders
		avgVisitValue = float64(ytdMetricsRevenue) / float64(ytdMetricsVisits)
	}
	return ytdMetricsOrders, ytdMetricsRevenue, ytdMetricsVisits, avgOrderValue, avgVisitValue
}

// Table for total Visits, Orders & Revenue
func tableTotalsVisitsOrdersRevenue() {

	// Generate HTML content
	htmlContent := generateTableTotalsHTML(totalVisits, totalOrders, totalRevenue)

	// Write HTML content to a file
	file, err := os.Create("./Utilities/seoChartsWeb/tableTotalsVisitsOrdersRevenue.html")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(htmlContent)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("HTML file generated successfully.")
}

func generateTableTotalsHTML(totalVisits, totalOrders int, totalRevenue int) string {

	//bloo
	totalVisitsFormatted := formatInt(totalVisits)
	totalOrdersFormatted := formatInt(totalOrders)
	totalRevenueFormatted := formatInt(totalRevenue)

	html := `
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
	return html
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
			Title:    "Visits per order",
			Subtitle: "On average, how many organic visits are needed to generate each order?",
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

func liquidBadges(badgeKPI string, badgeKPIValue float32) {

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
	liquid.AddSeries(badgeKPI, genLiquidItems([]float32{badgeKPIValue})).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show: opts.Bool(true),
			}),

			charts.WithLiquidChartOpts(opts.LiquidChart{
				IsWaveAnimation: opts.Bool(true),
				//IsShowOutline:   opts.Bool(true),
				Shape: "diamond",
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

// River chart
func riverCharRevenueVisits() {

	page := components.NewPage()
	page.AddCharts(
		generateRiverTime(),
	)
	f, err := os.Create("./Utilities/seoChartsWeb/seoRiver.html")
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
		charts.WithColorsOpts(opts.Colors{kpiColourVisits, kpiColourRevenue}), //bloo
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
func gaugeChartVisitsPerOrder(visitsPerOrder float64) {

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
	gauge.SetGlobalOptions(
	//  No options defined
	)

	gauge.AddSeries("Visits Per Order", []opts.GaugeData{{Name: "Visits per order", Value: visitsPerOrder}})

	return gauge
}

// CMGR
func getCMGR() {

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

	fmt.Printf(green + "\nCompound Monthly Growth Rate (CMGR)\n" + reset)
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
	// The final period value us not included as it is not a full month
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
func dataInsightsDetail() {
	var detailedKPIstableData [][]string

	for i := 0; i < len(startMthDates); i++ {
		formattedDate := formatDate(startMthDates[i])
		row := []string{
			//startMthDates[i],
			formattedDate,
			formatInt(seoMetricsOrders[i]),
			formatInt(seoMetricsRevenue[i]),
			formatInt(seoOrderValue[i]),
			formatInt(seoMetricsVisits[i]),
			formatFloat(seoVisitValue[i]),
		}
		detailedKPIstableData = append(detailedKPIstableData, row)
	}

	// Generate the table
	detailedKPIInsightsTable := generateHTML(detailedKPIstableData)

	// Write the HTML content
	file, err := os.Create("./Utilities/seoChartsWeb/dataInsightDetailKPIs.html")
	if err != nil {
		fmt.Println(red+"Error. dataInsightsDetail. Cannot create KPI detail table:"+reset, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(detailedKPIInsightsTable)
	if err != nil {
		fmt.Println(red+"Error. dataInsightsDetail. Cannot write KPI detail table:"+reset, err)
		return
	}
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
func generateHTML(data [][]string) string {
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

// Display the line break
func displaySeparator() {
	block := "█"
	fmt.Println()

	for i := 0; i < 130; i++ {
		fmt.Print(block)
	}

	fmt.Println()
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
	fmt.Println(checkmark + green + bold + " Revenue (YTD/monthly)" + reset)
	fmt.Println(checkmark + green + bold + " Visits (YTD/monthly)" + reset)
	fmt.Println(checkmark + green + bold + " Orders (YTD/monthly)" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) Average order value" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) Average visit value" + reset)
	fmt.Println(checkmark + green + bold + " (Computed) CMGR for Revenue, Visits, Orders, Order value, Visit value" + reset)

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
