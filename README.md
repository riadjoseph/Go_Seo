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
# segment2ndLevel
segment2ndLevel is used to generate the Botify segmentation regex for all second level folders found in the URL extract.   
**Usage:** segment2ndLevel URL_Extract_file Ex: segment2ndLevel siteurls.csv
# segmentifyLite
segmentifyLite combines segment1stLevel and segment2ndLevel as two functions in a single module.
**Usage:** segmentifyLite URL_Extract_file Ex: segmentifyLite siteurls.csv
# listAnalysis
listAnalysis uses the Botify API and returns key details for the most recent crawl in a specified project.   
**Note** Update the code to include your own Botify API key.   
**Usage:** listAnalysis Username Project_Slug. Ex: listAnalysis botify-org botify-project
 
