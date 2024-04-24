// listURLs: Export all URLs to a textfile called siteurlsExport.
// Maximum 1MM URLs will be exported
// Written by Jason Vicinanza

// To run this:
// go run listURLs
// Example: go run listURLs

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {

	//ANSI escape code for purple color
	purple := "\033[0;35m"
	//ANSI escape code to reset color
	reset := "\033[0m"

	clearScreen()

	//Display the welcome message
	fmt.Println(purple + "listURLs: Exporting URLs." + reset)

	// Create a file for writing
	file, err := os.Create("siteurlsExport.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Initialize total count
	totalCount := 0

	// Iterate through pages 1 through 10
	for page := 1; page <= 100; page++ {
		url := fmt.Sprintf("https://api.botify.com/v1/analyses/jason-org/rolex-project-name/20240421/urls?area=current&page=%d&size=1000", page)
		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "Token c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer res.Body.Close()

		// Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		// Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println("Results not found in response")
			return
		}

		// Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println("Error writing to file:", err)
						return
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
			fmt.Printf(purple+"\nTotal No. of URLs exported is %d.\n", page)

			break
		}

		fmt.Printf("\nPage %d: %d URLs exported\n", page, count)
	}

	// Print total number of URLs saved
	fmt.Printf("\nTotal URLs exported: %d\n", totalCount)
}

// Function to clear the screen
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
