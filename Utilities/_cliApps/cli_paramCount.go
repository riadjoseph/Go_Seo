// paramCount: A utility that counts the number of instances of all parameter keys
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

// Version
var version = "v0.1"

// APIToken should be replaced with your own Botify API token.
var APIToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Files used to store the extracted URLs
var urlExtractFile = "siteurlsExport.tmp"

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

type ValueCount struct {
	Text  string
	Count int
}

type botifyResponse struct {
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Count    int         `json:"count"`
	Results  []struct {
		Slug string `json:"slug"`
	} `json:"results"`
	Page int `json:"page"`
	Size int `json:"size"`
}

// ByCount implements sorting interface for ValueCount slice
type ByCount []ValueCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func main() {

	displayBanner()

	// Get the project credentials if they have not been specified on the command line
	checkCredentials()

	// If the credentials have been provided on the command line use them
	if !credentialsInput {
		orgName = os.Args[1]
		projectName = os.Args[2]
	} else {
		orgName = orgNameInput
		projectName = projectNameInput
	}

	// Get the latest analysis slug
	var analysisSlug = getAnalysis(orgName, projectName)

	fmt.Println(purple + "\nExporting URLs" + reset)
	fmt.Println("Organisation name:", orgName)
	fmt.Println("Project name:", projectName)
	fmt.Println("Analysis slug:", analysisSlug)
	urlEndpoint := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/", orgName, projectName, analysisSlug)
	fmt.Println("End point:", urlEndpoint)

	// Generate and display the folder stats
	generateFolderStats(analysisSlug, urlEndpoint)

	// Clean-up. Delete the temp. file
	os.Remove(urlExtractFile)
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {
	if len(os.Args) < 3 {
		credentialsInput = true
		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit paramCount" +
			"\n")

		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using paramCount. Goodbye!")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using paramCount. Goodbye!")
			os.Exit(0)
		}
	}
}

// Get the latest analysis
func getAnalysis(orgName, projectname string) string {
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError: Cannot create GET request:"+reset, err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+APIToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError: Cannot send request:"+reset, err)
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)

	if err != nil {
		log.Fatal(red+"\nError: Cannot read request body:"+reset, err)
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	//Display an error and exit if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project")
		fmt.Println(red+"End point:", url)
		os.Exit(1)
	}
	return responseObject.Results[0].Slug
}

func generateFolderStats(analysisSlug string, urlEndpoint string) {
	// Create a file for writing
	file, err := os.Create(urlExtractFile)
	if err != nil {
		fmt.Println(red+"\nError: Cannot create output file:"+reset, err)
		os.Exit(1)
	}
	defer file.Close()

	// Initialize total count
	totalCount := 0

	// Iterate through pages 1 through 100
	for page := 1; page <= 1000; page++ {
		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", orgName, projectName, analysisSlug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+APIToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"\nError: Cannot connect to the API:"+red, err)
			os.Exit(1)
		}
		defer res.Body.Close()

		// Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println(red+"\nError: Cannot decode JSON:"+reset, err)
			os.Exit(1)
		}

		// Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project" + reset)
			fmt.Println(red+"End point:", urlEndpoint)
			os.Exit(1)
		}

		// Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println(red+"\nError: Cannot write to the output file:"+reset, err)
						os.Exit(1)
					}
					count++
					totalCount++
					if count%10 == 0 {
						fmt.Print("#") // Print "#" every 10 URLs
					}
				}
			}
		}

		// If no URLs were saved for the page, exit the loop
		if count == 0 {
			break
		}
		fmt.Printf("\nPage %d: %d URLs processed\n", page, count)
	}

	// Print total number of URLs saved
	fmt.Printf(purple+"\nTotal no. of URLs processed: %d\n"+reset, totalCount)

	// Open the file
	file, errReturn := os.Open(urlExtractFile)
	if errReturn != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Map to keep track of counts of unique values
	valueCounts := make(map[string]int)

	// Variable to keep track of the total number of records processed
	totalRecords := 0

	// Counter to track the number of records scanned
	recordCounter := 0

	// Display welcome message
	fmt.Println(purple + "paramCount: Count the number of first level folders found." + reset)
	fmt.Println(purple+"Version:"+reset, version)

	// Variable to keep track of the number of records with at least one question mark
	questionMarkRecords := 0

	// Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		// Display a block for each 1000 records scanned
		if recordCounter%10000 == 0 {
			fmt.Print("#")
		}

		// Check if the URL contains at least one question mark
		if strings.Contains(line, "?") {
			questionMarkRecords++
		}

		// Split the line into substrings using question mark as delimiter
		parts := strings.Split(line, "?")

		// Iterate over the parts after each question mark
		for _, part := range parts[1:] {
			// Find the index of the equals sign
			equalsIndex := strings.Index(part, "=")
			if equalsIndex != -1 {
				// Extract the text between the question mark and the equals sign
				text := part[:equalsIndex]

				// Trim any leading or trailing whitespace
				text = strings.TrimSpace(text)

				// Update the count for this value
				valueCounts[text]++
			}
		}
	}

	// Calculate the percentage of records with at least one question mark
	percentage := float64(questionMarkRecords) / float64(totalRecords) * 100

	// Display the total number of records processed
	fmt.Printf("\n\nTotal URLs processed: %d\n", totalRecords)
	fmt.Printf("\n")

	// Display the percentage of records with at least one question mark
	fmt.Printf("Percentage of URLs that contain Parameters: %.2f%%\n", percentage)
	fmt.Println()

	// Create a slice to hold ValueCount structs
	var sortedCounts []ValueCount

	// Populate the slice with data from the map
	for value, count := range valueCounts {
		sortedCounts = append(sortedCounts, ValueCount{value, count})
	}

	// Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	// Display the counts for each unique value
	for _, vc := range sortedCounts {
		fmt.Printf("%s (count: %d)\n", vc.Text, vc.Count)
	}

	// We're done
	fmt.Println(purple + "\nParamCount: Done")
	fmt.Println(green + bold + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)
}

// Display the welcome banner
func displayBanner() {
	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite

	// Clear the screen
	fmt.Print(clearScreen)

	fmt.Print(green + `
██████╗  █████╗ ██████╗  █████╗ ███╗   ███╗ ██████╗ ██████╗ ██╗   ██╗███╗   ██╗████████╗
██╔══██╗██╔══██╗██╔══██╗██╔══██╗████╗ ████║██╔════╝██╔═══██╗██║   ██║████╗  ██║╚══██╔══╝
██████╔╝███████║██████╔╝███████║██╔████╔██║██║     ██║   ██║██║   ██║██╔██╗ ██║   ██║   
██╔═══╝ ██╔══██║██╔══██╗██╔══██║██║╚██╔╝██║██║     ██║   ██║██║   ██║██║╚██╗██║   ██║   
██║     ██║  ██║██║  ██║██║  ██║██║ ╚═╝ ██║╚██████╗╚██████╔╝╚██████╔╝██║ ╚████║   ██║   
╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝ ╚═════╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝   ╚═╝`)
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(purple + "paramCount: Count the number of parameter keys found.\n" + reset)
}
