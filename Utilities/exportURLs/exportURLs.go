// exportURLs: Export URLs from a project
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Version
var version = "v0.1"

// Token, log folder and cache folder taken from environment variables
var envBotifyAPIToken string
var envSegmentLogFolder string
var envSegmentFolder string

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
var yellow = "\033[0;33m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"
var lineSeparator = "█" + strings.Repeat("█", 130)

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
			fmt.Println(red+"Error. main. Cannot parse form:"+reset, err)
			return
		}
		organization = r.Form.Get("organization")
		project = r.Form.Get("project")

		// Generate a session ID used for grouping log entries
		sessionID, err := generateLogSessionID(8)

		if err != nil {
			fmt.Println(red+"Error. main. Failed generating session ID: %s"+reset, err)
			return
		}

		// Create the cache folder for the generated HTML if it does not exist
		cacheFolder = cacheFolderRoot + "/" + sessionID
		createCacheFolder()

		// Export the URLs
		dataStatus := generateExport(sessionID)

		// Manage errors
		// An invalid org/project name has been specified
		if dataStatus == "errorNoProjectFound" {
			writeLog(sessionID, organization, project, 0, "No project found")
			generateErrorPage("No project found. Try another organisation and project name. (" + organization + "/" + project + ")")
			http.Redirect(w, r, cacheFolder+"/"+"go_seo_exportURLs_error.html", http.StatusFound)
			return
		}

		// An invalid org/project name has been specified
		if dataStatus == "errorGenerateExport" {
			writeLog(sessionID, organization, project, 0, "No project found")
			generateErrorPage("Generate export error. (" + organization + "/" + project + ")")
			http.Redirect(w, r, cacheFolder+"/"+"go_seo_exportURLs_error.html", http.StatusFound)
			return
		}

		// Generate the page to display a preview of the exported URLs
		generatePreviewHTML()

		// Generate the export completion page
		generateCompletionPage()

		// Display the completion page
		http.Redirect(w, r, cacheFolder+"/go_seo_URLsExported.html", http.StatusFound)
	})

	// Start the HTTP server
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf(red+"Error. main. Failed to start HTTP server: %v\n"+reset, err)
	}

}

func generateExport(sessionID string) string {
	// Get the latest analysis
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", organization, project)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(red+"\nError. generateExport. Cannot create request:"+reset, err)
		return "errorGenerateExport"
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(red+"\nError. generateExport. Cannot send request:"+reset, err)
		return "errorGenerateExport"
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(red+"Error. executeBQL. Failed to close response body: %v\n"+reset, err)
		}
	}()

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(red+"\nError. generateExport. Cannot read request body:"+reset, err)
		return "errorGenerateExport"
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError. generateExport. Invalid credentials or no crawls found in the project" + reset)
		return "errorNoProjectFound"
	}

	if err != nil {
		fmt.Println(red+"\nError. generateExport. Cannot unmarshall JSON:"+reset, err)
		return "errorGenerateExport"
	}

	fmt.Println()
	fmt.Println(yellow + sessionID + purple + " Exporting URLs" + reset)
	fmt.Printf("\n%s%s%s Organisation: %s, Project: %s\n", yellow, sessionID, reset, organization, project)
	fmt.Println()

	// Create a file for writing
	file, err := os.Create(cacheFolder + "/exportedURLs.txt")
	if err != nil {
		fmt.Println(red+"\nError. generateExport. Cannot create output file:"+reset, err)
		return "errorGenerateExport"
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. generateExport. Failed to close file: %v\n", err)
			return
		}
	}()

	// Initialize total count
	totalCount := 0

	// Iterate through pages 1 through 10
	for page := 1; page <= 1000; page++ {
		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", organization, project, responseObject.Results[0].Slug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+envBotifyAPIToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"\nError. generateExport. Cannot connect to the API:"+reset, err)
			return "errorGenerateExport"
		}

		// Decode JSON response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println(red+"\nError. generateExport. Cannot decode JSON:"+reset, err)
			return "errorGenerateExport"
		}

		// Extract URLs from the "results" key
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Println(red + "\nError. generateExport. Invalid credentials or no crawls found in the project" + reset)
			return "errorGenerateExport"
		}

		// Write URLs to the file
		count := 0
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if url, ok := resultMap["url"].(string); ok {
					if _, err := file.WriteString(url + "\n"); err != nil {
						fmt.Println(red+"\nError. generateExport. Cannot write to the output file:"+reset, err)
						return "errorGenerateExport"
					}
					count++
					totalCount++
				}
			}
		}

		// If no URLs were saved for the page, exit the loop
		if count == 0 {
			break
		}

		fmt.Printf("%s%s%s Page %d: %d URLs exported\n", yellow, sessionID, reset, page, count)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println(red+"Error. executeBQL. Failed to close response body: %v\n"+reset, err)
		}
	}()

	// Print total number of URLs exported
	fmt.Printf("\nTotal URLs exported: %d\n"+reset, totalCount)

	writeLog(sessionID, organization, project, totalCount, "URL export complete")

	// We're done
	// Make a tidy display
	fmt.Println()
	fmt.Println(lineSeparator)

	// We're done
	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println("\nSession ID: " + yellow + sessionID + reset)
	fmt.Println("\nexportURLs: Done at " + formattedTime)
	fmt.Printf("\nOrganisation: %s, Project: %s\n"+reset, organization, project)

	// Make a tidy display
	fmt.Println()
	fmt.Println(lineSeparator)

	// Wait for the next request
	return "success"
}

