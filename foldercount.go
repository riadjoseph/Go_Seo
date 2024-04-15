// foldercount: A small utility that counts the number of instances of the first level folder
// Written by Jason Vicinanza
// First Github commit: 15/4/24

// To run this:
// go run foldercount.go file_name
// Example: go run foldercount.go siteurls.csv

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
	if len(os.Args) < 2 {
		fmt.Println("Please provide the filename as an argument.")
		return
	}
	filename := os.Args[1]

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("foldercount: Error opening file: %v\n", err)
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
	fmt.Println(purple + "foldercount: Count the number of first level folders found." + reset)

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

	// Subtract 2 in order to account for the two header records which are defaults in Botify URL extracts
	totalRecords -= 2

	// Clear the screen
	clearScreen()

	// Display welcome message
	fmt.Println(purple + "foldercount: Count the number First Level Folders found." + reset)

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

	fmt.Println(purple + "\nfoldercount: Done\n")

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("foldercount: Error scanning extract file: %v\n", err)
		return
	}
}

// Function to clear the screen
func clearScreen() {
	cmd := exec.Command("clear") // Use "cls" for Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}
