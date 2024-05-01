//segmentifyLite
//Written by Jason Vicinanza

//To run this:
//go run segmentifyLite.go org_name project_name
//Example: go run segmentifyLite.go jason-org jason-project-name (with complier)
//Example: segmentifyLite.go jason-org jason-project-name (with executable)
//Remember to use your own api_token

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

// Colours
var purple = "\033[0;35m"
var red = "\033[0;31m"
var reset = "\033[0m"

// Default input and output files
var inputFilename = "siteurlsExport.csv"
var outputFilename = "segment.txt"

// Unicode escape sequence for the checkmark symbol
var checkmark = "\u2713"

// Maximum No. of pages to export. 300 = 300k etc.
var maxURLsToExport = 300

// Percentage threshold
var thresholdPercent = 0.05

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

	//Display welcome message
	fmt.Println(purple + "\nsegmentifyLite: Fast segmentation regex generation\n" + reset)
	fmt.Println(purple+"Version:"+reset, version, "\n")

	//Generate the list of URLs
	urlExport()

	//Level1 folders
	//Get the threshold
	largestFolderSize, thresholdValueL1 := level1Threshold(inputFilename)
	fmt.Printf(purple + "Calculating folder threshold\n" + reset)
	fmt.Printf("Largest folder size found is %d URLs\n", largestFolderSize)
	fmt.Printf("Threshold folder size: %d\n", thresholdValueL1)

	//generate the regex
	segmentLevel1(thresholdValueL1)

	//Level2 folders
	segmentLevel2()

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

	//It's done! segmentifyList has left the building
	fmt.Println(purple+"Your regex can be found in:", outputFilename+reset)
	fmt.Println(purple + "Regex generation complete" + reset)
}

// Use the API to get the first 300k URLs and export them to a file
func urlExport() {

	//Get the command line arguments for the org and project name
	if len(os.Args) < 3 {
		fmt.Println(red + "Error. Please provide the organisation, project name as line arguments")
		os.Exit(1)
	}
	orgName := os.Args[1]
	projectName := os.Args[2]

	//Get the last analysis slug
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: "+reset, err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"Error. Check your network connection: "+reset, err)
	}
	defer res.Body.Close()

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(red+"Error reading response body: "+reset, err)
		os.Exit(1)
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		log.Fatal(red+"Error. Cnnot unmarshall JSON: "+reset, err)
		os.Exit(1)
	}

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "Error. Invalid crawl or no crawls found in the project")
		os.Exit(1)
	}

	//Display the welcome message
	fmt.Println(purple + "Exporting URLs" + reset)

	//Create a file for writing
	file, err := os.Create(inputFilename)
	if err != nil {
		fmt.Println(red+"Error creating file: "+reset, err)
		os.Exit(1)
	}
	defer file.Close()

	//Initialize total count
	totalCount := 0
	fmt.Println("Maximum No. of URLs to be exported is", maxURLsToExport, "k")
	fmt.Println("Organisation Name:", orgName)
	fmt.Println("Project Name:", projectName)
	fmt.Println("Latest analysis Slug:", responseObject.Results[0].Slug)
	analysisSlug := responseObject.Results[0].Slug
	urlEndpoint := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/", orgName, projectName, analysisSlug)
	fmt.Println("End point:", urlEndpoint, "\n")

	//Iterate through pages 1 through to the maximum no of pages defined by maxURLsToExport
	//Each page returns 1000 URLs
	for page := 1; page <= maxURLsToExport; page++ {

		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", orgName, projectName, analysisSlug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "Token c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"Error. Cannot connect to the API: "+reset, err)
			os.Exit(1)
		}
		defer res.Body.Close()

		//Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println(red+"Error. Cannot decode JSON: "+reset, err)
			os.Exit(1)
		}

		//Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "Error. Results not found in response. Check the specified organisation and project")
			os.Exit(1)
		}

		//Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println(red+"Error. Cannot write to file: "+reset, err)
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

		//If there are no more URLS export exit the function
		if count == 0 {
			//Print total number of URLs saved
			fmt.Printf("\nTotal URLs exported: %d\n", totalCount)
			fmt.Println(purple + "\nURL Extract complete. Generating regex...\n" + reset)
			break
		}

		//Max. number of URLs (200k) has been reached
		if totalCount > 190000 {
			fmt.Printf("\n\nExport limit of %d URLs reached. Generating regex...\n\n", totalCount)
			break
		}

		fmt.Printf("\nPage %d: %d URLs exported\n", page, count)
	}

}

