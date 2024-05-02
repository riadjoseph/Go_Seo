# Go_Seo
A series of command line utilities to support SEO actions.   

# folderCount
folderCount is used to identify all first level folders and count the number of URLs in each folder.   
**Usage:** folderCount URL_Extract_file
# paramCount
paramCount is used to identify and count all Parameter Keys used.   
**Usage:** paramCount URL_Extract_file
# listAnalysis
listAnalysis uses the Botify API and returns key details for the most recent crawl in a specified project.   
**Note** Update the code to include your own Botify API key.   
**Usage:** listAnalysis Username Project_Slug.
# listURLs
Export all URLs (up to a maximum of 1MM) to file (siteurlsExport.csv).  
**Usage:** listURLs      
# segmentifyLite (work in progress)
segmentifyLite generates the regex for the following segments: 
- first level folders
- second level folders
- parameter usage
- No. of parameters
- Parameter keys
- No. of folders
- Shopify (if detected)
- SFCC (if detected, and the site is not using "Search-Friendly URLs for B2C Commerce")
  
**Note** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  
**note** The number of URLs found in level 1 folders, level 2 folders and parameter key segments are included as comments after the generated regex. Use these insights to decide which segments slices to keep and which to remove.   
segmentifyLite will process a maximum of 300k URLs.  
**note** All level 1 and level 2 segments which are less than 5% of the size of the largest level 1 or level 2 folder found are excluded from the segmentation regex. To amend this threshold change the percentage threshold in the variable **thresholdPercent**.
**Usage:** segmentifyLite org_name project_name 
