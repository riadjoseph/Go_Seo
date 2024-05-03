#!/bin/bash

# Shell script to execute segmentifyLite.
# This script should be placed in the same folder as the segmentifyLite binary

GREEN='\033[0;32m'
NC='\033[0m' # No Color

clear
echo -e "${GREEN}"
echo "██████╗ ███████╗ ██████╗ ███╗   ███╗███████╗███╗   ██╗████████╗██╗███████╗██╗   ██╗██╗     ██╗████████╗███████╗"
echo "██╔════╝██╔════╝██╔════╝ ████╗ ████║██╔════╝████╗  ██║╚══██╔══╝██║██╔════╝╚██╗ ██╔╝██║     ██║╚══██╔══╝██╔════╝"
echo "███████╗█████╗  ██║  ███╗██╔████╔██║█████╗  ██╔██╗ ██║   ██║   ██║█████╗   ╚████╔╝ ██║     ██║   ██║   █████╗"
echo "╚════██║██╔══╝  ██║   ██║██║╚██╔╝██║██╔══╝  ██║╚██╗██║   ██║   ██║██╔══╝    ╚██╔╝  ██║     ██║   ██║   ██╔══╝"
echo "███████║███████╗╚██████╔╝██║ ╚═╝ ██║███████╗██║ ╚████║   ██║   ██║██║        ██║   ███████╗██║   ██║   ███████╗"
echo "╚══════╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝   ╚═╝   ╚══════╝"
echo -e "${NC}"

# Prompt for the "Organisation" input
read -p "Enter Organisation. Press Enter to exit: " organisation

# Check if Organisation input is empty, if so Exit
if [ -z "$organisation" ]; then
    echo "Thank you for using segmentifyLite. Goodbye!"
    exit 1
fi

read -p "Enter project. Press Enter to exit: " project

# Check if project input is empty, if so Exit
if [ -z "$project" ]; then
    echo "Thank you for using segmentifyLite. Goodbye!."
    exit 1
fi

# Execute the segmentifyLite passing the inputs as command-line arguments
./segmentifyLite "$organisation" "$project"
