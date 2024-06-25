// apiTester: Test Botify APIs
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
	"text/tabwriter"
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

// Project
type projectResponse struct {
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

	fmt.Println(bold+"\nOrganisation name:", orgName)
	fmt.Println(bold+"Project Name:", projectName+reset)
	fmt.Println()

	displaySeparator()

	// Datasource API
	datasourceApiTest()

	displaySeparator()

	// Collections API
	collectionsApiTest()

	displaySeparator()

	// Collections detail
	collectionsDetailApiTest()

	displaySeparator()

	// Project API
	projectApiTest()

	displaySeparator()

	apiTesterDone()
}

func datasourceApiTest() {

	var projectCount = 0

	fmt.Println(green + bold + "\nAPI: Datasource API" + reset)
	fmt.Println(bold + "List all projects created for the specified organisation name.\n" + reset)

	url := fmt.Sprintf("https://api.botify.com/v1/users/%s/datasources_summary_by_projects", orgName)

	fmt.Println(bold+"Endpoint:", url+reset+"\n")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError. datasourceApiTest. Cannot create request:"+reset, err)
		return
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError. datasourceApiTest. Cannot send request:"+reset, err)
		return
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(red+"\nError. datasourceApiTest. Cannot read response body. The specified credentials are probably invalid: "+reset, err)
		return
	}

	// Define a map with string keys for unmarshalling the JSON data
	var jsonData map[string]interface{}

	// Unmarshal the JSON data into the map
	err = json.Unmarshal([]byte(responseData), &jsonData)
	if err != nil {
		fmt.Println(red+"Error. datasourceApiTest. Cannot unmarshall JSON:"+reset, err)
		return
	}

	defer res.Body.Close()

	/*
		// Print only the URL
		for projectURL := range jsonData {

			projectCount++

			fmt.Printf("%s\n", projectURL)
		}

		fmt.Println(bold+"\nNo. of projects for this user:"+reset, projectCount)
	*/

	for projectURL, value := range jsonData {

		projectCount++

		fmt.Printf("\n%s ", projectURL)

		// Sitemaps
		sitemap, ok := value.(map[string]interface{})["sitemaps"].(map[string]interface{})
		if ok {
			fmt.Printf(green + "sitemaps " + reset)
		}

		// Search Console
		searchConsole, ok := value.(map[string]interface{})["search_console"].(map[string]interface{})
		if searchConsole == nil {
			continue
		} else {
			// Check if "runnable" field exists and is set to true
			if runnable, ok := searchConsole["runnable"].(bool); ok && runnable {
				fmt.Printf(green + "GSC " + reset)
			}
		}

		// Google Analytics 3
		googleAnalytics, ok := value.(map[string]interface{})["visits.ganalytics"].(map[string]interface{})
		if googleAnalytics == nil {
			continue
		} else {
			// Check if "runnable" field exists and is set to true
			if runnable, ok := googleAnalytics["runnable"].(bool); ok && runnable {
				fmt.Printf(green + "GA3 " + reset)
			}
		}

		// Google Analytics 4 visits
		googleAnalytics4Visits, ok := value.(map[string]interface{})["visits.dip"].(map[string]interface{})
		if googleAnalytics4Visits == nil {
			continue
		} else {
			// Check if "runnable" field exists and is set to true
			if runnable, ok := googleAnalytics4Visits["runnable"].(bool); ok && runnable {
				fmt.Printf(green + "GA4Visits " + reset)
			}
		}

		// Google Analytics 4 conversion
		googleAnalytics4Revenue, ok := value.(map[string]interface{})["conversion.dip"].(map[string]interface{})
		if googleAnalytics4Revenue == nil {
			continue
		} else {
			// Check if "runnable" field exists and is set to true
			if runnable, ok := googleAnalytics4Revenue["runnable"].(bool); ok && runnable {
				fmt.Printf(green + "GA4Revenue " + reset)
			}
		}

		// Google Analytics 4 paid date
		googleAnalytics4Paid, ok := value.(map[string]interface{})["paid_search.ga4.dip"].(map[string]interface{})
		if googleAnalytics4Paid == nil {
			continue
		} else {
			// Check if "runnable" field exists and is set to true
			if runnable, ok := googleAnalytics4Paid["runnable"].(bool); ok && runnable {
				fmt.Printf(green + "GA4Paid " + reset)
			}
		}

		// Adobe
		adobeAnalytics, ok := value.(map[string]interface{})["visits.adobe"].(map[string]interface{})
		if adobeAnalytics == nil {
			continue
		} else {
			// Check if "runnable" field exists and is set to true
			if runnable, ok := adobeAnalytics["runnable"].(bool); ok && runnable {
				fmt.Printf(green + "Adobe " + reset)
			}
		}

		// No. of URLs crawled. If not found the crawl is invalid
		stats, ok := sitemap["stats"].(map[string]interface{})
		if !ok {
			fmt.Printf(red + "(Crawl invalid)" + reset)
			continue
		}

		validUrls, ok := stats["ValidUrls"].(float64)
		if !ok {
			fmt.Printf(red+"Error. datasourceApiTest. Missing 'ValidUrls' key in %s\n"+reset, projectURL)
			continue
		}

		fmt.Printf(" (URLs: %.0f)", validUrls)
	}

	fmt.Println(bold+"\n\nNo. of projects for this user:"+reset, projectCount)

}

