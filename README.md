# Go_Seo
A series of command line utilities to support SEO actions.   

# folderCount
folderCount is used to identify and count all first level folders.  
Up to 1MM URLs maximum are used to produce the analysis.  
**Usage:** *./folderCount* (organisation and project name will be prompted for)    
**Usage:** *./folderCount org_name project_name* # paramCount  

# paramCount  
paramCount is used to identify and count all Parameter Keys in the crawl. 
**Usage:** *./paramCount* (organisation and project name will be prompted for)    
**Usage:** *./paramCount org_name project_name* 

# listAnalysis
listAnalysis displays the crawl meta data for the most recent crawl in a specified project.   
**Note:** Update the code to include your own Botify API key. Change the variable botify_api_token to reflect your token
**Usage:** *./listAnalysis* (organisation and project name will be prompted for)    
**Usage:** *./listAnalysis org_name project_name* # listURLs

# segmentifyLite   
Generates the segmentation regex for the following segments: 
- First level folders
- Second level folders
- Parameter usage
- No. of parameters
- Parameter keys
- No. of folders
- Shopify (if detected)
- SFCC (if detected, and the site is not using "Search-Friendly URLs for B2C Commerce")
**Note:** Update the code to include your own Botify API key. Change the variable **botify_api_token** to reflect your token  
**Note:** The number of URLs found in level 1 folders, level 2 folders and parameter key segments are included as comments after the generated regex. Use these insights to decide which segments slices to keep and which to remove.   
segmentifyLite will process a maximum of 300k URLs.  
**Note:** All level 1 and level 2 segments which are less than 5% of the size of the largest level 1 or level 2 folder found are excluded from the segmentation regex. To amend this threshold change the percentage threshold in the variable **thresholdPercent**.  
**Usage:** *./segmentifyLite* (organisation and project name will be prompted for)    
**Usage:** *./segmentifyLite org_name project_name*   

