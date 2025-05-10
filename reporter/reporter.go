package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"quake_log_parser/parser" // Import the parser package
)

// GameReport defines the structure for the JSON output for a single game.
// This includes the main report and the kills_by_means for the bonus.
// It now includes BSON tags for MongoDB storage.
type GameReport struct {
	TotalKills   int              `json:"total_kills" bson:"total_kills"`
	Players      []string         `json:"players" bson:"players"`
	Kills        map[string]int   `json:"kills" bson:"kills"`
	KillsByMeans map[string]int   `json:"kills_by_means,omitempty" bson:"kills_by_means,omitempty"`
}

// FormatGameData converts the raw parsed game data into a map of GameReport structs,
// suitable for JSON marshalling or database storage.
// It now accepts map[int]*parser.Game and returns map[int]GameReport.
func FormatGameData(games map[int]*parser.Game) map[int]GameReport {
	structuredGameReports := make(map[int]GameReport)

	for gameID, parsedGameData := range games {
		playerNames := make([]string, 0, len(parsedGameData.Players))
		for name := range parsedGameData.Players {
			if name != "<world>" { 
				playerNames = append(playerNames, name)
			}
		}
		sort.Strings(playerNames) 

		report := GameReport{
			TotalKills:   parsedGameData.TotalKills,
			Players:      playerNames,
			Kills:        parsedGameData.KillsByPlayer,
			KillsByMeans: parsedGameData.KillsByMeans,
		}
		structuredGameReports[gameID] = report
	}
	return structuredGameReports
}

// PrintGameReportsToConsole takes the formatted game reports and prints them to standard output as JSON.
// It now accepts map[int]GameReport.
func PrintGameReportsToConsole(reports map[int]GameReport) {
	fmt.Println("\n--- Game Reports (Console Output) ---")

	if len(reports) == 0 {
		fmt.Println("No game data to report.")
		return
	}

	jsonData, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling game reports to JSON for console: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))

	// TODO:
	// 1. Implement player ranking report (to be printed after the game reports).
} 