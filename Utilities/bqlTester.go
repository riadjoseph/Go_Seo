// bqlTester: Test Botify APIs
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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

// Array used to store the name of all collections found in collectionsApiTest. This array is used in collectionsDetailsApiTest
var collectionIdentifiers []string

// API STRUCTS
// The structs are defined in the order they are used in bqlTester

// DatasourceResponse represents the JSON structure
type datasourceResponse struct {
	SiteMapInfos map[string]SiteMapInfo `json:"sitemaps"`
}

type SiteMapInfo struct {
	Sitemaps struct {
		Runnable                   bool         `json:"runnable"`
		Datasource                 string       `json:"datasource"`
		DateLastSuccessfulRevision string       `json:"date_last_successful_revision"`
		LastRevisionStatus         string       `json:"last_revision_status"`
		Stats                      StatsInfo    `json:"stats"`
		Segments                   SegmentsInfo `json:"segments"`
	} `json:"sitemaps"`
}

type StatsInfo struct {
	Linkrels           int         `json:"Linkrels"`
	ValidUrls          int         `json:"ValidUrls"`
	InvalidUrls        int         `json:"InvalidUrls"`
	FileUploaded       int         `json:"FileUploaded"`
	UploadErrors       int         `json:"UploadErrors"`
	ExecutionTime      string      `json:"ExecutionTime"`
	ParsingErrors      int         `json:"ParsingErrors"`
	DownloadErrors     int         `json:"DownloadErrors"`
	SitemapsTreated    int         `json:"SitemapsTreated"`
	DownLoadErrorsUrls interface{} `json:"DownLoadErrorsUrls"`
}

type SegmentsInfo struct {
	Flags   []interface{} `json:"flags"`
	Names   []string      `json:"names"`
	Version int           `json:"version"`
	ID      int           `json:"id,omitempty"`
}

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
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits listURLs
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
