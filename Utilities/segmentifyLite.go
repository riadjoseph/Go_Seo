//segmentifyLite
//Written by Jason Vicinanza

//To run this:
//go run segmentifyLite.go org_name project_name
//Example: go run segmentifyLite.go jason-org jason-project-name

//Version
//version := "v0.1"

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Define a struct to hold text value and its associated count
type ValueCount struct {
	Text  string
	Count int
}

// Implement sorting interface for ValueCount slice
type ByCount []ValueCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func main() {

	//Version
	version := "v0.1"

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	clearScreen()

	displayBanner()

	//Display welcome message
	fmt.Println(purple + "\nsegmentifyLite: Fast segmentation regex generation\n" + reset)
	fmt.Println(purple+"Version:"+reset, version, "\n")

	//Functions to create the various regex statements

	//Generate the list of URLs
	urlExport()

	//Level1 folders
	segmentLevel1()

	//Level2 folders
	segmentLevel2()

	//Subdomains
	subDomains()

	//Parameter keys
	parameterKeys()

	//Pages containing parameters
	paramaterUsageRegex := `

[segment:sl_parameter_Usage]
@Parameters
query *=*

@Clean
path /*
# ----End of sl_parameter_Usage----`

	//Parameter usage message
	fmt.Println(purple + "\nParameter usage\n" + reset)
	errParamaterUsage := insertStaticRegex(paramaterUsageRegex)
	if errParamaterUsage != nil {
		panic(errParamaterUsage)
	}

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
	fmt.Println(purple + "Number of parameters\n" + reset)
	errParamaterNoRegex := insertStaticRegex(paramaterNoRegex)
	if errParamaterNoRegex != nil {
		panic(errParamaterNoRegex)
	}

	//Number of paramaters
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
	fmt.Println(purple + "Number of folders\n" + reset)
	errFolderNoRegex := insertStaticRegex(folderNoRegex)
	if errFolderNoRegex != nil {
		panic(errFolderNoRegex)
	}

	fmt.Println(purple + "segmentiftyLite. Regex generation complete" + reset)
}

func urlExport() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"
	//ANSI escape code for red color
	red := "\033[0;31m"

	//Get the command-line arguments for the org and project name
	if len(os.Args) < 4 {
		fmt.Println(red + "Error. Please provide the organisation, project name & analysis slug as command line arguments")
		os.Exit(1)
	}

	//Display the welcome message
	fmt.Println(purple + "Exporting URLs" + reset)

	// Create a file for writing
	file, err := os.Create("siteurlsExport.csv")
	if err != nil {
		fmt.Println(red+"Error creating file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Initialize total count
	totalCount := 0

	//Get the org and project name from the command line arguments
	orgName := os.Args[1]
	projectName := os.Args[2]
	analysisSlug := os.Args[3]
	fmt.Println("Organisation Name:", orgName)
	fmt.Println("Project Name:", projectName)
	fmt.Println("Analysis Slug:", analysisSlug, "\n")

	// Iterate through pages 1 through 10
	for page := 1; page <= 100; page++ {

		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", orgName, projectName, analysisSlug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "Token c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"Error. Cannot connect to the API:", err)
			os.Exit(1)
		}
		defer res.Body.Close()

		// Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println(red+"Error. Cannot decode JSON:", err)
			os.Exit(1)
		}

		// Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "Error. Results not found in response. Check the specified organisation and project names")
			os.Exit(1)
		}

		// Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println(red+"Error writing to file:", err)
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
			// Print total number of URLs saved
			fmt.Printf("\nTotal URLs exported: %d\n", totalCount)
			fmt.Println(purple + "\nURL Extract complete. Generating regex...\n" + reset)
			break
		}

		fmt.Printf("\nPage %d: %d URLs exported\n", page, count)
	}

}

// Regex for level 1 folders
func segmentLevel1() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code for red color
	red := "\033[0;31m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	inputFilename := "siteurlsExport.csv"
	outputFilename := "segment.txt"

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf(red+"segmentiftyLite. Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	valueCounts := make(map[string]int)

	//Variable to keep track of the total number of records processed
	totalRecords := 0

	//Counter to track the number of records scanned
	recordCounter := 0

	//Display welcome message
	fmt.Println(purple + "First level folders" + reset)

	//Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		//Display a block for each 1000 records scanned
		if recordCounter%1000 == 0 {
			fmt.Print("#")
		}

		// Check if the line contains a quotation mark, if yes, skip to the next line
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
				valueCounts[text]++
			}
		}
	}

	//Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	fmt.Printf("\n")

	//Create a slice to hold ValueCount structs
	var sortedCounts []ValueCount

	//Populate the slice with data from the map
	for value, count := range valueCounts {
		sortedCounts = append(sortedCounts, ValueCount{value, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	/*
		// Display the counts for each unique value
		for _, vc := range sortedCounts {
			fmt.Printf("%s (count: %d)\n", vc.Text, vc.Count)
		}
	*/

	//Open the output file for writing
	//Always create the file.
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("# Regex made with Go_SEO/segmentifyLite (level1)\n\n[segment:sl_level1_Folders]\n@Home\npath /\n\n"))

	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error writing header to output file: %v\n", err)
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
					fmt.Printf(red+"segment1stLevel. Error writing to output file: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines\
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of level1Folders Segment----\n")
	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error writing header to output file: %v\n", err)
		os.Exit(1)
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Level 1 Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		//fmt.Printf("%s (URLs found: %d)\n", vc.Text, vc.Count)
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"segment1stLevel. Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"segment1stLevel. Error flushing writer: %v\n", err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"segment1stLevel. Error scanning input file: %v\n", err)
		os.Exit(1)
	}
}

