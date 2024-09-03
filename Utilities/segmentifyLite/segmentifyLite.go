// segmentifyLite. Generate the segmentation regex for a specified crawl
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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Version
var version = "v0.2"

// Changelog v0.2
// UI updates & refinements (segmentifyLite)
// UI updates & refinements (index.html)
// Version & session displayed in UI
// Tooltips displayed on launch screen
// Added env. variable "envSegmentifyLiteHostingMode". Set to "local" or "docker"
// Fixed error in generated Shopify regex
// Fixed error in generated SFCC regex
// Delete temp file when regex generation is complete
// Identify PDP pages and generate segment (experimental)
// Comments to signal the end of Level 1 and Level 2 folders fixed
// Folders with 1OO URLs are ignored

// Token, log folder and cache folder acquired from environment variables
var envBotifyAPIToken string
var envSegmentifyLiteLogFolder string
var envSegmentifyLiteFolder string
var envSegmentifyLiteHostingMode string

// Colours & text formatting
var purple = "\033[0;35m"
var red = "\033[0;31m"
var green = "\033[0;32m"
var yellow = "\033[0;33m"
var reset = "\033[0m"
var lineSeparator = "█" + strings.Repeat("█", 129)
var clearScreen = "\033[H\033[2J"

// Default input and output files
var urlExtractFile = "siteurlsExport.tmp"
var regexOutputFile = "segment.txt"

// Maximum No. of URLs to process
var maxURLsToProcess = 100000

// Percentage threshold for level 1 & level 2 folders
var thresholdPercent = 0.00
var minFolderSize = 100

// Boolean to signal if SFCC has been detected
var sfccDetected = false

// Boolean to signal if Shopify has been detected
var shopifyDetected = false

// Strings used to store the project credentials for API access
var organisation string
var project string

// Number of forward-slashes in the URL to count in order to identify the folder level
// 4 = level 1
// 5 = level 2
var slashCountLevel1 = 4
var slashCountLevel2 = 5

// Host name and port the web server runs on
var hostname string
var port string
var fullHost string
var protocol string

// Name of the cache folder used to store the generated HTML
var cacheFolder string
var cacheFolderRoot string

// No of executions & generated session ID
var sessionIDCounter int

// PDP Regex
var generatePDPRegex bool
var isProductURL bool

// Declare the mutex
var mutex sync.Mutex

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

// ByCount implements a sorting interface for FolderCount slice
type ByCount []FolderCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func main() {

	startUp()

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
		organisation = r.Form.Get("organization")
		project = r.Form.Get("project")

		// Generate a session ID used for grouping log entries
		sessionID, err := generateSessionID(8)
		if err != nil {
			fmt.Println(red+"Error. writeLog. Failed generating a session ID: %s"+reset, err)
			os.Exit(0)
		}

		cacheFolderRoot = envSegmentifyLiteFolder
		cacheFolder = cacheFolderRoot + "/" + sessionID + organisation

		createCacheFolder()

		// Process URLs
		dataStatus := processURLs(sessionID)

		// Manage errors
		// An invalid org/project name has been specified
		if dataStatus == "errorNoProjectFound" {
			writeLog(sessionID, organisation, project, "No project found")
			generateErrorPage("No project found. Try another organisation and project name. (" + organisation + "/" + project + ")")
			http.Redirect(w, r, cacheFolder+"/"+"go_seo_segmentifyLiteError.html", http.StatusFound)
			return
		}

		// An error occurred in the process URLs function
		if dataStatus == "errorProcessURLs" {
			writeLog(sessionID, organisation, project, "No project found")
			generateErrorPage("Some kind of error occurred when processing URLs. Check the log for more information. (" + organisation + "/" + project + ")")
			http.Redirect(w, r, cacheFolder+"/"+"go_seo_segmentifyLiteError.html", http.StatusFound)
			return
		}

		writeLog(sessionID, organisation, project, "URLs acquired")

		// Generate the output file to store the regex
		generateRegexFile()

		//Level 1 and 2 folders
		level1and2Folders()

		// PDP pages. Only generate if PDP pages have been detected
		if isProductURL {
			insertPDPRegex()
		}

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
			writeLog(sessionID, organisation, project, "SFCC detected")
			sfccURLs()
		}

		// Shopify if detected
		if shopifyDetected {
			writeLog(sessionID, organisation, project, "Shopify detected")
			shopifyURLs()
		}

		//Static resources
		staticResources()

		writeLog(sessionID, organisation, project, "Regex generated successfully")

		// Generate the HTML used to present the regex
		generateSegmentationRegex(sessionID)

		// Display results and clean up
		finishUp(sessionID)

		// Respond to the client with a success message or redirect to another page
		http.Redirect(w, r, cacheFolder+"/go_seo_segmentifyLite.html", http.StatusFound)
	})

	// Start the HTTP server
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(red+"Error. main. Cannot start HTTP server.:"+reset, err)
		os.Exit(1)
	}
}

