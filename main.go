package main

import (
	"fmt"
	"log"

	"quake_log_parser/parser" // Assuming 'quake_log_parser' is your module name from go.mod
	"quake_log_parser/reporter" // Added import for reporter
)

func main() {
	fmt.Println("Quake Log Parser")

	logFilePath := "data/games.log" // Define log file path

	// Call parser.ParseLogFile()
	games, err := parser.ParseLogFile(logFilePath)
	if err != nil {
		log.Fatalf("Error parsing log file: %v", err)
	}

	fmt.Printf("Successfully parsed %d game(s) from %s\n", len(games), logFilePath)

	// Call reporter.GenerateReports()
	reporter.GenerateReports(games)
} 