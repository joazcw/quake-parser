package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"quake_log_parser/parser" // Import the parser package
)

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
			ID:           gameID,
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

// PlayerRankEntry defines the structure for a player's entry in the global ranking.
// This is used by the /playersranking endpoint.
type PlayerRankEntry struct {
	PlayerName string `json:"player_name"`
	TotalKills int    `json:"total_kills"`
} 