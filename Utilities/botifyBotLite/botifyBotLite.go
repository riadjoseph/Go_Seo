// botifyBotLite: Generate Botify crawl en-masse
// Written by Jason Vicinanza

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Version
var version = "v0.1"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"

// Strings used to store the project credentials for API access
var projectPrefix string
var urlCount int

// Strings used to store the input project credentials
var projectPrefixInput string
var urlCountInput int

// Boolean to signal if the project credentials have been entered by the user
var configInput = false

// Counter for the number of crawls to generate
var noCrawlsToGenerate = 0

// Used to store the project slug (region) based on the user selection
var projectSlug = ""

// Error detection
var err error

// No of URLs found in crawlme.txt
var urlsFound = 0

func main() {

	clearScreen()

	displayBanner()

	// Check to see if crawlme.txt exists. If not exit with an error
	checkCrawlmeTxt()

	// Check the formatting of the URLs in crawlme.txt
	validateCrawlmeTxt()

	// Get the crawl settings if they have not been specified on the command line
	checkCrawlParameters()

	// If the crawl settings have been provided on the command line use them
	if !configInput {
		projectPrefix = os.Args[1]
		urlCount, err = strconv.Atoi(os.Args[2])
	} else {
		projectPrefix = projectPrefixInput
		urlCount = urlCountInput
	}

	// Generate env.py.
	generateEnv()

	fmt.Println(purple + "\nGenerating crawls" + reset)
	fmt.Println("Project prefix name:", projectPrefix)
	fmt.Println("No. URLs to crawl:", urlCount)

	// Write a CSV file
	writeCSVContent()

	// Execute the bot.py script
	executeBotPY()

}

// Check of crawlme.txt exists. If not exit.
func checkCrawlmeTxt() {
	if _, err := os.Stat("./crawlme.txt"); os.IsNotExist(err) {
		fmt.Printf(red + "\nError. checkCrawlmeTxt. No " + bold + "crawlme.txt" + reset + red + " found. Crawls cannot be generated.\n" + reset)
		os.Exit(1)
	}
}

func validateCrawlmeTxt() {
	// Open the file
	file, err := os.Open("crawlme.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a new scanner
	scanner := bufio.NewScanner(file)

	// Flag to track if there are any invalid lines
	foundInvalid := false

	for scanner.Scan() {
		line := scanner.Text()
		// Ignore lines that start with # (comments)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, "https://") {
			// If a line does not start with "https://"
			foundInvalid = true
			fmt.Println("URL incorrectly formatted:", line)
		}
		urlsFound++

	}

	if err := scanner.Err(); err != nil {
		fmt.Println(red+"Error. validateCrawlmeTxt. Cannot read crawlme.txt:"+reset, err)
		os.Exit(1)
	}

	// Exit with error code 0 if any invalid lines were found
	if foundInvalid {
		fmt.Println(red + "\nCheck your crawlme.txt file, correct the formatting of the URLs above and try again. Remember all URLs must start with " + bold + "https://" + reset)
		os.Exit(0)
	} else {
		fmt.Println(green + "\ncrawlme.txt validated successfully\n" + reset)
		fmt.Printf("\nNo. of sites to crawl: %d\n", urlsFound)
	}
}

// Generate the env.py file used to determine which org the crawls will be created into
func generateEnv() {
	var organizationLine string

	// Default content of env.py
	baseContent := `TOKEN = "3a6f5c03366a24140f8614d8f26346856ecbf86e"
API_URL = "https://api.botify.com"
APP_URL = "https://app.botify.com"
DEBUG = True
USERNAME = "botifyprospectcrawls@gmail.com"
USERNAME2 = "botifyprospectcrawls"
PASSWORD = "BotifyParis75!"
`

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(purple + "\nWhich project organization should be used to store the generated crawls?" + reset)
		fmt.Println("\n1. RevOps")
		fmt.Println("2. North EMEA")
		fmt.Println("3. South EMEA")
		fmt.Println("4. USA")
		fmt.Print("\nSelect your regional project: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			fmt.Println("\nRevOps selected. Project:" + bold + " https://app.botify.com/revopsspeedworkers/")
			organizationLine = `ORGANIZATION = "revopsspeedworkers"`
			projectSlug = "revopsspeedworkers"
			break
		case "2":
			fmt.Println("\nNorth EMEA selected. Project:" + bold + " https://app.botify.com/uk-crawl-prospect/")
			organizationLine = `ORGANIZATION = "uk-crawl-prospect"`
			projectSlug = "uk-crawl-prospect"
			break
		case "3":
			fmt.Println("\nSouth EMEA selected. Project:" + bold + " https://app.botify.com/crawl-prospect/")
			organizationLine = `ORGANIZATION = "crawl-prospect"`
			projectSlug = "crawl-prospect"
			break
		case "4":
			fmt.Println("\nUSA selected. Project:" + bold + " https://app.botify.com/us-crawl-prospect/")
			organizationLine = `ORGANIZATION = "us-crawl-prospect"`
			projectSlug = "us-crawl-prospect"
			break
		default:
			fmt.Println("Invalid input. Please enter 1, 2, 3 or 4")
			continue
		}
		break
	}

	// Ask the user if they want to continue
	for {
		fmt.Print("\nAre you ready to generate the crawls now? (Y/N): ")
		generateCrawlsYN, _ := reader.ReadString('\n')
		generateCrawlsYN = strings.TrimSpace(strings.ToUpper(generateCrawlsYN))

		if generateCrawlsYN == "Y" {
			clearScreen()
			displayBanner()
			fmt.Println(green + bold + "\nLet's go!" + reset)
			break
		} else if generateCrawlsYN == "N" {
			fmt.Println(green + "\nThank you for using botifyBotLite. Goodbye!\n")
			os.Exit(0)
		} else {
			fmt.Println("Invalid input. Please enter Y or N.")
		}
	}

	// Generate the env.py file
	content := baseContent + organizationLine + "\n"
	file, err := os.Create("env.py")
	if err != nil {
		fmt.Printf("Failed to create file: %s\n", err)
		return
	}
	defer file.Close()

	// Write the content to the file
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Failed to write to file: %s\n", err)
		return
	}
}

