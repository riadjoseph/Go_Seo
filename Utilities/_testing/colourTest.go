// colourTest: Display an example of all HTML standard colours
// Written by Jason Vicinanza

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var green = "\033[0;32m"

// Color represents a color name and its ANSI escape sequence for background color
type Color struct {
	Name    string
	Display string
}

func main() {

	clearScreen()

	displayBanner()

	// Define a slice of Color structs with all the colors
	colors := []Color{
		{"AliceBlue", "\033[48;5;231m"},
		{"AntiqueWhite", "\033[48;5;255m"},
		{"Aqua", "\033[48;5;51m"},
		{"Aquamarine", "\033[48;5;122m"},
		{"Azure", "\033[48;5;39m"},
		{"Beige", "\033[48;5;230m"},
		{"Bisque", "\033[48;5;223m"},
		{"Black", "\033[48;5;0m"},
		{"BlanchedAlmond", "\033[48;5;231m"},
		{"Blue", "\033[48;5;21m"},
		{"BlueViolet", "\033[48;5;93m"},
		{"Brown", "\033[48;5;124m"},
		{"BurlyWood", "\033[48;5;138m"},
		{"CadetBlue", "\033[48;5;73m"},
		{"Chartreuse", "\033[48;5;118m"},
		{"Chocolate", "\033[48;5;130m"},
		{"Coral", "\033[48;5;209m"},
		{"CornflowerBlue", "\033[48;5;69m"},
		{"Cornsilk", "\033[48;5;230m"},
		{"Crimson", "\033[48;5;160m"},
		{"Cyan", "\033[48;5;51m"},
		{"DarkBlue", "\033[48;5;18m"},
		{"DarkCyan", "\033[48;5;36m"},
		{"DarkGoldenRod", "\033[48;5;136m"},
		{"DarkGray", "\033[48;5;242m"},
		{"DarkGreen", "\033[48;5;22m"},
		{"DarkKhaki", "\033[48;5;143m"},
		{"DarkMagenta", "\033[48;5;90m"},
		{"DarkOliveGreen", "\033[48;5;58m"},
		{"DarkOrange", "\033[48;5;208m"},
		{"DarkOrchid", "\033[48;5;98m"},
		{"DarkRed", "\033[48;5;52m"},
		{"DarkSalmon", "\033[48;5;173m"},
		{"DarkSeaGreen", "\033[48;5;108m"},
		{"DarkSlateBlue", "\033[48;5;62m"},
		{"DarkSlateGray", "\033[48;5;238m"},
		{"DarkTurquoise", "\033[48;5;44m"},
		{"DarkViolet", "\033[48;5;92m"},
		{"DeepPink", "\033[48;5;198m"},
		{"DeepSkyBlue", "\033[48;5;38m"},
		{"DimGray", "\033[48;5;243m"},
		{"DodgerBlue", "\033[48;5;33m"},
		{"FireBrick", "\033[48;5;124m"},
		{"FloralWhite", "\033[48;5;255m"},
		{"ForestGreen", "\033[48;5;28m"},
		{"Fuchsia", "\033[48;5;201m"},
		{"Gainsboro", "\033[48;5;232m"},
		{"GhostWhite", "\033[48;5;255m"},
		{"Gold", "\033[48;5;220m"},
		{"GoldenRod", "\033[48;5;178m"},
		{"Gray", "\033[48;5;242m"},
		{"Green", "\033[48;5;28m"},
		{"GreenYellow", "\033[48;5;154m"},
		{"HoneyDew", "\033[48;5;248m"},
		{"HotPink", "\033[48;5;205m"},
		{"IndianRed", "\033[48;5;167m"},
		{"Indigo", "\033[48;5;54m"},
		{"Ivory", "\033[48;5;255m"},
		{"Khaki", "\033[48;5;143m"},
		{"Lavender", "\033[48;5;231m"},
		{"LavenderBlush", "\033[48;5;255m"},
		{"LawnGreen", "\033[48;5;118m"},
		{"LemonChiffon", "\033[48;5;230m"},
		{"LightBlue", "\033[48;5;117m"},
		{"LightCoral", "\033[48;5;203m"},
		{"LightCyan", "\033[48;5;195m"},
		{"LightGoldenRodYellow", "\033[48;5;227m"},
		{"LightGray", "\033[48;5;250m"},
		{"LightGreen", "\033[48;5;119m"},
		{"LightPink", "\033[48;5;218m"},
		{"LightSalmon", "\033[48;5;216m"},
		{"LightSeaGreen", "\033[48;5;37m"},
		{"LightSkyBlue", "\033[48;5;111m"},
		{"LightSlateGray", "\033[48;5;103m"},
		{"LightSteelBlue", "\033[48;5;146m"},
		{"LightYellow", "\033[48;5;230m"},
		{"Lime", "\033[48;5;46m"},
		{"LimeGreen", "\033[48;5;77m"},
		{"Linen", "\033[48;5;231m"},
		{"Magenta", "\033[48;5;201m"},
		{"Maroon", "\033[48;5;52m"},
		{"MediumAquaMarine", "\033[48;5;79m"},
		{"MediumBlue", "\033[48;5;19m"},
		{"MediumOrchid", "\033[48;5;134m"},
		{"MediumPurple", "\033[48;5;98m"},
		{"MediumSeaGreen", "\033[48;5;71m"},
		{"MediumSlateBlue", "\033[48;5;63m"},
		// Add more colors here...
	}

	// Print the text field with each color
	for _, color := range colors {
		fmt.Printf("%s%s\033[0m\n", color.Display, color.Name)
	}
}

// Display the welcome banner
func displayBanner() {

	//Banner
	//https://patorjk.com/software/taag/#p=display&c=bash&f=ANSI%20Shadow&t=SegmentifyLite
	fmt.Println(green + `
██████╗ ██████╗ ██╗      ██████╗ ██╗   ██╗██████╗ ████████╗███████╗███████╗████████╗
██╔════╝██╔═══██╗██║     ██╔═══██╗██║   ██║██╔══██╗╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝
██║     ██║   ██║██║     ██║   ██║██║   ██║██████╔╝   ██║   █████╗  ███████╗   ██║   
██║     ██║   ██║██║     ██║   ██║██║   ██║██╔══██╗   ██║   ██╔══╝  ╚════██║   ██║   
╚██████╗╚██████╔╝███████╗╚██████╔╝╚██████╔╝██║  ██║   ██║   ███████╗███████║   ██║   
 ╚═════╝ ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚══════╝   ╚═╝
`)
}

// Function to clear the screen
func clearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
