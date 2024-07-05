// listURLs: Export URLs to a text file called siteurlsExport.txt
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

// Version
var version = "v0.1"

// Specify your Botify API token here
var APIToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

// Colours
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var bold = "\033[1m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"

type botifyResponse struct {
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Count    int         `json:"count"`
	Results  []struct {
		User struct {
			Login          string      `json:"login"`
			Email          string      `json:"email"`
			IsOrganization bool        `json:"is_organization"`
			URL            string      `json:"url"`
			DateJoined     string      `json:"date_joined"`
			Status         *string     `json:"status,omitempty"`
			FirstName      string      `json:"first_name"`
			LastName       string      `json:"last_name"`
			CompanyName    interface{} `json:"company_name"`
		} `json:"user"`
		Owner struct {
			Login          string      `json:"login"`
			Email          string      `json:"email"`
			IsOrganisation bool        `json:"is_organisation"`
			URL            string      `json:"url"`
			DateJoined     string      `json:"date_joined"`
			Status         interface{} `json:"status"`
			FirstName      string      `json:"first_name"`
			LastName       string      `json:"last_name"`
			CompanyName    interface{} `json:"company_name"`
		} `json:"owner"`
		Slug              string      `json:"slug"`
		Name              string      `json:"name"`
		FriendlyName      interface{} `json:"friendly_name"`
		URL               string      `json:"url"`
		Status            string      `json:"status"`
		ComputingRevision interface{} `json:"computing_revision"`
		Features          struct {
			Js struct {
				Version int `json:"version"`
			} `json:"js"`
			Rel struct {
				ProcessRelAmp       bool `json:"process_rel_amp"`
				ProcessRelApp       bool `json:"process_rel_app"`
				ProcessRelAlternate bool `json:"process_rel_alternate"`
				ProcessRelCanonical bool `json:"process_rel_canonical"`
				ProcessRelPrevNext  bool `json:"process_rel_prev_next"`
			} `json:"rel"`
			Main struct {
				Lang                          bool  `json:"lang"`
				HasSw                         *bool `json:"has_sw,omitempty"`
				ProcessDevice                 bool  `json:"process_device"`
				CompliantExcludeBadCanonicals bool  `json:"compliant_exclude_bad_canonicals"`
			} `json:"main"`
			Links struct {
				Chains              bool `json:"chains"`
				PageRank            bool `json:"page_rank"`
				PrevNext            bool `json:"prev_next"`
				LinksGraph          bool `json:"links_graph"`
				TopAnchors          bool `json:"top_anchors"`
				TopDomains          bool `json:"top_domains"`
				LinksToNoindex      bool `json:"links_to_noindex"`
				LinksSegmentGraph   bool `json:"links_segment_graph"`
				LinksToNonCanonical bool `json:"links_to_non_canonical"`
			} `json:"links"`
			Scoring struct {
				Version       int  `json:"version"`
				ActionsHash   int  `json:"actions_hash"`
				ActionsCount  int  `json:"actions_count"`
				HaveMlActions bool `json:"have_ml_actions"`
			} `json:"scoring"`
			Segments struct {
				Flags  []string `json:"flags"`
				Names  []string `json:"names"`
				Values []struct {
					Name  string `json:"name"`
					Field string `json:"field"`
				} `json:"values"`
				DateCreated string `json:"date_created"`
			} `json:"segments"`
			Sitemaps struct {
				Urls           []string `json:"urls"`
				DateRetrieved  string   `json:"date_retrieved"`
				HasOrphansArea bool     `json:"has_orphans_area"`
			} `json:"sitemaps"`
			MainImage     interface{} `json:"main_image"`
			SearchConsole struct {
				DateEnd   interface{} `json:"date_end"`
				DateStart interface{} `json:"date_start"`
			} `json:"search_console"`
			ContentQuality struct {
				Samples bool `json:"samples"`
			} `json:"content_quality"`
			SemanticMetadata struct {
				Length         bool `json:"length"`
				StructuredData struct {
					Stats struct {
						Offer      int `json:"offer"`
						Address    int `json:"address"`
						Product    int `json:"product"`
						Breadcrumb int `json:"breadcrumb"`
					} `json:"stats"`
					Versions struct {
						Qa             string `json:"qa"`
						Car            string `json:"car"`
						Faq            string `json:"faq"`
						Book           string `json:"book"`
						News           string `json:"news"`
						Dates          string `json:"dates"`
						Event          string `json:"event"`
						Movie          string `json:"movie"`
						Offer          string `json:"offer"`
						Course         string `json:"course"`
						Person         string `json:"person"`
						Rating         string `json:"rating"`
						Recipe         string `json:"recipe"`
						Review         string `json:"review"`
						Address        string `json:"address"`
						Product        string `json:"product"`
						AudioBook      string `json:"audio_book"`
						Breadcrumb     string `json:"breadcrumb"`
						Restaurant     string `json:"restaurant"`
						TrainTrip      string `json:"train_trip"`
						JobPosting     string `json:"job_posting"`
						VideoObject    string `json:"video_object"`
						EducationEvent string `json:"education_event"`
					} `json:"versions"`
					Currencies struct {
						Offer []string `json:"offer"`
					} `json:"currencies"`
				} `json:"structured_data"`
			} `json:"semantic_metadata"`
			DuplicateQueryKvs bool `json:"duplicate_query_kvs"`
		} `json:"features"`
		UrlsDone    int `json:"urls_done"`
		UrlsInQueue int `json:"urls_in_queue"`
		Config      struct {
			MaxUrls          int         `json:"max_urls"`
			MaxUrlsPerSec    int         `json:"max_urls_per_sec"`
			MaxDepth         interface{} `json:"max_depth"`
			VirtualRobotsTxt interface{} `json:"virtual_robots_txt"`
			AllowedDomains   []struct {
				Domain          string `json:"domain"`
				Mobile          bool   `json:"mobile"`
				Protocol        string `json:"protocol"`
				UserAgent       string `json:"user_agent"`
				AllowSubdomains bool   `json:"allow_subdomains"`
			} `json:"allowed_domains"`
			BlacklistedDomains []string `json:"blacklisted_domains"`
			StartUrls          []string `json:"start_urls"`
			StartUrlsURL       []string `json:"start_urls_url"`
			ExportLimit        int      `json:"export_limit"`
		} `json:"config"`
		DateLaunched                string      `json:"date_launched"`
		DateFinished                string      `json:"date_finished"`
		DateLastModified            string      `json:"date_last_modified"`
		DateCreated                 string      `json:"date_created"`
		DateCrawlDone               string      `json:"date_crawl_done"`
		Failures                    []string    `json:"failures"`
		RedButtonDomain             interface{} `json:"red_button_domain"`
		ImportKeywordsData          bool        `json:"import_keywords_data"`
		ImportKeywordsDataByCountry bool        `json:"import_keywords_data_by_country"`
		CrawlLaunchType             string      `json:"crawl_launch_type"`
		ToBeDeletedAt               string      `json:"to_be_deleted_at"`
		Comparable                  bool        `json:"comparable"`
		ExcludedFromTrends          bool        `json:"excluded_from_trends"`
		Pk                          int         `json:"pk"`
		HasRawPages                 bool        `json:"has_raw_pages"`
	} `json:"results"`
	Page int `json:"page"`
	Size int `json:"size"`
}