// Generate the error page
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
    <title>exportURLs</title>
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
    <span style="font-size: 20px;">exportURLs</span>
</header>

<!-- Back Button -->
<button class="back-button" onclick="goHome()">Try again</button>

<!-- Error message -->
<div class="error-message" id="error-message">
    %s
</div>

<script>
    function goHome() {
        window.open('http://%s/', '_blank');
    }
</script>

</body>
</html>`, displayMessage, fullHost)

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_exportURLs_error.html")

}

// This page is presented when the export is complete
func generateCompletionPage() {

	// Using these two variables to replace width values in the HTML below because string interpolation confuses the percent signs as variables
	size80 := "80%"
	size100 := "100%"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>exportURLs</title>
 <style>
        body {
            margin: 0;
            font-family: Arial, sans-serif;
            background-color: Cornsilk;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
        }
        .banner {
            background-color: DeepSkyBlue;
            color: white;
            text-align: center;
            padding: 15px 0;
            position: absolute;
            top: 0;
            width: %s;
        }
        .banner.top {
            font-size: 24px;
        }
        .container {
            display: flex;
            flex-direction: row;
            justify-content: flex-start; 
            align-items: center;
            width: %s;
            height: %s;
            background-color: Khaki;
            color: white;
        }
        iframe {
            width: %s;
            height: %s;
            border: 2px solid LightGray;
            border-radius: 10px;
        }
        .no-border iframe {
            border: none;
        }
        .download-button {
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
        .download-button:hover {
            background-color: #0056b3;
            box-shadow: 0 6px 8px rgba(0, 0, 0, 0.15);
        }
       }
    </style>
</head>
<body>

<!-- Top Banner -->
<header class="banner top">
    <span>Go_Seo</span><br>
    <span style="font-size: 20px;">exportURLs</span>
</header>

<!-- Download Button for exportedURLs.txt -->
<button class="download-button" onclick="downloadFile()">Download export file</button>

<!-- Sections with iframes -->
<section class="container row no-border">
    <iframe src="go_seo_URLPreview.html" title="exportURLs"></iframe>
</section>

<script>
function downloadFile() {
    const fileName = 'exportedURLs.txt';
    const filePath = './' + fileName;  // Adjust this path to the actual location of the file
    const a = document.createElement('a');
    a.href = filePath;
    a.download = fileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    alert("Your URL export file has been downloaded.");
}
</script>

</body>
</html>
`, size100, size100, size100, size100, size80)

	// Save the HTML to a file
	saveHTML(htmlContent, "/go_seo_URLsExported.html")
}

