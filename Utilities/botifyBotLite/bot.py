import argparse

import csv
from display import bcolors, error_display, success_display
from functions import create_project, launch_crawl

CREATE_PROJECT_ENDPOINT = "/v1/projects"

parser = argparse.ArgumentParser(description="Crawl")
parser.add_argument(
    "-i", dest="input", type=str, help="Path to the CSV file to parse"
)

args = parser.parse_args()

with open(args.input) as file:
    reader = csv.reader(file)
    next(reader)  # Header line
    for row in reader:
        project_name = row[1]
        url = row[0]
        max_urls = row[2]
        # creating the project
        res = create_project(project_name=project_name, start_url=url, max_url=max_urls)
        print(
            f"##### {bcolors.BOLD}{bcolors.UNDERLINE}{project_name}:{url}{bcolors.ENDC} #####"
        )
        print("|")

        if res.status_code != 201:
            error_display(
                message="Error. bot.py. Cannot create the crawl. You may have provided a duplicate project prefix. Try again and use a different one.", status_code=res.status_code
            )
            continue

        create_res = res.json()
        success_display(message="Project was successfully created")

        # Launching a crawl
        launch_res = launch_crawl(create_res["slug"])

        if launch_res.status_code != 201:
            error_display(
                message="Error. bot.py. Cannot launch the crawl.", status_code=launch_res.status_code
            )
            continue

        launch_res_as_json = launch_res.json()
        success_display(
            message="Crawl was successfully launched {analysis_slug}".format(
                analysis_slug=launch_res_as_json["analysis_slug"]
            )
        )
        print("----------------------------------------------------------")
