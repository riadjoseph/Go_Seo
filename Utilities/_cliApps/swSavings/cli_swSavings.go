// swSavings. Calculate how much page load time has been removed from the origin server with SW
// Written by Jason Vicinanza

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Version
var version = "v0.1"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"

// APIToken should be replaced with your own Botify API token.
var APIToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

// Struct to store page load speed info
type LoadSpeed struct {
	Results []struct {
		Dimensions []string  `json:"dimensions"`
		Metrics    []float64 `json:"metrics"`
	}
}

func main() {

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

	// Run the BQL etc.
	swSavings()

	displaySeparator()

	swSavingsDone()
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit bqlTester" +
			"\n")

		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using bqlTester. Goodbye!")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using bqlTester. Goodbye!")
			os.Exit(0)
		}
	}
}

func swSavings() {

	var urlGetLoadData string
	urlGetLoadData = "https://api.botify.com/v1/projects/" + orgName + "/" + projectName + "/query"
	req, errorCheck := http.NewRequest("GET", urlGetLoadData, nil)
	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+APIToken)
	req.Header.Add("Content-Type", "application/json")
	// Create HTTP client and execute the request
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, errorCheck := client.Do(req)
	if errorCheck != nil {
		log.Fatal(red+"\nError. swSavings. Cannot create request: "+reset, errorCheck)
	}
	defer resp.Body.Close()
	// Read the response body
	responseData, errorCheck := io.ReadAll(resp.Body)
	if errorCheck != nil {
		log.Fatal(red+"Error. swSavings. Cannot read response body: "+reset, errorCheck)
	}
	// Unmarshal the JSON data into the struct
	var loadSpeedData LoadSpeed
	if err := json.Unmarshal(responseData, &loadSpeedData); err != nil {
		log.Fatal(red+"Error. swSavings. Cannot unmarshall the JSON: "+reset, err)
	}

	// Get the load speed data and compute the total savings
	executeLoadSpeedBQL()
}

