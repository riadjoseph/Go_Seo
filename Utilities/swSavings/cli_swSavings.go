// swSavings. Calculate how much page load time has been shifted from the origin server with SW
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

// Token, log folder and cache folder acquired from environment variables
var envBotifyAPIToken string

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var reset = "\033[0m"

// Strings used to store the project credentials for API access
var organisation string
var project string

// Strings used to store the input project credentials
var organisationInput string
var projectInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

// Struct to store page load speed info
type LoadSpeed struct {
	Results []struct {
		Dimensions []string  `json:"dimensions"`
		Metrics    []float64 `json:"metrics"`
	} `json:"results"`
	Count int `json:"count"`
}

var totalURLsCached int
var totalPages int
var remainingSeconds float64
var firstPass bool = true
var finishingIndicator bool = true
var seconds float64
var totalURLsToProcess = 5000000

// Define the layout for parsing the time input
const layout = "2006-01-02 15:04:05"

func main() {

	clearScreen()
	fmt.Print(green + "\nInitializing swSavings. One moment please." + reset)

	// Get the environment variables for token, log folder & cache folder
	envBotifyAPIToken = getEnvVariables()

	// Delete the CSV from the previous analysis
	deleteCSV()

	// Get the credentials if they have not been specified on the command line
	checkCredentials()

	// If the credentials have been provided on the command line use them
	if !credentialsInput {
		organisation = os.Args[1]
		project = os.Args[2]
	} else {
		organisation = organisationInput
		project = projectInput
	}

	executeLoadSpeedBQL()

	displaySeparator()

	swSavingsDone()
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit swSavings" +
			"\n")
		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&organisationInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(organisationInput) == "" {
			fmt.Println(green + "\nThank you for using swSavings. Goodbye!")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		fmt.Scanln(&projectInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectInput) == "" {
			fmt.Println(green + "\nThank you for using swSavings. Goodbye!")
			os.Exit(0)
		}
	}
}

func executeLoadSpeedBQL() {

	projectExists := checkProjectExists()
	if !projectExists {
		os.Exit(1)
	}

	// Now get the total URLs cached for this date period
	totalURLsCached, totalPages = getURLCount()

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

	// Will process a max of 2500 pages. 5 million total URLs.
	for i := 0; i < 2500; i++ {

		// get start time. Used to calculate the remaing analysis time
		startTime := time.Now()

		page++

		displayBanner()

		url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query?size=2000&page=%d", organisation, project, page)

		fmt.Println()
		fmt.Println(purple+"Total URLs cached (Max processing limit is 5MM pages): "+reset, totalURLsCached)

		// If there are more URLs cached than will be processed
		if totalURLsCached > totalURLsToProcess {
			totalURLsCached = totalURLsToProcess
		}

		// Calculate the %age progress
		percentProgress := float64(totalURLs) / float64(totalURLsCached) * 100
		percentProgressInt := int(percentProgress)

		fmt.Println(purple+"URLs processed: "+reset, totalURLs)
		fmt.Println(purple+"Progress: "+reset, percentProgressInt, "%")
		fmt.Println(purple+"Computed time saving (ms): "+reset, totalOriginalTimeOrigin)
		totalOriginalTimeOriginHours := float64(totalOriginalTimeOrigin) / 1000 / 60 / 60
		fmt.Printf(purple+"Computed time saving (hours): "+reset+"%.2f", totalOriginalTimeOriginHours)
		fmt.Println()

		// As the time remaining is approx. it could be negative. If negative set to zero.
		if remainingSeconds < 0 {
			//remainingSeconds = 0
			finishingIndicator = true
		}

		if remainingSeconds > 0 {
			fmt.Printf(purple+"Approx. analysis time remaining (seconds):  "+reset+"%.0f", remainingSeconds)
		}

		if finishingIndicator && remainingSeconds < 0 {
			fmt.Println()
			fmt.Println(green + "Finishing up..." + reset)
		}

		// GET the HTTP request
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			log.Fatal(red+"\nError. executeLoadSpeedBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, err)
		}

		httpBody := []byte(bqlLoadSpeed)

		req, err = http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
		if err != nil {
			log.Fatal(red+"Error. executeLoadSpeedBQL. Cannot create request. Perhaps the provided credentials are invalid: "+reset, err)
		}

		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", "token "+envBotifyAPIToken)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(red+"Error.  executeLoadSpeedBQL. Error: "+reset, err)
		}
		defer resp.Body.Close()

		responseData, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(red+"Error.  executeLoadSpeedBQL. Cannot read response body: "+reset, err)
			return
		}

		var response LoadSpeed
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			log.Fatalf(red+"Error.  executeLoadSpeedBQL. Cannot unmarshal the JSON: %v"+reset, err)
		}

		responseCount := len(response.Results)
		if responseCount == 0 {
			fmt.Println()
			fmt.Println()
			fmt.Println(green + "Analysis complete" + reset)
			break
		}

		for i := 0; i < responseCount; i++ {
			url := response.Results[i].Dimensions[0]
			avgDeliveryTimeSW := int(response.Results[i].Metrics[0])
			avgOriginalTimeOrigin := int(response.Results[i].Metrics[1])
			noRequests := int(response.Results[i].Metrics[2])
			totalURLs = totalURLs + 1
			// totalDeliveryTimeSW is the total amount of page load time managed by SW
			totalDeliveryTimeSW += avgDeliveryTimeSW
			// totalOriginalTimeOrigin is the total amount of page load time shifted to SW from the origin site
			totalOriginalTimeOrigin += avgOriginalTimeOrigin
			// totalBotRequests is the total number of times the URL was requested
			totalBotRequests += noRequests
			// Write the results to a CSV file
			writeToCSV(url, avgOriginalTimeOrigin, avgDeliveryTimeSW, noRequests)
		}

		if firstPass {
			// Get the end time
			endTime := time.Now()
			// Get the execution duration
			duration := endTime.Sub(startTime)
			// Convert the duration to seconds
			seconds = duration.Seconds()
			// Reduce the number of pages yet to process by 1
			firstPass = false
		}
		totalPages = totalPages - 1

		// Calculate the time remaining
		remainingSeconds = float64(totalPages) * seconds
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(purple+"Total savings (ms):"+reset, totalOriginalTimeOrigin)
	totalOriginalTimeOriginHours := float64(totalOriginalTimeOrigin) / 1000 / 60 / 60
	fmt.Print(purple + "Total savings (hours): " + reset)
	fmt.Printf("%.2f\n", totalOriginalTimeOriginHours)
	fmt.Println(purple+"Total URLs processed: "+reset, totalURLs)
	fmt.Println()
}

