// projectInfo: Get key crawl attributes for the latest crawl of a specific project
// Written by Jason Vicinanza

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Version
var version = "v0.1"

// Token, log folder and cache folder taken from environment variables
var envBotifyAPIToken string
var envProjectInfoLogFolder string
var envProjectInfoFolder string

// Host name and port the web server runs on
var hostname string
var port string
var fullHost string

// Colours
var purple = "\033[0;35m"
var blue = "\033[34m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var reset = "\033[0m"
var clearScreen = "\033[H\033[2J"

// No of executions
var sessionIDCounter int

// Strings used to store the project credentials for API access
var organisation string
var project string

// Name of the cache folder used to store the generated HTML
var cacheFolder string
var cacheFolderRoot string

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

func main() {

	startUp()

	// Serve static files from the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Define a handler function for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the form data from the request
		err := r.ParseForm()
		if err != nil {
			fmt.Println("Error. Cannot parse form:", err)
			return
		}
		organisation = r.Form.Get("organization")
		project = r.Form.Get("project")

		// Generate a session ID used for grouping log entries
		var sessionID string

		sessionID, err = generateSessionID(8)
		if err != nil {
			fmt.Println("Error. writeLog. Failed generating session ID: %s", err)
		}

		cacheFolderRoot = envProjectInfoFolder
		cacheFolder = cacheFolderRoot + "/" + sessionID + organisation

		createCacheFolder()

		// Process URLs
		dataStatus := acquireProjectSettings(sessionID)

		// Manage errors
		// An invalid org/project name has been specified
		if dataStatus == "errorNoProjectFound" {
			writeLog(sessionID, organisation, project, "No project found")
			generateErrorPage("No project found. Try another organisation and project name. (" + organisation + "/" + project + ")")
			http.Redirect(w, r, cacheFolder+"/"+"go_seo_projectInfoError.html", http.StatusFound)
			return
		}

		writeLog(sessionID, organisation, project, "Project loaded")
	})

	// Start the HTTP server
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error. Main. Failed to start HTTP server: %v\n", err)
	}
}

