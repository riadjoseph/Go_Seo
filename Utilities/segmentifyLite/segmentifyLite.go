// segmentifyLite. Generate the regex for a specified crawl
// See the readme for details on segments generated
// Segmentation based on the first 300k records
// Written by Jason Vicinanza

package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"math/rand"
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
var botifyApiToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Colours & text formatting
var purple = "\033[0;35m"
var red = "\033[0;31m"
var green = "\033[0;32m"
var reset = "\033[0m"
var lineSeparator = "█" + strings.Repeat("█", 129)

// Default input and output files
var urlExtractFile = "siteurlsExport.tmp"
var regexOutputFile = "segment.txt"

// Maximum No. of URLs to process. (300 = 300k).
var maxURLsToProcess = 300000

// Percentage threshold for level 1 & level 2 folders
var thresholdPercent = 0.05

// Boolean to signal if SFCC has been detected
var sfccDetected = false

// Boolean to signal if Shopify has been detected
var shopifyDetected = false

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Number of forward-slashes in the URL to count in order to identify the folder level
// 4 = level 1
// 5 = level 2
var slashCountLevel1 = 4
var slashCountLevel2 = 5

// Host name and port the web server runs on
var hostname string
var port string
var fullHost = hostname + ":" + port

type botifyResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Slug string `json:"slug"`
	} `json:"results"`
}

// FolderCount defines a struct to hold text value and its associated count
type FolderCount struct {
	Text  string
	Count int
}

// byCount implements a sorting interface for FolderCount slice
type ByCount []FolderCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func main() {

	clearScreen()

	displayBanner()

	// Serve static files from the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Define a handler function for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the form data from the request
		err := r.ParseForm()
		if err != nil {
			// Handle the error appropriately, e.g., log it and return an error response
			fmt.Println(red+"Error. Cannot parse form:"+reset, err)
			return
		}
		orgName = r.Form.Get("organization")
		projectName = r.Form.Get("project")

		// Generate a session ID used for grouping log entries
		var sessionID string

		sessionID, err = generateLogSessionID(16)
		if err != nil {
			log.Fatalf(red+"Error. writeLog. Failed generating session ID: %s"+reset, err)
		}

		// Process URLs
		processURLsInProject(sessionID)

		// Write to the log
		writeLog(sessionID, orgName, projectName, "URLs acquired")

		// Generate the output file to store the regex
		generateRegexFile()

		//Level 1 and 2 folders
		level1andFolders()

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
			// Write to the log
			writeLog(sessionID, orgName, projectName, "SFCC detected")
			sfccURLs()
		}

		// Shopify if detected
		if shopifyDetected {
			// Write to the log
			writeLog(sessionID, orgName, projectName, "Shopify detected")
			shopifyURLs()
		}

		//Static resources
		staticResources()

		// Write to the log
		writeLog(sessionID, orgName, projectName, "Regex generated successfully")

		// Generate the HTML used to present the regex
		generateSegmentationRegex()

		// Display results and clean up
		cleanUp()

		// Respond to the client with a success message or redirect to another page
		http.Redirect(w, r, "go_seo_segmentifyLite.html", http.StatusFound)
	})

	// Start the HTTP server
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(red+"Error. main. Cannot start HTTP server.:"+red, err)
		os.Exit(1)
	}
}

