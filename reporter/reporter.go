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
type GameReport struct {
	TotalKills   int              `json:"total_kills"`
	Players      []string         `json:"players"`
	Kills        map[string]int   `json:"kills"`
	KillsByMeans map[string]int   `json:"kills_by_means,omitempty"` // omitempty if we only want to show it if present
}

// GenerateReports prints the game reports to standard output in the specified JSON format.
func GenerateReports(games map[string]*parser.Game) {
	fmt.Println("\n--- Game Reports ---")

	if len(games) == 0 {
		fmt.Println("No game data to report.")
		return
	}

	// Prepare a map to hold the structured reports for JSON marshalling
	structuredGameReports := make(map[string]GameReport)

	for gameID, parsedGameData := range games {
		playerNames := make([]string, 0, len(parsedGameData.Players)) // Get players from Players map for consistency
		for name := range parsedGameData.Players {
			if name != "<world>" { // Ensure <world> is not listed as a player
				playerNames = append(playerNames, name)
			}
		}
		sort.Strings(playerNames) // Sort player names for consistent output

		report := GameReport{
			TotalKills:   parsedGameData.TotalKills,
			Players:      playerNames,
			Kills:        parsedGameData.KillsByPlayer, // This already has player scores
			KillsByMeans: parsedGameData.KillsByMeans,
		}
		structuredGameReports[gameID] = report
	}

	jsonData, err := json.MarshalIndent(structuredGameReports, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling game reports to JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))

	// TODO:
	// 1. Implement player ranking report (to be printed after the game reports).
} 