// Use the API to get the first 300k URLs and export them to a temp file
func processURLs(sessionID string) string {

	//Get the last analysis slug
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", organisation, project)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(red+"\nError. processURLs Cannot create request:"+reset, err)
		return "errorProcessURLs"
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(red+"\nError. processURLs. Check your network connection: "+reset, err)
		return "errorProcessURLs"
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(red+"Error. processURLs. Closing (1):"+reset, err)
		}
	}()

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(red+"\nError. processURLs. Cannot read response body: "+reset, err)
		return "errorProcessURLs"
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		fmt.Println(red+"\nError. processURLs. Cannot unmarshall JSON: "+reset, err)
		return "errorProcessURLs"
	}

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError. processURLs. Invalid credentials or no crawls found in the project (1)" + reset)
		return "errorNoProjectFound"
	}

	//Display the welcome message
	fmt.Println()
	fmt.Printf(yellow + sessionID + purple + " Generating segmentation regex" + reset)
	fmt.Printf("\n%s%s%s Organisation: %s, Project: %s\n", yellow, sessionID, reset, organisation, project)

	//Create a file for writing
	file, err := os.Create(urlExtractFile)
	if err != nil {
		fmt.Println(red+"\nError. processURLs. Cannot create file: "+reset, err)
		return "errorProcessURLs"

	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. processURLs. Closing (2):"+reset, err)
		}
	}()

	//Initialize total count
	totalCount := 0
	fmt.Println(yellow+sessionID+reset+" Latest analysis slug:", responseObject.Results[0].Slug)

	analysisSlug := responseObject.Results[0].Slug

	//Iterate through pages 1 through to the maximum no of pages defined by maxURLsToProcess
	//Each page returns 1000 URLs
	for page := 1; page <= maxURLsToProcess; page++ {

		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", organisation, project, analysisSlug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)
		//bloo
		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+envBotifyAPIToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"\nError. processURLs. Cannot connect to the API: "+reset, err)
			return "errorProcessURLs"
		}

		//Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println(red+"\nError. processURLs. Cannot decode JSON: "+reset, err)
			return "errorProcessURLs"
		}

		//Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "\nError. processURLs. Invalid credentials or no crawls found in the project (2)" + reset)
			return "errorNoProjectFound"
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
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println(red+"\nError. processURLs. Cannot write to file: "+reset, err)
						return "errorProcessURLs"
					}
					count++
					totalCount++
				}
			}
		}

		//If there are no more URLS process exit the function
		if count == 0 {
			break
		}

		//Max. number of URLs has been reached
		if totalCount > maxURLsToProcess {
			break
		}

		fmt.Printf("%s%s%s Page %d: %d URLs processed\n", yellow, sessionID, reset, page, count)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(red+"Error. processURLs. Closing (3):"+reset, err)
		}
	}()

	return "success"
}

// Generate regex for level 1 and 2 folders
func level1and2Folders() {

	//Level1 folders
	//Get the threshold. Use the level 1 slashCount
	_, thresholdValueL1 := levelThreshold(urlExtractFile, slashCountLevel1)

	//Generate the regex
	segmentFolders(thresholdValueL1, slashCountLevel1, "Level 1 Folders")

	//Level2 folders
	//Get the threshold. Use the level 2 slashCount
	_, thresholdValueL2 := levelThreshold(urlExtractFile, slashCountLevel2)

	//Level2 folders
	segmentFolders(thresholdValueL2, slashCountLevel2, "Level 2 Folders")
}

