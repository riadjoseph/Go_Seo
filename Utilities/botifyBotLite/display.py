from env import DEBUG


class bcolors:
    HEADER = "\033[95m"
    OKBLUE = "\033[94m"
    OKCYAN = "\033[96m"
    OKGREEN = "\033[92m"
    WARNING = "\033[93m"
    FAIL = "\033[91m"
    ENDC = "\033[0m"
    BOLD = "\033[1m"
    UNDERLINE = "\033[4m"


def error_display(message, status_code):
    print(f"|{bcolors.FAIL}  {message} {bcolors.ENDC}")
    if DEBUG:
        print(f"|{bcolors.OKBLUE}  [DEBUG] Status code {status_code}{bcolors.ENDC}")
        print(f"|")


def success_display(message):
    print(f"|{bcolors.OKGREEN}  {message}{bcolors.ENDC}")
    print("|")