// Get environment variables for the token
func getEnvVariables() (envBotifyAPIToken string) {

	// Botify API token from the env. variable getbotifyenvBotifyAPIToken
	envBotifyAPIToken = os.Getenv("envBotifyAPIToken")
	if envBotifyAPIToken == "" {
		fmt.Println()
		fmt.Println(red + "Error. getEnvVariables. envBotifyAPIToken environment variable is not set." + reset)
		fmt.Println(red + "Cannot launch swSavings analysis." + reset)
		os.Exit(0)
	}
	return envBotifyAPIToken
}

func getURLCount() (int, int) {

	// Get the dates for the BQL. Assume the last 28 days
	today, daysAgo := getDates()

	bqlGetURLCount := fmt.Sprintf(`
{
    "collections": [
        "activation.delivery"
    ],
    "periods": [
        [
            "%s",
            "%s"
        ]
    ],
    "query": {
        "dimensions": [
            "url"
        ],
        "metrics": [
        ]
    }
}
`, daysAgo, today)

	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query?count=true", organisation, project)

	// GET the HTTP request
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatal(red+"\nError. getURLCount. Cannot create request. Perhaps the provided credentials are invalid: "+reset, err)
	}

	httpBody := []byte(bqlGetURLCount)

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
	if err != nil {
		log.Fatal(red+"Error. getURLCount. Cannot create request. Perhaps the provided credentials are invalid: "+reset, err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(red+"Error.  getURLCount. Error: "+reset, err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(red+"Error.  getURLCount. Cannot read response body: "+reset, err)
	}

	var response LoadSpeed
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		log.Fatalf(red+"Error.  getURLCount. Cannot unmarshal the JSON: %v"+reset, err)
	}

	totalPages = response.Count / 2000

	return response.Count, totalPages
}

func writeToCSV(url string, originalSpeed, deliverySpeed, noRequests int) error {
	file, err := os.OpenFile("swSavings.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf(red+"Error. failed to open file: %w"+reset, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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

func checkProjectExists() bool {

	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/collections", organisation, project)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(red+"\nError. swSavings. Cannot create request:"+reset, err)
		return false
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(red+"\nError. swSavings. Cannot send request:"+reset, err)
		return false
	}
	defer res.Body.Close()
	var responseData []map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&responseData)
	if err != nil {
		fmt.Println()
		fmt.Print(red + "Error. swSavings. The specified credentials are probably invalid.")
		fmt.Println()
		return false
	}
	return true
}

// getDates returns today's date and the date 28 days ago in YYYYMMDD format. Used for BQL date range.
func getDates() (string, string) {
	today := time.Now()
	todayFormatted := today.Format("20060102")
	daysAgo := today.AddDate(0, 0, -28)
	daysAgoFormatted := daysAgo.Format("20060102")

	return todayFormatted, daysAgoFormatted
}

func swSavingsDone() {
	fmt.Println(green + "\nswSavings: Done!")
	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println()
	fmt.Println(green + "Analysis finished at " + formattedTime + reset)
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

func deleteCSV() {
	fileName := "swSavings.csv"

	_ = os.Remove(fileName)
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

	fmt.Println("\nOrganisation name:", organisation)
	fmt.Println("Project name:", project+reset)
	fmt.Println()

	displaySeparator()
}