// Write the content in the CSV file
func writeCSVContent() {
	// Create crawlme.csv
	file, err := os.OpenFile("crawlme.csv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf(red+"Error. writeCSVContent. Failed to open file: %s\n"+reset, err)
		os.Exit(1)
	}

	// Create project_list.txt
	projectListFile, err := os.OpenFile("project_list.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf(red+"Error. writeCSVContent. Failed to open or create project_list.txt: %s"+reset, err)
	}
	defer projectListFile.Close()

	// Open crawlme.txt
	file, err = os.Open("crawlme.txt")
	if err != nil {
		fmt.Printf(red+"Error. writeCSVContent. Failed to open crawlme.txt: %s"+reset, err)
	}

	// Create a new writer and write the header in crawlme.csv
	writer := csv.NewWriter(file)
	record := []string{"URL", "Project Name", "Max URLs"}
	err = writer.Write(record)
	if err != nil {
		fmt.Printf(red+"Error. writeCSVContent. Failed to write record to file: %s\n", err)
		return
	}

	// Create a writer for project_list.txt
	projectListWriter := bufio.NewWriter(projectListFile)
	defer projectListWriter.Flush()

	// Open the input file
	inputFile, err := os.Open("crawlme.txt")
	if err != nil {
		log.Fatalf(red+"Error. writeCSVContent. Failed to open crawlme.txt: %s"+reset, err)
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create("crawlme.csv")
	if err != nil {
		log.Fatalf(red+"Error. writeCSVContent. Failed to create crawlme.csv: %s"+reset, err)
	}
	defer outputFile.Close()

	// Create a writer for the CSV file
	writer = csv.NewWriter(outputFile)
	defer writer.Flush()

	// Create a new scanner for crawlme.txt
	scanner := bufio.NewScanner(inputFile)

	// Read each line from crawlme.txt and write to crawlme.csv
	for scanner.Scan() {
		record := scanner.Text()

		// Skip lines starting with # (comments)
		//if strings.HasPrefix(record, "#") {
		//	continue
		//}

		// Ensure the record ends with a "/", if it does not end with a "/" add one
		if !strings.HasSuffix(record, "/") {
			record += "/"
		}

		// Extract the domain from the record
		domain := extractDomain(record)

		// Combine the fragments to make the record in crawlme.csv
		newRecord := []string{record, projectPrefix + "_" + domain + "__bbl", fmt.Sprintf("%d", urlCount)}
		err := writer.Write(newRecord)

		// Build and write the record to project_list.txt
		if !strings.HasPrefix(record, "#") {
			_, err := projectListWriter.WriteString(record + "," + "https://app.botify.com/" + projectSlug + "/" + projectPrefix + "_" + domain + "__bbl" + "\n")
			if err != nil {
				log.Fatalf(red+"Error. writeCSVContent. Failed to write to project_list.txt: %s"+reset, err)
			}
		}

		writer.Flush()
		projectListWriter.Flush()

		noCrawlsToGenerate += 1

		if err != nil {
			log.Fatalf(red+"Error. writeCSVContent. Failed to write to crawlme.csv: %s"+reset, err)
		}
	}

	defer file.Close()

	// Calculate the run duration of the script and round it up
	noCrawlsToGenerate = noCrawlsToGenerate - 1
	fmt.Println("No. crawls to generate:"+reset, noCrawlsToGenerate)
	fmt.Printf("Your crawls will be available in the following project: https://app.botify.com/%s\n", projectSlug)

	// Each crawl should take approx. 40 seconds to complete
	estimatedRunTime := float64(noCrawlsToGenerate) * 40 / 60
	roundedRunTime := math.Ceil(estimatedRunTime)
	fmt.Printf("Estimated time to generate all crawls is %.0f minutes\n", roundedRunTime)
	// Display the current time
	currentTime := time.Now()
	formattedTime := currentTime.Format("15:04")
	fmt.Println("Started at:", formattedTime+"\n")

	fmt.Println("\n")
	fmt.Printf(bold + "The crawls are currently being generated. Information indicating the progress of crawl generation will be displayed in a moment.\n" + reset)
	fmt.Printf(bold + "\nPlease stand-by.\n" + reset)
	fmt.Println("\n")
}

// Used to extract the domain from the record
func extractDomain(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 4 {
		return parts[2]
	}
	return ""
}

// Check that the crawl settings have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCrawlParameters() {
	if len(os.Args) < 3 {
		configInput = true
		fmt.Print("\n\nEnter your crawl settings. Press" + green + " Enter " + reset + "to exit botifyBotLite" +
			"\n")
		fmt.Print(purple + "\nEnter the project prefix: " + reset)
		fmt.Scanln(&projectPrefixInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectPrefixInput) == "" {
			fmt.Println(green + "\nThank you for using botifyBotLite. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter the No. of URLs to crawl: " + reset)
		fmt.Scanln(&urlCountInput)

		// Check if input is 0 if so exit
		if urlCountInput == 0 {
			fmt.Println(green + "\nThank you for using botifyBotLite. Goodbye!\n")
			os.Exit(0)
		}
		// Default to 100000 if the entered number of urls is than 100000
		// This is to ensure all crawls are 100k URLs maximum
		if urlCountInput > 100000 {
			urlCountInput = 100000
		}
	}
}

// Execute the python script bot.py
func executeBotPY() {
	// Get the current directory, this is done to ensure that the scripts do not have to be in the PATH
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(red+"Error. executeBotPY. Cannot determine current directory:"+reset, err)
		return
	}

	// Construct the path to bot.py
	botPath := filepath.Join(currentDir, "bot.py")

	// Create the command with the full path to bot.py
	cmd := exec.Command("python3", botPath, "-i", "crawlme.csv")

	// Set the command's stdout and stderr to the program's stdout and stderr to ensure that the script output is displayed in real time
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Launch the bot!
	err = cmd.Run()
	if err != nil {
		fmt.Printf(red+"Error. executeBotPY. Cannot execute bot.py: %s\n"+reset, err)
		return
	}

	// For security reasons delete env.py
	os.Remove("env.py")
	// Keep things tidy. Delete crawlme.csv
	os.Remove("crawlme.csv") //bloo

	fmt.Println(green + "\nCrawl generation process complete!\n" + reset)

	// Display the current time
	currentTime := time.Now()
	formattedTime := currentTime.Format("15:04")
	fmt.Println(green+"botifyBotLite finished at:"+reset, formattedTime+"\n")

	fmt.Println("\nThe start pages crawled and the generated Botify project URL can be found in" + bold + " projects_list.txt" + reset)
	fmt.Println(bold + "\nThank you for using botifyBotLite. Goodbye!\n" + reset)
}

