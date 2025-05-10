# Quake Log Parser

This project parses log files from Quake 3 Arena to generate game reports and player statistics.

## Requirements

- Go (version 1.21 or higher recommended)

## Setup

1.  Clone the repository.
2.  Navigate to the project directory: `cd quake_log_parser`
3.  Download the Quake game log (e.g., `qgames.log`) and place it in the `data/` directory (or provide the path via command line).

## Usage

Build the project:
```sh
go build
```

Run the parser:
```sh
./quake_log_parser -logfile=data/qgames.log 
# Or if your main.go handles a default path or no flags yet:
# ./quake_log_parser
```

(Command-line flags for log file path will be added) 