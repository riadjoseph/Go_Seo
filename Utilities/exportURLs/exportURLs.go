// exportURLs: Export URLs to a text file called siteurlsExport.txt
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Version
var version = "v0.1"

// Specify your Botify API token here
var botify_api_token = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Declare the mutex
var mutex sync.Mutex

// Host name and port the web server runs on
var hostname string
var port string
var fullHost string

// Name of the cache folder used to store the generated HTML
var cacheFolder string
var cacheFolderRoot = "./_cache"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"

type botifyResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Slug string `json:"slug"`
	} `json:"results"`
}

// Used to store the project credentials for API access
var organization string
var project string

func main() {

	displayBanner()

	// Serve static files from the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Define a handler function for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {

		// Lock the function until it's complete to prevent race conditions
		mutex.Lock()
		defer mutex.Unlock()

		// Retrieve the form data from the request (org and username)
		err := r.ParseForm()
		if err != nil {
			fmt.Println(red+"Error. Cannot parse form:"+reset, err)
			return
		}
		organization = r.Form.Get("organization")
		project = r.Form.Get("project")

		fmt.Printf("\nOrganization: %s, Project: %s\n", organization, project)

		// Generate a session ID used for grouping log entries
		sessionID, err := generateLogSessionID(8)

		if err != nil {
			fmt.Println(red+"Error. writeLog. Failed generating session ID: %s"+reset, err)
			return
		}

		generateExport(sessionID)

	})

	// Start the HTTP server
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf(red+"Error. Main. Failed to start HTTP server: %v\n"+reset, err)
	}

}

// bloo

func generateExport(sessionID string) {
	// Get the latest analysis
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", organization, project)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError: Cannot create request:"+reset, err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError: Cannot send request:"+reset, err)
	}
	defer res.Body.Close()

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(red+"\nError: Cannot read request bodY:"+reset, err)
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project")
		os.Exit(1)
	}

	if err != nil {
		log.Fatal(red+"\nError: Cannot unmarshall JSON:"+reset, err)
		os.Exit(1)
	}

	fmt.Println(purple + "(" + sessionID + ") Exporting URLs" + reset)
	fmt.Println("Organisation name:", organization)
	fmt.Println("Project name:", project)
	fmt.Println("Analysis slug:", responseObject.Results[0].Slug)

	// Create a file for writing
	file, err := os.Create("siteurlsExport.txt")
	if err != nil {
		fmt.Println(red+"\nError: Cannot create output file:"+reset, err)
		os.Exit(1)
	}
	defer file.Close()

	// Initialize total count
	totalCount := 0

	// Iterate through pages 1 through 10
	for page := 1; page <= 1000; page++ {
		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", organization, project, responseObject.Results[0].Slug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+botify_api_token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"\nError: Cannot connect to the API:"+red, err)
			os.Exit(1)
		}
		defer res.Body.Close()

		// Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println(red+"\nError: Cannot decode JSON:"+reset, err)
			os.Exit(1)
		}

		// Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "\nError: Invalid credentials or no crawls found in the project" + reset)
			os.Exit(1)
		}

		// Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println(red+"\nError: Cannot write to the output file:"+reset, err)
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
			break
		}

		fmt.Printf("\nPage %d: %d URLs exported\n", page, count)
	}

	// Print total number of URLs saved
	fmt.Printf(purple+"\nTotal no. of URLs exported: %d\n"+reset, totalCount)

	// We're done
	return
}