// Display the welcome banner
func displayBanner() {
	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Println(green + `
██████╗  ██████╗ ████████╗██╗███████╗██╗   ██╗██████╗  ██████╗ ████████╗██╗     ██╗████████╗███████╗
██╔══██╗██╔═══██╗╚══██╔══╝██║██╔════╝╚██╗ ██╔╝██╔══██╗██╔═══██╗╚══██╔══╝██║     ██║╚══██╔══╝██╔════╝
██████╔╝██║   ██║   ██║   ██║█████╗   ╚████╔╝ ██████╔╝██║   ██║   ██║   ██║     ██║   ██║   █████╗  
██╔══██╗██║   ██║   ██║   ██║██╔══╝    ╚██╔╝  ██╔══██╗██║   ██║   ██║   ██║     ██║   ██║   ██╔══╝  
██████╔╝╚██████╔╝   ██║   ██║██║        ██║   ██████╔╝╚██████╔╝   ██║   ███████╗██║   ██║   ███████╗
╚═════╝  ╚═════╝    ╚═╝   ╚═╝╚═╝        ╚═╝   ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝╚═╝   ╚═╝   ╚══════╝
`)

	//Display welcome message
	fmt.Println(purple + "botifyBotLite: Generate and launch Botify crawls, en masse!\n" + reset)
	fmt.Println(purple+"Version:"+reset, version+"\n")
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
