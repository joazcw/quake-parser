package parser

// Game stores all information for a single game match.
type Game struct {
	ID            string             // e.g., "game_1"
	TotalKills    int                // Total kills in the game
	Players       map[string]*Player // Player name -> Player details (stores latest known player object by name)
	KillsByPlayer map[string]int     // Player name -> kill count (net score)
	KillsByMeans  map[string]int     // Death cause -> count (for bonus)

	// Internal tracking, not directly part of the JSON output structure
	ClientNames map[string]string // clientID -> most recent playerName for this game
}

// Player stores information about a player.
type Player struct {
	Name  string
	Kills int // Net kills (actual kills - deaths by <world> or suicides) - This might be redundant if KillsByPlayer is the source of truth for scores.
} 