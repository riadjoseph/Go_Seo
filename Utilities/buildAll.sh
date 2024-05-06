#!/bin/bash

clear
echo "** Building Go_SEO utilities"

echo - building folderCount.go
go build folderCount.go
echo - building paramCount.go
go build paramCount.go
echo - building listAnalysis.go
go build listAnalysis.go
echo - building listURLs.go
go build listURLs.go
echo - building segmentifyLite.go
go build segmentifyLite.go

echo "** Building Go_SEO utilities - Done!"



