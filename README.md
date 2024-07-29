# Go_Seo
A series of tools used to support SEO actions. Developed in Go.   

## seoBusinessInsights   
Undercover the value of organic traffic with this business insights broadsheet for Botify.

Charts included:
- Compound growth indicators
- Revenue & visits
- Visits per order
- Order volume
- Order value
- Revenue & visits river chart
- Visit value
- Branded & non branded keyword cloud
- Winning branded and non branded keywords
- Detailed KPI insights 

**Usage:**  
Required environment variables:  

export envBotifyAPIToken="_your_botify_token_"  
export envInsightsFolder="./seoInsightsCache"  
export envInsightsLogFolder="."  
export envInsightsHostingMode="local" (if hosting on a Docker container, change to _docker_)  

Initialization file. The default configuration is as follows, adjust if hosting in a location other than localhost:   

protocol=http  
port=8080  
hostname=localhost   

## segmentifyLite   
Generates the segmentation regex for the following segments: 

- First level folders
- Second level folders
- Parameter usage
- No. of parameters
- Parameter keys
- No. of folders
- Static resources
- Shopify (if detected)
- SFCC (if detected, and the site is not using "Search-Friendly URLs for B2C Commerce")

**Usage:**  

Required environment variables:  

export envBotifyAPIToken="your_botify_token"  
export envSegmentifyLiteFolder="./segmentifyLiteCache"  
export envSegmentifyLiteLogFolder="."  
export envSegmentifyLiteHostingMode="local"  

Initialization file. The default configuration is as follows, adjust if hosting in a location other than localhost:   

protocol=http  
port=8080  
hostname=localhost   

![1_Ifpd_HtDiK9u6h68SZgNuA](https://github.com/user-attachments/assets/6bc49ad4-312e-4cc6-abdb-9fccbf0e7ca3)