// Use the API to get the first 300k URLs and export them to a temp file
func processURLsInProject(sessionID string) {

	//Get the last analysis slug
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal("\nError creating request: "+reset, errorCheck)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botifyApiToken)

	res, errorCheck := http.DefaultClient.Do(req)
	if errorCheck != nil {
		log.Fatal(red+"\nError. processURLsInProject. Check your network connection: "+reset, errorCheck)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(red+"Error. Closing (1):"+reset, err)
		}
	}()

	responseData, errorCheck := io.ReadAll(res.Body)
	if errorCheck != nil {
		log.Fatal(red+"\nError. generateURLsInProject. Cannot read response body: "+reset, errorCheck)
	}

	var responseObject botifyResponse
	errorCheck = json.Unmarshal(responseData, &responseObject)

	if errorCheck != nil {
		log.Fatal(red+"\nError. generateURLsInProject. Cannot unmarshall JSON: "+reset, errorCheck)
	}

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError. processURLsInProject. Invalid credentials or no crawls found in the project" + reset)
		return
	}

	//Display the welcome message
	fmt.Println(purple + "\nRequest received" + reset)

	//Create a file for writing
	file, errorCheck := os.Create(urlExtractFile)
	if errorCheck != nil {
		fmt.Println(red+"\nError creating file: "+reset, errorCheck)
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (2):"+reset, err)
		}
	}()

	//Initialize total count
	totalCount := 0
	fmt.Println(sessionID+": Organisation name:", orgName)
	fmt.Println(sessionID+": Project name:", projectName)
	fmt.Println(sessionID+": Latest analysis slug:", responseObject.Results[0].Slug)
	println()

	analysisSlug := responseObject.Results[0].Slug

	//Iterate through pages 1 through to the maximum no of pages defined by maxURLsToProcess
	//Each page returns 1000 URLs
	for page := 1; page <= maxURLsToProcess; page++ {

		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", orgName, projectName, analysisSlug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+botifyApiToken)

		res, errorCheck := http.DefaultClient.Do(req)
		if errorCheck != nil {
			fmt.Println(red+"\nError. processURLsInProject. Cannot connect to the API: "+reset, errorCheck)
			os.Exit(1)
		}

		//Decode JSON response
		var response map[string]interface{}
		if errorCheck := json.NewDecoder(res.Body).Decode(&response); errorCheck != nil {
			fmt.Println(red+"\nError. processURLsInProject. Cannot decode JSON: "+reset, errorCheck)
			os.Exit(1)
		}

		//Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "\nError. processURLsInProject. Invalid credentials or no crawls found in the project" + reset)
			os.Exit(1)
		}

		//Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					// Check if SFCC is used. This bool us used to determine if the SFCC regex is generated
					if strings.Contains(url, "/demandware/") {
						sfccDetected = true
					}
					// Check if Shopify is used. This bool us used to determine if the Shopify regex is generated
					if strings.Contains(url, "/collections/") && strings.Contains(url, "/products/") {
						shopifyDetected = true
					}
					if _, errorCheck := file.WriteString(url + "\n"); errorCheck != nil {
						fmt.Println(red+"\nError. processURLsInProject. Cannot write to file: "+reset, errorCheck)
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
			break
		}

		//Max. number of URLs has been reached
		if totalCount > maxURLsToProcess {
			fmt.Printf("\n\nLimit of %d URLs reached. Generating regex...\n\n", totalCount)
			break
		}

		fmt.Printf(" Page %d: %d URLs processed\n", page, count)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(red+"Error. Closing (3):"+reset, err)
		}
	}()
}

// Generate regex for level 1 and 2 folders
func level1andFolders() {

	//Level1 folders
	//Get the threshold. Use the level 1 slashCount
	_, thresholdValueL1 := levelThreshold(urlExtractFile, slashCountLevel1)

	//generate the regex
	segmentFolders(thresholdValueL1, slashCountLevel1)

	//Level2 folders
	//Get the threshold. Use the level 2 slashCount
	_, thresholdValueL2 := levelThreshold(urlExtractFile, slashCountLevel2)

	//Level2 folders
	segmentFolders(thresholdValueL2, slashCountLevel2)
}

func generateRegexFile() {

	//Always create the file.
	outputFile, errorCheck := os.Create(regexOutputFile)
	if errorCheck != nil {
		fmt.Printf(red+"\nError. generateRegexFile. Cannot create output file: %v\n"+reset, errorCheck)
		os.Exit(1)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. Closing (4):"+reset, err)
		}
	}()

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

	_, errorCheck = writer.WriteString(fmt.Sprintf("# Regex made with love using segmentifyLite %s\n", version))

	if errorCheck != nil {
		fmt.Printf(red+"\nError. generateRegexFile. Cannot write header to output file: %v\n"+reset, errorCheck)
		os.Exit(1)
	}

	_, err := writer.WriteString(fmt.Sprintf("# Organisation name: %s\n", orgName))
	if err != nil {
		errMsg := fmt.Errorf(red+"Error. Cannot write organisation name in Regex file: %w", err)
		println(errMsg)
	}
	_, err = writer.WriteString(fmt.Sprintf("# Project name: %s\n", projectName))
	if err != nil {
		errMsg := fmt.Errorf(red+"Error. Cannot write project name in Regex file: %w", err)
		println(errMsg)
	}
	_, err = writer.WriteString(fmt.Sprintf("# Generated %s", currentTime.Format(time.RFC1123)))
	if err != nil {
		errMsg := fmt.Errorf(red+"Error. Cannot write generate date/time name in Regex file: %w", err)
		println(errMsg)
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nError. generateRegexFile. Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
}