func generateRegexFile() {

	//Always create the file.
	outputFile, err := os.Create(regexOutputFile)
	if err != nil {
		fmt.Printf(red+"\nError. generateRegexFile. Cannot create output file: %v\n"+reset, err)
		os.Exit(1)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. generateRegexFile. Closing (4):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	// Get the user's local time zone for the header
	userLocation, err := time.LoadLocation("") // Load the default local time zone
	if err != nil {
		fmt.Println("\nError loading user's location:", err)
		return
	}
	// Get the current date and time in the user's local time zone
	currentTime := time.Now().In(userLocation)

	_, err = writer.WriteString(fmt.Sprintf("# Regex made with love using segmentifyLite %s\n", version))

	if err != nil {
		fmt.Printf(red+"\nError. generateRegexFile. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	_, err = writer.WriteString(fmt.Sprintf("# Organisation name: %s\n", organisation))
	if err != nil {
		errMsg := fmt.Errorf(red+"Error. Cannot write organisation name in Regex file: %w"+reset, err)
		println(errMsg)
	}
	_, err = writer.WriteString(fmt.Sprintf("# Project name: %s\n", project))
	if err != nil {
		errMsg := fmt.Errorf(red+"Error. Cannot write project name in Regex file: %w"+reset, err)
		println(errMsg)
	}
	_, err = writer.WriteString(fmt.Sprintf("# Generated %s", currentTime.Format(time.RFC1123)))
	if err != nil {
		errMsg := fmt.Errorf(red+"Error. Cannot write generate date/time name in Regex file: %w"+reset, err)
		println(errMsg)
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"\nError. generateRegexFile. Cannot flush writer: %v\n"+reset, err)
		os.Exit(1)
	}
}

func segmentFolders(thresholdValue int, slashCount int, folderLevel string) {

	//Open the input file for reading
	file, err := os.Open(urlExtractFile)
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. segmentFolders. Closing (5):"+reset, err)
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

		//Check if the line contains a question mark, if yes, skip to the next line
		//if strings.Contains(line, "?") {
		//	continue
		//}

		// Is this a product URL?
		isProductURL = isValidisProductURL(line)
		if isProductURL {
			println(line)
			generatePDPRegex = true
		}

		//Split the line into substrings using a forward-slash as delimiter
		// slashCount
		// = 4 for Level 1 folders
		// slashCount
		// = 5 for Level 2 folders
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
			if count > minFolderSize {
				sortedCounts = append(sortedCounts, FolderCount{folderName, count})
			} else {
				noFoldersExcluded++
			}
		} else {
			// Count the number of folders excluded because they didn't meet the thresholdValue
			noFoldersExcluded++
		}
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. segmentFolders. Closing (6):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the segment name
	// SlashCount = 4 signals level 1 folders
	// SlashCount = 5 signals level 2 folders

	// Level 1
	if slashCount == 4 {
		if _, err := writer.WriteString(fmt.Sprintf("\n\n[segment:sl_level1_folders]\n@Home\npath /\n\n")); err != nil {
			fmt.Printf(red+"Error. segmentFolders. Cannot write segment to writer. Level 1 folders: %v\n"+reset, err)
		}
	}

	// Level 2
	if slashCount == 5 {
		if _, err := writer.WriteString(fmt.Sprintf("\n\n[segment:sl_level2_folders]\n@Home\npath /\n\n")); err != nil {
			fmt.Printf(red+"Error. segmentFolders. Cannot write segment to writer. Level 2 folders: %v\n"+reset, err)
		}
	}

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		if folderValueCount.Text != "" {
			//Extract the text between the third and fourth forward-slashes
			parts := strings.SplitN(folderValueCount.Text, "/", 4)
			if len(parts) >= 4 && parts[3] != "" {
				folderLabel := parts[3] //Extract the text between the third and fourth forward-slashes
				_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, folderValueCount.Text))
				if err != nil {
					fmt.Printf(red+"\nError. segmentFolders. Cannot write to output file: %v\n"+reset, err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	//_, err = writer.WriteString("@~Other\npath /*\n# ----End of level2Folders Segment----\n")
	formattedString := fmt.Sprintf("@~Other\npath /*\n# ----End of %s Segment----\n", folderLevel)
	_, err = writer.WriteString(formattedString)
	if err != nil {
		fmt.Printf(red+"Error. segmentFolders. Cannot write segment to writer: %v\n"+reset, err)
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Folder URL analysis----\n")
	if err != nil {
		fmt.Printf(red+"Error. segmentFolders. Cannot write segment to writer: %v\n"+reset, err)
	}
	for _, folderValueCount := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", folderValueCount.Text, folderValueCount.Count))
		if err != nil {
			fmt.Printf(red+"Error. segmentFolders. Cannot write segment to writer: %v\n"+reset, err)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"\nError. segmentFolders. Cannot flush writer: %v\n"+reset, err)
		os.Exit(1)
	}
}

// Regex for subdomains
func subDomains() {

	//Open the input file
	file, err := os.Open(urlExtractFile)
	if err != nil {
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. subDomains. Closing (7):"+reset, err)
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
	outputFile, err := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. subDomains. Closing (8):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_subdomains]\n@Home\npath /\n\n"))
	if err != nil {
		fmt.Printf(red+"Error. subDomains. Cannot write segment to writer: %v\n"+reset, err)
	}

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		if folderValueCount.Text != "" {
			//Extract the text between the third and fourth forward-slashes
			parts := strings.SplitN(folderValueCount.Text, "/", 4)
			if len(parts) >= 3 && parts[2] != "" {
				folderLabel := parts[2] //Extract the text between the third and fourth forward-slashes
				_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, folderValueCount.Text))
				if err != nil {
					fmt.Printf(red+"Error. subDomains. Cannot write segment to writer: %v\n"+reset, err)
					// Handle or return the error as needed
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of subDomains Segment----\n")
	if err != nil {
		fmt.Printf(red+"Error. subDomains. Cannot write segment to writer: %v\n"+reset, err)
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----subDomains Folder URL analysis----\n")
	if err != nil {
		fmt.Printf(red+"Error. subDomains. Cannot write segment to writer: %v\n"+reset, err)
	}
	for _, folderValueCount := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", folderValueCount.Text, folderValueCount.Count))
		if err != nil {
			fmt.Printf(red+"\nError. subDomains. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"\nError. subDomains. Cannot flush writer: %v\n"+reset, err)
		os.Exit(1)
	}
}

// Regex to identify which parameter keys are used
func parameterKeys() {

	//Open the input file
	file, err := os.Open(urlExtractFile)
	if err != nil {
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. parameterKeys. Closing (9):"+reset, err)
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
	outputFile, err := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. parameterKeys. Closing (10):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_parameter_keys]\n"))
	if err != nil {
		fmt.Printf(red+"Error. parameterKeys. Cannot write segment to writer: %v\n"+reset, err)
	}

	//Write the regex
	for _, folderValueCount := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("@%s\nquery *%s=*\n\n", folderValueCount.Text, folderValueCount.Text))
		if err != nil {
			fmt.Printf(red+"\nError. parameterKeys. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of parameterKeys Segment----\n")
	if err != nil {
		fmt.Printf(red+"Error. parameterKeys. Cannot write segment to writer: %v\n"+reset, err)
		// Handle or return the error as needed
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----parameterKeys URL analysis----\n")
	if err != nil {
		fmt.Printf(red+"Error. parameterKeys. Cannot write segment to writer: %v\n"+reset, err)
		// Handle or return the error as needed
	}
	for _, folderValueCount := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", folderValueCount.Text, folderValueCount.Count))
		if err != nil {
			fmt.Printf(red+"\nError. parameterKeys. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"\nError. parameterKeys. Cannot flush writer: %v\n"+reset, err)
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

# ----End of sl_parameter_usage----
`

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

# ----End of sl_no_of_parameters----
`

	errParameterNoRegex := insertStaticRegex(parameterNoRegex)
	if errParameterNoRegex != nil {
		panic(errParameterNoRegex)
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

# ----End of sl_no_of_folders----
`

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

# ----End of sl_sfcc----
`

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
# ----End of sl_shopify----
`

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
[segment:sl_Static_Resources]  
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

# ----End of sl_static_resources----
`

	errStaticResources := insertStaticRegex(staticResources)
	if errStaticResources != nil {
		panic(errStaticResources)
	}

}

// PDP Regex
func insertPDPRegex() {

	generatePDPRegex := `
[segment:sl_PDP]  
@pdp
path rx:[a-zA-Z0-9\-]+-\d+\.html$

@Other
path /*

# ----End of sl_PDP segment----
`
	errStaticResources := insertStaticRegex(generatePDPRegex)
	if errStaticResources != nil {
		panic(errStaticResources)
	}
}

// Get the folder size threshold for level 1 & 2 folders
func levelThreshold(inputFilename string, slashCount int) (largestValueSize, fivePercentValue int) {

	// Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf(red+"\nError. levelThreshhold. Cannot open input file: %v\n"+reset, err)
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. levelThreshold. Closing (11):"+reset, err)
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

// Display the results and finishUp
func finishUp(sessionID string) {

	// We're done
	fmt.Println(lineSeparator)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println("\nSession ID: " + sessionID)
	fmt.Println("\nsegmentifyLite: Done at " + formattedTime)
	fmt.Printf("\n%s%s%s Organisation: %s, Project: %s\n", yellow, sessionID, reset, organisation, project)

	// Make a tidy display
	fmt.Println()
	fmt.Println(lineSeparator)

	// Delete the temp. file
	_ = os.Remove(urlExtractFile)
}

// Write the static Regex to the segments file
func insertStaticRegex(regexText string) error {

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(regexOutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println(red+"Error. insertStaticRegex. Closing (12):"+reset, err)
		}
	}()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	_, err = writer.WriteString(regexText)
	if err != nil {
		fmt.Printf(red+"\nError. insertStaticRegex. Cannot write to outputfile: %v\n"+reset, err)
		panic(err)
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"\nError. insertStaticRegex. Cannot flush writer: %v\n"+reset, err)
		os.Exit(1)
	}

	return err
}

func writeLog(sessionID, organisation, project, statusDescription string) {

	// Define log file name
	fileName := envSegmentifyLiteLogFolder + "/_segmentifyLite.log"

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
			fmt.Println(red+"Error. writeLog. Closing (13):"+reset, err)
		}
	}()

	// Get current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Construct log record
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%s\n",
		sessionID, currentTime, organisation, project, statusDescription)

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

func generateSessionID(length int) (string, error) {
	// Generate random sessionID
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

// Generate the HTML pages used to present the segmentation regex
func generateSegmentationRegex(sessionID string) {

	// Using these two variables to replace width values in the HTML below because string interpolation confuses the percent signs as variables
	width50 := "50%"
	width100 := "100%"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>segmentifyLite</title>
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
            background-color: DeepSkyBlue;
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
            background-color: Green;
            box-shadow: 0 6px 8px rgba(0, 0, 0, 0.15);
        }
		.header-info {
            position: absolute;
            top: 100px;
            right: 20px;
            color: DeepSkyBlue;
            font-size: 15px;
        }
 		.header-info {
            position: absolute;
            top: 100px;
            right: 20px;
            text-align: right;
        }
        .darkgrey {
            color: #00796b;
        }
    </style>
</head>
<body>

<!-- Top Banner -->
<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">segmentifyLite</span>
</header>

<!-- Back Button to create a new dashboard -->
<button class="back-button" onclick="goHome()">New segmentation</button>
<div class="header-info">
    <span class="deepskyblue">Version: </span><span class="darkgrey">`+fmt.Sprintf("%s", version)+`</span><br>
    <span class="deepskyblue">Session: </span><span class="darkgrey">`+fmt.Sprintf("%s", sessionID)+`</span>
</div>
<script>
    function goHome() {
        window.open('%s://%s/', '_blank');
    }
async function copyFileToClipboard() {
        try {
            const response = await fetch('../../segment.txt');
            if (!response.ok) {
                throw new Error('Network response was not ok ' + response.statusText);
            }
            const text = await response.text();
            await navigator.clipboard.writeText(text);
            console.log('Text copied to clipboard');
        } catch (error) {
            console.error('Failed to copy text to clipboard: ', error);
        }
    }

    document.addEventListener('DOMContentLoaded', (event) => {
        copyFileToClipboard();
    });
</script>

<section class="container row no-border">
    <iframe src="go_seo_segmentationRegex.html" title="Segmentation regex"></iframe>
</section>

</body>
</html>
`, width100, width50, width100, protocol, fullHost)

	// Generate the URL to link to the segment editor in the project
	projectURL := "https://app.botify.com/" + organisation + "/" + project + "/segmentation"

	htmlContent += fmt.Sprintf("<div style='text-align: center;'>\n")
	htmlContent += fmt.Sprintf("<h2 style='color: deepskyblue;'>Segmentation regex generation is complete</h2>\n")
	htmlContent += fmt.Sprintf("<h3 style='color: dimgray; padding-left: 20px; padding-right: 20px;'>The regex has been copied to the clipboard ready for pasting directly into your Botify project.</h3>\n")
	htmlContent += fmt.Sprintf("<h4 style='color: dimgray;'><a href='%s' target='_blank'>Click here to open the segment editor for %s</a></h4>\n", projectURL, project)
	htmlContent += fmt.Sprintf("</div>\n")

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_segmentifyLite.html")

	// Generate the HTML containing the segmentation regex
	generateSegmentHTML()

	// Copy the regex to the clipboard
	// Not used, unable to do this when segmentifyLite is hosted by Botify.
	//copyRegexToClipboard()
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
	file, err := os.Create(cacheFolder + "/go_seo_segmentationRegex.html")
	if err != nil {
		log.Fatalf(red+"Error. generateSegmentHTML. Failed to create HTML file: %v"+reset, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. generateSegmentHTML. Closing (14):"+reset, err)
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
    <title>segmentifyLite</title>
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
    <span style="font-size: 20px;">segmentifyLite</span>
</header>

<!-- Back Button -->
<button class="back-button" onclick="goHome()">Try again</button>

<!-- Error message -->
<div class="error-message" id="error-message">
    %s
</div>

<script>
    function goHome() {
        window.open('%s://%s/', '_blank');
    }

</script>

</body>
</html>`, displayMessage, protocol, fullHost)

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_segmentifyLiteError.html")

}

// Function used to generate and save the HTML content to a file
func saveHTML(genHTML string, genFilename string) {

	file, err := os.Create(cacheFolder + genFilename)
	if err != nil {
		fmt.Printf(red+"Error. saveHTML. Can create %s: "+reset+"%s\n", genFilename, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. saveHTML. Closing (15):"+reset, err)
			return
		}
	}()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fmt.Printf(red+"Error. saveHTML. Can write %s: "+reset+"%s\n", genFilename, err)
		return
	}
}

// Create the cache folder
func createCacheFolder() {

	cacheDir := cacheFolder

	// Check if the directory already exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Create the directory and any necessary parents
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			log.Fatalf(red+"Error. Failed to create the cache directory: %v"+reset, err)
		}
	}
}

func getHostnamePort() {

	// Load the INI file
	cfg, err := ini.Load("segmentifyLite.ini")
	if err != nil {
		fmt.Printf(red+"Error. getHostnamePort. Failed to read segmentifyLite.ini file: %v"+reset, err)
	}

	// Get values from the .ini file
	if !cfg.Section("").HasKey("protocol") {
		fmt.Println(yellow + "Warning: 'protocol' not found in configuration file. Will default to HTTPS." + reset)
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
	if envSegmentifyLiteHostingMode == "local" {
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

func startUp() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite

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
███████╗███████╗ ██████╗ ███╗   ███╗███████╗███╗   ██╗████████╗██╗███████╗██╗   ██╗██╗     ██╗████████╗███████╗
██╔════╝██╔════╝██╔════╝ ████╗ ████║██╔════╝████╗  ██║╚══██╔══╝██║██╔════╝╚██╗ ██╔╝██║     ██║╚══██╔══╝██╔════╝
███████╗█████╗  ██║  ███╗██╔████╔██║█████╗  ██╔██╗ ██║   ██║   ██║█████╗   ╚████╔╝ ██║     ██║   ██║   █████╗
╚════██║██╔══╝  ██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║   ██║██╔══╝    ╚██╔╝  ██║     ██║   ██║   ██╔══╝
███████║███████╗╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║   ██║██║        ██║   ███████╗██║   ██║   ███████╗
╚══════╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝   ╚═╝   ╚══════╝`)

	fmt.Println()
	fmt.Println(purple+"\nVersion:"+reset, version)
	fmt.Println(green + "\nsegmentifyLite server is ON\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Server started at " + formattedTime + reset)
	fmt.Println(green+"Maximum No. of URLs to be processed is", maxURLsToProcess, "k")

	// Get the environment variables for token, log & cache folder
	envBotifyAPIToken, envSegmentifyLiteLogFolder, envSegmentifyLiteFolder, envSegmentifyLiteHostingMode = getEnvVariables()

	// Get the hostname and port
	getHostnamePort()

	fmt.Println(green + "\n... waiting for requests\n" + reset)
}

// Get environment variables for token and cache folders
func getEnvVariables() (envBotifyAPIToken string, envSegmentifyLiteLogFolder string, envSegmentifyLiteFolder string, envSegmentifyLiteHostingMode string) {

	// Botify API token from the env. variable getbotifyAPIToken
	envBotifyAPIToken = os.Getenv("envBotifyAPIToken")
	if envBotifyAPIToken == "" {
		fmt.Println(red + "Error. getEnvVariables. envBotifyAPIToken environment variable not set." + reset)
		fmt.Println(red + "Cannot start segmentifyLite server." + reset)
		os.Exit(0)
	}

	// Storage folder for the log file
	envSegmentifyLiteLogFolder = os.Getenv("envSegmentifyLiteLogFolder")
	if envSegmentifyLiteLogFolder == "" {
		fmt.Println(red + "Error. getEnvVariables. envSegmentifyLiteLogFolder environment variable not set." + reset)
		fmt.Println(red + "Cannot start segmentifyLite server." + reset)
		os.Exit(0)
	} else {
		fmt.Println()
		fmt.Println(green + "Log folder: " + envSegmentifyLiteLogFolder + reset)
	}

	// Storage folder
	envSegmentifyLiteFolder = os.Getenv("envSegmentifyLiteFolder")
	if envSegmentifyLiteFolder == "" {
		fmt.Println(red + "Error. getEnvVariables. envSegmentifyLiteFolder environment variable not set." + reset)
		fmt.Println(red + "Cannot start segmentifyLite server." + reset)
		os.Exit(0)
	} else {
		fmt.Println(green + "segmentifyLite cache folder: " + envSegmentifyLiteFolder + reset)
	}

	// Hosting mode. This will be either "local" or "docker"
	envSegmentifyLiteHostingMode = os.Getenv("envSegmentifyLiteHostingMode")
	if envSegmentifyLiteHostingMode == "" {
		fmt.Println(red + "Error. getEnvVariables. envSegmentifyLiteHostingMode environment variable not set." + reset)
		fmt.Println(red + "Cannot start segmentifyLite server." + reset)
		os.Exit(0)
	} else {
		fmt.Println(green + "segmentifyLite hosting mode: " + envSegmentifyLiteHostingMode + reset)
	}

	return envBotifyAPIToken, envSegmentifyLiteLogFolder, envSegmentifyLiteFolder, envSegmentifyLiteHostingMode
}

// isValidisProductURL checks if a URL ends with a product name followed by a numeric identifier
func isValidisProductURL(url string) bool {
	// Define the regex pattern for matching URLs ending with a product name and numeric identifier
	pattern := `[a-zA-Z0-9\-]+-\d+\.html$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(url)
}