// Define the error page
func generateErrorPage(displayMessage string) {

	// If displayMessage is empty or nil display a default error message.
	if displayMessage == "" {
		displayMessage = "An Unknown error has occurred"
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go_Seo Dashboard</title>
    <style>
        body {
            margin: 0;
            font-family: Arial, sans-serif;
            background-color: Cornsilk;
        }
        .banner {
            background-color: DeepSkyBlue;
            color: white;
            text-align: center;
            padding: 15px 0;
        }
        .banner.top {
            font-size: 24px;
        }
        .back-button {
            padding: 12px 24px;
            font-size: 18px;
            color: white;
            background-color: #007BFF;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            position: fixed;
            bottom: 20px;
            right: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            transition: background-color 0.3s, box-shadow 0.3s;
        }
        .back-button:hover {
            background-color: #0056b3;
            box-shadow: 0 6px 8px rgba(0, 0, 0, 0.15);
        }
		.error-container {
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        .error-message {
            color: red;
            font-weight: bold;
            text-align: center;
            padding: 100px;
        }
    </style>
</head>
<body>

<!-- Top Banner -->
<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">Business insights broadsheet (Beta)</span>
</header>

<!-- Back Button -->
<button class="back-button" onclick="goHome()">Try again</button>

<!-- Error message -->
<div class="error-message" id="error-message">
    %s
</div>

<script>
    function goHome() {
        window.open('http://localhost:8080/', '_blank');
    }
</script>

</body>
</html>`, displayMessage)

	// Save the HTML to a file
	saveHTML(htmlContent, "/seoBusinessInsights_error.html")

}

// Write the log entry
func writeLog(sessionID, organization, project, analyticsID, statusDescription string) {

	// Define log file name
	fileName := "_seoBusinessInsights.log"

	// Check if the log file exists
	fileExists := true
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fileExists = false
	}

	// Open or create the log file
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf(red+"Error. writeLog. Cannot open log file: %s"+reset, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf(red+"Error. writeLog. Failed to close log file: %v\n"+reset, err)
		}
	}()

	// Get current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Construct log record
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
		sessionID, currentTime, organization, project, analyticsID, statusDescription)

	// If the file doesn't exist, write header first
	if !fileExists {
		header := "SessionID,Date,Organisation,Project,AnalyticsID,Status\n"
		if _, err := file.WriteString(header); err != nil {
			fmt.Printf(red+"Error. writeLog. Failed to write log header: %s"+reset, err)
		}
	}

	// Write log record to file
	if _, err := file.WriteString(logRecord); err != nil {
		fmt.Printf(red+"Error. writeLog. Cannot write to log file: %s"+reset, err)
	}
}

// Function used to generate and save the HTML content to a file
func saveHTML(genHTML string, genFilename string) {

	file, err := os.Create(cacheFolder + genFilename)
	if err != nil {
		fmt.Println(red+"Error. saveHTML. Cannot create:"+reset, genFilename, err)
		return
	}

	//defer file.Close()
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. saveHTML. Failed to close file: %v\n"+reset, err)
			return
		}
	}()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fullDirectory := cacheFolder + genFilename
		fmt.Printf(red+"Error. saveHTML. Cannot write HTML file: %s"+reset, fullDirectory)
		fmt.Printf(red+"Error. saveHTML. Error %s:"+reset, err)
		return
	}
}

// Generate the sessionID
func generateLogSessionID(length int) (string, error) {

	// Generate random bytes
	sessionIDLength := make([]byte, length)
	if _, err := rand.Read(sessionIDLength); err != nil {
		return "", err
	}

	// Encode bytes to base64 string
	return base64.URLEncoding.EncodeToString(sessionIDLength), nil
}

func getHostnamePort() {

	// Load the INI file
	cfg, err := ini.Load("exportURLs.ini")
	if err != nil {
		fmt.Printf(red+"Error. getHostnamePort. Failed to read seoBusinessInsights.ini file: %v"+reset, err)
	}

	// Get values from the INI file
	hostname = cfg.Section("").Key("hostname").String()
	port = cfg.Section("").Key("port").String()
	port = cfg.Section("").Key("port").String()
	fullHost = hostname + ":" + port

	// Save the values to variables
	var serverHostname, serverPort string
	serverHostname = hostname
	serverPort = port

	// Display the hostname and port
	fmt.Printf(green+"\nHostname: %s\n"+reset, serverHostname)
	fmt.Printf(green+"Port: %s\n"+reset, serverPort)

}

// Display the welcome banner
func displayBanner() {

	// Clear the screen
	fmt.Print(clearScreen)

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Print(green + `
██╗   ██╗██████╗ ██╗     ███████╗██╗  ██╗██████╗  ██████╗ ██████╗ ████████╗
██║   ██║██╔══██╗██║     ██╔════╝╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝
██║   ██║██████╔╝██║     █████╗   ╚███╔╝ ██████╔╝██║   ██║██████╔╝   ██║   
██║   ██║██╔══██╗██║     ██╔══╝   ██╔██╗ ██╔═══╝ ██║   ██║██╔══██╗   ██║   
╚██████╔╝██║  ██║███████╗███████╗██╔╝ ██╗██║     ╚██████╔╝██║  ██║   ██║   
 ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝
`)

	fmt.Println()
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(purple + "\nexportURLs server.\n" + reset)
	fmt.Println(green + "\nThe exportURLs server is ON.\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Server started at " + formattedTime + reset)

	// Get the hostname and port
	getHostnamePort()

	fmt.Println(green + "\n... waiting for requests\n" + reset)
}