func segmentFolders(thresholdValue int, slashCount int) {

	//Open the input file
	file, errorCheck := os.Open(urlExtractFile)
	if errorCheck != nil {
		os.Exit(1)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (5):"+reset, err)
		}
	}()

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

	//Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

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

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. Closing (6):"+reset, err)
		}
	}()

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
					fmt.Printf(red+"\nError. segmentFolders. Cannot write to output file: %v\n"+reset, errorCheck)
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
		fmt.Printf(red+"\nError. segmentFolders. Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
}

// Regex for subdomains
func subDomains() {

	//Open the input file
	file, errorCheck := os.Open(urlExtractFile)
	if errorCheck != nil {
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (7):"+reset, err)
		}
	}()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	FolderCounts := make(map[string]int)

	//Variable to keep track of the total number of records processed
	totalRecords := 0

	//Counter to track the number of records scanned
	recordCounter := 0

	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

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

	//Create a slice to hold FolderCount structs
	var sortedCounts []FolderCount

	//Populate the slice with data from the map
	for folderName, count := range FolderCounts {
		sortedCounts = append(sortedCounts, FolderCount{folderName, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. Closing (8):"+reset, err)
		}
	}()

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
					fmt.Printf(red+"\nError. subDomains. Cannot write to output file: %v\n"+reset, errorCheck)
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
			fmt.Printf(red+"\nError. subDomains. Cannot write to output file: %v\n"+reset, errorCheck)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nError. subDomains. Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
}

// Regex to identify which parameter keys are used
func parameterKeys() {

	//Open the input file
	file, errorCheck := os.Open(urlExtractFile)
	if errorCheck != nil {
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (9):"+reset, err)
		}
	}()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	FolderCounts := make(map[string]int)

	//Variable to keep track of the total number of records processed
	totalRecords := 0

	//Counter to track the number of records scanned
	recordCounter := 0

	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

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

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. Closing (10):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	writer.WriteString(fmt.Sprintf("\n\n[segment:sl_parameter_keys]\n"))

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		_, errorCheck := writer.WriteString(fmt.Sprintf("@%s\nquery *%s=*\n\n", folderValueCount.Text, folderValueCount.Text))
		if errorCheck != nil {
			fmt.Printf(red+"\nError. parameterKeys. Cannot write to output file: %v\n"+reset, errorCheck)
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
			fmt.Printf(red+"\nError. parameterKeys. Cannot write to output file: %v\n"+reset, errorCheck)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nError. parameterKeys. Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}
}

// Regex to identify of a parameter key is used in the URL
func parameterUsage() {

	//URLs containing parameters
	parameterUsageRegex := `

[segment:sl_parameter_usage]
@Parameters
query *=*

@Clean
path /*

# ----End of sl_parameter_usage----`

	errParameterUsage := insertStaticRegex(parameterUsageRegex)
	if errParameterUsage != nil {
		panic(errParameterUsage)
	}

}

