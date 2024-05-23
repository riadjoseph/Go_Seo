import requests

from env import TOKEN, ORGANIZATION, API_URL

HEADERS = {"Content-Type": "application/json", "Authorization": "Token " + TOKEN}


"""
    example response:
    {
        "slug": "my-awesome-project",
        "name": "My awesome project",
        "creationDate": "2022-06-29T10:57:07.455074+01:00",
        "settings": {
            "domains": [
                {
                    "protocol": "both",
                    "domain": "google.com",
                    "allow_subdomains": true,
                    "mobile": false
                }
            ],
            "blacklistedDomains": null
        }
    }
    """
def create_project(project_name, start_url, max_url):
    payload = {
        "name": project_name,
        "start_url": start_url,
        "max_nb_pages": max_url,
        "owner": ORGANIZATION,
        "with_scheduling": "off",
        "crawl_subdomains": "on",
    }
    return requests.post(API_URL + "/v1/projects", json=payload, headers=HEADERS)


"""
    example response:
    {
        "status": 201,
        "analysis_slug": "20220920",
        "message": "Analysis has been created and launched successfully"
    }
    """
def launch_crawl(project_name):
    return requests.post(
        "{API_URL}/v1/analyses/{ORGANIZATION}/{project_name}/create/launch".format(
            API_URL=API_URL,
            ORGANIZATION=ORGANIZATION,
            project_name=project_name
        ),
        headers=HEADERS
    )