// Display the collections API results
func collectionsApiTest() {
	var collectionsCount = 0

	fmt.Println(green + bold + "\nAPI: Collections API" + reset)
	fmt.Println(bold + "List all collections found in a specified project.\n" + reset) //bloo

	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/collections", orgName, projectName)

	fmt.Println(bold+"Endpoint:"+reset, url, "\n")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError. collectionsApitest. collectionsApiTest. Cannot create request:"+reset, err)
		return
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError. collectionsApitest. Cannot send request:"+reset, err)
		return
	}

	defer res.Body.Close()

	var responseData []map[string]interface{}

	err = json.NewDecoder(res.Body).Decode(&responseData)
	if err != nil {
		log.Fatal(red+"\nError. collectionsApitest. Cannot read response body. The specified credentials are probably invalid: "+reset, err)
		return
	}

	// Display all collections
	for _, collection := range responseData {
		collectionsCount++

		id, ok := collection["id"].(string)
		if !ok {
			fmt.Printf(red+"Error. collectionsApiTest. Invalid collection: %+v\n"+reset, collection)
			continue
		}
		fmt.Printf("id: %s\n", id)

		name, ok := collection["name"].(string)
		if !ok {
			fmt.Printf(red+"Error. collectionsApiTest. Invalid collection: %+v\n"+reset, collection)
			continue
		}
		fmt.Printf("Name: %s\n", name)

		// Save the name of the collection in an array, for use in collectionDetailApiTest
		collectionIdentifiers = append(collectionIdentifiers, id)

		date, ok := collection["date"].(string)
		if !ok {
			fmt.Printf(red+"Error. collectionsApiTest. Invalid collection: %+v\n"+reset, collection)
			continue
		}
		fmt.Printf("Date: %s\n", date)

		genericName, ok := collection["generic_name"].(string)
		if !ok {
			fmt.Printf("Generic Name: Not found\n")
		} else {
			fmt.Printf("Generic Name: %s\n", genericName)
		}
		fmt.Printf("\n")
	}

	fmt.Println(bold+"No. of collections found for this project:"+reset, collectionsCount)
}

