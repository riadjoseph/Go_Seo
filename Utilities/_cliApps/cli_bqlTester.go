// bqlTester: Test Botify APIs
// Analysis based on 1MM URL maximum
// Written by Jason Vicinanza

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Used to get the latest slug
type latestSlug struct {
	Count   int `json:"count"`
	Results []struct {
		Slug string `json:"slug"`
	} `json:"results"`
}

// Used to store the site crawler KPIs
type basicKPIs struct {
	Results []struct {
		KPI []int `json:"metrics"`
	} `json:"results"`
}

// DateRanges struct used to hold the monthly date ranges and the YTD date range
// Used for revenue and visits data
type DateRanges struct {
	MonthlyRanges [][2]time.Time
	YTDRange      [2]time.Time
}

// AnalyticsID is used to identify which analytics tool is in use
type AnalyticsID struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Date        string      `json:"date"`
	Timestamped bool        `json:"timestamped"`
	DateStart   string      `json:"date_start"`
	DateEnd     string      `json:"date_end"`
	GenericName interface{} `json:"generic_name"`
}

type transRevID struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Multiple bool          `json:"multiple"`
	Fields   []field       `json:"fields"`
	Groups   []interface{} `json:"groups"`
	Category []interface{} `json:"category"`
}

type field struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Subtype        string      `json:"subtype"`
	Multiple       bool        `json:"multiple"`
	Permissions    []string    `json:"permissions"`
	Optional       bool        `json:"optional"`
	Kind           string      `json:"kind"`
	GlobalField    string      `json:"global_field"`
	DiffReturnType interface{} `json:"diff_return_type"`
	ApiOnly        bool        `json:"api_only"`
	Meta           meta        `json:"meta"`
	Suggestion     bool        `json:"suggestion"`
}

type meta struct {
	RequiredFields []string `json:"required_fields"`
}

// Result used to store the revenue, transactions and visits
type Result struct {
	Dimensions []interface{} `json:"dimensions"`
	Metrics    []float64     `json:"metrics"`
}

type Response struct {
	Results  []Result    `json:"results"`
	Previous interface{} `json:"previous"`
	Next     string      `json:"next"`
	Page     int         `json:"page"`
	Size     int         `json:"size"`
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

// APIToken should be replaced with your own Botify API token.
var APIToken = "c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205"

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

	fmt.Println(bold+"\nOrganisation name:", orgName)
	fmt.Println(bold+"Project name:", projectName+reset)
	fmt.Println()

	displaySeparator()

	// Crawl stats KPIs
	crawlStats()

	displaySeparator()

	// Revenue for the last 12 months
	seoRevenue()

	displaySeparator()

	// Visits for the last 12 months
	//seoVisits()

	bqlTesterDone()
}

