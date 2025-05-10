package main

import (
	"context"
	"flag" // Added for command-line flags
	"fmt"
	"log"
	"time"

	"quake_log_parser/database" // Import the new database package
	"quake_log_parser/parser"
	"quake_log_parser/reporter"
	// MongoDB driver imports are now handled in database/database.go
)

const (
	// mongoDBURI         = "mongodb://localhost:27017" // Defined in database package
	databaseName       = "quake_reports_db"
	gameReportsCollection = "game_reports"
)

func main() {
	// Define command-line flags
	noStore := flag.Bool("no-store", false, "Set to true to skip storing reports in MongoDB")
	printStored := flag.Bool("print-stored", true, "Set to false to skip printing all stored reports from MongoDB")
	deleteAll := flag.Bool("delete-all", false, "Set to true to delete all reports from MongoDB before other operations") // New flag
	flag.Parse()

	fmt.Println("Quake Log Parser")

	// Setup MongoDB connection
	// Context for MongoDB operations
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout slightly
	defer dbCancel()

	// Connect to MongoDB using the function from the database package
	mongoClient, err := database.ConnectDB(dbCtx, database.DefaultMongoDBURI) // Using default URI from database pkg
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelDisconnect()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err) // Use Printf for non-fatal defer errors
		}
		fmt.Println("Disconnected from MongoDB.")
	}()

	// Handle -delete-all flag first if set
	if *deleteAll {
		fmt.Println("\nProcessing -delete-all flag...")
		err = database.DeleteAllGameReportsFromDB(dbCtx, mongoClient.Database(databaseName).Collection(gameReportsCollection))
		if err != nil {
			log.Printf("Error deleting all game reports from MongoDB: %v", err)
			// Decide if we should exit here or continue. For now, we log and continue.
		} else {
			fmt.Println("Successfully processed -delete-all flag.")
		}
		// If -delete-all is the only intended operation, you might want to exit here.
		// For now, it will continue with other operations.
	}

	// --- Log Parsing ---
	logFilePath := "data/games.log"
	parsedGames, err := parser.ParseLogFile(logFilePath)
	if err != nil {
		log.Fatalf("Error parsing log file: %v", err)
	}
	fmt.Printf("Successfully parsed %d game(s) from %s\n", len(parsedGames), logFilePath)

	// --- Format Data for Reporting/Storage ---
	reportsToProcess := reporter.FormatGameData(parsedGames)

	// --- Print Reports to Console (Optional) ---
	if len(reportsToProcess) > 0 {
		reporter.PrintGameReportsToConsole(reportsToProcess)
	}

	gameCollection := mongoClient.Database(databaseName).Collection(gameReportsCollection)

	if !*noStore { // Check the flag before storing
		if len(reportsToProcess) > 0 {
			// Convert map[int]reporter.GameReport to map[int]interface{} for StoreGameReports
			reportsForDB := make(map[int]interface{}, len(reportsToProcess))
			for k, v := range reportsToProcess {
				reportsForDB[k] = v
			}

			err = database.StoreGameReports(dbCtx, gameCollection, reportsForDB)
			if err != nil {
				log.Printf("Error storing game reports in MongoDB: %v", err) // Log as non-fatal for now
			} else {
				fmt.Println("Game reports processing for MongoDB completed.")
			}
		}
	} else {
		fmt.Println("Skipping database storage due to -no-store flag.")
	}

	if *printStored { // Check the flag before printing stored reports
		fmt.Println("\nAttempting to retrieve and print all stored game reports from MongoDB...")
		err = database.PrintAllStoredGames(dbCtx, gameCollection)
		if err != nil {
			log.Printf("Error printing stored game reports from MongoDB: %v", err)
		}
	} else {
		fmt.Println("\nSkipping printing stored reports from MongoDB due to -print-stored=false flag.")
	}

	fmt.Println("\nApplication finished.")
} 