// Regex for level 2 folders
func segmentLevel2() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"
	//ANSI escape code for red color
	red := "\033[0;31m"

	inputFilename := "siteurlsExport.csv"
	outputFilename := "segment.txt"

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	valueCounts := make(map[string]int)

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

		// Check if the line contains a quotation mark, if yes, skip to the next line
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
				valueCounts[text]++
			}
		}
	}

	//Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	fmt.Printf("\n")

	//Create a slice to hold ValueCount structs
	var sortedCounts []ValueCount

	//Populate the slice with data from the map
	for value, count := range valueCounts {
		sortedCounts = append(sortedCounts, ValueCount{value, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

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
		fmt.Printf(red+"segment2ndLevel. Error writing header to output file: %v\n", err)
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
					fmt.Printf(red+"segment2ndLevel. Error writing to output file: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of level2Folders Segment----\n")
	if err != nil {
		fmt.Printf(red+"segment2ndLevel. Error writing header to output file: %v\n", err)
		os.Exit(1)
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Level 2 Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		//fmt.Printf("%s (URLs found: %d)\n", vc.Text, vc.Count)
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"segment1stLevel. Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"segment2ndLevel. Error flushing writer: %v\n", err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"segment2ndLevel. Error scanning input file: %v\n", err)
		os.Exit(1)
	}
}

// Subdomains
// Regex for subdomains
func subDomains() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"
	//ANSI escape code for red color
	red := "\033[0;31m"

	inputFilename := "siteurlsExport.csv"
	outputFilename := "segment.txt"

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	valueCounts := make(map[string]int)

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

		// Display a block for each 1000 records scanned
		if recordCounter%1000 == 0 {
			fmt.Print("#")
		}

		// Check if the line contains a quotation mark, if yes, skip to the next line
		if strings.Contains(line, "\"") {
			continue
		}

		// Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")
		// Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			// Extract the text between the third and fourth forward slashes
			text := strings.Join(parts[:3], "/")

			// Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			// Update the count for this value if it's not empty
			if text != "" {
				// Update the count for this value if it's not empty
				valueCounts[text]++
			}
		}
	}

	//Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	fmt.Printf("\n")

	//Create a slice to hold ValueCount structs
	var sortedCounts []ValueCount

	//Populate the slice with data from the map
	for value, count := range valueCounts {
		sortedCounts = append(sortedCounts, ValueCount{value, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

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
		fmt.Printf(red+"subDomains. Error writing header to output file: %v\n", err)
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
					fmt.Printf(red+"subDomains. Error writing to output file: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of subDomains Segment----\n")
	if err != nil {
		fmt.Printf(red+"subDomains. Error writing header to output file: %v\n", err)
		os.Exit(1)
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----subDomains Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		//fmt.Printf("%s (URLs found: %d)\n", vc.Text, vc.Count)
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"subDomains. Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"subDomains. Error flushing writer: %v\n", err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"subDomains. Error scanning input file: %v\n", err)
		os.Exit(1)
	}
}

func parameterKeys() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"
	//ANSI escape code for red color
	red := "\033[0;31m"

	inputFilename := "siteurlsExport.csv"
	outputFilename := "segment.txt"

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()

	//Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	//Map to keep track of counts of unique values
	valueCounts := make(map[string]int)

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

		// Display a block for each 1000 records scanned
		if recordCounter%1000 == 0 {
			fmt.Print("#")
		}

		// Check if the line contains a quotation mark, if yes, skip to the next line
		if strings.Contains(line, "\"") {
			continue
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

	//Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	fmt.Printf("\n")

	//Create a slice to hold ValueCount structs
	var sortedCounts []ValueCount

	//Populate the slice with data from the map
	for value, count := range valueCounts {
		sortedCounts = append(sortedCounts, ValueCount{value, count})
	}

	//Sort the slice based on counts
	sort.Sort(ByCount(sortedCounts))

	//Open the file in append mode, create if it doesn't exist
	outputFile, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("\n\n[segment:sl_parameterKeys]\n@Home\npath /\n\n"))

	if err != nil {
		fmt.Printf(red+"parameterKeys. Error writing header to output file: %v\n", err)
		os.Exit(1)
	}

	//Write the regex
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s*\n\n", vc.Text, vc.Text))
		if err != nil {
			fmt.Printf(red+"parameterKeys. Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of parameterKeys Segment----\n")
	if err != nil {
		fmt.Printf(red+"parameterKeys. Error writing header to output file: %v\n", err)
		os.Exit(1)
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----parameterKeys URL analysis----\n")
	for _, vc := range sortedCounts {
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf(red+"parameterKeys. Error writing to output file: %v\n", err)
			os.Exit(1)
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf(red+"parameterKeys. Error flushing writer: %v\n", err)
		os.Exit(1)
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf(red+"parameterKeys. Error scanning input file: %v\n", err)
		os.Exit(1)
	}
}

// Regex hard coded regex entries
func insertStaticRegex(regexText string) error {

	//ANSI escape code for red color
	red := "\033[0;31m"

	outputFilename := "segment.txt"

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
		fmt.Printf(red+"parameterUsage. Error flushing writer: %v\n", err)
		return (err)
	}

	return (err)
}

// Function to display the banner
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

// Function to clear the screen
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