// This page contains a preview of the first 100 URLs exported
func generatePreviewHTML() {

	file, err := os.Open(cacheFolder + "/exportedURLs.txt")
	if err != nil {
		fmt.Println(red+"Error. generateCompletionPage. Cannot open file:"+reset, err)
		return
	}

	scanner := bufio.NewScanner(file)
	var content []string
	lineCount := 0

	// Read up to the first 100 lines for preview purposes
	for scanner.Scan() {
		content = append(content, scanner.Text())
		lineCount++
		if lineCount >= 100 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(red+"Error. generatePreviewHTML. Cannot read file:"+reset, err)
		return
	}

	// Join the content with newlines
	finalContent := strings.Join(content, "\n")

	// HTML template with the content
	htmlContent := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Segment Content</title>
</head>
<body>
    <pre>%s</pre>
</body>
</html>`

	// Create the HTML file
	file, err = os.Create(cacheFolder + "/go_seo_URLPreview.html")
	if err != nil {
		fmt.Printf(red+"Error. generatePreviewHTML. Failed to create HTML file: %v"+reset, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(red+"Error. generatePreviewHTML. Closing (14):"+reset, err)
		}
	}()

	// Write the formatted HTML content to the file
	_, err = file.WriteString(
		fmt.Sprintf(htmlContent, finalContent),
	)
	if err != nil {
		fmt.Printf("Failed to write to HTML file: %v", err)
	}
}

// Create the cache folder
func createCacheFolder() {

	cacheDir := cacheFolder

	// Check if the directory already exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Create the directory and any necessary parents
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			fmt.Printf(red+"Error. createCacheFolder. Failed to create the cache directory: %v"+reset, err)
		}
	}
}

// Write the log entry
func writeLog(sessionID string, organization string, project string, URLCount int, statusDescription string) {

	// Define log file name
	fileName := "_exportURLs.log"

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
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%d,%s\n",
		sessionID, currentTime, organization, project, URLCount, statusDescription)

	// If the file doesn't exist, write header first
	if !fileExists {
		header := "SessionID,Date,Organisation,Project,URLCount,Status\n"
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
		fmt.Printf(red+"Error. getHostnamePort. Failed to read exportURLs.ini file: %v"+reset, err)
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

// Get environment variables for token and storage folders
func getEnvVariables() (envBotifyAPIToken string, envExportURLsLogFolder string, envExportURLsFolder string) {

	// Botify API token from the env. variable getbotifyenvBotifyAPIToken
	envBotifyAPIToken = os.Getenv("envBotifyAPIToken")
	if envBotifyAPIToken == "" {
		fmt.Println(red + "Error. getEnvVariables. envBotifyAPIToken environment variable is not set." + reset)
		fmt.Println(red + "Cannot start exportURLs server." + reset)
		os.Exit(0)
	}

	// Storage folder for the log file
	envExportURLsLogFolder = os.Getenv("envExportURLsLogFolder")
	if envExportURLsLogFolder == "" {
		fmt.Println(red + "Error. getEnvVariables. envExportURLsLogFolder environment variable is not set." + reset)
		fmt.Println(red + "Cannot start exportURLs server." + reset)
		os.Exit(0)
	} else {
		fmt.Println()
		fmt.Println(green + "Log folder: " + envExportURLsLogFolder + reset)
	}

	// Storage folder for the cached insights
	envExportURLsFolder = os.Getenv("envExportURLsFolder")
	if envExportURLsFolder == "" {
		fmt.Println(red + "Error. getEnvVariables. envExportURLsFolder environment variable is not set." + reset)
		fmt.Println(red + "Cannot start exportURLs server." + reset)
		os.Exit(0)
	} else {
		fmt.Println(green + "URL export folder: " + envExportURLsFolder + reset)
	}

	return envBotifyAPIToken, envExportURLsLogFolder, envExportURLsFolder
}

// Display the welcome banner
func displayBanner() {

	// Clear the screen
	fmt.Print(clearScreen)

	fmt.Print(green + `
 ██████╗  ██████╗         ███████╗███████╗ ██████╗ 
██╔════╝ ██╔═══██╗        ██╔════╝██╔════╝██╔═══██╗
██║  ███╗██║   ██║        ███████╗█████╗  ██║   ██║
██║   ██║██║   ██║        ╚════██║██╔══╝  ██║   ██║
╚██████╔╝╚██████╔╝███████╗███████║███████╗╚██████╔╝
 ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚══════╝ ╚═════╝`)
	fmt.Print(purple + `

███████╗██╗  ██╗██████╗  ██████╗ ██████╗ ████████╗██╗   ██╗██████╗ ██╗     ███████╗
██╔════╝╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝██║   ██║██╔══██╗██║     ██╔════╝
█████╗   ╚███╔╝ ██████╔╝██║   ██║██████╔╝   ██║   ██║   ██║██████╔╝██║     ███████╗
██╔══╝   ██╔██╗ ██╔═══╝ ██║   ██║██╔══██╗   ██║   ██║   ██║██╔══██╗██║     ╚════██║
███████╗██╔╝ ██╗██║     ╚██████╔╝██║  ██║   ██║   ╚██████╔╝██║  ██║███████╗███████║
╚══════╝╚═╝  ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝
`)

	fmt.Println()
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(green + "\nThe exportURLs server is ON\n" + reset)

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println(green + "Server started at " + formattedTime + reset)

	// Get the hostname and port
	getHostnamePort()

	// Get the environment variables for token, log folder & cache folder
	envBotifyAPIToken, envSegmentLogFolder, envSegmentFolder = getEnvVariables()

	fmt.Println(green + "\n... waiting for requests\n" + reset)
}
