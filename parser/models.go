package parser

// Game struct now uses int for ID
type Game struct {
	ID            int // Changed from string to int
	TotalKills    int
	Players       map[string]*Player
	KillsByPlayer map[string]int
	KillsByMeans  map[string]int
	ClientNames   map[string]string
}

// Player stores information about a player.
type Player struct {
	Name  string
	Kills int // Net kills (actual kills - deaths by <world> or suicides) - This might be redundant if KillsByPlayer is the source of truth for scores.
} 