package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings" // Added for strings.Contains
	// "regexp"
	// "strings"
)

// Pre-compile regexes for efficiency
var (
	reClientUserinfoChanged = regexp.MustCompile(`^.*?ClientUserinfoChanged: (\d+) n\\([^\\\\]+)\\.*playerNameIsHere>([^<]+)<\\x{005c}*t\\(\\d+\\).*$|^.*?ClientUserinfoChanged: (\d+) n\\(([^\\\\]+))\\\\t.*`)
	reKill = regexp.MustCompile(`^.*?Kill: (\d+) (\d+) (\d+): (.*) killed (.*) by (MOD_[A-Z_]+)$`)
)

// ParseLogFile reads and parses the Quake log file.
// It returns a map of game data, keyed by game ID, and an error if any occurs.
func ParseLogFile(filePath string) (map[string]*Game, error) {
	// fmt.Println("Parsing log file:", filePath) // Can be noisy

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	games := make(map[string]*Game)
	var currentGame *Game
	gameCounter := 0

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.Contains(line, "InitGame:") {
			// Finalize previous game if any (though ShutdownGame should handle this)
			if currentGame != nil {
				// This case should ideally not be hit if logs are well-formed with ShutdownGame
				// log.Printf("Warning: InitGame encountered while a game (%s) was already in progress. Finalizing previous.", currentGame.ID)
				currentGame = nil
			}
			gameCounter++
			gameID := fmt.Sprintf("game_%d", gameCounter)
			currentGame = &Game{
				ID:            gameID,
				TotalKills:    0,
				Players:       make(map[string]*Player),
				KillsByPlayer: make(map[string]int),
				KillsByMeans:  make(map[string]int),
				ClientNames:   make(map[string]string),
			}
			games[gameID] = currentGame
			// fmt.Printf("Started %s\n", gameID) // Optional: for debugging
		} else if strings.Contains(line, "ShutdownGame:") {
			if currentGame != nil {
				currentGame = nil // End of current game processing
				// fmt.Printf("Ended %s\n", currentGame.ID) // Optional: for debugging
			}
		} else if currentGame != nil { // Process lines only if we are inside a game
			// Attempt to parse ClientUserinfoChanged
			matches := reClientUserinfoChanged.FindStringSubmatch(line)
			if len(matches) > 0 {
				var clientID, playerName string
				// The regex has two alternate patterns due to observed variations in log format.
				// Check which group of capturing parentheses has the match.
				if matches[1] != "" && matches[2] != "" { //playerNameIsHere variant
					clientID = matches[1]
					playerName = matches[2]
				} else if matches[3] != "" && matches[4] != "" { // standard n\\NAME\\t variant
					clientID = matches[3]
					playerName = matches[4]
				}

				if clientID != "" && playerName != "" {
					playerName = strings.TrimSpace(playerName)
					currentGame.ClientNames[clientID] = playerName
					// Ensure player is in KillsByPlayer and Players map
					if _, ok := currentGame.KillsByPlayer[playerName]; !ok {
						currentGame.KillsByPlayer[playerName] = 0
					}
					if _, ok := currentGame.Players[playerName]; !ok {
						currentGame.Players[playerName] = &Player{Name: playerName, Kills: 0} // Kills here might be redundant
					}
				} else {
					// log.Printf("Line %d: Failed to extract ClientID/PlayerName from ClientUserinfoChanged: %s (Matches: %v)", lineNumber, line, matches)
				}
			} else {
				// Attempt to parse Kill line if not ClientUserinfoChanged
				killMatches := reKill.FindStringSubmatch(line)
				if len(killMatches) == 7 {
					// killerClientID := killMatches[1] // not directly used for scoring logic with current approach
					// victimClientID := killMatches[2]   // not directly used
					// meansOfDeathID := killMatches[3] // useful if mapping ID to string is needed later
					killerName := strings.TrimSpace(killMatches[4])
					victimName := strings.TrimSpace(killMatches[5])
					mod := strings.TrimSpace(killMatches[6])

					currentGame.TotalKills++
					currentGame.KillsByMeans[mod]++

					// Ensure victim is in score tracking, even if <world> killed them first
					if _, ok := currentGame.KillsByPlayer[victimName]; !ok && victimName != "<world>" {
						currentGame.KillsByPlayer[victimName] = 0
						currentGame.Players[victimName] = &Player{Name: victimName, Kills: 0}
					}

					if killerName == "<world>" {
						if victimName != "<world>" { // Should not happen but good check
							currentGame.KillsByPlayer[victimName]--
						}
					} else {
						// Ensure killer is in score tracking
						if _, ok := currentGame.KillsByPlayer[killerName]; !ok {
							currentGame.KillsByPlayer[killerName] = 0
							currentGame.Players[killerName] = &Player{Name: killerName, Kills: 0}
						}

						if killerName == victimName { // Suicide
							currentGame.KillsByPlayer[killerName]--
						} else { // Player killed another player
							currentGame.KillsByPlayer[killerName]++
						}
					}
				} else if len(killMatches) > 0 && len(killMatches) != 7 {
					// log.Printf("Line %d: Kill line regex matched but got %d groups, expected 7: %s", lineNumber, len(killMatches), line)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s (line %d): %w", filePath, lineNumber, err)
	}

	return games, nil
} 