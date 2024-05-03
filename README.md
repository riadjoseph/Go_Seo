# Go_Seo
A series of command line utilities to support SEO actions.   

# folderCount
folderCount is used to identify and count all first level folders.   
**Usage:** folderCount URL_Extract_file
# paramCount
paramCount is used to identify and count all Parameter Keys.   
**Usage:** paramCount URL_Extract_file
# listAnalysis
listAnalysis uses the Botify API and returns crawl meta data for the most recent crawl in a specified project.   
**Note** Update the code to include your own Botify API key.   
**Usage:** listAnalysis Username Project_Slug.
# listURLs
Export all URLs (up to a maximum of 1MM) to file (siteurlsExport.csv).  
**Usage:** listURLs      
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
**Usage:** ./segmentifyLite org_name project_name   

# execute_segmentifyLite    
Shell script used to execute segmentifyLite. Prompts for the organisation and project names.  