// Display the collections attributes API results
// This example only displays the first 30 attributes for the first collection found
// Iterate through collectionIdentifiers[] to display the attributes for all collections
func collectionsDetailApiTest() {
	fmt.Println(green + bold + "\nAPI: Collections detail API" + reset)
	fmt.Println(bold + "Display the Name, Identifier, Kind, and Type for the first collection." + reset)
	fmt.Println(bold + "For demonstration purposes only the first 30 attributes of the first collection are displayed." + reset)

	// Iterate through the collectionIdentifiers array
	collectionName := collectionIdentifiers[0] // Only return attributes for the first collection only
	fmt.Println(bold+"\nCollection:"+reset, collectionName, "\n")

	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/collections/%s", orgName, projectName, collectionName)
	fmt.Println(bold+"Endpoint:"+reset, url, "\n")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError. collectionsDetailApiTest. Cannot create request:"+reset, err)
		return
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError. collectionsDetailApiTest. Cannot send request:"+reset, err)
		return
	}
	defer res.Body.Close()

	var responseData map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&responseData)
	if err != nil {
		log.Fatal(red+"\nError. collectionsDetailApiTest. Cannot read response body. The specified credentials are probably invalid: "+reset, err)
		return
	}

	datasets, ok := responseData["datasets"].([]interface{})
	if !ok || len(datasets) == 0 {
		fmt.Println(red + "\nNo datasets found in the response." + reset)
		return
	}

	firstDataset, ok := datasets[0].(map[string]interface{})
	if !ok {
		log.Println(red+"Error. collectionsDetailApiTest. Invalid dataset format:"+reset, datasets[0])
		return
	}

	fields, ok := firstDataset["fields"].([]interface{})
	if !ok {
		fmt.Println(red + "Error. collectionsDetailApiTest. Fields not found for dataset." + reset)
		return
	}

	// Set up tab writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Name", "Identifier", "Kind", "Type")

	// Loop through the first 30 fields.
	for i, field := range fields {
		if i >= 30 {
			break
		}
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			log.Println(red+"Error. collectionsDetailApiTest. Invalid field format:"+reset, field)
			continue
		}

		id, ok := fieldMap["id"].(string)
		if !ok {
			fmt.Println(red + "Error. collectionsDetailApiTest. Invalid field: ID not found" + reset)
			continue
		}

		name, ok := fieldMap["name"].(string)
		if !ok {
			fmt.Println(red + "Error. collectionsDetailApiTest. Invalid field: Name not found" + reset)
			continue
		}

		kind, ok := fieldMap["kind"].(string)
		if !ok {
			fmt.Println(red + "Error. collectionsDetailApiTest. Invalid field: Kind not found" + reset)
			continue
		}

		typeData, ok := fieldMap["type"].(string)
		if !ok {
			fmt.Println(red + "Error. collectionsDetailApiTest. Invalid field: Type not found" + reset)
			continue
		}

		// Print fields with tabwriter
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, id, kind, typeData)
	}
}

