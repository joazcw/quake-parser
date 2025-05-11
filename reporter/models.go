package reporter

// RankedPlayer stores a player's name and their score for ranking.
type RankedPlayer struct {
	Name  string `json:"name" bson:"name"`
	Score int    `json:"score" bson:"score"`
}

// GameReport defines the structure for the JSON output for a single game.
// This includes the main report and the kills_by_means for the bonus.
// It now includes BSON tags for MongoDB storage.
type GameReport struct {
	ID            int              `json:"id" bson:"_id"`
	TotalKills    int              `json:"total_kills" bson:"total_kills"`
	Players       []string         `json:"players" bson:"players"`
	Kills         map[string]int   `json:"kills" bson:"kills"`
	KillsByMeans  map[string]int   `json:"kills_by_means,omitempty" bson:"kills_by_means,omitempty"`
	// PlayerRanking []RankedPlayer `json:"player_ranking" bson:"player_ranking"` // Removed per-game ranking
}