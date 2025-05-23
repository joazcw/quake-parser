package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware

	"go.mongodb.org/mongo-driver/mongo"
	"quake_log_parser/database"
	_ "quake_log_parser/docs" // docs is generated by Swag CLI, you need to import it.
	"quake_log_parser/parser"
	"quake_log_parser/reporter"
)

// SetupRouter initializes and configures the Gin router with all API endpoints.
// It takes the MongoDB collection for game reports as an argument.
func SetupRouter(gameCollection *mongo.Collection) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	// This allows requests from http://localhost:8000 (your frontend)
	// and specifies allowed methods, headers, etc.
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8000", "http://localhost:8080"} // Added localhost:8080 for Swagger UI access from browser
	// You can also use config.AllowAllOrigins = true for wider access, but specific origins are safer.
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"} // Explicitly allow methods
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	// config.ExposeHeaders = []string{"Content-Length"}
	// config.AllowCredentials = true // If you were using cookies or auth headers that need credentials
	// config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))

	// Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API endpoint annotations start here

	// GetGameByID godoc
	// @Summary Get a single game report by its ID
	// @Description Retrieves full details for a specific game based on its unique ID.
	// @Tags games
	// @Accept json
	// @Produce json
	// @Param id path int true "Game ID"
	// @Success 200 {object} reporter.GameReport "Successfully retrieved game report"
	// @Failure 400 {object} ErrorResponse "Invalid game ID format"
	// @Failure 404 {object} ErrorResponse "Game not found"
	// @Failure 500 {object} ErrorResponse "Failed to retrieve game data"
	// @Router /games/{id} [get]
	router.GET("/games/:id", func(c *gin.Context) {
		gameIDStr := c.Param("id")
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
			return
		}

		// Create a new context for this specific request
		reqCtx, reqCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer reqCancel()

		report, err := database.GetGameReportByID(reqCtx, gameCollection, gameID)
		if err != nil {
			log.Printf("Error retrieving game ID %d from database: %v", gameID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve game data"})
			return
		}

		if report == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Game with ID %d not found", gameID)})
			return
		}

		c.JSON(http.StatusOK, report)
	})

	// GetAllGames godoc
	// @Summary Get all game reports
	// @Description Retrieves a list of all game reports stored in the database, sorted by game ID.
	// @Tags games
	// @Accept json
	// @Produce json
	// @Success 200 {array} reporter.GameReport "Successfully retrieved list of game reports"
	// @Failure 500 {object} ErrorResponse "Failed to retrieve game reports"
	// @Router /games [get]
	router.GET("/games", func(c *gin.Context) {
		// Create a new context for this specific request
		reqCtx, reqCancel := context.WithTimeout(context.Background(), 15*time.Second) // Slightly longer timeout for potentially larger data
		defer reqCancel()

		reports, err := database.GetAllGameReports(reqCtx, gameCollection)
		if err != nil {
			log.Printf("Error retrieving all game reports from database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve game reports"})
			return
		}

		// If reports is nil (which GetAllGameReports now prevents, returning [] instead),
		// or if it's empty, return an empty JSON array [] instead of null.
		// Gin handles empty slices correctly, marshalling them to [].
		c.JSON(http.StatusOK, reports)
	})

	// UploadLogFile godoc
	// @Summary Upload a Quake log file for processing
	// @Description Uploads a game log file (.log). The server parses it, generates game reports, and stores them.
	// @Tags games
	// @Accept multipart/form-data
	// @Produce json
	// @Param logFile formData file true "The Quake log file to upload"
	// @Success 201 {object} UploadResponse "Log file processed and game(s) stored successfully"
	// @Success 200 {object} UploadResponse "Log file processed. No games found to report (already OK if file is valid but empty of games)"
	// @Failure 400 {object} ErrorResponse "Error retrieving/parsing uploaded file or invalid file format"
	// @Failure 500 {object} ErrorResponse "Server error during file processing or storage"
	// @Router /games/upload [post]
	router.POST("/games/upload", func(c *gin.Context) {
		// Source
		fileHeader, err := c.FormFile("logFile")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error retrieving uploaded file: %v", err)})
			return
		}

		uploadedFile, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error opening uploaded file: %v", err)})
			return
		}
		defer uploadedFile.Close()

		// Create a temporary file
		// The first argument "" means use the default directory for temporary files.
		// The second argument "upload-*.log" is a pattern for the temporary file name.
		tempFile, err := os.CreateTemp("", "upload-*.log")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creating temporary file: %v", err)})
			return
		}
		// Crucial: Ensure the temporary file is closed and removed.
		// Close it first, then remove. Defer runs in LIFO order.
		defer os.Remove(tempFile.Name()) // Remove the file
		defer tempFile.Close()          // Close the file

		// Copy uploaded file content to the temporary file
		_, err = io.Copy(tempFile, uploadedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error copying to temporary file: %v", err)})
			return
		}

		// Ensure all content is written to disk before parsing
		if err := tempFile.Sync(); err != nil {
		    log.Printf("Warning: failed to sync temporary file %s: %v", tempFile.Name(), err)
		    // Depending on the OS, this might not be strictly necessary after Close(),
		    // but explicit sync can be safer before another process (parser) reads it.
		}

		tempFilePath := tempFile.Name()

		// --- Parse the Log File ---
		// Use a new context for potentially long-running operations
		procCtx, procCancel := context.WithTimeout(c.Request.Context(), 60*time.Second) // e.g., 1 minute timeout for processing
		defer procCancel()

		parsedGames, err := parser.ParseLogFile(tempFilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error parsing log file: %v", err)})
			return
		}

		if len(parsedGames) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "Log file processed. No games found to report.", "games_processed": 0})
			return
		}

		// --- Format the Report ---
		reportsToStore := reporter.FormatGameData(parsedGames)

		// --- Store in Database ---
		reportsForDB := make(map[int]interface{}, len(reportsToStore))
		for k, v := range reportsToStore {
			reportsForDB[k] = v
		}

		err = database.StoreGameReports(procCtx, gameCollection, reportsForDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error storing game reports: %v", err)})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":         fmt.Sprintf("Log file processed and %d game(s) stored successfully.", len(reportsToStore)),
			"games_processed": len(reportsToStore),
		})
	})

	// DeleteAllGames godoc
	// @Summary Delete all game reports
	// @Description Permanently removes all game reports from the database.
	// @Tags games
	// @Accept json
	// @Produce json
	// @Success 200 {object} SuccessResponse "All game reports deleted successfully"
	// @Failure 500 {object} ErrorResponse "Failed to delete all game reports"
	// @Router /games [delete]
	router.DELETE("/games", func(c *gin.Context) {
		// Use a new context for the database operation
		dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 30*time.Second) // Context for the database call
		defer dbCancel()

		err := database.DeleteAllGameReportsFromDB(dbCtx, gameCollection)
		if err != nil {
			log.Printf("Error deleting all game reports from database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete all game reports"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "All game reports deleted successfully"})
		// Alternatively, could use http.StatusNoContent and send no body:
		// c.Status(http.StatusNoContent)
	})

	// DeleteGameByID godoc
	// @Summary Delete a specific game report by its ID
	// @Description Permanently removes a single game report based on its unique ID.
	// @Tags games
	// @Accept json
	// @Produce json
	// @Param id path int true "Game ID"
	// @Success 200 {object} SuccessResponse "Game deleted successfully"
	// @Failure 400 {object} ErrorResponse "Invalid game ID format"
	// @Failure 404 {object} ErrorResponse "Game not found"
	// @Failure 500 {object} ErrorResponse "Failed to delete game report"
	// @Router /games/{id} [delete]
	router.DELETE("/games/:id", func(c *gin.Context) {
		gameIDStr := c.Param("id")
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
			return
		}

		// Use a new context for the database operation
		dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer dbCancel()

		deleted, err := database.DeleteGameReportByID(dbCtx, gameCollection, gameID)
		if err != nil {
			log.Printf("Error deleting game ID %d from database: %v", gameID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game report"})
			return
		}
		if deleted == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Game with ID %d not found", gameID)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Game with ID %d deleted successfully", gameID)})
	})

	// GetPlayersRanking godoc
	// @Summary Get aggregated player rankings across all games
	// @Description Retrieves a list of players ranked by their total kills across all recorded games.
	// @Tags rankings
	// @Accept json
	// @Produce json
	// @Success 200 {array} reporter.PlayerRankEntry "Successfully retrieved player rankings"
	// @Failure 500 {object} ErrorResponse "Failed to retrieve player rankings"
	// @Router /playersranking [get]
	router.GET("/playersranking", func(c *gin.Context) {
		// Create a new context for this specific request
		reqCtx, reqCancel := context.WithTimeout(context.Background(), 30*time.Second) // Longer timeout for aggregation
		defer reqCancel()

		allReports, err := database.GetAllGameReports(reqCtx, gameCollection)
		if err != nil {
			log.Printf("Error retrieving all game reports for ranking: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data for player rankings"})
			return
		}

		aggregatedKills := make(map[string]int)
		for _, report := range allReports {
			// The report.Kills map is player -> kills in that specific game
			for player, killsInGame := range report.Kills {
				aggregatedKills[player] += killsInGame
			}
		}

		// Convert map to slice of PlayerRankEntry for sorting and JSON response
		var playerRanks []reporter.PlayerRankEntry
		for player, totalKills := range aggregatedKills {
			playerRanks = append(playerRanks, reporter.PlayerRankEntry{PlayerName: player, TotalKills: totalKills})
		}

		// Sort players by total kills in descending order
		sort.Slice(playerRanks, func(i, j int) bool {
			return playerRanks[i].TotalKills > playerRanks[j].TotalKills
		})

		c.JSON(http.StatusOK, playerRanks)
	})

	return router
} 