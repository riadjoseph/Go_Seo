// segmentifyLite. Generate the regex for a specified crawl
// See the readme for details on segments generated
// Segmentation based on the first 300k records
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
	"time"
)

// Version
var version = "v0.1"

// Specify your Botify API token here
var botify_api_token = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Colours & text formatting
var purple = "\033[0;35m"
var red = "\033[0;31m"
var green = "\033[0;32m"
var bold = "\033[1m"
var reset = "\033[0m"
var checkmark = "\u2713"

// Default input and output files
var urlExtractFile = "siteurlsExport.tmp"
var regexOutputFile = "segment.txt"

// Maximum No. of URLs to process. (300 = 300k).
var maxURLsToProcess = 300

// Percentage threshold for level 1 & level 2 folders
var thresholdPercent = 0.05

// Boolean to signal if SFCC has been detected
var sfccDetected = false

// Boolean to signal if Shopify has been detected
var shopifyDetected = false

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

// Number of forward-slashes in the URL to count in order to identify the folder level
// 4 = level 1
// 5 = level 2
var slashCountLevel1 = 4
var slashCountLevel2 = 5

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

// Define a struct to hold text value and its associated count
type FolderCount struct {
	Text  string
	Count int
}

// Implement sorting interface for FolderCount slice
type ByCount []FolderCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func main() {

	clearScreen()

	displayBanner()

	checkCredentials()

	//Generate the list of URLs
	processURLsInProject()

	// Generate the output file to store the regex
	generateRegexFile()

	//Level1 folders
	//Get the threshold. Use the level 1 slashCount
	largestFolderSize, thresholdValueL1 := levelThreshold(urlExtractFile, slashCountLevel1)
	fmt.Printf(purple + "Calculating level 1 folder threshold\n" + reset)
	fmt.Printf("Largest level 1 folder size found is %d URLs\n", largestFolderSize)
	fmt.Printf("Threshold folder size: %d\n", thresholdValueL1)

	//generate the regex
	fmt.Println(purple + "\nFirst level folders" + reset)
	segmentFolders(thresholdValueL1, slashCountLevel1)

	//Level2 folders
	//Get the threshold. Use the level 2 slashCount
	largestFolderSize, thresholdValueL2 := levelThreshold(urlExtractFile, slashCountLevel2)
	fmt.Printf(purple + "\nCalculating level 2 folder threshold\n" + reset)
	fmt.Printf("Largest level 2 folder size found is %d URLs\n", largestFolderSize)
	fmt.Printf("Threshold folder size: %d\n", thresholdValueL2)

	//Level2 folders
	fmt.Println(purple + "\nSecond level folders" + reset)
	segmentFolders(thresholdValueL2, slashCountLevel2)

	//Subdomains
	subDomains()

	//Parameter keys
	parameterKeys()

	//Parameter keys utilization
	parameterUsage()

	//No. of parameter keys
	noOfParameters()

	//No. of folders
	noOfFolders()

	// Salesforce Commerce Cloud if detected
	if sfccDetected {
		sfccURLs()
	}

	// Shopify if detected
	if shopifyDetected {
		shopifyURLs()
	}

	//Static resources
	staticResources()

	// Copy the regex to the clipboard
	copyRegexToClipboard()

	//It's done! segmentifyList has left the building
	fmt.Println(green+bold+"Your regex can be found in:", regexOutputFile+reset)
	fmt.Println(green + bold + "The regex is also in your clipboard ready to paste directly into Botify's segment editor\n" + reset)

	fmt.Println(green + bold + checkmark + reset + " First level folders" + reset)
	fmt.Println(green + bold + checkmark + reset + " Second level folders" + reset)
	fmt.Println(green + bold + checkmark + reset + " Parameter usage" + reset)
	fmt.Println(green + bold + checkmark + reset + " No. of parameters" + reset)
	fmt.Println(green + bold + checkmark + reset + " Parameter keys" + reset)
	fmt.Println(green + bold + checkmark + reset + " No. of folders" + reset)
	fmt.Println(green + bold + checkmark + reset + " Static resources" + reset)

	if sfccDetected {
		fmt.Println(green + checkmark + reset + " Salesforce Commerce Cloud" + reset)
	}
	if shopifyDetected {
		fmt.Println(green + checkmark + reset + " Shopify" + reset)
	}
	fmt.Println(purple + "\nRegex generation complete" + reset)

	// We're done
	// Clean-up. Delete the temp. file
	os.Remove(urlExtractFile)

	fmt.Println(purple + "\nsegmentifyLite: Done!\n")
	fmt.Println(green + bold + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits segmentifyLite
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit Segmentify Lite\n")

		fmt.Print(purple + "\nEnter Organization Name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using Segmentify Lite. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter Project Name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using Segmentify Lite. Goodbye!\n")
			os.Exit(0)
		}
	}
}

