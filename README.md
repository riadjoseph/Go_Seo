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
segmentifyLite generates the regex for the following segments: first level folders, second level folders, parameter usage, no. of parameters, parameter keys and number of folders.  
**Usage:** segmentifyLite org_name project_name 
**note** The number of URLs found in each level 1 and 2 folder are included as comments after the generated regex. Use these insights to decide which segments to keep and which to remove.   
