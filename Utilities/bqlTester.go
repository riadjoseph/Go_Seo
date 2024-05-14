// bqlTester: Test Botify APIs
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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

// Version
var version = "v0.1"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"
var checkmark = "\u2713"

// Specify your Botify API token here
var botify_api_token = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

func main() {

	clearScreen()

	displayBanner()

	// Get the credentials if they have not been specified on the command line
	checkCredentials()

	// If the credentials have been provided on the command line use them
	if !credentialsInput {
		orgName = os.Args[1]
		projectName = os.Args[2]
	} else {
		orgName = orgNameInput
		projectName = projectNameInput
	}

	fmt.Println(bold+"\nOrganisation Name:", orgName)
	fmt.Println(bold+"Project Name:", projectName+reset)
	fmt.Println()

	displaySeparator()

	seoFunnel()

	bqlTesterDone()
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit apiTester" +
			"\n")

		fmt.Print(purple + "\nEnter Organisation Name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using listURLs. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter Project Name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using listURLs. Goodbye!\n")
			os.Exit(0)
		}
	}
}

func seoFunnel() {
	//Display the welcome message
	fmt.Println(purple + "\nGetting the latest funnel insights." + reset)

	// Get the latest analysis slug
	var latestSlug = getLatestSlug()

	// Define the indexable URLs BQL
	bqlIndexableUrls := fmt.Sprintf(`
	{
		"field": "crawl.%s.count_urls_crawl",
		"filters": {
			"and": [
				{
					"field": "crawl.%s.compliant.is_compliant",
					"predicate": "eq",
					"value": true
				}
			]
		}
	}`, latestSlug, latestSlug)

	// Define the non indexable URLs BQL
	bqlNonIndexableUrls := fmt.Sprintf(`
	{
		"field": "crawl.%s.count_urls_crawl",
		"filters": {
			"and": [
				{
					"field": "crawl.%s.compliant.is_compliant",
					"predicate": "eq",
					"value": false
				}
			]
		}
	}`, latestSlug, latestSlug)

	// Define the slow pages speed URLs (greater than 500ms)
	bqlSlowPageSpeedUrls := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.delay_last_byte",
                            "predicate": "gt",
                            "value": 500
                    },
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	bqlFewInlinksUrls := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.inlinks_internal.nb.unique",
                            "predicate": "gt",
                            "value": 500
                    },
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Define the deep links URLs BQL (greater than 4)
	bqlDeepUrls := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.depth",
                            "predicate": "gt",
                            "value": 5
                    },
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Bring the BQL fragments into a single query
	bqlFunnelBody := fmt.Sprintf(`
	{
		"collections": [
			"crawl.%s"
		],
		"query": {
			"dimensions": [],
			"metrics": [
				%s,
				%s,
				%s,
				%s,
				%s
			]
		}
	}`, latestSlug, bqlIndexableUrls, bqlNonIndexableUrls, bqlSlowPageSpeedUrls, bqlFewInlinksUrls, bqlDeepUrls)

	fmt.Println(bqlFunnelBody)
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(bqlFunnelBody)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error copying to clipboard:", err)
		return
	}

	/*
		url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query", orgName, projectName)
		fmt.Println("End point:", url, "\n")

		req, errorCheck := http.NewRequest("GET", url, nil)
		if errorCheck != nil {
			log.Fatal(red+"\nError. seoFunnel. Cannot create request: "+reset, errorCheck)
		}
		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", "token "+botify_api_token)
		req.Header.Add("httpbody" + bqlFunnelBody)

		res, errorCheck := http.DefaultClient.Do(req)
		if errorCheck != nil {
			log.Fatal(red+"\nError. seoFunnel. Check your network connection: "+reset, errorCheck)
		}
		defer res.Body.Close()

		responseData, errorCheck := ioutil.ReadAll(res.Body)
		if errorCheck != nil {
			log.Fatal(red+"\nError. seoFunnel. Cannot read request body: "+reset, errorCheck)
			os.Exit(1)
		}

		var responseObject botifyResponse
		errorCheck = json.Unmarshal(responseData, &responseObject)

		if errorCheck != nil {
			log.Fatal(red+"\nError. seoFunnel. Cannot unmarshall JSON: "+reset, errorCheck)
			os.Exit(1)
		}
	*/

}

func getLatestSlug() string {
	//Get the last analysis slug
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot create request: "+reset, errorCheck)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, errorCheck := http.DefaultClient.Do(req)
	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Check your network connection: "+reset, errorCheck)
	}
	defer res.Body.Close()

	responseData, errorCheck := ioutil.ReadAll(res.Body)
	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot read request body: "+reset, errorCheck)
		os.Exit(1)
	}

	var responseObject botifyResponse
	errorCheck = json.Unmarshal(responseData, &responseObject)

	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot unmarshall JSON: "+reset, errorCheck)
		os.Exit(1)
	}

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError: seoFunnel. Invalid credentials or no crawls found in the project")
		os.Exit(1)
	}

	fmt.Println("Organisation Name:", orgName)
	fmt.Println("Project Name:", projectName)
	fmt.Println("Latest analysis Slug:", responseObject.Results[0].Slug)

	return (responseObject.Results[0].Slug)
}

func bqlTesterDone() {

	// We're done
	fmt.Println(purple + "\nbqlTester: Done!\n")
	fmt.Println(bold + green + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)
}

// Display the welcome banner
func displayBanner() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Println(green + `

██████╗  ██████╗ ██╗  ████████╗███████╗███████╗████████╗███████╗██████╗ 
██╔══██╗██╔═══██╗██║  ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝██╔══██╗
██████╔╝██║   ██║██║     ██║   █████╗  ███████╗   ██║   █████╗  ██████╔╝
██╔══██╗██║▄▄ ██║██║     ██║   ██╔══╝  ╚════██║   ██║   ██╔══╝  ██╔══██╗
██████╔╝╚██████╔╝███████╗██║   ███████╗███████║   ██║   ███████╗██║  ██║
╚═════╝  ╚══▀▀═╝ ╚══════╝╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
`)

	//Display welcome message
	fmt.Println(purple+"Version:"+reset, version+"\n")

	fmt.Println(purple + "bqlTester: Test Botify BQL.\n" + reset)
	fmt.Println(purple + "Use it as a template for your Botify integration needs.\n" + reset)
	fmt.Println(purple + "BQL tests performed in this version.\n" + reset)
	fmt.Println(checkmark + green + bold + " Funnel statistics" + reset)
	fmt.Println(checkmark + green + bold + " Revenue" + reset)
	fmt.Println(checkmark + green + bold + " Visits" + reset)
	fmt.Println(checkmark + green + bold + " ActionBoard\n" + reset)
}

// Display the seperator

func displaySeparator() {
	block := "█"
	fmt.Println()

	for i := 0; i < 130; i++ {
		fmt.Print(block)
	}

	fmt.Println()
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
