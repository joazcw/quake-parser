package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"quake_log_parser/database" // Assuming database package is accessible
	"quake_log_parser/reporter" // Assuming reporter package is accessible
)

const (
	testMongoDBURI         = "mongodb://localhost:27017" // Or use an env variable
	testDatabaseName       = "quake_test_db"
	testGameReportsCollection = "test_game_reports"
)

var testMongoClient *mongo.Client
var testGameCollection *mongo.Collection

// TestMain sets up the MongoDB connection before tests run and closes it after.
func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testMongoClient, err = database.ConnectDB(ctx, testMongoDBURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB for testing: %v", err)
	}
	defer func() {
		if testMongoClient != nil {
			disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelDisconnect()
			if err := testMongoClient.Disconnect(disconnectCtx); err != nil {
				log.Printf("Failed to disconnect test MongoDB client: %v", err)
			}
		}
	}()

	testGameCollection = testMongoClient.Database(testDatabaseName).Collection(testGameReportsCollection)

	// Run tests
	exitVal := m.Run()

	// Clean up the test database after all tests in the package are done (optional, but good practice)
	// You might want to drop the entire collection or database if it's purely for testing.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()
	err = testGameCollection.Drop(cleanupCtx)
	if err != nil {
		log.Printf("Warning: failed to drop test collection %s: %v", testGameCollection.Name(), err)
	}


	os.Exit(exitVal)
}

func TestGetGameByID_Success(t *testing.T) {
	// 1. Prepare test data
	gameID := 101
	expectedReport := reporter.GameReport{
		TotalKills:   5,
		Players:      []string{"Player1", "Player2"},
		Kills:        map[string]int{"Player1": 3, "Player2": 2},
		KillsByMeans: map[string]int{"MOD_RAILGUN": 5},
	}

	// 2. Insert test data into the test collection
	// We need to wrap it in a document that includes the _id for MongoDB
	docToInsert := bson.M{
		"_id":              gameID, // This will be the game ID
		"total_kills":    expectedReport.TotalKills,
		"players":        expectedReport.Players,
		"kills":          expectedReport.Kills,
		"kills_by_means": expectedReport.KillsByMeans,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := testGameCollection.InsertOne(ctx, docToInsert)
	if err != nil {
		t.Fatalf("Failed to insert test game report: %v", err)
	}
	// Defer cleanup for this specific test item
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, delErr := testGameCollection.DeleteOne(cleanupCtx, bson.M{"_id": gameID})
		if delErr != nil {
			t.Logf("Warning: failed to delete test game report with ID %d: %v", gameID, delErr)
		}
	}()

	// 3. Setup router
	router := SetupRouter(testGameCollection) // SetupRouter is from your routers.go

	// 4. Create a new HTTP request
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/games/%d", gameID), nil)
	w := httptest.NewRecorder()

	// 5. Serve HTTP
	router.ServeHTTP(w, req)

	// 6. Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var actualReport reporter.GameReport
	err = json.Unmarshal(w.Body.Bytes(), &actualReport)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Basic check - can be more comprehensive
	if actualReport.TotalKills != expectedReport.TotalKills {
		t.Errorf("Expected TotalKills %d, got %d", expectedReport.TotalKills, actualReport.TotalKills)
	}
	if len(actualReport.Players) != len(expectedReport.Players) {
		t.Errorf("Expected %d players, got %d", len(expectedReport.Players), len(actualReport.Players))
	}
	// Add more detailed comparisons for maps and slices if necessary
}

func TestGetGameByID_NotFound(t *testing.T) {
	router := SetupRouter(testGameCollection)
	nonExistentGameID := 99999

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/games/%d", nonExistentGameID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d for non-existent ID, got %d", http.StatusNotFound, w.Code)
	}

	var errorResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response body: %v", err)
	}
	expectedErrorMsg := fmt.Sprintf("Game with ID %d not found", nonExistentGameID)
	if errorResponse["error"] != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, errorResponse["error"])
	}
}

func TestGetGameByID_InvalidIDFormat(t *testing.T) {
	router := SetupRouter(testGameCollection)

	req, _ := http.NewRequest(http.MethodGet, "/games/not-an-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid ID format, got %d", http.StatusBadRequest, w.Code)
	}
	var errorResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response body: %v", err)
	}
	expectedErrorMsg := "Invalid game ID format"
	if errorResponse["error"] != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, errorResponse["error"])
	}
} 