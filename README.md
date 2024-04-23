# Go_Seo
A series of command line utilities to support SEO actions.   

# folderCount
folderCount is used to identify all first level folders and count the number of URLs in each folder.   
**Usage:** folderCount URL_Extract_file **Ex: folderCount siteurls.csv**
# paramCount
paramCount is used to identify and count all Parameter Keys used.   
**Usage:** paramCount URL_Extract_file **Ex: paramCount siteurls.csv**
# segment1stLevel
segment1stLevel is used to generate the Botify segmentation regex for all first level folders found in the URL extract.   
**Usage:** segment1stLevel URL_Extract_file Regex_Output_File  **Ex: segment1stLevel siteurls.csv segment.txt**
# segment2ndLevel
segment2ndLevel is used to generate the Botify segmentation regex for all second level folders found in the URL extract.   
**Usage:** segment2ndLevel URL_Extract_file Regex_Output_File **Ex: segment2ndLevel siteurls.csv segment.txt**
# listAnalysis
listAnalysis uses the Botify API and returns key details for the most recent crawl in a specified project.   
**Note** Update the code to include your own Botify API key.   
**Usage:** listAnalysis Username Project_Slug. **Ex: listAnalysis botify-org botify-project**
# segmentifyLite
segmentifyLite genertates the regex for the following segments: first level folders, second level folders, parameter usage, no. of parameters and number of folders.  
**Usage:** segmentifyLite URL_Extract_file **Ex: segmentifyLite siteurls.csv**  
**note** The number of URLs found in each level 1 and 2 folder are included as comments after the generated regex. Use these insights to decide which segments to keep and which to remove.   
