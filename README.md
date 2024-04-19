# Go_Seo
A series of command line utilities to support SEO actions.   

# folderCount
folderCount is used to identify all first level folders and count the number of URLs in each folder.   
**Usage:** folderCount URL_Extract_file Ex: folderCount siteurls.csv
# paramCount
paramCount is used to identify and count all Parameter Keys used.   
**Usage:** paramCount URL_Extract_file Ex: paramCount siteurls.csv
# segment1stLevel
segment1stLevel is used to generate the Botify segmentation regex for all first level folders found in the URL extract.   
**Usage:** segment1stLevel URL_Extract_file Ex: segment1stLevel siteurls.csv
# listAnalysis
listAnalysis uses the Botify API and returns key details for the most recent crawl in a specified project.   
**Usage:** listAnalysis Username Project_Slug. Ex: listAnalysis botify-org botify-project
