//segment1stLevel: Generate the regex for all first level folders found in the URL extract
//Written by Jason Vicinanza

//To run this:
//go run segmentifyLite.go url_extract regex_output
//Example: go run segmentifyLite.go siteurls.csv myregex.txt

//Version
//version := "v0.1"

package main

import (
	"bufio"
	"fmt"
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

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	//Functions to create the various regex statements

	//Level1 folders
	segmentLevel1()

	//Level2 folders
	segmentLevel2()

	//Subdomains
	subDomains()

	//Pages containing parameters
	paramaterUsageRegex := `

[segment:sl_parameter_Usage]
@Parameters
query *=*

@Clean
path /*
# ----End of sl_parameter_Usage----`

	//Parameter usage message
	fmt.Println(purple + "paramaterUsage: Regex for pages with/without parameters." + reset)
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
	fmt.Println(purple + "noOfParameters: Regex for number of parameters on the URL." + reset)
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
	fmt.Println(purple + "noOfParameters: Regex for number of folders on the URL." + reset)
	errFolderNoRegex := insertStaticRegex(folderNoRegex)
	if errFolderNoRegex != nil {
		panic(errFolderNoRegex)
	}

	fmt.Println("\nsegmentiftyLite. Regex generation complete.")

}

// Regex for level 1 folders
func segmentLevel1() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	//Clear the screen
	clearScreen()

	//Get the filename from the command-line arguments
	if len(os.Args) < 2 {
		clearScreen()
		fmt.Println("segmentifyLite (level1). Error. Please provide the input filename (URL Extract file) as an argument.")
		return
	}
	inputFilename := os.Args[1]
	outputFilename := "segment.txt"

	fmt.Println("segmentiftyLite. Generating segmentation regex for all Level 1 and 2 folders.\n")

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("segmentiftyLite. Error opening input file: %v\n", err)
		return
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
	fmt.Println(purple + "segment1stLevel: Regex for all first level folders." + reset)

	//Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		//Check the first record of the specified URL Extract file
		//If the file does not start with "sep=", consider the file invalid
		if recordCounter == 1 && !strings.HasPrefix(line, "sep=") {
			fmt.Printf("segmentifyLite. Error. Invalid URL extract file specified.\n")
			os.Exit(1)
		}

		//Display a block for each 10000 records scanned
		if recordCounter%10000 == 0 {
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

	//Display the total number of records processed
	fmt.Printf("\nTotal URLs processed: %d\n", totalRecords)
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
		fmt.Printf("segment1stLevel. Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	//Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	//Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("# Regex made with Go_SEO/segmentifyLite (level1)\n\n[segment:sl_level1_Folders]\n@Home\npath /\n\n"))

	if err != nil {
		fmt.Printf("segment1stLevel. Error writing header to output file: %v\n", err)
		return
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
					fmt.Printf("segment1stLevel. Error writing to output file: %v\n", err)
					return

				}
			}
		}
	}

	//Write the footer lines\
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of level1Folders Segment----\n")
	if err != nil {
		fmt.Printf("segment1stLevel. Error writing header to output file: %v\n", err)
		return
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Level 1 Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		//fmt.Printf("%s (URLs found: %d)\n", vc.Text, vc.Count)
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf("segment1stLevel. Error writing to output file: %v\n", err)
			return
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf("segment1stLevel. Error flushing writer: %v\n", err)
		return
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("segment1stLevel. Error scanning input file: %v\n", err)
		return
	}
}

// Regex for level 2 folders
func segmentLevel2() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	inputFilename := os.Args[1]
	outputFilename := "segment.txt"

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		return
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
	fmt.Println(purple + "\nsegment2ndLevel: Regex for all second level folders." + reset)

	//Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		//Display a block for each 10000 records scanned
		if recordCounter%10000 == 0 {
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

	//Display the total number of records processed
	fmt.Printf("\nTotal URLs processed: %d\n", totalRecords)
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
		fmt.Printf("segment2ndLevel. Error writing header to output file: %v\n", err)
		return
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
					fmt.Printf("segment2ndLevel. Error writing to output file: %v\n", err)
					return
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of level2Folders Segment----\n")
	if err != nil {
		fmt.Printf("segment2ndLevel. Error writing header to output file: %v\n", err)
		return
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----Level 2 Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		//fmt.Printf("%s (URLs found: %d)\n", vc.Text, vc.Count)
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf("segment1stLevel. Error writing to output file: %v\n", err)
			return
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf("segment2ndLevel. Error flushing writer: %v\n", err)
		return
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("segment2ndLevel. Error scanning input file: %v\n", err)
		return
	}
}

// Subdomains
// Regex for subdomains
func subDomains() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	inputFilename := os.Args[1]
	outputFilename := "segment.txt"

	//Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		return
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
	fmt.Println(purple + "\nsubDomains: Regex for subdomains." + reset)

	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		// Display a block for each 10000 records scanned
		if recordCounter%10000 == 0 {
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

	//Display the total number of records processed
	fmt.Printf("\nTotal URLs processed: %d\n", totalRecords)
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
		fmt.Printf("subDomains. Error writing header to output file: %v\n", err)
		return
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
					fmt.Printf("subDomains. Error writing to output file: %v\n", err)
					return
				}
			}
		}
	}

	//Write the footer lines
	_, err = writer.WriteString("@~Other\npath /*\n# ----End of subDomains Segment----\n")
	if err != nil {
		fmt.Printf("subDomains. Error writing header to output file: %v\n", err)
		return
	}

	// Insert the number of URLs found in each folder as comments
	_, err = writer.WriteString("\n# ----subDomains Folder URL analysis----\n")
	for _, vc := range sortedCounts {
		//fmt.Printf("%s (URLs found: %d)\n", vc.Text, vc.Count)
		_, err := writer.WriteString(fmt.Sprintf("# --%s (URLs found: %d)\n", vc.Text, vc.Count))
		if err != nil {
			fmt.Printf("subDomains. Error writing to output file: %v\n", err)
			return
		}
	}

	//Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf("subDomains. Error flushing writer: %v\n", err)
		return
	}

	//Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("subDomains. Error scanning input file: %v\n", err)
		return
	}
}

// Regex hard coded regex entries
func insertStaticRegex(regexText string) error {

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
		fmt.Printf("parameterUsage. Error flushing writer: %v\n", err)
		return (err)
	}

	return (err)
}

// Function to clear the screen
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