// Use the API to get the first 300k URLs and export them to a temp file
func processURLsInProject() {

	// If the credentials have been provided on the command line use them
	if !credentialsInput {
		orgName = os.Args[1]
		projectName = os.Args[2]
	} else {
		orgName = orgNameInput
		projectName = projectNameInput
	}

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
		os.Exit(1)
	}

	//Display the welcome message
	fmt.Println(purple + "\nProcessing URLs" + reset)

	//Create a file for writing
	file, errorCheck := os.Create(urlExtractFile)
	if errorCheck != nil {
		fmt.Println(red+"\nError creating file: "+reset, errorCheck)
		os.Exit(1)
	}
	defer file.Close()

	//Initialize total count
	totalCount := 0
	fmt.Println("Maximum No. of URLs to be processed is", maxURLsToProcess, "k")
	fmt.Println("Organisation Name:", orgName)
	fmt.Println("Project Name:", projectName)
	fmt.Println("Latest analysis Slug:", responseObject.Results[0].Slug)
	analysisSlug := responseObject.Results[0].Slug
	urlEndpoint := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/", orgName, projectName, analysisSlug)
	fmt.Println("End point:", urlEndpoint, "\n")

	//Iterate through pages 1 through to the maximum no of pages defined by maxURLsToProcess
	//Each page returns 1000 URLs
	for page := 1; page <= maxURLsToProcess; page++ {

		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", orgName, projectName, analysisSlug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+botify_api_token)

		res, errorCheck := http.DefaultClient.Do(req)
		if errorCheck != nil {
			fmt.Println(red+"\nError: Cannot connect to the API: "+reset, errorCheck)
			os.Exit(1)
		}
		defer res.Body.Close()

		//Decode JSON response
		var response map[string]interface{}
		if errorCheck := json.NewDecoder(res.Body).Decode(&response); errorCheck != nil {
			fmt.Println(red+"\nError: Cannot decode JSON: "+reset, errorCheck)
			os.Exit(1)
		}

		//Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project")
			os.Exit(1)
		}

		//Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					// Check if SFCC is used. This bool us used to deterline if the SFCC regex is generated
					if strings.Contains(url, "/demandware/") {
						sfccDetected = true
					}
					// Check if Shopify is used. This bool us used to deterline if the Shopify regex is generated
					if strings.Contains(url, "/collections/") && strings.Contains(url, "/products/") {
						shopifyDetected = true
					}
					if _, errorCheck := file.WriteString(url + "\n"); errorCheck != nil {
						fmt.Println(red+"\nError: Cannot write to file: "+reset, errorCheck)
						os.Exit(1)
					}
					count++
					totalCount++
					if count%10 == 0 {
						fmt.Print("#") //Print "#" every 10 URLs. Used as a progress indicator
					}
				}
			}
		}

		//If there are no more URLS process exit the function
		if count == 0 {
			//Print total number of URLs saved
			fmt.Printf("\nTotal URLs processed: %d\n", totalCount)
			if sfccDetected {
				fmt.Printf(bold + "\nNote: Salesforce Commerce Cloud has been detected. Regex will be generated\n" + reset)
			}
			if shopifyDetected {
				fmt.Printf(bold + "\nNote: Shopify has been detected. Regex will be generated\n" + reset)
			}
			fmt.Println(purple + "\nURLs processed. Generating regex...\n" + reset)
			// Check if SFCC is used. This bool us used to deterline if SFCC regex is generated
			break
		}

		//Max. number of URLs (200k) has been reached
		if totalCount > 190000 {
			fmt.Printf("\n\nLimit of %d URLs reached. Generating regex...\n\n", totalCount)
			break
		}

		fmt.Printf("\nPage %d: %d URLs processed\n", page, count)
	}

}

