// folderCount: A utility that counts the number of instances of the first level folder
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// Version
var version = "v0.1"

// Specify your Botify API token here
var botify_api_token = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"

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

// Implement sorting interface for ValueCount slice
type ByCount []ValueCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func main() {

	clearScreen()

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
	fmt.Println("End point:", urlEndpoint, "\n")

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
		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit folderCount" +
			"\n")

		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using folderCount. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using folderCount. Goodbye!\n")
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
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError: Cannot send request:"+reset, err)
	}
	defer res.Body.Close()

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(red+"\nError: Cannot read request body:"+reset, err)
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	//Display an error and exit if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project")
		fmt.Println(red+"End point:", url, "\n")
		os.Exit(1)
	}
	return responseObject.Results[0].Slug
}

// Use the API to get the first 300k URLs and export them to a file
func exportURLsFromProject() {
	//Get the last analysis slug
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal("\nError creating request: "+reset, errorCheck)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, errorCheck := http.DefaultClient.Do(req)
	if errorCheck != nil {
		log.Fatal(red+"\nError: Check your network connection: "+reset, errorCheck)
	}
	defer res.Body.Close()

	responseData, errorCheck := ioutil.ReadAll(res.Body)
	if errorCheck != nil {
		log.Fatal(red+"\nError reading response body: "+reset, errorCheck)
		os.Exit(1)
	}

	var responseObject botifyResponse
	errorCheck = json.Unmarshal(responseData, &responseObject)

	if errorCheck != nil {
		log.Fatal(red+"\nError: Cannot unmarshall JSON: "+reset, errorCheck)
		os.Exit(1)
	}

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project")
		fmt.Println(red+"End point:", url, "\n")
		os.Exit(1)
	}
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
		req.Header.Add("Authorization", "token "+botify_api_token)

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
			fmt.Println(red+"End point:", urlEndpoint, "\n")
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
	fmt.Println(purple + "folderCount: Count the number of first level folders found." + reset)
	fmt.Println(purple+"Version:"+reset, version, "\n")

	// Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		// Display a block for each 10000 records scanned
		if recordCounter%10000 == 0 {
			fmt.Print("#")
		}

		// Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")

		// Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			// Extract the text between the third and fourth forward slashes
			text := parts[3]

			// Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			// Update the count for this value
			valueCounts[text]++
		}
	}

	// Display the total number of records processed
	fmt.Printf("\n\nTotal URLs processed: %d\n", totalRecords)
	fmt.Printf("\n")

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
	fmt.Println(purple + "\nfolderCount: Done\n")
	fmt.Println(green + bold + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning extract file: %v\n", err)
		os.Exit(1)
	}
}

// Display the welcome banner
func displayBanner() {
	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Print(green + `
███████╗ ██████╗ ██╗     ██████╗ ███████╗██████╗  ██████╗ ██████╗ ██╗   ██╗███╗   ██╗████████╗
██╔════╝██╔═══██╗██║     ██╔══██╗██╔════╝██╔══██╗██╔════╝██╔═══██╗██║   ██║████╗  ██║╚══██╔══╝
█████╗  ██║   ██║██║     ██║  ██║█████╗  ██████╔╝██║     ██║   ██║██║   ██║██╔██╗ ██║   ██║   
██╔══╝  ██║   ██║██║     ██║  ██║██╔══╝  ██╔══██╗██║     ██║   ██║██║   ██║██║╚██╗██║   ██║   
██║     ╚██████╔╝███████╗██████╔╝███████╗██║  ██║╚██████╗╚██████╔╝╚██████╔╝██║ ╚████║   ██║   
╚═╝      ╚═════╝ ╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝   ╚═╝
`)
	fmt.Println(purple + "folderCount: Count the number of URLs found in each first level folder.\n" + reset)
	fmt.Println(purple+"Version:"+reset, version, "\n")
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