// Display the project API results
func projectApiTest() {
	fmt.Println(green + bold + "\nAPI: Project API" + reset)
	fmt.Println(bold + "Display all crawls found for the specified project. Also display the detailed information for the latest crawl.\n" + reset)

	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)
	fmt.Println(bold+"Endpoint:", url+reset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(red+"\nError. projectApiTest. Cannot create request:"+reset, err)
		return
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+botify_api_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(red+"\nError. projectApiTest. Cannot sent request:"+reset, err)
	}
	defer res.Body.Close()

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(red+"\nError. projectApiTest. Cannot read response body. The specified credentials are probably invalid: "+reset, err)
		return
	}

	var responseObject projectResponse
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		log.Fatal(red+"\nError. projectApiTest. Cannot unmarshall JSON:"+reset, err)
		return
	}

	// Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError. projectApiTest. Invalid crawl or no crawls found in the project." + reset)
		return
	}

	fmt.Println(bold+"\nNo. of crawls in this project:"+reset, responseObject.Count)

	fmt.Println(bold + "\nAll crawls found - Summary\n" + reset)

	fmt.Println(bold + "Slug\t\tFriendly Name\tURL\t\t\t\t\t\t\tURLs\tActionB\tStatus" + reset)
	for crawlIndex := range responseObject.Results {
		friendlyName := "Not specified"
		if responseObject.Results[crawlIndex].FriendlyName != nil {
			friendlyName = fmt.Sprintf("%v", responseObject.Results[crawlIndex].FriendlyName)
		}

		// Format the string with tabs between the fields
		line := fmt.Sprintf("%s\t%v\t%s\t%d\t%d\t%s",
			responseObject.Results[crawlIndex].Slug,
			friendlyName,
			responseObject.Results[crawlIndex].URL,
			responseObject.Results[crawlIndex].UrlsDone,
			responseObject.Results[crawlIndex].Features.Scoring.ActionsCount,
			responseObject.Results[crawlIndex].Status)
		fmt.Println(line)
	}

	fmt.Println(bold + "\nLatest crawl detail" + reset)

	if len(responseObject.Results) > 0 {
		user := responseObject.Results[0].User
		fmt.Println(bold + "\nUser" + reset)
		fmt.Println("User Login:", user.Login)
		fmt.Println("User Email:", user.Email)
		fmt.Println("Is Organization:", user.IsOrganization)
		fmt.Println("URL:", user.URL)
		fmt.Println("Date Joined:", user.DateJoined)
		fmt.Println("First Name:", user.FirstName)
		fmt.Println("Last Name:", user.LastName)
		fmt.Println("Company Name:", user.CompanyName)

		owner := responseObject.Results[0].Owner
		fmt.Println(bold + "\nOwner" + reset)
		fmt.Println("Owner:", owner.Login)
		fmt.Println("Email:", owner.Email)
		fmt.Println("Is Organization:", owner.IsOrganisation)
		fmt.Println("URL:", owner.URL)
		fmt.Println("Date Joined:", owner.DateJoined)
		fmt.Println("Status:", owner.Status)
		fmt.Println("First Name:", owner.FirstName)
		fmt.Println("Last Name:", owner.LastName)
		fmt.Println("Company Name:", owner.CompanyName)

		fmt.Println(bold + "\nCrawl Details" + reset)
		fmt.Println("Analysis Slug:", responseObject.Results[0].Slug)
		fmt.Println("Friendly Name:", responseObject.Results[0].FriendlyName)
		fmt.Println("URL:", responseObject.Results[0].URL)
		fmt.Println("Status:", responseObject.Results[0].Status)
		fmt.Println("Computing Revision:", responseObject.Results[0].ComputingRevision)

		//Crawl Configuration
		fmt.Println(bold + "\nCrawl Configuration" + reset)
		fmt.Println("MaxUrls:", responseObject.Results[0].Config.MaxUrls)
		fmt.Println("Crawl Speed:", responseObject.Results[0].Config.MaxUrlsPerSec)
		fmt.Println("Max Depth:", responseObject.Results[0].Config.MaxDepth)
		fmt.Println("Virtual Robots:", responseObject.Results[0].Config.VirtualRobotsTxt)

		fmt.Println(bold + "\nCrawled URLs" + reset)
		fmt.Println("Crawl Schedule:", responseObject.Results[0].CrawlLaunchType)
		fmt.Println("Latest URLs Crawled:", responseObject.Results[0].UrlsDone)
		fmt.Println("URLs in Queue:", responseObject.Results[0].UrlsInQueue)

		fmt.Println(bold + "\nActionBoard" + reset)
		fmt.Println("No. Recommendations:", responseObject.Results[0].Features.Scoring.ActionsCount)

		//Allowed domains
		fmt.Println(bold + "\nAllowed Domains" + reset)
		for _, AllowedDomains := range responseObject.Results[0].Config.AllowedDomains {
			fmt.Println(green+"Domain:", AllowedDomains.Domain+reset)
			fmt.Println("Mobile?:", AllowedDomains.Mobile)
			fmt.Println("Protocol:", AllowedDomains.Protocol)
			fmt.Println("User Agent:", AllowedDomains.UserAgent)
			fmt.Println("Allow Subdomains:", AllowedDomains.AllowSubdomains)
		}

		//Start URLs
		fmt.Println(bold + "\nStart URLs" + reset)
		for _, StartUrls := range responseObject.Results[0].Config.StartUrls {
			fmt.Println(green+"URLs:", StartUrls)
		}
		fmt.Println(reset+"Export Limit:", responseObject.Results[0].Config.ExportLimit)
		fmt.Println("Date Launched:", responseObject.Results[0].DateLaunched)
		fmt.Println("Date Finished:", responseObject.Results[0].DateFinished)
		fmt.Println("Date Last Modified:", responseObject.Results[0].DateLastModified)
		fmt.Println("Date Created:", responseObject.Results[0].DateCreated)
		fmt.Println("Date Crawl Done:", responseObject.Results[0].DateCrawlDone)
		for _, failure := range responseObject.Results[0].Failures {
			fmt.Println("Failure:", failure)
		}

		//Blacklisted domains
		fmt.Println(bold + "\nBlacklisted Domains, if any" + reset)
		for _, BlacklistedDomains := range responseObject.Results[0].Config.BlacklistedDomains {
			fmt.Println(green+"Domain:", BlacklistedDomains)
		}

		//Segments
		fmt.Println(bold + "\nSegments" + reset)
		fmt.Println("Date Created:", responseObject.Results[0].Features.Segments.DateCreated)
		// Iterate over Values and print Name and Field
		for _, segment := range responseObject.Results[0].Features.Segments.Values {
			fmt.Println(segment.Name)
			//fmt.Println("Field:", segment.Field)
		}

		//Sitemaps
		fmt.Println(bold + "\nSitemaps" + reset)
		// Iterate over values and print the URLs
		for _, url := range responseObject.Results[0].Features.Sitemaps.Urls {
			fmt.Println(green+"URL:", url)
		}

		fmt.Println(reset+"Date Retrieved:", responseObject.Results[0].Features.Sitemaps.DateRetrieved)
		fmt.Println("Has Orphans Area:", responseObject.Results[0].Features.Sitemaps.HasOrphansArea)

		//Search console
		fmt.Println(bold + "\nSearch Console" + reset)
		fmt.Println("Date Start:", responseObject.Results[0].Features.SearchConsole.DateStart)
		fmt.Println("Date End:", responseObject.Results[0].Features.SearchConsole.DateEnd)

		//Additional crawl attributes
		fmt.Println(bold + "\n\nFEATURES" + reset)
		//Rel
		fmt.Println(bold + "Rel" + reset)
		fmt.Println("ProcessRelAmp:", responseObject.Results[0].Features.Rel.ProcessRelAmp)
		fmt.Println("ProcessRelApp:", responseObject.Results[0].Features.Rel.ProcessRelApp)
		fmt.Println("ProcessRelAlternate:", responseObject.Results[0].Features.Rel.ProcessRelAlternate)
		fmt.Println("ProcessRelCanonical:", responseObject.Results[0].Features.Rel.ProcessRelCanonical)
		fmt.Println("ProcessRelPrevNext:", responseObject.Results[0].Features.Rel.ProcessRelPrevNext)

		//Main
		fmt.Println(bold + "\nMain" + reset)
		fmt.Println("Lang:", responseObject.Results[0].Features.Main.Lang)
		fmt.Println("ProcessDevice:", responseObject.Results[0].Features.Main.ProcessDevice)
		fmt.Println("CompliantExcludeBadCanonicals:", responseObject.Results[0].Features.Main.CompliantExcludeBadCanonicals)

		//Links
		fmt.Println(bold + "\nLinks" + reset)
		fmt.Println("Chains:", responseObject.Results[0].Features.Links.Chains)
		fmt.Println("PageRank:", responseObject.Results[0].Features.Links.PageRank)
		fmt.Println("PrevNext:", responseObject.Results[0].Features.Links.PrevNext)
		fmt.Println("LinksGraph:", responseObject.Results[0].Features.Links.LinksGraph)
		fmt.Println("TopAnchors:", responseObject.Results[0].Features.Links.TopAnchors)
		fmt.Println("TopDomains:", responseObject.Results[0].Features.Links.TopDomains)
		fmt.Println("LinksToNoindex:", responseObject.Results[0].Features.Links.LinksToNoindex)
		fmt.Println("LinksToNoindex:", responseObject.Results[0].Features.Links.LinksToNoindex)
		fmt.Println("LinksSegmentGraph:", responseObject.Results[0].Features.Links.LinksSegmentGraph)
		fmt.Println("LinksToNonCanonical:", responseObject.Results[0].Features.Links.LinksToNonCanonical)

		fmt.Println("LinksToNonCanonical:", responseObject.Results[0].Features.Scoring.Version)
		fmt.Println("ActionsHash:", responseObject.Results[0].Features.Scoring.ActionsHash)
		fmt.Println("No. Actionboard Recos:", responseObject.Results[0].Features.Scoring.ActionsCount)
		fmt.Println("HaveMlActions:", responseObject.Results[0].Features.Scoring.HaveMlActions)
		fmt.Println("MainImage:", responseObject.Results[0].Features.MainImage)

		//Content quality
		fmt.Println(bold + "\nContent quality" + reset)
		fmt.Println("Samples:", responseObject.Results[0].Features.ContentQuality.Samples)

		fmt.Println(bold + "\nSemantic metadata" + reset)
		fmt.Println("Length:", responseObject.Results[0].Features.SemanticMetadata.Length)
		fmt.Println("Address:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Stats.Address)
		fmt.Println("Product:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Stats.Product)
		fmt.Println("Breadcrumb:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Stats.Breadcrumb)
		fmt.Println("Version: QA:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Qa)
		fmt.Println("Version: Car:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Car)
		fmt.Println("Version: FAQ:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Faq)
		fmt.Println("Version: Book:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Book)
		fmt.Println("Version: News:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.News)
		fmt.Println("Version: Dates:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Dates)
		fmt.Println("Version: Event:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Event)
		fmt.Println("Version: Movie:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Movie)
		fmt.Println("Version: Offer:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Offer)
		fmt.Println("Version: Course:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Course)
		fmt.Println("Version: Person:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Person)
		fmt.Println("Version: Rating:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Rating)
		fmt.Println("Version: Recipe:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Recipe)
		fmt.Println("Version: Review:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Review)
		fmt.Println("Version: Address:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Address)
		fmt.Println("Version: Product:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Product)
		fmt.Println("Version: AudioBook:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.AudioBook)
		fmt.Println("Version: Breadcrumb:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Breadcrumb)
		fmt.Println("Version: Restaurant:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.Restaurant)
		fmt.Println("Version: TrainTrip:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.TrainTrip)
		fmt.Println("Version: JobPosting:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.JobPosting)
		fmt.Println("Version: VideoObject:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.VideoObject)
		fmt.Println("Version: EducationEvent:", responseObject.Results[0].Features.SemanticMetadata.StructuredData.Versions.EducationEvent)

		fmt.Println(bold + "\nCurrency" + reset)
		for _, currency := range responseObject.Results[0].Features.SemanticMetadata.StructuredData.Currencies.Offer {
			fmt.Println("Currency Offer:", currency)
		}

		fmt.Println("\nDuplicateQueryKvs:", responseObject.Results[0].Features.DuplicateQueryKvs)
		fmt.Println("RedButtonDomain:", responseObject.Results[0].RedButtonDomain)
		fmt.Println("ImportKeywordsData:", responseObject.Results[0].ImportKeywordsData)
		fmt.Println("ImportKeywordsDataByCountry:", responseObject.Results[0].ImportKeywordsDataByCountry)
		fmt.Println("CrawlLaunchType:", responseObject.Results[0].CrawlLaunchType)
		fmt.Println("ToBeDeletedAt:", responseObject.Results[0].ToBeDeletedAt)
		fmt.Println("Comparable:", responseObject.Results[0].Comparable)
		fmt.Println("ExcludedFromTrends:", responseObject.Results[0].ExcludedFromTrends)
		fmt.Println("Pk:", responseObject.Results[0].Pk)
		fmt.Println("HasRawPages:", responseObject.Results[0].HasRawPages)
	}
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit apiTester" +
			"\n")

		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using listURLs. Goodbye!\n")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project Name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using listURLs. Goodbye!\n")
			os.Exit(0)
		}
	}
}

func apiTesterDone() {

	// We're done
	fmt.Println(purple + "\napiTester: Done!\n")
	fmt.Println(bold + green + "\nPress any key to exit..." + reset)
	var input string
	fmt.Scanln(&input)
	os.Exit(0)
}

// Display the welcome banner
func displayBanner() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Print(green + `
█████╗ ██████╗ ██╗████████╗███████╗███████╗████████╗███████╗██████╗ 
██╔══██╗██╔══██╗██║╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝██╔══██╗
███████║██████╔╝██║   ██║   █████╗  ███████╗   ██║   █████╗  ██████╔╝
██╔══██║██╔═══╝ ██║   ██║   ██╔══╝  ╚════██║   ██║   ██╔══╝  ██╔══██╗
██║  ██║██║     ██║   ██║   ███████╗███████║   ██║   ███████╗██║  ██║
╚═╝  ╚═╝╚═╝     ╚═╝   ╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
`)
	fmt.Println(purple+"Version:"+reset, version+"\n")
	fmt.Println(purple + "apiTester: Test Botify APIs\n" + reset)
	fmt.Println(purple + "This utility calls a range of Botify APIs and displays the results.\n" + reset)
	fmt.Println(purple + "Use it as a template for your Botify integration needs.\n" + reset)
	fmt.Println(purple + "APIs used in this version.\n" + reset)
	fmt.Println(checkmark + green + bold + " Datasource API" + reset)
	fmt.Println(checkmark + green + bold + " Collections API" + reset)
	fmt.Println(checkmark + green + bold + " Collections Attributes API" + reset)
	fmt.Println(checkmark + green + bold + " Project API\n" + reset)
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