// Acquire the project information
func acquireProjectSettings(sessionID string) string {

	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", organisation, project)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("\nError. acquireProjectSettings. Invalid credentials or no crawls found in the project (1)")
		return "errorNoProjectFound"
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+envBotifyAPIToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("\nError. acquireProjectSettings. Invalid credentials or no crawls found in the project (2)")
		return "errorNoProjectFound"
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("\nError. acquireProjectSettings. Invalid credentials or no crawls found in the project (3)")
		return "errorNoProjectFound"
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		log.Fatal("\nError: Cannot unmarshall JSON:", err)
	}

	fmt.Println("\nOrganisation name:", organisation)
	fmt.Println("Project name:", project)

	addLineToHTML("Organisation: "+organisation, true, true)
	addLineToHTML("Project: "+project, true, true)

	// Return an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println("\nError. acquireProjectSettings. Invalid credentials or no crawls found in the project (4)")
		return "errorNoProjectFound"
	}

	// User details
	user := responseObject.Results[0].User
	addLineToHTML("User", true, true)
	addLineToHTML("User Login: "+user.Login, false, true)
	addLineToHTML("User Email: "+user.Email, false, true)
	//addLineToHTML("Is Organization: "+user.IsOrganization, false, true)
	addLineToHTML("URL: "+user.URL, false, true)
	addLineToHTML("Date Joined: "+user.DateJoined, false, true)
	addLineToHTML("First Name: "+user.FirstName, false, true)
	addLineToHTML("Last Name: "+user.LastName, false, true)
	//addLineToHTML("Company Name: "+user.CompanyName, false, true)

	// Owner details
	owner := responseObject.Results[0].Owner
	addLineToHTML("Owner", true, true)
	addLineToHTML("Owner: "+owner.Login, false, true)
	addLineToHTML("Email: "+owner.Email, false, true)
	addLineToHTML("Is Organization: "+fmt.Sprintf("%t", user.IsOrganization), false, true)
	addLineToHTML("URL: "+owner.URL, false, true)
	addLineToHTML("Date Joined: "+owner.DateJoined, false, true)
	addLineToHTML("First Name: "+owner.FirstName, false, true)
	addLineToHTML("Last Name: "+owner.LastName, false, true)
	//addLineToHTML("Company Name: "+user.CompanyName, false, true)

	// Crawl details
	addLineToHTML("Crawl Details", true, true)
	addLineToHTML("Analysis Slug: "+responseObject.Results[0].Slug, false, true)
	addLineToHTML("Friendly Name: "+fmt.Sprintf("%v", responseObject.Results[0].FriendlyName), false, true) //check interface

	addLineToHTML("URL: "+responseObject.Results[0].URL, false, true)
	addLineToHTML("Status: "+responseObject.Results[0].Status, false, true)
	//addLineToHTML("Computing Revision:: "+responseObject.Results[0].ComputingRevision, false, true)

	//Crawl Configuration
	addLineToHTML("Crawl Configuration", true, true)
	addLineToHTML("MaxUrls: "+fmt.Sprintf("%d", responseObject.Results[0].Config.MaxUrls), false, true) //check

	addLineToHTML("Crawl Speed: "+fmt.Sprintf("%d", responseObject.Results[0].Config.MaxUrlsPerSec), false, true)
	addLineToHTML("Max Depth: "+fmt.Sprintf("%v", responseObject.Results[0].Config.MaxDepth), false, true) //check

	//addLineToHTML("Virtual Robots: "+responseObject.Results[0].Config.VirtualRobotsTxt, false, true)

	// Crawls
	addLineToHTML("Crawled URLs", true, true)
	addLineToHTML("Crawl Schedule: "+responseObject.Results[0].CrawlLaunchType, false, true)
	addLineToHTML("Latest URLs Crawled: "+fmt.Sprintf("%d", responseObject.Results[0].UrlsDone), false, true) //check
	addLineToHTML("URLs in Queue: "+fmt.Sprintf("%d", responseObject.Results[0].UrlsInQueue), false, true)    //check

	// ActionBoard
	fmt.Println("\nActionBoard") // Title
	addLineToHTML("ActionBoard", true, true)
	//addLineToHTML("No. Recommendations:: "+responseObject.Results[0].Features.Scoring.ActionsCount, false, true)

	//Allowed domains
	addLineToHTML("Allowed Domains", true, true)

	for _, AllowedDomains := range responseObject.Results[0].Config.AllowedDomains {
		addLineToHTML("Domain: "+AllowedDomains.Domain, false, true)
		//addLineToHTML("Mobile: "+AllowedDomains.Mobile, false, true)
		addLineToHTML("Protocol: "+AllowedDomains.Protocol, false, true)
		addLineToHTML("User Agent: "+AllowedDomains.UserAgent, false, true)
		//addLineToHTML("Allow SubDomains: "+AllowedDomains.AllowSubdomains, false, true)
	}

	//Start URLs
	addLineToHTML("Start URLs", true, true)
	for _, StartUrls := range responseObject.Results[0].Config.StartUrls {
		fmt.Println("URLs:", StartUrls)
		addLineToHTML("URLs: "+StartUrls, false, true)

	}
	//addLineToHTML("Export Limit: "+responseObject.Results[0].Config.ExportLimit, false, true)
	addLineToHTML("Date Launched: "+responseObject.Results[0].DateLaunched, false, true)
	addLineToHTML("Date Finished: "+responseObject.Results[0].DateFinished, false, true)
	addLineToHTML("Date Last Modified: "+responseObject.Results[0].DateLastModified, false, true)
	addLineToHTML("Date Created: "+responseObject.Results[0].DateCreated, false, true)
	addLineToHTML("Date Crawl Done: "+responseObject.Results[0].DateCrawlDone, false, true)
	for _, failure := range responseObject.Results[0].Failures {
		fmt.Println("Failure:", failure)
		addLineToHTML("Failure: "+failure, false, true)
	}

	// Blacklisted domains
	addLineToHTML("Blacklisted Domains, if any", true, true)
	for _, BlacklistedDomains := range responseObject.Results[0].Config.BlacklistedDomains {
		addLineToHTML("Domain: "+BlacklistedDomains, false, true)
	}

	// Segments
	addLineToHTML("Segments", true, true)
	addLineToHTML("Date Created: "+responseObject.Results[0].Features.Segments.DateCreated, false, true)
	//fmt.Println("Flags:", responseObject.Results[0].Features.Segments.Flags)
	//fmt.Println("Segment Names:", responseObject.Results[0].Features.Segments.Names)
	// Iterate over Values and print Name and Field
	for _, segment := range responseObject.Results[0].Features.Segments.Values {
		addLineToHTML(segment.Name, false, true)
		//fmt.Println("Field:", segment.Field)
	}

	// Sitemaps
	addLineToHTML("Sitemaps", true, true)
	// Iterate over values and print the URLs
	for _, url := range responseObject.Results[0].Features.Sitemaps.Urls {
		addLineToHTML("URL: "+url, false, true)
	}

	addLineToHTML("Date Retrieved: "+responseObject.Results[0].Features.Sitemaps.DateRetrieved, false, true)
	//addLineToHTML("Has Orphans Area: "+responseObject.Results[0].Features.Sitemaps.HasOrphansArea, false, true)

	//Search console
	addLineToHTML("Search Console", true, true)
	//addLineToHTML("Date Start: "+responseObject.Results[0].Features.SearchConsole.DateStart, false, true)
	//addLineToHTML("Date End: "+responseObject.Results[0].Features.SearchConsole.DateEnd, false, true)

	// Additional crawl attributes
	addLineToHTML("FEATURES", true, true)

	//Rel
	addLineToHTML("Rel", true, true)
	//addLineToHTML("ProcessRelAmp: "+responseObject.Results[0].Features.Rel.ProcessRelAmp, false, true)
	//addLineToHTML("ProcessRelAlternate: "+responseObject.Results[0].Features.Rel.ProcessRelAlternate, false, true)
	//addLineToHTML("ProcessRelCanonical: "+responseObject.Results[0].Features.Rel.ProcessRelCanonical, false, true)
	//addLineToHTML("ProcessRelPrevNext: "+responseObject.Results[0].Features.Rel.ProcessRelPrevNext, false, true)

	//Main
	//addLineToHTML("Main", true, true)
	//addLineToHTML("Lang: "+responseObject.Results[0].Features.Main.Lang, false, true)
	//addLineToHTML("ProcessDevice: "+responseObject.Results[0].Features.Main.ProcessDevice, false, true)
	//addLineToHTML("CompliantExcludeBadCanonicals: "+responseObject.Results[0].Features.Main.CompliantExcludeBadCanonicals, false, true)

	//Links
	//addLineToHTML("Links", true, true)
	//addLineToHTML("Chains: "+responseObject.Results[0].Features.Links.Chains, false, true)
	//addLineToHTML("PageRank: "+responseObject.Results[0].Features.Links.PageRank, false, true)
	//addLineToHTML("PrevNext: "+responseObject.Results[0].Features.Links.PrevNext, false, true)
	//addLineToHTML("LinksGraph: "+responseObject.Results[0].Features.Links.LinksGraph, false, true)
	//addLineToHTML("TopAnchors: "+responseObject.Results[0].Features.Links.TopAnchors, false, true)
	//addLineToHTML("TopDomains: "+responseObject.Results[0].Features.Links.TopDomains, false, true)
	//addLineToHTML("LinksToNoindex: "+responseObject.Results[0].Features.Links.LinksToNoindex, false, true)
	//addLineToHTML("LinksSegmentGraph: "+responseObject.Results[0].Features.Links.LinksSegmentGraph, false, true)
	//addLineToHTML("LinksToNonCanonical: "+responseObject.Results[0].Features.Links.LinksToNonCanonical, false, true)
	//addLineToHTML("Version: "+responseObject.Results[0].Features.Scoring.Version, false, true)
	//addLineToHTML("ActionsHash: "+responseObject.Results[0].Features.Scoring.ActionsHash, false, true)
	//addLineToHTML("No. Actionboard Recos: "+responseObject.Results[0].Features.Scoring.ActionsCount, false, true)
	//addLineToHTML("HaveMlActions: "+responseObject.Results[0].Features.Scoring.HaveMlActions, false, true)
	//addLineToHTML("MainImage: "+responseObject.Results[0].Features.MainImage, false, true)

	//Content quality
	addLineToHTML("Content Quality", true, true)
	//addLineToHTML("Samples: "+responseObject.Results[0].Features.ContentQuality.Samples, false, true)

	addLineToHTML("Semantic Metadata", true, true)
	//addLineToHTML("Length: "+responseObject.Results[0].Features.SemanticMetadata.Length, false, true)
	//addLineToHTML("Address: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Stats.Address, false, true)
	//addLineToHTML("Product: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Stats.Product, false, true)
	//addLineToHTML("Breadcrumb: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Stats.Breadcrumb, false, true)
	addLineToHTML("Version: QA: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Qa, false, true)
	addLineToHTML("Version: Car: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Car, false, true)
	addLineToHTML("Version: FAQ: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Faq, false, true)
	addLineToHTML("Version: Book: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Book, false, true)
	addLineToHTML("Version: News: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.News, false, true)
	addLineToHTML("Version: Dates: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Dates, false, true)
	addLineToHTML("Version: Offer: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Offer, false, true)
	addLineToHTML("Version: Course: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Course, false, true)
	addLineToHTML("Version: Person: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Person, false, true)
	addLineToHTML("Version: Rating: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Rating, false, true)
	addLineToHTML("Version: Recipe: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Recipe, false, true)
	addLineToHTML("Version: Review: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Review, false, true)
	addLineToHTML("Version: Address: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Address, false, true)
	addLineToHTML("Version: Product: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Product, false, true)
	addLineToHTML("Version: AudioBook: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.AudioBook, false, true)
	addLineToHTML("Version: Breadcrumb: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Breadcrumb, false, true)
	addLineToHTML("Version: Restaurant: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Restaurant, false, true)
	addLineToHTML("Version: TrainTrip: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.TrainTrip, false, true)
	addLineToHTML("Version: JobPosting: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.JobPosting, false, true)
	addLineToHTML("Version: VideoObject: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.VideoObject, false, true)
	addLineToHTML("Version: EducationEvent: "+responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.EducationEvent, false, true)

	addLineToHTML("Currency", true, true)
	for _, currency := range responseObject.Results[0].Features.SemanticMetadata.StructuredData.Currencies.Offer {
		addLineToHTML("Currency Offer: "+currency, false, true)
	}
	//addLineToHTML("DuplicateQueryKvs: "+responseObject.Results[0].Features.DuplicateQueryKvs, false, true)
	//addLineToHTML("RedButtonDomain: "+responseObject.Results[0].RedButtonDomain, false, true)
	//addLineToHTML("ImportKeywordsData: "+responseObject.Results[0].ImportKeywordsData, false, true)
	//addLineToHTML("ImportKeywordsDataByCountry: "+responseObject.Results[0].ImportKeywordsDataByCountry, false, true)
	//addLineToHTML("CrawlLaunchType: "+responseObject.Results[0].CrawlLaunchType, false, true)
	//addLineToHTML("ToBeDeletedAt: "+responseObject.Results[0].ToBeDeletedAt, false, true)
	//addLineToHTML("Comparable: "+responseObject.Results[0].Comparable, false, true)
	//addLineToHTML("ExcludedFromTrends: "+responseObject.Results[0].ExcludedFromTrends, false, true)
	//addLineToHTML("Pk: "+responseObject.Results[0].Pk, false, true)
	//addLineToHTML("HasRawPages: "+responseObject.Results[0].HasRawPages, false, true)

	return "success"
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
    <title>projectInfo</title>
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
    <span style="font-size: 20px;">projectInfo</span>
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
	saveHTML(htmlContent, "/go_seo_projectInfoError.html")

}

// Function used to generate and save the HTML content to a file
func saveHTML(genHTML string, genFilename string) {

	file, err := os.Create(cacheFolder + genFilename)
	if err != nil {
		fmt.Printf("Error. saveHTML. Can create %s: "+"%s\n", genFilename, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error. saveHTML. Closing:", err)
			return
		}
	}()

	_, err = file.WriteString(genHTML)
	if err != nil {
		fmt.Printf("Error. saveHTML. Can write %s: "+"%s\n", genFilename, err)
		return
	}
}

func writeLog(sessionID, organisation, project, statusDescription string) {

	// Define log file name
	//fileName := "_seoprojectInfo.log"
	fileName := envProjectInfoLogFolder + "/_projectInfo.log"

	// Check if the log file exists
	fileExists := true
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fileExists = false
	}

	// Open or create the log file
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error. writeLog. Cannot oprn log file: %s", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error. writeLog. Closing:", err)
		}
	}()

	// Get current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Construct log record
	logRecord := fmt.Sprintf("%s,%s,%s,%s,%s\n",
		sessionID, currentTime, organisation, project, statusDescription)

	// If the file doesn't exist, write header first
	if !fileExists {
		header := "SessionID,Date,Organisation,Project,Status\n"
		if _, err := file.WriteString(header); err != nil {
			log.Fatalf("Error. writeLog. Failed to write log header: %s", err)
		}
	}

	// Write log record to file
	if _, err := file.WriteString(logRecord); err != nil {
		log.Fatalf("Error. writeLog. Cannot write to log file: %s", err)
	}
}

// Create the cache folder
func createCacheFolder() {

	println("folder")
	cacheDir := cacheFolder

	// Check if the directory already exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Create the directory and any necessary parents
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			log.Fatalf("Error. Failed to create the cache directory: %v", err)
		}
	}
}

func generateSessionID(length int) (string, error) {
	// Generate random sessionID
	sessionID := make([]byte, length)
	if _, err := rand.Read(sessionID); err != nil {
		return "", err
	}

	// Add to the execution increment
	sessionIDCounter++

	var builder strings.Builder
	builder.WriteString(strconv.Itoa(sessionIDCounter))
	builder.WriteString("-")
	builder.WriteString(base64.URLEncoding.EncodeToString(sessionID))

	// Convert the builder to a string and return
	return builder.String(), nil
}

// Get the hostname and port from the .ini
func getHostnamePort() {

	// Load the INI file
	cfg, err := ini.Load("projectInfo.ini")
	if err != nil {
		log.Fatalf("Error. getHostnamePort. Failed to read go_seo_projectInfo.ini file: %v", err)
	}

	// Get values from the .ini file
	hostname = cfg.Section("").Key("hostname").String()
	port = cfg.Section("").Key("port").String()
	fullHost = hostname + ":" + port

	// Save the values to variables
	var serverHostname, serverPort string
	serverHostname = hostname
	serverPort = port

	// Print the values (for demonstration purposes)
	fmt.Printf("\nHostname: %s\n", serverHostname)
	fmt.Printf("Port: %s\n", serverPort)
}

// Get environment variables for token and storage folders
func getEnvVariables() (envBotifyAPIToken string, envProjectInfoLogFolder string, envProjectInfoFolder string) {

	// Botify API token from the env. variable getbotifyenvBotifyAPIToken
	envBotifyAPIToken = os.Getenv("envBotifyAPIToken")
	if envBotifyAPIToken == "" {
		fmt.Println("Error. getEnvVariables. envBotifyAPIToken environment variable is not set.")
		fmt.Println("Cannot start projectInfo server.")
		os.Exit(0)
	}

	// Storage folder for the log file
	envProjectInfoLogFolder = os.Getenv("envProjectInfoLogFolder")
	if envProjectInfoLogFolder == "" {
		fmt.Println("Error. getEnvVariables. envProjectInfoLogFolder environment variable is not set.")
		fmt.Println("Cannot start projectInfo server.")
		os.Exit(0)
	} else {
		fmt.Println()
		fmt.Println("Log folder: " + envProjectInfoLogFolder)
	}

	// Storage folder for the cached insights
	envProjectInfoFolder = os.Getenv("envProjectInfoFolder")
	if envProjectInfoFolder == "" {
		fmt.Println("Error. getEnvVariables. envProjectInfoFolder environment variable is not set.")
		fmt.Println("Cannot start projectInfo server.")
		os.Exit(0)
	} else {
		fmt.Println("projectInfo cache folder: " + envProjectInfoFolder)
	}

	return envBotifyAPIToken, envProjectInfoLogFolder, envProjectInfoFolder
}

func startUp() {

	// Clear the screen
	fmt.Print(clearScreen)

	fmt.Print(`
 ██████╗  ██████╗         ███████╗███████╗ ██████╗ 
██╔════╝ ██╔═══██╗        ██╔════╝██╔════╝██╔═══██╗
██║  ███╗██║   ██║        ███████╗█████╗  ██║   ██║
██║   ██║██║   ██║        ╚════██║██╔══╝  ██║   ██║
╚██████╔╝╚██████╔╝███████╗███████║███████╗╚██████╔╝
 ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚══════╝ ╚═════╝`)

	fmt.Print(`
██████╗ ██████╗  ██████╗      ██╗███████╗ ██████╗████████╗██╗███╗   ██╗███████╗ ██████╗ 
██╔══██╗██╔══██╗██╔═══██╗     ██║██╔════╝██╔════╝╚══██╔══╝██║████╗  ██║██╔════╝██╔═══██╗
██████╔╝██████╔╝██║   ██║     ██║█████╗  ██║        ██║   ██║██╔██╗ ██║█████╗  ██║   ██║
██╔═══╝ ██╔══██╗██║   ██║██   ██║██╔══╝  ██║        ██║   ██║██║╚██╗██║██╔══╝  ██║   ██║
██║     ██║  ██║╚██████╔╝╚█████╔╝███████╗╚██████╗   ██║   ██║██║ ╚████║██║     ╚██████╔╝
╚═╝     ╚═╝  ╚═╝ ╚═════╝  ╚════╝ ╚══════╝ ╚═════╝   ╚═╝   ╚═╝╚═╝  ╚═══╝╚═╝      ╚═════╝
`)

	fmt.Println()
	fmt.Println("\nVersion:", version)
	fmt.Println("\nprojectInfo server is ON\n")

	now := time.Now()
	formattedTime := now.Format("15:04 02/01/2006")
	fmt.Println("Server started at " + formattedTime)

	// Get the hostname and port
	getHostnamePort()

	// Get the environment variables for token, log folder & cache folder
	envBotifyAPIToken, envProjectInfoLogFolder, envProjectInfoFolder = getEnvVariables()

	fmt.Println("\n... waiting for requests\n")
}

func addLineToHTML(line interface{}, titleType bool, createHTML bool) {
	// Convert the line to a string based on its type
	var lineStr string
	switch v := line.(type) {
	case string:
		lineStr = v
	case int, int32, int64, float32, float64:
		lineStr = fmt.Sprintf("%v", v)
	case bool:
		lineStr = fmt.Sprintf("%t", v)
	default:
		lineStr = fmt.Sprintf("%v", v)
	}

	// Define the line content based on titleType
	var content string
	if titleType {
		content = fmt.Sprintf(`<h1 class="title">%s</h1>`, lineStr)
	} else {
		content = fmt.Sprintf(`<p>%s</p>`, lineStr)
	}

	// Initialize the full content with the HTML template
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Generated HTML</title>
    <style>
        .title {
            font-size: 24px;
            color: DeepSkyBlue;
        }
    </style>
</head>
<body>
%s
</body>
</html>`

	// Check if the HTML file already exists
	existingContent := ""
	if _, err := os.Stat("output.html"); err == nil {
		data, err := os.ReadFile("output.html")
		if err == nil {
			existingContent = string(data)
			existingContent = existingContent[:len(existingContent)-len("</body>\n</html>")] // Remove closing tags
		}
	}

	// Combine the existing content with the new content
	fullContent := fmt.Sprintf(htmlTemplate, existingContent+content)

	// Write to the HTML file if createHTML is true
	if createHTML {
		err := os.WriteFile("output.html", []byte(fullContent), 0644)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		} else {
			fmt.Println("HTML file updated successfully.")
		}
	}
}