func generateRegexFile() {

	//Always create the file.
	outputFile, errorCheck := os.Create(regexOutputFile)
	if errorCheck != nil {
		fmt.Printf(red+"\nsegment1stLevel. Error: Cannot create output file: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	// Get the user's local time zone for the header
	userLocation, errorCheck := time.LoadLocation("") // Load the default local time zone
	if errorCheck != nil {
		fmt.Println("\nError loading user's location:", errorCheck)
		return
	}
	// Get the current date and time in the user's local time zone
	currentTime := time.Now().In(userLocation)

	_, errorCheck = writer.WriteString(fmt.Sprintf("# Regex made with love using Go_SEO/segmentifyLite %s\n", version))

	if errorCheck != nil {
		fmt.Printf(red+"\nsegment1stLevel. Error: Cannot write header to output file: %v\n"+reset, errorCheck)
		os.Exit(1)
	}

	writer.WriteString(fmt.Sprintf("# Organisation Name: %s\n", orgName))
	writer.WriteString(fmt.Sprintf("# Project Name: %s\n", projectName))
	writer.WriteString(fmt.Sprintf("# Generated %s", currentTime.Format(time.RFC1123)))

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\ngenerateRegexFile. Error: Cannot flush writer: %v\n", errorCheck)
		os.Exit(1)
	}
}

func segmentFolders(thresholdValue int, slashCount int) {

	//Open the input file
	file, errorCheck := os.Open(urlExtractFile)
	if errorCheck != nil {
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	FolderCounts := make(map[string]int)

	//Variable to keep track of the total number of records processed
	totalRecords := 0

	//Counter to track the number of records scanned
	recordCounter := 0

	//Counter to track the number of folders excluded from the regex
	noFoldersExcluded := 0

	//Display welcome message
	fmt.Printf("Folders with less than %d URLs will be excluded\n", thresholdValue)

	//Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		//Display a block for each 1000 records scanned
		if recordCounter%1000 == 0 {
			fmt.Print("#")
		}

		//Check if the line contains a quotation mark, if yes, skip to the next line
		if strings.Contains(line, "\"") {
			continue
		}

		//Split the line into substrings using a forward-slash as delimiter
		// slashCount = 4 for Level 1 folders
		// slashCount = 5 for Level 2 folders
		parts := strings.Split(line, "/")

		if len(parts) >= slashCount {
			//Extract the text
			text := strings.Join(parts[:slashCount], "/")

			//Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			//Update the count for this value if it's not empty
			if text != "" {
				FolderCounts[text]++
			}
		}
	}

	fmt.Printf("\n")

	//Create a slice to hold FolderCount structs
	var sortedCounts []FolderCount

	//Populate the slice with data from the map
	for folderName, count := range FolderCounts {
		if count > thresholdValue {
			sortedCounts = append(sortedCounts, FolderCount{folderName, count})
		} else {
			// Count the number of folders excluded
			noFoldersExcluded++
		}
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Display the counts for each unique value
	for _, folderValueCount := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", folderValueCount.Text, folderValueCount.Count)
	}

	fmt.Printf("\nNo. of folders excluded %d\n", noFoldersExcluded)

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the segment name
	// SlashCount = 4 signals level 1 folders
	// SlashCount = 5 signals level 2 folders
	if slashCount == 4 {
		writer.WriteString(fmt.Sprintf("\n\n[segment:sl_level1_folders]\n@Home\npath /\n\n"))
	}

	if slashCount == 5 {
		writer.WriteString(fmt.Sprintf("\n\n[segment:sl_level2_folders]\n@Home\npath /\n\n"))
	}

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		if folderValueCount.Text != "" {
			//Extract the text between the third and fourth forward-slashes
			parts := strings.SplitN(folderValueCount.Text, "/", 4)
			if len(parts) >= 4 && parts[3] != "" {
				folderLabel := parts[3] //Extract the text between the third and fourth forward-slashes
				_, errorCheck := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, folderValueCount.Text))
				if errorCheck != nil {
					fmt.Printf(red+"\nsegmentFolders. Error: Cannot write to output file: %v\n"+reset, errorCheck)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	writer.WriteString("@~Other\npath /*\n# ----End of level2Folders Segment----\n")

	//Insert the number of URLs found in each folder as comments
	writer.WriteString("\n# ----Folder URL analysis----\n")
	for _, folderValueCount := range sortedCounts {
		writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", folderValueCount.Text, folderValueCount.Count))
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nsegmentFolders. Error: Cannot flush writer: %v\n", errorCheck)
		os.Exit(1)
	}
	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")
}

// Regex for subdomains
func subDomains() {

	//Open the input file
	file, errorCheck := os.Open(urlExtractFile)
	if errorCheck != nil {
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	FolderCounts := make(map[string]int)

	//Variable to keep track of the total number of records processed
	totalRecords := 0

	//Counter to track the number of records scanned
	recordCounter := 0

	//Display welcome message
	fmt.Println(purple + "\nSubdomains" + reset)

	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		//Display a block for each 1000 records scanned
		if recordCounter%1000 == 0 {
			fmt.Print("#")
		}

		//Check if the line contains a quotation mark, if yes, skip to the next line
		if strings.Contains(line, "\"") {
			continue
		}

		//Split the line into substrings using a forward-slash as delimiter
		parts := strings.Split(line, "/")
		//Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			//Extract the text between the third and fourth forward-slashes
			text := strings.Join(parts[:3], "/")

			//Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			//Update the count for this value if it's not empty
			if text != "" {
				//Update the count for this value if it's not empty
				FolderCounts[text]++
			}
		}
	}

	//Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	fmt.Printf("\n")

	//Create a slice to hold FolderCount structs
	var sortedCounts []FolderCount

	//Populate the slice with data from the map
	for folderName, count := range FolderCounts {
		sortedCounts = append(sortedCounts, FolderCount{folderName, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Display the counts for each unique value
	for _, folderValueCount := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", folderValueCount.Text, folderValueCount.Count)
	}

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	writer.WriteString(fmt.Sprintf("\n\n[segment:sl_subdomains]\n@Home\npath /\n\n"))

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		if folderValueCount.Text != "" {
			//Extract the text between the third and fourth forward-slashes
			parts := strings.SplitN(folderValueCount.Text, "/", 4)
			if len(parts) >= 3 && parts[2] != "" {
				folderLabel := parts[2] //Extract the text between the third and fourth forward-slashes
				writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, folderValueCount.Text))
				if errorCheck != nil {
					fmt.Printf(red+"\nsubDomains. Error: Cannot write to output file: %v\n"+reset, errorCheck)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	writer.WriteString("@~Other\npath /*\n# ----End of subDomains Segment----\n")

	//Insert the number of URLs found in each folder as comments
	writer.WriteString("\n# ----subDomains Folder URL analysis----\n")
	for _, folderValueCount := range sortedCounts {
		_, errorCheck := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", folderValueCount.Text, folderValueCount.Count))
		if errorCheck != nil {
			fmt.Printf(red+"\nsubDomains. Error: Cannot write to output file: %v\n"+reset, errorCheck)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nsubDomains. Error: Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")
}

// Regex to identify which parameter keys are used
func parameterKeys() {

	//Open the input file
	file, errorCheck := os.Open(urlExtractFile)
	if errorCheck != nil {
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	FolderCounts := make(map[string]int)

	//Variable to keep track of the total number of records processed
	totalRecords := 0

	//Counter to track the number of records scanned
	recordCounter := 0

	//Display welcome message
	fmt.Println(purple + "\nParameter keys" + reset)

	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		//Display a block for each 1000 records scanned
		if recordCounter%1000 == 0 {
			fmt.Print("#")
		}

		//Check if the line contains a quotation mark, if yes, skip to the next line
		if strings.Contains(line, "\"") {
			continue
		}

		//Split the line into substrings using question mark as delimiter
		parts := strings.Split(line, "?")

		//Iterate over the parts after each question mark
		for _, part := range parts[1:] {
			//Find the index of the equals sign
			equalsIndex := strings.Index(part, "=")
			if equalsIndex != -1 {
				//Extract the text between the question mark and the equals sign
				text := part[:equalsIndex]

				//Trim any leading or trailing whitespace
				text = strings.TrimSpace(text)

				//Update the count for this value
				FolderCounts[text]++
			}
		}
	}

	//Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	fmt.Printf("\n")

	//Create a slice to hold FolderCount structs
	var sortedCounts []FolderCount

	//Populate the slice with data from the map
	for folderName, count := range FolderCounts {
		sortedCounts = append(sortedCounts, FolderCount{folderName, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Display the counts for each unique value
	for _, folderValueCount := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", folderValueCount.Text, folderValueCount.Count)
	}

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	writer.WriteString(fmt.Sprintf("\n\n[segment:sl_parameter_keys]\n"))

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		_, errorCheck := writer.WriteString(fmt.Sprintf("@%s\nquery *%s=*\n\n", folderValueCount.Text, folderValueCount.Text))
		if errorCheck != nil {
			fmt.Printf(red+"\nparameterKeys. Error: Cannot write to output file: %v\n"+reset, errorCheck)
			os.Exit(1)
		}
	}

	//Write the footer lines
	writer.WriteString("@~Other\npath /*\n# ----End of parameterKeys Segment----\n")

	//Insert the number of URLs found in each folder as comments
	writer.WriteString("\n# ----parameterKeys URL analysis----\n")
	for _, folderValueCount := range sortedCounts {
		_, errorCheck := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", folderValueCount.Text, folderValueCount.Count))
		if errorCheck != nil {
			fmt.Printf(red+"\nparameterKeys. Error: Cannot write to output file: %v\n"+reset, errorCheck)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nparameterKeys. Error: Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")
}

// Regex to identify of a parameter key is used in the URL
func parameterUsage() {

	//URLs containing parameters
	paramaterUsageRegex := `

[segment:sl_parameter_usage]
@Parameters
query *=*

@Clean
path /*

# ----End of sl_parameter_usage----`

	//Parameter usage message
	fmt.Println(purple + "\nParameter usage" + reset)
	errParamaterUsage := insertStaticRegex(paramaterUsageRegex)
	if errParamaterUsage != nil {
		panic(errParamaterUsage)
	}

	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")
}

// Regex to count the number of parameters in the URL
func noOfParameters() {

	//Number of paramaters
	paramaterNoRegex := `


[segment:sl_no_of_parameters]
@Home
path /

@5_Parameters
query rx:=(.)+=(.)+=(.)+(.)+(.)+

@4_Parameters
query rx:=(.)+=(.)+=(.)+(.)+

@3_Parameters
query rx:=(.)+=(.)+=(.)+

@2_Parameters
query rx:=(.)+=(.)+

@1_Parameter
query rx:=(.)+

@~Other
path /*

# ----End of sl_no_of_parameters----`

	//No. of parameters message
	fmt.Println(purple + "Number of parameters" + reset)
	errParamaterNoRegex := insertStaticRegex(paramaterNoRegex)
	if errParamaterNoRegex != nil {
		panic(errParamaterNoRegex)
	}

	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")

}

// Regex to count the number of folders in the URL
func noOfFolders() {

	//Number of folders
	folderNoRegex := `


[segment:sl_no_of_folders]
@Home
path /

@Folders/5
path rx:^/[^/]+/[^/]+/[^/]+/[^/]+/[^/]+

@Folders/4
path rx:^/[^/]+/[^/]+/[^/]+/[^/]+

@Folders/3
path rx:^/[^/]+/[^/]+/[^/]+

@Folders/2
path rx:^/[^/]+/[^/]+

@Folders/1
path rx:^/[^/]+

@~Other
path /*

# ----End of sl_no_of_folders----`

	//No. of folders message
	fmt.Println(purple + "Number of folders" + reset)
	errFolderNoRegex := insertStaticRegex(folderNoRegex)
	if errFolderNoRegex != nil {
		panic(errFolderNoRegex)
	}

	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")

}

// SFCC Regex
func sfccURLs() {

	//SFCC
	sfccURLs := `


[segment:sl_sfcc]
@Home
path /

@SFCC
path */demandware*

@~Other
path /*

# ----End of sl_sfcc----`

	// SFCC message
	fmt.Println(purple + "Salesforce Commerce Cloud (Demandware)" + reset)
	errSfccURLs := insertStaticRegex(sfccURLs)
	if errSfccURLs != nil {
		panic(errSfccURLs)
	}

	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")

}

// Shopify Regex
func shopifyURLs() {

	// Shopify
	shopifyURLs := `


[segment:sl_shopify]
@Home
path /

@PDP/Products/Variants
path */products/*
URL *variant=*

@PDP/Products
path */products/*

@PLP/Collections
path */collections/*

@Pages
path */pages/*

@~Other
path /*
# ----End of sl_shopify----`

	// Shopify message
	fmt.Println(purple + "Shopify" + reset)
	errShopify := insertStaticRegex(shopifyURLs)
	if errShopify != nil {
		panic(errShopify)
	}

	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")

}

// Static resources
func staticResources() {

	// Static resources
	staticResources := `


[segment:s_Static_Resources]  
@true  
or (  
path *.bmp
path *.css
path *.doc
path *.gif
path *.ief
path *.jpe
path *.jpeg
path *.jpg
path *.js
path *.m1v
path *.mov
path *.mp2
path *.mp3
path *.mp4
path *.mpa
path *.mpe
path *.mpeg
path *.mpg
path *.pbm
path *.pdf
path *.pgm
path *.png
path *.pnm
path *.ppm
path *.pps
path *.ppt
path *.ps
path *.qt
path *.ras
path *.rgb
path *.swf
path *.tif
path *.tiff
path *.tsv
path *.txt
path *.vcf
path *.wav
path *.xbm
path *.xls
path *.xml
path *.xpdl
path *.xpm
path *.xwd
path */api/*
)

@~Other
path /*

# ----End of sl_static_resources----`

	// Static resources message
	fmt.Println(purple + "Static resources" + reset)
	errStaticResources := insertStaticRegex(staticResources)
	if errStaticResources != nil {
		panic(errStaticResources)
	}

	//Finished
	fmt.Println("Done!", green+checkmark+reset, "\n")
}

// Get the folder size threshold for level 1 & 2 folders
func levelThreshold(inputFilename string, slashCount int) (largestValueSize, fivePercentValue int) {
	// Open the input file
	file, errorCheck := os.Open(inputFilename)
	if errorCheck != nil {
		fmt.Printf("\nError: Cannot open input file: %v\n", errorCheck)
		os.Exit(1)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Map to keep track of counts of unique values
	FolderCounts := make(map[string]int)

	// Variable to keep track of the total number of records processed
	totalRecords := 0

	// Counter to track the number of records scanned
	recordCounter := 0

	// Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		// Check if the line contains a quotation mark, if yes, skip to the next line
		if strings.Contains(line, "\"") {
			continue
		}

		// Split the line into substrings using a forward-slash as delimiter
		parts := strings.Split(line, "/")

		// Check if there are at least slashCount parts in the URL
		// See slashCount variable declaration comments for more information
		if len(parts) >= slashCount {
			// Extract the text between the forward-slashes
			text := strings.Join(parts[:slashCount], "/")

			// Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			// Update the count for this value if it's not empty
			if text != "" {
				FolderCounts[text]++
			}
		}
	}

	// Subtract 2 in order to account for the two header records which are defaults in URL extract
	totalRecords -= 2

	// Create a slice to hold FolderCount structs
	var sortedCounts []FolderCount

	// Populate the slice with data from the map
	for folderName, count := range FolderCounts {
		sortedCounts = append(sortedCounts, FolderCount{folderName, count})
	}

	// Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	// Get the largest value size
	if len(sortedCounts) > 0 {
		largestValueSize = sortedCounts[0].Count
	}

	// Calculate 5% of the largest value
	fivePercentValue = int(float64(largestValueSize) * thresholdPercent)

	return largestValueSize, fivePercentValue
}

// Write the static Regex to the segments file
func insertStaticRegex(regexText string) error {

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	_, errorCheck = writer.WriteString(regexText)
	if errorCheck != nil {
		fmt.Printf(red+"\ninsertStaticRegex. Error: Cannot write to outputfile: %v\n"+reset, errorCheck)
		panic(errorCheck)
		os.Exit(1)
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\ninsertStaticRegex. Error: Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}

	return errorCheck
}

// Copy Regex to the clipboard
func copyRegexToClipboard() {
	content, err := ioutil.ReadFile(regexOutputFile)
	if err != nil {
		panic(err)
	}

	// Copy the content to the clipboard using pbcopy (macOS) or type to clip (Windows)
	var copyCmd string

	switch runtime.GOOS {
	case "windows":
		copyCmd = "type segment.txt | clip"
	default:
		copyCmd = "pbcopy"
	}

	cmd := exec.Command(copyCmd)
	cmd.Stdin = strings.NewReader(string(content))
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

// Display the welcome banner
func displayBanner() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Println(green + `

███████╗███████╗ ██████╗ ███╗   ███╗███████╗███╗   ██╗████████╗██╗███████╗██╗   ██╗██╗     ██╗████████╗███████╗
██╔════╝██╔════╝██╔════╝ ████╗ ████║██╔════╝████╗  ██║╚══██╔══╝██║██╔════╝╚██╗ ██╔╝██║     ██║╚══██╔══╝██╔════╝
███████╗█████╗  ██║  ███╗██╔████╔██║█████╗  ██╔██╗ ██║   ██║   ██║█████╗   ╚████╔╝ ██║     ██║   ██║   █████╗
╚════██║██╔══╝  ██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║   ██║██╔══╝    ╚██╔╝  ██║     ██║   ██║   ██╔══╝
███████║███████╗╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║   ██║██║        ██║   ███████╗██║   ██║   ███████╗
╚══════╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝   ╚═╝   ╚══════╝
`)

	//Display welcome message
	fmt.Println(purple + "\nsegmentifyLite: Fast segmentation regex generation\n" + reset)
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
