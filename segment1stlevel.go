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

	// ANSI escape code for purple color
	purple := "\033[0;35m"
	// ANSI escape code to reset color
	reset := "\033[0m"

	// Clear the screen
	clearScreen()

	// Get the filename from the command-line arguments
	if len(os.Args) < 3 {
		clearScreen()
		fmt.Println("foldercount")
		fmt.Println("Error: Please provide the input filename and output filename as arguments.")
		return
	}
	inputFilename := os.Args[1]
	outputFilename := os.Args[2]

	// Declare a variable to store the domain name
	var domain string

	// Open the input file
	file, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("segment1stlevel: Error opening input file: %v\n", err)
		return
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
	fmt.Println(purple + "segment1stlevel: Generate the regex for all first level folders." + reset)

	// Iterate through each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		totalRecords++
		recordCounter++

		// Display a block for each 10000 records scanned
		if recordCounter%10000 == 0 {
			fmt.Print("#")
		}

		// Call extractDomain when reading the third record
		//if recordCounter == 3 {
		domain = extractDomain(line)
		//if domain != "" {
		//fmt.Println("Domain extracted:", domain)
		// You can save the domain in a variable for later use if needed
		//}
		//}

		// Split the line into substrings using a forward slash as delimiter
		parts := strings.Split(line, "/")

		// Check if there are at least 4 parts in the line
		if len(parts) >= 4 {
			// Extract the text between the third and fourth forward slashes
			text := parts[3]

			// Trim any leading or trailing whitespace
			text = strings.TrimSpace(text)

			// Update the count for this value if it's not empty
			if text != "" {
				valueCounts[text]++
			}
		}
	}

	// Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	// Clear the screen
	clearScreen()

	// Display welcome message
	fmt.Println(purple + "segment1stlevel: Count the number First Level Folders found." + reset)

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

	// Open the output file for writing
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf("foldercount: Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	// Create a writer to write to the output file
	writer := bufio.NewWriter(outputFile)

	// Write the header lines
	_, err = writer.WriteString(fmt.Sprintf("# Regex made with Go_SEO/segmentfirstlevel for domain: %s\n\n[segment:level1Folders]\n@Home\npath /\n\n", domain))
	if err != nil {
		fmt.Printf("segment1stlevel: Error writing header to output file: %v\n", err)
		return
	}

	// Write the counts for each unique value to the output file
	for _, vc := range sortedCounts {
		if vc.Text != "" {
			var folderLabel = vc.Text
			_, err := writer.WriteString(fmt.Sprintf("@%s\nurl *%s/%s/*\n\n", folderLabel, domain, vc.Text))
			if err != nil {
				fmt.Printf("segment1stlevel: Error writing to output file: %v\n", err)
				return
			}
		}
	}

	// Write the footer lines
	_, err = writer.WriteString("\n@~Other\npath /*\n# ----End of level1Folders Segment----\n")
	if err != nil {
		fmt.Printf("segment1stlevel: Error writing header to output file: %v\n", err)
		return
	}

	// Flush the writer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		fmt.Printf("segment1stlevel: Error flushing writer: %v\n", err)
		return
	}

	fmt.Println(purple + "\nsegment1stlevel: Done\n")

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("segment1stlevel: Error scanning input file: %v\n", err)
		return
	}
}

// Function to clear the screen
func clearScreen() {
	cmd := exec.Command("clear") // Use "cls" for Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Function to extract domain from URL
func extractDomain(url string) string {
	parts := strings.Split(url, "//")
	if len(parts) >= 2 {
		domainParts := strings.Split(parts[1], "/")
		if len(domainParts) >= 1 {
			return domainParts[0]
		}
	}
	return ""
}
