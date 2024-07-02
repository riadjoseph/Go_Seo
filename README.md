# Go_Seo
A series of tools used to support SEO actions. Developed in Go.   

## seoBusinessInsights   
Undercover the value of organic traffic with this business insights dashboard for Botify.

Charts included:
- Compound growth indicators
- Revenue & visits
- Visits per order
- Order volume
- Order value
- Revenue & Visits river
- Visit value
- Branded & non branded keyword cloud
- Winning branded and non branded keywords
- Detailed KPI insights 

**Note:** The API token for seoBusinessInsights is acquired from the environment variable BotifyAPIToken.

## segmentifyLite   
Generates the segmentation regex for the following segments: 

- First level folders
- Second level folders
- Parameter usage
- No. of parameters
- Parameter keys
- No. of folders
- Static resources
- Shopify (if detected)
- SFCC (if detected, and the site is not using "Search-Friendly URLs for B2C Commerce")

**Note:** The API token for seoBusinessInsights is acquired from the environment variable BotifyAPIToken

**Note:** segmentifyLite will process a maximum of 300k URLs.   
**Note:** The number of URLs found in level 1 folders, level 2 folders and parameter key segments are included as comments after the generated regex. Use these insights to decide which segments slices to keep and which to remove.   
**Note:** All level 1 and level 2 segments which are less than 5% of the size of the largest level 1 or level 2 folder found are excluded from the segmentation regex. To amend this threshold change the percentage threshold in the variable **thresholdPercent**.  
**Note:** Update the code to include your own Botify API key. Change the variable **APIToken** to reflect your token  

## exportURLs  
Export all URLs from the latest crawl in a specified project (up to a maximum of 1MM) to a file (siteurlsExport.csv)


## botifyBotLite   
Generate Botify crawls en-masse.    

**Note:** Populate the file crawlme.txt with the start page URL for each of the sites you want to crawl.  
**Note:** When botifyBotLite has completed a list of the crawled start pages and the generated Botify project URLs can be found in **project_list.txt**.  

```
Usage: ./botifyBotLite (project prefix and no. urls to crawl will be prompted for)    
Usage: ./botifyBotLite project_prefix no_urls_to_crawl    
```

## apiTester   
Example utilisation of a range of Botify APIs. Included in this version are: 
- Datasource API
- Collections API
- Collections attributes API (First 30 attributes for the first collection only)
- Project API

```
Usage: ./apiTester (organisation and project name will be prompted for)    
Usage: ./apiTester org_name project_name    
```
**Note:** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  

## bqlTester   
Demonstartion of using BQL in Go. Included in this version are BQL queries for: 
- Site crawler stats
- Revenue
- Vists   
```
Usage: ./bqlTester (organisation and project name will be prompted for)    
Usage: ./bqlTester org_name project_name    
```
**Note:** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  

## folderCount
folderCount is used to identify and count all first level folders.  
Up to 1MM URLs maximum are used to produce the analysis.  
```
Usage: ./folderCount (organisation and project name will be prompted for)    
Usage: ./folderCount org_name project_name  
```
**Note:** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  

## paramCount  
paramCount is used to identify and count all Parameter Keys in the crawl.  
```
Usage: ./paramCount (organisation and project name will be prompted for)    
Usage: ./paramCount org_name project_name  
```
**Note:** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  

## listAnalysis
listAnalysis displays the crawl metadata for the most recent crawl in a specified project.   
```
Usage: ./listAnalysis (organisation and project name will be prompted for)    
Usage: ./listAnalysis org_name project_name
```
**Note:** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  