// Strings used to store the project credentials for API access
var orgName string
var projectName string

// Strings used to store the input project credentials
var orgNameInput string
var projectNameInput string

// Boolean to signal if the project credentials have been entered by the user
var credentialsInput = false

func main() {

	displayBanner()

	// Get the project credentials if they have not been specified on the command line
	checkCredentials()

	// If the credentials have been provided on the command line use them
	if !credentialsInput {
		orgName = os.Args[1]
		projectName = os.Args[2]
	} else {
		orgName = orgNameInput
		projectName = projectNameInput
	}

	// Get the latest analysis
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError: Cannot create request:"+reset, err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+APIToken)

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

	fmt.Println(purple + "\nExporting URLs" + reset)
	fmt.Println("Organisation name:", orgName)
	fmt.Println("Project name:", projectName)
	fmt.Println("Analysis slug:", responseObject.Results[0].Slug)
	urlEndpoint := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/", orgName, projectName, responseObject.Results[0].Slug)
	fmt.Println("End point:", urlEndpoint, "\n")

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
		url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s/%s/urls?area=current&page=%d&size=1000", orgName, projectName, responseObject.Results[0].Slug, page)

		payload := strings.NewReader("{\"fields\":[\"url\"]}")

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("Authorization", "token "+APIToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(red+"\nError: Cannot connect to the API:"+red, err)
			os.Exit(1)
		}

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

	defer res.Body.Close()

	// Print total number of URLs saved
	fmt.Printf(purple+"\nTotal no. of URLs exported: %d\n"+reset, totalCount)

	// Prompt to save generated list in the clipboard
	copyURLsToClipboard()

	// We're done
	fmt.Println(purple + "\nlistURLs: Done\n")
	fmt.Println(green + bold + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit listURLs" +
			"\n")

		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using listURLs. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using listURLs. Goodbye!\n")
			os.Exit(0)
		}
	}
}

func copyURLsToClipboard() {

	var response string

	validInputs := map[string]bool{"Y": true, "N": true}

	for {
		fmt.Print("\nCopy exported URLs to the clipboard? (Y/N): ")
		fmt.Scanln(&response)
		response = strings.ToUpper(strings.TrimSpace(response))

		if _, valid := validInputs[response]; valid {
			break
		} else {
			fmt.Println("Invalid input. Please enter Y or N.")
		}
	}

	if response == "Y" {
		content, err := ioutil.ReadFile("siteurlsExport.txt")
		if err != nil {
			panic(err)
		}

		// Copy the content to the clipboard using pbcopy (macOS) or type to clip (Windows)
		var copyCmd string

		switch runtime.GOOS {
		case "windows":
			copyCmd = "type siteurlsExport.txt | clip"
		default:
			copyCmd = "pbcopy"
		}

		cmd := exec.Command(copyCmd)
		cmd.Stdin = strings.NewReader(string(content))
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		fmt.Println(green + bold + "\nThe exported URLs can be found in siteurlsExport.txt" + reset)
		fmt.Println(green + bold + "The URLs have also been copied to your clipboard" + reset)
	} else {
		fmt.Println(green + bold + "\nThe exported URLs can be found in siteurlsExport.txt" + reset)
	}
}

// Display the welcome banner
func displayBanner() {

	// Clear the screen
	fmt.Print(clearScreen)

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Print(green + `
██╗     ██╗███████╗████████╗██╗   ██╗██████╗ ██╗     ███████╗
██║     ██║██╔════╝╚══██╔══╝██║   ██║██╔══██╗██║     ██╔════╝
██║     ██║███████╗   ██║   ██║   ██║██████╔╝██║     ███████╗
██║     ██║╚════██║   ██║   ██║   ██║██╔══██╗██║     ╚════██║
███████╗██║███████║   ██║   ╚██████╔╝██║  ██║███████╗███████║
╚══════╝╚═╝╚══════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝
`)
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(purple + "listURLs: Export URLs from a project\n" + reset)
}