func executeLoadSpeedBQL() {

	// Get the dates for the BQL. Assume the last 28 days
	today, daysAgo := getDates()

	bqlLoadSpeed := fmt.Sprintf(`
	{
    "collections": [
        "activation.delivery",
        "activation.indexation"
    ],
    "periods": [
        [
            "%s",
            "%s"
        ]
    ],
    "query": {
        "dimensions": [
            "activation.delivery.period_0.url"
        ],
        "metrics": [
            {
                "field": "activation.delivery.period_0.avg_delivery_time",
                "name": "Avg. Delivery Time"
            },
                        {
                "field": "activation.delivery.period_0.avg_original_time",
                "name": "Avg. Original Time"
            },
            {
                "field": "activation.delivery.period_0.count_hits",
                "name": "No. of requests"
            }
        ],
        "sort": []
    }
}
`, daysAgo, today)

	totalDeliveryTimeSW := 0
	totalOriginalTimeOrigin := 0
	totalBotRequests := 0
	totalURLs := 0
	page := 0

	for i := 0; i < 2000; i++ {

		page++

		url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query?size=2000&page=%d", orgName, projectName, page)

		displayBanner()

		fmt.Println()
		fmt.Println(purple+"Page"+reset, page)
		fmt.Println(purple+"URL Count"+reset, totalURLs)
		fmt.Println(purple+"Savings so far (ms):"+reset, totalOriginalTimeOrigin)
		totalOriginalTimeOriginHours := float64(totalOriginalTimeOrigin) / 1000 / 60 / 60
		fmt.Printf(purple+"In hours: %.2f"+reset, totalOriginalTimeOriginHours)

		// GET the HTTP request
		req, errorCheck := http.NewRequest("GET", url, nil)
		if errorCheck != nil {
			log.Fatal(red+"\nError. executeLoadSpeedBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
		}

		// Define the body
		httpBody := []byte(bqlLoadSpeed)

		// Create the POST request
		req, errorCheck = http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
		if errorCheck != nil {
			log.Fatal(red+"Error. executeLoadSpeedBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
		}

		// Define the headers
		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", "token "+APIToken)
		req.Header.Add("Content-Type", "application/json")

		// Create HTTP client and execute the request
		client := &http.Client{
			Timeout: 20 * time.Second,
		}
		resp, errorCheck := client.Do(req)
		if errorCheck != nil {
			log.Fatal(red+"Error.  executeLoadSpeedBQL. Error: "+reset, errorCheck)
		}
		defer resp.Body.Close()

		// Read the response body
		responseData, errorCheck := io.ReadAll(resp.Body)

		if errorCheck != nil {
			log.Fatal(red+"Error.  executeLoadSpeedBQL. Cannot read response body: "+reset, errorCheck)
			return
		}

		// Unmarshal the JSON data into the struct
		var response LoadSpeed
		err := json.Unmarshal(responseData, &response)
		if err != nil {
			log.Fatalf(red+"Error.  executeLoadSpeedBQL. Cannot unmarshal the JSON: %v"+reset, err)
		}

		// Check if any data has been returned from the API. Count the number of elements in the Results array
		responseCount := len(response.Results)
		if responseCount == 0 {
			fmt.Println()
			fmt.Println(bold + green + "Analysis complete" + reset)
			break
		}
		for i := 0; i < responseCount; i++ {
			url := response.Results[i].Dimensions[0]
			avgDeliveryTimeSW := int(response.Results[i].Metrics[0])
			avgOriginalTimeOrigin := int(response.Results[i].Metrics[1])
			noRequests := int(response.Results[i].Metrics[2])
			totalURLs = totalURLs + 1
			//fmt.Println("URL ", url)
			//fmt.Println("Average delivery time with SW", avgDeliveryTimeSW)
			//fmt.Println("Average original time", avgOriginalTimeOrigin)
			//fmt.Println("No. of bot requests", noRequests)
			//  totalDeliveryTimeSW is the total amount of page load time managed by SW
			totalDeliveryTimeSW = totalDeliveryTimeSW + avgDeliveryTimeSW
			//  totalOriginalTimeOrigin is the total amount of page load time shifted to SW from the origin site
			totalOriginalTimeOrigin = totalOriginalTimeOrigin + avgOriginalTimeOrigin
			// totalBotRequests is the total number of times the URL was requested
			totalBotRequests = totalBotRequests + noRequests
			writeToCSV(url, avgOriginalTimeOrigin, avgDeliveryTimeSW, noRequests)
		}
	}
	fmt.Println(purple+"Grand Final total savings (ms):"+reset, totalOriginalTimeOrigin)
	totalOriginalTimeOriginHours := float64(totalOriginalTimeOrigin) / 1000 / 60 / 60
	fmt.Printf(purple+"Grand Final total savings (hours): %.2f"+reset, totalOriginalTimeOriginHours)
	fmt.Println()
	fmt.Println(purple+"Total URLs processed"+reset, totalURLs)
}

// writeToCSV writes the given parameters to a CSV file.
func writeToCSV(url string, originalSpeed, deliverySpeed, noRequests int) error {
	// Create or open the CSV file
	file, err := os.OpenFile("swSavings.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf(red+"Error. failed to open file: %w"+reset, err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the headers if the file is new
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf(red+"Error. failed to get file info: %w"+reset, err)
	}
	if fileInfo.Size() == 0 {
		headers := []string{"URL", "Original Speed (Origin)", "Delivery Speed (SW)", "No. of Requests"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf(red+"Error. failed to write headers: %w"+reset, err)
		}
	}

	// Write the data
	record := []string{
		url,
		fmt.Sprintf("%d", originalSpeed),
		fmt.Sprintf("%d", deliverySpeed),
		fmt.Sprintf("%d", noRequests),
	}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf(red+"Error. failed to write record: %w"+reset, err)
	}

	return nil
}

// getDates returns today's date and the date 28 days ago in YYYYMMDD format.
func getDates() (string, string) {
	// Get today's date
	today := time.Now()

	// Format today's date in YYYYMMDD format
	todayFormatted := today.Format("20060102")

	// Calculate the date 28 days ago
	daysAgo := today.AddDate(0, 0, -28)

	// Format the date 28 days ago in YYYYMMDD format
	daysAgoFormatted := daysAgo.Format("20060102")

	return todayFormatted, daysAgoFormatted
}

func swSavingsDone() {
	// We're done
	fmt.Println(purple + "\nswSavings: Done!")
	os.Exit(0)
}

func displaySeparator() {
	block := "█"
	fmt.Println()

	for i := 0; i < 130; i++ {
		fmt.Print(block)
	}
	fmt.Println()
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

func displayBanner() {

	clearScreen()

	fmt.Print(green + `
 ██████╗  ██████╗         ███████╗███████╗ ██████╗ 
██╔════╝ ██╔═══██╗        ██╔════╝██╔════╝██╔═══██╗
██║  ███╗██║   ██║        ███████╗█████╗  ██║   ██║
██║   ██║██║   ██║        ╚════██║██╔══╝  ██║   ██║
╚██████╔╝╚██████╔╝███████╗███████║███████╗╚██████╔╝
 ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚══════╝ ╚═════╝`)

	fmt.Print(purple + `

███████╗██╗    ██╗███████╗ █████╗ ██╗   ██╗██╗███╗   ██╗ ██████╗ ███████╗
██╔════╝██║    ██║██╔════╝██╔══██╗██║   ██║██║████╗  ██║██╔════╝ ██╔════╝
███████╗██║ █╗ ██║███████╗███████║██║   ██║██║██╔██╗ ██║██║  ███╗███████╗
╚════██║██║███╗██║╚════██║██╔══██║╚██╗ ██╔╝██║██║╚██╗██║██║   ██║╚════██║
███████║╚███╔███╔╝███████║██║  ██║ ╚████╔╝ ██║██║ ╚████║╚██████╔╝███████║
╚══════╝ ╚══╝╚══╝ ╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝
`)

	fmt.Println()
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(green + "\nswSavings processing\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Analysis started at " + formattedTime + reset)

	fmt.Println(bold+"\nOrganisation name:", orgName)
	fmt.Println(bold+"Project name:", projectName+reset)
	fmt.Println()

	displaySeparator()
}
