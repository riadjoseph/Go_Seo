// listAnalysis: Get key crawl attributes for the latest crawl of a specific project.
// Written by Jason Vicinanza

// To run this:
// go run listAnalysis username project_slug
// Example: go run listAnalysis botify-project botify-project-name

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

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

	// Version
	version := "v0.1"

	// ANSI escape code for purple color
	purple := "\033[0;35m"
	// ANSI escape code for blue color
	blue := "\033[34m"
	// ANSI escape code to reset color
	reset := "\033[0m"

	// Clear the screen
	clearScreen()

	// Get the username anf project slug from the command-line arguments
	if len(os.Args) < 3 {
		clearScreen()
		fmt.Println("listAnalysis")
		fmt.Println("listAnalysis. Error. Please provide the Username and Project Slug as arguments.")
		return
	}
	argUserName := os.Args[1]
	argProjectSlug := os.Args[2]

	url := fmt.Sprintf("https://api.botify.com/v1/analyses/%s/%s?page=1&only_success=true", argUserName, argProjectSlug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("listAnalysis. Error creating request:", err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "token c1e6c5ab4a8dc6a16620fd0a885dd4bee7647205")
	//req.Header.Add("Authorization", "token your_Botify_API_Token_Here")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("listAnalysis. Error sending request:", err)
	}
	defer res.Body.Close()

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("listAnalysis. Error reading response body:", err)
	}

	var responseObject botifyResponse
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	// Display welcome message
	fmt.Println(purple + "listAnalysis: Report the project metadata for the last crawl." + reset)
	fmt.Println(purple+"Version:"+reset, version, "\n")

	fmt.Println(purple+"Username:", argUserName)
	fmt.Println("Project Slug:", argProjectSlug+reset)

	// Display an error if no crawls found
	if responseObject.Count == 0 {
		fmt.Println("listAnalysis. Error. Invalid crawl or no crawls found in the project.")
		// Handle error condition here
	}

	fmt.Println("\nNo. Crawls in project:", responseObject.Count)

	if len(responseObject.Results) > 0 {
		user := responseObject.Results[0].User
		fmt.Println(purple + "\nUser" + reset)
		fmt.Println("User Login:", user.Login)
		fmt.Println("User Email:", user.Email)
		fmt.Println("Is Organization:", user.IsOrganization)
		fmt.Println("URL:", user.URL)
		fmt.Println("Date Joined:", user.DateJoined)
		fmt.Println("First Name:", user.FirstName)
		fmt.Println("Last Name:", user.LastName)
		fmt.Println("Company Name:", user.CompanyName)

		owner := responseObject.Results[0].Owner
		fmt.Println(purple + "\nOwner" + reset)
		fmt.Println("Owner:", owner.Login)
		fmt.Println("Email:", owner.Email)
		fmt.Println("Is Organization:", owner.IsOrganisation)
		fmt.Println("URL:", owner.URL)
		fmt.Println("Date Joined:", owner.DateJoined)
		fmt.Println("Status:", owner.Status)
		fmt.Println("First Name:", owner.FirstName)
		fmt.Println("Last Name:", owner.LastName)
		fmt.Println("Company Name:", owner.CompanyName)

		fmt.Println(purple + "\nCrawl Details" + reset)
		fmt.Println("Analysis Slug:", responseObject.Results[0].Slug)
		fmt.Println("Friendly Name:", responseObject.Results[0].FriendlyName)
		fmt.Println("URL:", responseObject.Results[0].URL)
		fmt.Println("Status:", responseObject.Results[0].Status)
		fmt.Println("Computing Revision:", responseObject.Results[0].ComputingRevision)

		//Crawl Configuration
		fmt.Println(purple + "\nCrawl Configuration" + reset)
		fmt.Println("MaxUrls:", responseObject.Results[0].Config.MaxUrls)
		fmt.Println("Crawl Speed:", responseObject.Results[0].Config.MaxUrlsPerSec)
		fmt.Println("Max Depth:", responseObject.Results[0].Config.MaxDepth)
		fmt.Println("Virtual Robots:", responseObject.Results[0].Config.VirtualRobotsTxt)

		fmt.Println(purple + "\nCrawled URLs" + reset)
		fmt.Println("Crawl Schedule:", responseObject.Results[0].CrawlLaunchType)
		fmt.Println("Latest URLs Crawled:", responseObject.Results[0].UrlsDone)
		fmt.Println("URLs in Queue:", responseObject.Results[0].UrlsInQueue)

		fmt.Println(purple + "\nActionBoard" + reset)
		fmt.Println("No. Recommendations:", responseObject.Results[0].Features.Scoring.ActionsCount)

		//Allowed domains
		fmt.Println(purple + "\nAllowed Domains" + reset)
		for _, AllowedDomains := range responseObject.Results[0].Config.AllowedDomains {
			fmt.Println(blue+"Domain:", AllowedDomains.Domain)
			fmt.Println(reset+"Mobile?:", AllowedDomains.Mobile)
			fmt.Println("Protocol:", AllowedDomains.Protocol)
			fmt.Println("User Agent:", AllowedDomains.UserAgent)
			fmt.Println("Allow Subdomains:", AllowedDomains.AllowSubdomains)
		}

		//Start URLs
		fmt.Println(purple + "\nStart URLs" + reset)
		for _, StartUrls := range responseObject.Results[0].Config.StartUrls {
			fmt.Println(blue+"URLs:", StartUrls)
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
		fmt.Println(purple + "\nBlacklisted Domains, if any" + reset)
		for _, BlacklistedDomains := range responseObject.Results[0].Config.BlacklistedDomains {
			fmt.Println(blue+"Domain:", BlacklistedDomains)
		}

		//Segments
		fmt.Println(purple + "\nSegments" + reset)
		fmt.Println("Date Created:", responseObject.Results[0].Features.Segments.DateCreated)
		//fmt.Println("Flags:", responseObject.Results[0].Features.Segments.Flags)
		//fmt.Println("Segment Names:", responseObject.Results[0].Features.Segments.Names)
		// Iterate over Values and print Name and Field
		for _, segment := range responseObject.Results[0].Features.Segments.Values {
			fmt.Println(segment.Name)
			//fmt.Println("Field:", segment.Field)
		}

		//Sitemaps
		fmt.Println(purple + "\nSitemaps" + reset)
		// Iterate over values and print the URLs
		for _, url := range responseObject.Results[0].Features.Sitemaps.Urls {
			fmt.Println(blue+"URL:", url)
		}

		fmt.Println(reset+"Date Retrieved:", responseObject.Results[0].Features.Sitemaps.DateRetrieved)
		fmt.Println("Has Orphans Area:", responseObject.Results[0].Features.Sitemaps.HasOrphansArea)

		//Search console
		fmt.Println(purple + "\nSearch Console" + reset)
		fmt.Println("Date Start:", responseObject.Results[0].Features.SearchConsole.DateStart)
		fmt.Println("Date End:", responseObject.Results[0].Features.SearchConsole.DateEnd)

		//Additional crawl attributes
		fmt.Println(purple + "\n\nFEATURES" + reset)
		//Rel
		fmt.Println(purple + "Rel" + reset)
		fmt.Println("ProcessRelAmp:", responseObject.Results[0].Features.Rel.ProcessRelAmp)
		fmt.Println("ProcessRelApp:", responseObject.Results[0].Features.Rel.ProcessRelApp)
		fmt.Println("ProcessRelAlternate:", responseObject.Results[0].Features.Rel.ProcessRelAlternate)
		fmt.Println("ProcessRelCanonical:", responseObject.Results[0].Features.Rel.ProcessRelCanonical)
		fmt.Println("ProcessRelPrevNext:", responseObject.Results[0].Features.Rel.ProcessRelPrevNext)

		//Main
		fmt.Println(purple + "\nMain" + reset)
		fmt.Println("Lang:", responseObject.Results[0].Features.Main.Lang)
		fmt.Println("ProcessDevice:", responseObject.Results[0].Features.Main.ProcessDevice)
		fmt.Println("CompliantExcludeBadCanonicals:", responseObject.Results[0].Features.Main.CompliantExcludeBadCanonicals)

		//Links
		fmt.Println(purple + "\nLinks" + reset)
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
		fmt.Println(purple + "\nContent quality" + reset)
		fmt.Println("Samples:", responseObject.Results[0].Features.ContentQuality.Samples)

		fmt.Println(purple + "\nSemantic metadata" + reset)
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

		fmt.Println(purple + "\nCurrency" + reset)
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
	fmt.Println(purple + "\nlistAnalysis: Done\n")
}

// Function to clear the screen
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
