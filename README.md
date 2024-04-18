# Go_Seo
A series of command line utilities to support SEO actions.   

# foldercount
foldercount is used to identify the first level folder in the URL and to count the number of instances found in the URL file.   
Usage: go run foldercount.go URL_Extract_file Ex: go run foldercount.go siteurls.csv
# paramcount
paramcount is used to identify and count all Parameter Keys used.   
Usage: go run paramcount.go URL_Extract_file Ex: go run paramcount.go siteurls.csv
# segment1stlevel
segment1stlevel is used to generate the Botify segmentation regex for all first level folders found in the URL extract.   
Usage: go run segment1stlevel.go URL_Extract_file Ex: go run segment1stlevel.go siteurls.csv
# listAnalysis
listAnalysis uses the Botify API and returns key details for the latest crawl in a specified project.   
Usage: listAnalysis Username Project_Slug. Ex: listAnalysis botify-org botify-project