// Regex to count the number of parameters in the URL
func noOfParameters() {

	//Number of parameters
	parameterNoRegex := `


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

	errParamaterNoRegex := insertStaticRegex(parameterNoRegex)
	if errParamaterNoRegex != nil {
		panic(errParamaterNoRegex)
	}

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
	errFolderNoRegex := insertStaticRegex(folderNoRegex)
	if errFolderNoRegex != nil {
		panic(errFolderNoRegex)
	}

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

	errStaticResources := insertStaticRegex(staticResources)
	if errStaticResources != nil {
		panic(errStaticResources)
	}

}

// Get the folder size threshold for level 1 & 2 folders
func levelThreshold(inputFilename string, slashCount int) (largestValueSize, fivePercentValue int) {
	// Open the input file
	file, errorCheck := os.Open(inputFilename)
	if errorCheck != nil {
		fmt.Printf(red+"\nError. levelThreshhold. Cannot open input file: %v\n"+reset, errorCheck)
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (11):"+reset, err)
		}
	}()

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

// Display the results and cleanup
func cleanUp() {

	// Clean-up. Delete the temp. file
	os.Remove(urlExtractFile)
	os.Remove("segment.txt")

	// We're done
	fmt.Println(lineSeparator)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(purple + "\nsegmentifyLite: Done at " + formattedTime)
	fmt.Printf("\nOrganization: %s, Project: %s\n"+reset, orgName, projectName)

	// Make a tidy display
	fmt.Println()
	fmt.Println(lineSeparator)

	return
}

// Write the static Regex to the segments file
func insertStaticRegex(regexText string) error {

	//Open the file in append mode, create if it doesn't exist
	outputFile, errorCheck := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if errorCheck != nil {
		panic(errorCheck)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. Closing (12):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	_, errorCheck = writer.WriteString(regexText)
	if errorCheck != nil {
		fmt.Printf(red+"\nError. insertStaticRegex. Cannot write to outputfile: %v\n"+reset, errorCheck)
		panic(errorCheck)
	}

	//Flush the writer to ensure all data is written to the file
	errorCheck = writer.Flush()
	if errorCheck != nil {
		fmt.Printf(red+"\nError. insertStaticRegex. Cannot flush writer: %v\n"+reset, errorCheck)
		os.Exit(1)
	}

	return errorCheck
}

func writeLog(sessionID, orgName, projectName, statusDescription string) {
	// Define log file name
	fileName := "_seoSegmentifyLitelogfile.log"

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

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (13):"+reset, err)
		}
	}()

	// Get current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Construct log record
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%s\n",
		sessionID, currentTime, orgName, projectName, statusDescription)

	// If the file doesn't exist, write header first
	if !fileExists {
		header := "SessionID,Date,Organisation,Project,Status\n"
		if _, err := file.WriteString(header); err != nil {
			log.Fatalf(red+"Error. writeLog. Failed to write log header: %s"+reset, err)
		}
	}

	// Write log record to file
	if _, err := file.WriteString(logRecord); err != nil {
		log.Fatalf(red+"Error. writeLog. Cannot write to log file: %s"+reset, err)
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

// Generate the HTML pages used to present the segmentation regex
func generateSegmentationRegex() {

	// Using these two variables to replace width values in the HTML below because string interpolation confuses the percent signs as variables
	width50 := "50%"
	width100 := "100%"

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
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
        }
        .banner {
            background-color: DeepSkyBlue;
            color: white;
            text-align: center;
            padding: 15px 0;
            position: absolute;
            top: 0;
			/* Width set to 100percent */
            width: %s;
        }
        .banner.top {
            font-size: 24px;
        }
        .container {
            display: flex;
            justify-content: center;
            align-items: center;
			/* Width set to 100percent */
            width: %s;
			/* Width set to 100percent */
            height: %s;
            background-color: Khaki;
            color: white;
        }
        iframe {
            width: 80vw;
            height: 80vh;
            border: 2px solid LightGray;
            border-radius: 10px;
        }
        .no-border iframe {
            border: none;
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
    <span style="font-size: 20px;">segmentifyLite. Segment regex generation</span>
</header>

<!-- Back Button to create a new dashboard -->
<button class="back-button" onclick="goHome()">New segmentation</button>

<script>
    function goHome() {
        window.open('http://%s/', '_blank');
    }
</script>

<!-- Sections with Iframes -->
<section class="container row no-border">
    <iframe src="seo_segmentHTML.html" title="Segmentation regex"></iframe>
</section>

</body>
</html>
`, width100, width50, width100, fullHost)

	// Generate the URL to link to the segment editor in the project
	projectURL := "https://app.botify.com/" + orgName + "/" + projectName + "/segmentation"

	htmlContent += fmt.Sprintf("<div style='text-align: center;'>\n")
	htmlContent += fmt.Sprintf("<h2 style='color: deepskyblue;'>Segmentation regex generation is complete</h2>\n")
	htmlContent += fmt.Sprintf("<h3 style='color: dimgray; padding-left: 20px; padding-right: 20px;'>The regex has been copied into the clipboard ready for pasting directly into your Botify project.</h3>\n")
	htmlContent += fmt.Sprintf("<h4 style='color: dimgray;'><a href='%s' target='_blank'>Click here to open the segment editor for %s</a></h4>\n", projectURL, orgName)

	htmlContent += fmt.Sprintf("</div>\n")

	// Save the HTML to a file
	saveHTML(htmlContent, "./go_seo_segmentifyLite.html")

	// Generate the HTML containing the segmentation regex
	generateSegmentHTML()

	// Copy the regex to the clipboard
	copyRegexToClipboard()

}