// Check that the org and project names have been specified as command line arguments
// if not prompt for them
// Pressing Enter exits
func checkCredentials() {

	if len(os.Args) < 3 {

		credentialsInput = true

		fmt.Print("\nEnter your project credentials. Press" + green + " Enter " + reset + "to exit bqlTester" +
			"\n")

		fmt.Print(purple + "\nEnter organisation name: " + reset)
		fmt.Scanln(&orgNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(orgNameInput) == "" {
			fmt.Println(green + "\nThank you for using bqlTester. Goodbye!")
			os.Exit(0)
		}

		fmt.Print(purple + "Enter project name: " + reset)
		fmt.Scanln(&projectNameInput)
		// Check if input is empty if so exit
		if strings.TrimSpace(projectNameInput) == "" {
			fmt.Println(green + "\nThank you for using bqlTester. Goodbye!")
			os.Exit(0)
		}
	}
}

// Basic KPIs
func crawlStats() {
	fmt.Println(purple + bold + "\nGetting the site crawler insights\n" + reset)

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

	// Define the BQL to get the pages with few inlinks (< 10 inlinks)
	bqlFewInlinksUrls := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.inlinks_internal.nb.unique",
                            "predicate": "lt",
                            "value": 10
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

	// Duplicate titles
	bqlDuplicateTitles := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
                    {
                        "field": "crawl.%s.metadata.title.duplicates.context_aware.nb",
                            "predicate": "gt",
                            "value": 0
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Unique titles
	bqlUniqueTitles := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    }, 
      				{
                        "field": "crawl.%s.metadata.title.duplicates.context_aware.nb",
                            "value": 0
                    },
   					 {
						"field": "crawl.%s.metadata.title.nb",
						"predicate": "gt",
						"value": 0
  				    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug, latestSlug)

	// Missing H1
	bqlMissingTitles := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
                    {
                        "field": "crawl.%s.metadata.title.nb",
                            "value": 0
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Duplicate H1
	bqlDuplicateH1 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
                    {
                        "field": "crawl.%s.metadata.h1.duplicates.context_aware.nb",
                            "predicate": "gt",
                            "value": 0
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Unique H1
	bqlUniqueH1 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    }, 
      				{
                        "field": "crawl.%s.metadata.h1.duplicates.context_aware.nb",
                            "value": 0
                    },
   					 {
						"field": "crawl.%s.metadata.h1.nb",
						"predicate": "gt",
						"value": 0
  				    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug, latestSlug)

	// Missing H1
	bqlMissingH1 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
                    {
                        "field": "crawl.%s.metadata.h1.nb",
                            "value": 0
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Duplicate Description
	bqlDuplicateDescription := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
                    {
                        "field": "crawl.%s.metadata.description.duplicates.context_aware.nb",
                            "predicate": "gt",
                            "value": 0
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Unique Description
	bqlUniqueDescription := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    }, 
      				{
                        "field": "crawl.%s.metadata.description.duplicates.context_aware.nb",
                            "value": 0
                    },
   					 {
						"field": "crawl.%s.metadata.description.nb",
						"predicate": "gt",
						"value": 0
  				    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug, latestSlug)

	// Missing Description
	bqlMissingDescription := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
                    {
                        "field": "crawl.%s.metadata.description.nb",
                            "value": 0
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// HTTP 100: Informational responses (100 – 199)
	bqlHttp100 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.http_code",
                         "predicate": "between",
						 "value": [100, 199]
                    }
                ]
            }
	}`, latestSlug, latestSlug)

	// HTTP 200: Successful responses (200 – 299)
	bqlHttp200 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.http_code",
                         "predicate": "between",
						 "value": [200, 299]
                    }
                ]
            }
	}`, latestSlug, latestSlug)

	// HTTP 300: Redirection messages (300 – 399)
	bqlHttp300 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.http_code",
                         "predicate": "between",
						 "value": [300, 399]
                    }
                ]
            }
	}`, latestSlug, latestSlug)

	// HTTP 400: Client error messages (400 – 499)
	bqlHttp400 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.http_code",
                         "predicate": "between",
						 "value": [400, 499]
                    }
                ]
            }
	}`, latestSlug, latestSlug)

	// Page load speed - Fast
	bqlHttp500 := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
					{
                        "field": "crawl.%s.http_code",
                         "predicate": "between",
						 "value": [500, 599]
                    }
                ]
            }
	}`, latestSlug, latestSlug)

	// Fast URLs
	bqlFastURLs := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
					{
                        "field": "crawl.%s.delay_last_byte",
                         "predicate": "between",
						 "value": [0, 499]
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Medium speed URLs
	bqlMediumURLs := fmt.Sprintf(`
	{
            "field": "crawl.%s.count_urls_crawl",
            "filters": {
                "and": [
                    {
                        "field": "crawl.%s.compliant.is_compliant",
                        "predicate": "eq",
                        "value": true
                    },
					{
                        "field": "crawl.%s.delay_last_byte",
                         "predicate": "between",
						 "value": [500, 999]
                    }
                ]
            }
	}`, latestSlug, latestSlug, latestSlug)

	// Slow speed URLs
	bqlSlowSpeedURLs := fmt.Sprintf(`
{
        "field": "crawl.%s.count_urls_crawl",
        "filters": {
            "and": [
                {
                    "field": "crawl.%s.compliant.is_compliant",
                    "predicate": "eq",
                    "value": true
                },
                {
                    "field": "crawl.%s.delay_last_byte",
                     "predicate": "between",
                     "value": [1000, 1999]
                }
            ]
        }
}`, latestSlug, latestSlug, latestSlug)

	// Slowest speed URLs
	bqlSlowestURLs := fmt.Sprintf(`
{
        "field": "crawl.%s.count_urls_crawl",
        "filters": {
            "and": [
                {
                    "field": "crawl.%s.compliant.is_compliant",
                    "predicate": "eq",
                    "value": true
                },
                {
                    "field": "crawl.%s.delay_last_byte",
                    "predicate": "gt",
				    "value": 2000
                }
            ]
        }
}`, latestSlug, latestSlug, latestSlug)

	// Generate the bqlQueriesDepth for the depth (text/HTML only)
	depths := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	bqlQueriesDepth := make([]string, 0, len(depths)+1)

	for _, depth := range depths {
		bqlDepthQuery := generateQuery(latestSlug, depth, "eq", depth)
		bqlQueriesDepth = append(bqlQueriesDepth, bqlDepthQuery)
	}

	queryDepth10Plus := generateQuery(latestSlug, 10, "gt", 10)
	bqlQueriesDepth = append(bqlQueriesDepth, queryDepth10Plus)

	// Indexable URLs in Sitemap
	bqlIndexableInSitemapURLs := fmt.Sprintf(`
{
        "field": "crawl.%s.count_urls_crawl",
        "filters": {
            "and": [
                {
                    "field": "crawl.%s.compliant.is_compliant",
                    "predicate": "eq",
                    "value": true
                },
                {
                    "field": "crawl.%s.sitemaps.present",
                    "predicate": "eq",
                    "value": true
                }
            ]
        }
}`, latestSlug, latestSlug, latestSlug)

	// Non Indexable URLs in Sitemap
	bqlNonIndexableInSitemapURLs := fmt.Sprintf(`
{
        "field": "crawl.%s.count_urls_crawl",
        "filters": {
            "and": [
                {
                    "field": "crawl.%s.compliant.is_compliant",
                    "predicate": "eq",
                    "value": false
                },
                {
                    "field": "crawl.%s.sitemaps.present",
                    "predicate": "eq",
                    "value": true
                }
            ]
        }
}`, latestSlug, latestSlug, latestSlug)

	// URLs crawled by Botify (aka Known pages)
	bqlCrawlByBotifyURLs := fmt.Sprintf(`
{
					"field": "crawl.%s.count_urls_crawl"
}`, latestSlug)

	// END OF BQL GENERATION

	// Array of BQL fragments used to construct the final BQL
	metrics := []string{
		bqlIndexableUrls,
		bqlNonIndexableUrls,
		bqlSlowPageSpeedUrls,
		bqlFewInlinksUrls,
		bqlDeepUrls,
		bqlDuplicateTitles,
		bqlUniqueTitles,
		bqlMissingTitles,
		bqlDuplicateH1,
		bqlUniqueH1,
		bqlMissingH1,
		bqlDuplicateDescription,
		bqlUniqueDescription,
		bqlMissingDescription,
		bqlHttp100,
		bqlHttp200,
		bqlHttp300,
		bqlHttp400,
		bqlHttp500,
		bqlFastURLs,
		bqlMediumURLs,
		bqlSlowSpeedURLs,
		bqlSlowestURLs,
		bqlQueriesDepth[0],
		bqlQueriesDepth[1],
		bqlQueriesDepth[2],
		bqlQueriesDepth[3],
		bqlQueriesDepth[4],
		bqlQueriesDepth[5],
		bqlQueriesDepth[6],
		bqlQueriesDepth[7],
		bqlQueriesDepth[8],
		bqlQueriesDepth[9],
		bqlQueriesDepth[10],
		bqlIndexableInSitemapURLs,
		bqlNonIndexableInSitemapURLs,
		bqlCrawlByBotifyURLs,
	}

	// Join the metrics BQL fragments together
	metricsString := strings.Join(metrics, ",\n")

	// Bring the BQL fragments into a single query
	bqlFunnelBody := fmt.Sprintf(`
{
	"collections": [
		"crawl.%s"
	],
	"query": {
		"dimensions": [],
		"metrics": [
			%s
		]
	}
}`, latestSlug, metricsString)

	// Define the URL
	url := fmt.Sprintf("https://api.botify.com/v1/projects/%s/%s/query", orgName, projectName)
	fmt.Println("End point:", url)

	// GET the HTTP request
	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot create request. Perhaps the provided credentials are invalid: "+reset, errorCheck)
	}

	// Define the body
	httpBody := []byte(bqlFunnelBody)

	// Create the POST request
	req, errorCheck = http.NewRequest("POST", url, bytes.NewBuffer(httpBody))
	if errorCheck != nil {
		log.Fatal("Error. seoFunnel. Cannot create request. Perhaps the provided credentials are invalid: ", errorCheck)
	}

	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+APIToken)
	req.Header.Add("Content-Type", "application/json")

	// Create HTTP client and execute the request
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, errorCheck := client.Do(req)
	if errorCheck != nil {
		log.Fatal("Error. seoFunnel. Error: ", errorCheck)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, errorCheck := io.ReadAll(resp.Body)

	if errorCheck != nil {
		log.Fatal(red+"Error. seoFunnel. Cannot read response body: "+reset, errorCheck)
		return
	}

	// Unmarshal the JSON data into the struct
	var responseObject basicKPIs
	errorCheck = json.Unmarshal(responseData, &responseObject)
	if errorCheck != nil {
		log.Fatal(red+"Error. seoFunnel. Cannot unmarshal JSON: "+reset, errorCheck)
	}

	firstResult := responseObject.Results[0]
	indexableURLs := firstResult.KPI[0]
	nonIndexableURLs := firstResult.KPI[1]
	slowURLs := firstResult.KPI[2]
	fewInlinksURLs := firstResult.KPI[3]
	deepURLs := firstResult.KPI[4]
	duplicateTitles := firstResult.KPI[5]
	uniqueTitles := firstResult.KPI[6]
	missingTitles := firstResult.KPI[7]
	duplicateH1 := firstResult.KPI[8]
	uniqueH1 := firstResult.KPI[9]
	missingH1 := firstResult.KPI[10]
	duplicateDescription := firstResult.KPI[11]
	uniqueDescription := firstResult.KPI[12]
	missingDescription := firstResult.KPI[13]
	http100 := firstResult.KPI[14]
	http200 := firstResult.KPI[15]
	http300 := firstResult.KPI[16]
	http400 := firstResult.KPI[17]
	http500 := firstResult.KPI[18]
	fastURLs := firstResult.KPI[19]
	mediumURLs := firstResult.KPI[20]
	slowSpeedURLs := firstResult.KPI[21]
	slowestURLs := firstResult.KPI[22]
	depth0URLs := firstResult.KPI[23]
	depth1URLs := firstResult.KPI[24]
	depth2URLs := firstResult.KPI[25]
	depth3URLs := firstResult.KPI[26]
	depth4URLs := firstResult.KPI[27]
	depth5URLs := firstResult.KPI[28]
	depth6URLs := firstResult.KPI[29]
	depth7URLs := firstResult.KPI[30]
	depth8URLs := firstResult.KPI[31]
	depth9URLs := firstResult.KPI[32]
	depth10PlusURLs := firstResult.KPI[33]
	indexableInSitemap := firstResult.KPI[34]
	nonIndexableInSitemap := firstResult.KPI[35]
	crawledByBotify := firstResult.KPI[36]

	// Print the results
	fmt.Println(bold+"Total crawled by Botify:"+reset, crawledByBotify)
	fmt.Println(bold+"Indexable:"+reset, indexableURLs)
	fmt.Println(bold+"Non indexable:"+reset, nonIndexableURLs)
	fmt.Println(green + "\nKPIs for Indexable pages" + reset)
	fmt.Println(bold+"Slow pages (> 500 ms):"+reset, slowURLs)
	fmt.Println(bold+"Pages with few inlinks (< 10 inlinks):"+reset, fewInlinksURLs)
	fmt.Println(bold+"Deep pages (> depth 5):"+reset, deepURLs)
	fmt.Println(green + "\nHTML Tags Performance For Indexable URLs" + reset)
	fmt.Println(bold+"Duplicate titles:"+reset, duplicateTitles)
	fmt.Println(bold+"Unique titles:"+reset, uniqueTitles)
	fmt.Println(bold+"Missing titles:"+reset, missingTitles)
	fmt.Println(bold+"Duplicate H1:"+reset, duplicateH1)
	fmt.Println(bold+"Unique H1:"+reset, uniqueH1)
	fmt.Println(bold+"Missing H1:"+reset, missingH1)
	fmt.Println(bold+"Duplicate description:"+reset, duplicateDescription)
	fmt.Println(bold+"Unique description:"+reset, uniqueDescription)
	fmt.Println(bold+"Missing description:"+reset, missingDescription)
	fmt.Println(green + "\nHTTP Status Codes Distribution (Indexable URLs)" + reset)
	fmt.Println(bold+"HTTP 100 class. Informational:"+reset, http100)
	fmt.Println(bold+"HTTP 200 class. Successful response:"+reset, http200)
	fmt.Println(bold+"HTTP 300 class. Redirect response:"+reset, http300)
	fmt.Println(bold+"HTTP 400 class. Client error response:"+reset, http400)
	fmt.Println(bold+"HTTP 500 class. Server error response:"+reset, http500)
	fmt.Println(green + "\nLoad time Distribution (Indexable URLs)" + reset)
	fmt.Println(bold+"Fast URLs (0-499ms):"+reset, fastURLs)
	fmt.Println(bold+"Medium URLs (500-999ms):"+reset, mediumURLs)
	fmt.Println(bold+"Slow URLs (1000-1999ms):"+reset, slowSpeedURLs)
	fmt.Println(bold+"Slowest URLs (>2000ms):"+reset, slowestURLs)
	fmt.Println(green + "\nURLs By Depth for text/HTML content" + reset)
	fmt.Println(bold+"Depth 0:"+reset, depth0URLs)
	fmt.Println(bold+"Depth 1:"+reset, depth1URLs)
	fmt.Println(bold+"Depth 2:"+reset, depth2URLs)
	fmt.Println(bold+"Depth 3:"+reset, depth3URLs)
	fmt.Println(bold+"Depth 4:"+reset, depth4URLs)
	fmt.Println(bold+"Depth 5:"+reset, depth5URLs)
	fmt.Println(bold+"Depth 6:"+reset, depth6URLs)
	fmt.Println(bold+"Depth 7:"+reset, depth7URLs)
	fmt.Println(bold+"Depth 8:"+reset, depth8URLs)
	fmt.Println(bold+"Depth 9:"+reset, depth9URLs)
	fmt.Println(bold+"Depth 10+:"+reset, depth10PlusURLs)
	fmt.Println(green + "\nURL Distribution In Sitemaps" + reset)
	fmt.Println(bold+"Indexable URLs in sitemap:"+reset, indexableInSitemap)
	fmt.Println(bold+"Non indexable URLs in sitemap:"+reset, nonIndexableInSitemap)
}

// Function to generate the query for a given depth
func generateQuery(slug string, depth int, predicate string, value interface{}) string {
	return fmt.Sprintf(`
	{
		"field": "crawl.%s.count_urls_crawl",
		"filters": {
			"and": [
				{
					"field": "crawl.%s.compliant.is_compliant",
					"predicate": "eq",
					"value": true
				},
				{
					"field": "crawl.%s.depth",
					"predicate": "%s",
					"value": %v
				},
				{
					"field": "crawl.%s.content_type",
					"predicate": "eq",
					"value": "text/html"
				}
			]
		}
	}`, slug, slug, slug, predicate, value, slug)
}

func getLatestSlug() string {
	//Get the last analysis slug
	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", orgName, projectName)

	req, errorCheck := http.NewRequest("GET", url, nil)
	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot create request: "+reset, errorCheck)
	}

	// Define the headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token "+APIToken)
	req.Header.Add("Content-Type", "application/json")

	res, errorCheck := http.DefaultClient.Do(req)
	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Check your network connection: "+reset, errorCheck)
	}
	defer res.Body.Close()

	responseData, errorCheck := io.ReadAll(res.Body)

	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot read request body: "+reset, errorCheck)
	}

	var responseObject latestSlug
	errorCheck = json.Unmarshal(responseData, &responseObject)

	if errorCheck != nil {
		log.Fatal(red+"\nError. seoFunnel. Cannot unmarshall JSON: "+reset, errorCheck)
	}

	//Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println(red + "\nError: seoFunnel. Invalid credentials or no crawls found in the project")
		os.Exit(1)
	}

	fmt.Println("Organisation name:", orgName)
	fmt.Println("Project name:", projectName)
	fmt.Println("Latest analysis Slug:", responseObject.Results[0].Slug)

	return responseObject.Results[0].Slug
}

func bqlTesterDone() {

	// We're done
	fmt.Println(purple + "\nbqlTester: Done!")
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
██████╗  ██████╗ ██╗  ████████╗███████╗███████╗████████╗███████╗██████╗ 
██╔══██╗██╔═══██╗██║  ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝██╔══██╗
██████╔╝██║   ██║██║     ██║   █████╗  ███████╗   ██║   █████╗  ██████╔╝
██╔══██╗██║▄▄ ██║██║     ██║   ██╔══╝  ╚════██║   ██║   ██╔══╝  ██╔══██╗
██████╔╝╚██████╔╝███████╗██║   ███████╗███████║   ██║   ███████╗██║  ██║
╚═════╝  ╚══▀▀═╝ ╚══════╝╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
`)
	fmt.Println(purple+"Version:"+reset, version)
	fmt.Println(purple + "bqlTester: Test Botify BQL.\n" + reset)
	fmt.Println(purple + "Use it as a template for your Botify integration needs.\n" + reset)
	fmt.Println(purple + "BQL tests performed in this version.\n" + reset)
	fmt.Println(checkmark + green + bold + " Site crawler insights (examples of site crawler KPI retrieval)" + reset)
	fmt.Println(checkmark + green + bold + " Revenue" + reset)
	fmt.Println(checkmark + green + bold + " Visits" + reset)
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

// Function to format an integer with comma separation
func formatWithCommas(n int) string {
	s := strconv.Itoa(n)
	nLen := len(s)
	if nLen <= 3 {
		return s
	}

	var result strings.Builder
	commaOffset := nLen % 3
	if commaOffset > 0 {
		result.WriteString(s[:commaOffset])
		if nLen > commaOffset {
			result.WriteString(",")
		}
	}

	for i := commaOffset; i < nLen; i += 3 {
		result.WriteString(s[i : i+3])
		if i+3 < nLen {
			result.WriteString(",")
		}
	}

	return result.String()
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
