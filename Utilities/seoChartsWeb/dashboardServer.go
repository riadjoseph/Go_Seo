package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// Version
var version = "v0.1"

// Colours, symbols etc
var purple = "\033[0;35m"
var green = "\033[0;32m"
var red = "\033[0;31m"
var yellow = "\033[33m"
var bold = "\033[1m"
var reset = "\033[0m"
var checkmark = "\u2713"
var clearScreen = "\033[H\033[2J"

func main() {

	// Display the welcome banner
	displayBanner()

	// Define a handler function for serving the HTML file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Define a handler function for form submission
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the form data from the request
		r.ParseForm()
		organization := r.Form.Get("organization")
		project := r.Form.Get("project")

		// Process the form data as needed (e.g., print it)
		fmt.Printf("Organization: %s, Project: %s\n", organization, project)

		// Respond to the client with a success message or redirect to another page
		http.Redirect(w, r, "/success.html", http.StatusFound)
	})

	// Launch the Chrome browser with the default URL: http://localhost:8080
	launchChrome()

	// Start the HTTP server
	http.ListenAndServe(":8080", nil)

}

func launchChrome() {
	url := "http://localhost:8080"

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		fmt.Println(red + "Error. dashboardServer. You are attempting to launch the server on an unsupported platform" + reset)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error. dashboardServer. Failed to launch Chrome: %v\n", err)
	}
}

// Display the welcome banner
func displayBanner() {

	// Clear the screen
	fmt.Print(clearScreen)

	fmt.Println(green + `
██████╗  █████╗ ███████╗██╗  ██╗██████╗  ██████╗  █████╗ ██████╗ ██████╗ ███████╗███████╗██████╗ ██╗   ██╗███████╗██████╗ 
██╔══██╗██╔══██╗██╔════╝██║  ██║██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
██║  ██║███████║███████╗███████║██████╔╝██║   ██║███████║██████╔╝██║  ██║███████╗█████╗  ██████╔╝██║   ██║█████╗  ██████╔╝
██║  ██║██╔══██║╚════██║██╔══██║██╔══██╗██║   ██║██╔══██║██╔══██╗██║  ██║╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗
██████╔╝██║  ██║███████║██║  ██║██████╔╝╚██████╔╝██║  ██║██║  ██║██████╔╝███████║███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║
╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝
`)
	//Display welcome message
	fmt.Println(purple+"Version:"+reset, version)

	fmt.Println(purple + "\ndashboardServer\n" + reset)
	fmt.Println(purple + "The Go_Seo dashboard server is ON.\n" + reset)
	fmt.Println("Go to http://localhost:8080 to view the dashboard")

}