// Copy Regex to the clipboard
func generateSegmentHTML() {
	// Read the contents of segment.txt
	content, err := os.ReadFile("segment.txt")

	if err != nil {
		log.Fatalf(red+"Error. generateSegmentationRegex. Failed to read segment.txt: %v"+reset, err)
	}

	// HTML template with the content
	htmlContent := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Segment Content</title>
</head>
<body>
    <pre>%s</pre>
</body>
</html>`

	// Create the HTML file
	file, err := os.Create("seo_segmentHTML.html")
	if err != nil {
		log.Fatalf("Failed to create HTML file: %v", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (14):"+reset, err)
		}
	}()

	// Write the formatted HTML content to the file
	_, err = file.WriteString(
		fmt.Sprintf(htmlContent, content),
	)
	if err != nil {
		log.Fatalf("Failed to write to HTML file: %v", err)
	}
}

// Function used to generate and save the HTML content to a file
func saveHTML(genHTML string, genFilename string) {

	file, err := os.Create(genFilename)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Can create %s:"+reset, genFilename, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. Closing (15):"+reset, err)
		}
	}()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Can write %s:"+reset, genFilename, err)
		return
	}
}

// Copy Regex to the clipboard
func copyRegexToClipboard() {
	content, err := os.ReadFile(regexOutputFile)
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

func getHostnamePort() {
	// Load the INI file
	cfg, err := ini.Load("go_seo_segmentifyLite.ini")
	if err != nil {
		log.Fatalf(red+"Error. getHostnamePort. Failed to read go_seo_segmentifyLite.ini file: %v"+reset, err)
	}

	// Get values from the INI file
	hostname = cfg.Section("").Key("hostname").String()
	port = cfg.Section("").Key("port").String()

	// Save the values to variables
	var serverHostname, serverPort string
	serverHostname = hostname
	serverPort = port

	// Print the values (for demonstration purposes)
	fmt.Printf(green+"\nHostname: %s\n"+reset, serverHostname)
	fmt.Printf(green+"Port: %s\n"+reset, serverPort)
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

// Display the welcome banner
func displayBanner() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite

	fmt.Println(green + `
 ██████╗  ██████╗         ███████╗███████╗ ██████╗ 
██╔════╝ ██╔═══██╗        ██╔════╝██╔════╝██╔═══██╗
██║  ███╗██║   ██║        ███████╗█████╗  ██║   ██║
██║   ██║██║   ██║        ╚════██║██╔══╝  ██║   ██║
╚██████╔╝╚██████╔╝███████╗███████║███████╗╚██████╔╝
 ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚══════╝ ╚═════╝`)

	fmt.Println(purple + `
███████╗███████╗ ██████╗ ███╗   ███╗███████╗███╗   ██╗████████╗██╗███████╗██╗   ██╗██╗     ██╗████████╗███████╗
██╔════╝██╔════╝██╔════╝ ████╗ ████║██╔════╝████╗  ██║╚══██╔══╝██║██╔════╝╚██╗ ██╔╝██║     ██║╚══██╔══╝██╔════╝
███████╗█████╗  ██║  ███╗██╔████╔██║█████╗  ██╔██╗ ██║   ██║   ██║█████╗   ╚████╔╝ ██║     ██║   ██║   █████╗
╚════██║██╔══╝  ██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║   ██║██╔══╝    ╚██╔╝  ██║     ██║   ██║   ██╔══╝
███████║███████╗╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║   ██║██║        ██║   ███████╗██║   ██║   ███████╗
╚══════╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝   ╚═╝   ╚══════╝`)

	//Display welcome message
	fmt.Println(purple + "\nsegmentifyLite: Fast segmentation regex generation\n" + reset)
	fmt.Println(purple+"Version:"+reset, version, "\n")

	fmt.Println(green + "\nThe Go_Seo segmentifyLite server is ON.\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Server started at " + formattedTime + reset)
	fmt.Println(green+"Maximum No. of URLs to be processed is"+reset, maxURLsToProcess, "k")

	// Get the hostname and port
	getHostnamePort()

	fmt.Println(green + "\n... waiting for requests\n" + reset)

}
