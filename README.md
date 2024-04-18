# Go_Seo
A series of command line utilities to support SEO actions.   

# folderCount
folderCount is used to identify the first level folder in the URL and to count the number of instances found in the URL file.   
**Usage:** go run folderCount.go URL_Extract_file Ex: go run folderCount.go siteurls.csv
# paramCount
paramCount is used to identify and count all Parameter Keys used.   
**Usage:** go run paramCount.go URL_Extract_file Ex: go run paramCount.go siteurls.csv
# segment1stLevel
segment1stLevel is used to generate the Botify segmentation regex for all first level folders found in the URL extract.   
**Usage:** go run segment1stLevel.go URL_Extract_file Ex: go run segment1stLevel.go siteurls.csv
# listAnalysis
listAnalysis uses the Botify API and returns key details for the latest crawl in a specified project.   
**Usage:** listAnalysis Username Project_Slug. Ex: listAnalysis botify-org botify-project
