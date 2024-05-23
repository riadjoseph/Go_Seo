// botifyBotLite: Generate Botify crawl en-masse
// Written by Jason Vicinanza

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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

// Strings used to store the project credentials for API access
var projectPrefix string
var urlCount int

// Strings used to store the input project credentials
var projectPrefixInput string
var urlCountInput int

// Boolean to signal if the project credentials have been entered by the user
var configInput = false

var err error

func main() {

	clearScreen()

	displayBanner()

	// Check to see if crawlme.txt exists. If not exit with an error
	checkCrawlmeTxt()

	// Get the crawl settings if they have not been specified on the command line
	checkCredentials()

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
	fmt.Println("\n")

	// Write a CSV file
	writeCSVContent()

	// Execute the bot.py script

	// Done

}

// Check of crawlme.txt exists. If not exit.
func checkCrawlmeTxt() {
	if _, err := os.Stat("crawlme.txt"); os.IsNotExist(err) {
		fmt.Printf(red + "\nError. botifyBotLite. No " + bold + "crawlme.txt" + reset + red + " found. Crawls cannot be generated.\n" + reset)
		os.Exit(1)
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
		fmt.Print("\nEnter the number corresponding to your choice: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

			switch input {
			case "1":
				fmt.Println("RevOps selected. Project:" + bold + " https://app.botify.com/revopsspeedworkers/")
				organizationLine = `ORGANIZATION = "revopsspeedworkers"`
				break
			case "2":
				fmt.Println("North EMEA selected. Project:" + bold + " https://app.botify.com/uk-crawl-prospect/")
				organizationLine = `ORGANIZATION = "uk-crawl-prospect"`
				break
			case "3":
				fmt.Println("South EMEA selected. Project:" + bold + " https://app.botify.com/crawl-prospect/")
				organizationLine = `ORGANIZATION = "crawl-prospect"`
				break
			case "4":
				fmt.Println("USA selected. Project:" + bold + " https://app.botify.com/us-crawl-prospect/")
				organizationLine = `ORGANIZATION = "us-crawl-prospect"`
				break
			default:
				fmt.Println("Invalid input. Please enter 1, 2, 3 or 4")
				continue
			}
			break
		}

		// Ask the user if they want to continue
		for {
			fmt.Print("Are you ready to generate the crawl now? (Y/N): ")
			contInput, _ := reader.ReadString('\n')
			contInput = strings.TrimSpace(strings.ToUpper(contInput))

			if contInput == "Y" {
				fmt.Println(bold + "Let's go!" + reset)
				break
			} else if contInput == "N" {
				fmt.Println(green + "\nThank you for using botifyBotLite. Goodbye!\n")
				os.Exit(0)
			} else {
				fmt.Println("Invalid input. Please enter Y or N.")
			}
		}

		// Generate the env file
		content := baseContent + organizationLine + "\n"

		// Create a file called env.py
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
}

// Write the content in the CSV file
func writeCSVContent() {
	file, err := os.OpenFile("crawlme.csv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf(red+"Error. botifyBotLite. Failed to open file: %s\n"+reset, err)
		os.Exit(1)
	}

	// Open the file
	file, err = os.Open("crawlme.txt")
	if err != nil {
		fmt.Printf(red+"Error. writeCSVContent. Failed to open crawlme.txt: %s"+reset, err)
	}

	// Create a new writer and write the header
	writer := csv.NewWriter(file)
	record := []string{"URL", "Project Name", "Max URLs"}
	err = writer.Write(record)
	if err != nil {
		fmt.Printf(red+"Error. writeCSVContent. Failed to write record to file: %s\n", err)
		return
	}

	/*
		// Write the content
		// Create a new scanner for the file
		scanner := bufio.NewScanner(file)
		// Read each line and print it
		for scanner.Scan() {
			record := scanner.Text()
			fmt.Println(record)
		}
	*/

	// Open the input file
	inputFile, err := os.Open("crawlme.txt")
	if err != nil {
		log.Fatalf("failed to open input file: %s", err)
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create("crawlme.csv")
	if err != nil {
		log.Fatalf("failed to create output file: %s", err)
	}
	defer outputFile.Close()

	// Create a CSV writer
	writer = csv.NewWriter(outputFile)
	defer writer.Flush()

	// Create a new scanner for the input file
	scanner := bufio.NewScanner(inputFile)

	// Read each line from the input file and write to the CSV file
	for scanner.Scan() {
		record := scanner.Text()
		fmt.Printf(record + "\n")
		newRecord := []string{record, projectPrefix + "_lbb", fmt.Sprintf("%d", urlCount)}
		err := writer.Write(newRecord)
		if err != nil {
			log.Fatalf("failed to write to CSV file: %s", err)
		}
	}

	// Check for errors during the scan
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading input file: %s", err)
	}

	// Check for errors during the scan
	if err := scanner.Err(); err != nil {
		log.Fatalf(red+"Error. Cannot read crawlme.txt: %s"+reset, err)
	}

	writer.Flush()
	defer file.Close()

}

// Check that the crawl settings have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {
		configInput = true
		fmt.Print("\nEnter your crawl settings. Press" + green + " Enter " + reset + "to exit botifyBotLite" +
			"\n")

		fmt.Print(purple + "\nEnter the project prefix: " + reset)
		fmt.Scanln(&projectPrefixInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectPrefixInput) == "" {
			fmt.Println(green + "\nThank you for using botifyBotLite. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter the no. of URLs to crawl: " + reset)
		fmt.Scanln(&urlCountInput)
		// Check if input is 0 if so exit
		if urlCountInput == 0 {
			fmt.Println(green + "\nThank you for using botifyBotLite. Goodbye!\n")
			os.Exit(0)
		}
		// Default to 100000 if the entered number of urls is than 100000
		// This is to ensure all crawls are are 100k URLs maximum
		if urlCountInput > 100000 {
			urlCountInput = 100000
		}
	}
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
	fmt.Println(purple+"Version:"+reset, version)
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