// Get the folder size threshold for level 1 folders
func level1Threshold(inputFilename string) (largestValueSize, fivePercentValue int) {
	// Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error: Cannot open input file: %v\n", err)
		return 0, 0
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

		// Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")

		// Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			// Extract the text between the third and fourth forward slashes
			text := strings.Join(parts[:4], "/")

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

// Regex for level 1 folders
func segmentLevel1(thresholdValueL1 int) {

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf(red+"segmentiftyLite. Error. Cannot open input file: %v\n "+reset, err)
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

	//Counter to track the number of folders excluded from teh regex
	noFoldersExcludedL1 := 0

	//Display welcome message
	fmt.Println(purple + "\nFirst level folders" + reset)
	fmt.Printf("Folders with less than %d URLs will be excluded\n", thresholdValueL1)

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

		//Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")

		//Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			//Extract the text between the third and fourth forward slashes
			text := strings.Join(parts[:4], "/")

			//Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			//Update the count for this value if it's not empty
			if text != "" {
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
		if count > thresholdValueL1 {
			sortedCounts = append(sortedCounts, FolderCount{folderName, count})
		} else {
			// Count the number of folders excluded
			noFoldersExcludedL1++
		}
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Display the counts for each unique value
	for _, vc := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", vc.Text, vc.Count)
	}

	fmt.Printf("\nNo. of level 1 folders excluded %d\n", noFoldersExcludedL1)

	//Open the output file for writing
	//Always create the file.
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error. Cannot create output file: %v\n"+reset, err)
		os.Exit(1)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	// Get the user's local time zone for the header
	userLocation, err := time.LoadLocation("") // Load the default local time zone
	if err != nil {
		fmt.Println("Error loading user's location:", err)
		return
	}
	// Get the current date and time in the user's local time zone
	currentTime := time.Now().In(userLocation)

	_, err = writer.WriteString(fmt.Sprintf("# Regex made with Go_SEO/segmentifyLite %s\n", version))
	_, err = writer.WriteString(fmt.Sprintf("# Generated on %s\n", currentTime.Format(time.RFC1123)))

	// Start of regex
	_, err = writer.WriteString(fmt.Sprintf("\n[segment:sl_level1_Folders]\n@Home\npath /\n\n"))

	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Write the regex
	for _, vc := range sortedCounts {
		if vc.Text != "" {
			//Extract the text between the third and fourth forward slashes
			parts := strings.SplitN(vc.Text, "/", 4)
			if len(parts) >= 4 && parts[3] != "" {
				folderLabel := parts[3] //Extract the text between the third and fourth forward slashes
				_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, vc.Text))

				if err != nil {
					fmt.Printf(red+"segment1stLevel. Error. Cannot write to output file: %v\n"+reset, err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines\
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of level1Folders Segment----\n")
	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Level 1 Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"segment1stLevel. Error. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error. Cannot flush writer: %v\n", err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"segment1stLevel. Error. Cannot scan input file: %v\n"+reset, err)
		os.Exit(1)
	}
}

// Regex for level 2 folders
func segmentLevel2() {

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
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
	fmt.Println(purple + "\nSecond level folders" + reset)

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

		//Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")

		//Check if there are at least 4 parts in the line
		if len(parts) >= 5 {
			//Extract the text between the third and fourth forward slashes
			text := strings.Join(parts[:5], "/")

			//Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			//Update the count for this value if it's not empty
			if text != "" {
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
	for _, vc := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", vc.Text, vc.Count)
	}

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_level2_Folders]\n@Home\npath /\n\n"))

	if err != nil {
		fmt.Printf(red+"segment2ndLevel. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Write the regex
	for _, vc := range sortedCounts {
		if vc.Text != "" {
			//Extract the text between the third and fourth forward slashes
			parts := strings.SplitN(vc.Text, "/", 4)
			if len(parts) >= 4 && parts[3] != "" {
				folderLabel := parts[3] //Extract the text between the third and fourth forward slashes
				_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, vc.Text))
				if err != nil {
					fmt.Printf(red+"segment2ndLevel. Error. Cannot write to output file: %v\n"+reset, err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of level2Folders Segment----\n")
	if err != nil {
		fmt.Printf(red+"segment2ndLevel. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Level 2 Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"segment1stLevel. Error. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"segment2ndLevel. Error. Cannot flush writer: %v\n", err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"segment2ndLevel. Error. Cannot scan input file: %v\n"+reset, err)
		os.Exit(1)
	}
}

// Regex for subdomains
func subDomains() {

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
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

		//Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")
		//Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			//Extract the text between the third and fourth forward slashes
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
	for _, vc := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", vc.Text, vc.Count)
	}

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_subdomains]\n@Home\npath /\n\n"))

	if err != nil {
		fmt.Printf(red+"subDomains. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Write the regex
	for _, vc := range sortedCounts {
		if vc.Text != "" {
			//Extract the text between the third and fourth forward slashes
			parts := strings.SplitN(vc.Text, "/", 4)
			if len(parts) >= 3 && parts[2] != "" {
				folderLabel := parts[2] //Extract the text between the third and fourth forward slashes
				_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/*\n\n", folderLabel, vc.Text))
				if err != nil {
					fmt.Printf(red+"subDomains. Error. Cannot write to output file: %v\n"+reset, err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of subDomains Segment----\n")
	if err != nil {
		fmt.Printf(red+"subDomains. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----subDomains Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"subDomains. Error. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"subDomains. Error. Cannot flush writer: %v\n"+reset, err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"subDomains. Error. Cannot scan input file: %v\n"+reset, err)
		os.Exit(1)
	}
}

// Regex to identify which parameter keys are used
func parameterKeys() {

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
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
	for _, vc := range sortedCounts {
		fmt.Printf("%s (URLs: %d)\n", vc.Text, vc.Count)
	}

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	//_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_parameterKeys]\n@Home\npath /\n\n"))
	_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_parameterKeys]\n"))

	if err != nil {
		fmt.Printf(red+"parameterKeys. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Write the regex
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("@%s\nquery *%s=*\n\n", vc.Text, vc.Text))
		if err != nil {
			fmt.Printf(red+"parameterKeys. Error. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of parameterKeys Segment----\n")
	if err != nil {
		fmt.Printf(red+"parameterKeys. Error. Cannot write header to output file: %v\n"+reset, err)
		os.Exit(1)
	}

	//Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----parameterKeys URL analysis----\n")
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"parameterKeys. Error. Cannot write to output file: %v\n"+reset, err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"parameterKeys. Error. Cannot flush writer: %v\n"+reset, err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"parameterKeys. Error. Canot scan input file: %v\n"+reset, err)
		os.Exit(1)
	}
}

// Regex to identify of a parameter key is used in the URL
func parameterUsage() {

	//Pages containing parameters
	paramaterUsageRegex := `

[segment:sl_parameter_Usage]
@Parameters
query *=*

@Clean
path /*
# ----End of sl_parameter_Usage----`

	//Parameter usage message
	fmt.Println(purple + "\nParameter usage" + reset)
	errParamaterUsage := insertStaticRegex(paramaterUsageRegex)
	if errParamaterUsage != nil {
		panic(errParamaterUsage)
	}

	//Finished
	fmt.Println("Done!", checkmark, "\n")

}

// Regex to count the number of parameters in the URL
func noOfParameters() {

	//Number of paramaters
	paramaterNoRegex := `


[segment:sl_no_Of_Parameters]
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
# ----End of sl_no_Of_Parameters----`

	//No. of parameters message
	fmt.Println(purple + "Number of parameters" + reset)
	errParamaterNoRegex := insertStaticRegex(paramaterNoRegex)
	if errParamaterNoRegex != nil {
		panic(errParamaterNoRegex)
	}

	//Finished
	fmt.Println("Done!", checkmark, "\n")

}

// Regex to count the number of folders in the URL
func noOfFolders() {

	//Number of folders
	folderNoRegex := `


[segment:sl_no_Of_Folders]
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
# ----End of sl_no_Of_Folders----`

	//No. of folders message
	fmt.Println(purple + "Number of folders" + reset)
	errFolderNoRegex := insertStaticRegex(folderNoRegex)
	if errFolderNoRegex != nil {
		panic(errFolderNoRegex)
	}

	//Finished
	fmt.Println("Done!", checkmark, "\n")

}

// Write the static Regex to the segments file
func insertStaticRegex(regexText string) error {

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	_, err = writer.WriteString(regexText)
	if err != nil {
		panic(err)
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"parameterUsage. Error. Cannot flush writer: %v\n"+reset, err)
		return err
	}

	return err
}

// Display the welcome banner
func displayBanner() {

	//ANSI escape code for Green
	green := "\033[0;32m"

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
}

// Clear the screen
func clearScreen() {

	// Determine the appropriate command based on the operating system used
	var clearCmd string

	switch runtime.GOOS {
	case "windows":
		clearCmd = "cls"
	default:
		clearCmd = "clear"
	}

	cmd := exec.Command(clearCmd)
	cmd.Stdout = os.Stdout
	cmd.Run()
}
