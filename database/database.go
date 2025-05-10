package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	// "time" // Removed as it's not currently used in this file

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// We will need "quake_log_parser/reporter" for GameReport type in StoreGameReports
)

const (
	DefaultMongoDBURI = "mongodb://localhost:27017"
)

// ConnectDB establishes a connection to MongoDB and returns the client.
// The caller is responsible for deferring client.Disconnect().
func ConnectDB(ctx context.Context, uri string) (*mongo.Client, error) {
	if uri == "" {
		uri = DefaultMongoDBURI
	}
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB at %s: %w", uri, err)
	}

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		// Disconnect if ping fails to avoid lingering connections
		if dErr := client.Disconnect(context.Background()); dErr != nil {
			log.Printf("Failed to disconnect after ping failure: %v", dErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB at %s: %w", uri, err)
	}
	fmt.Println("Successfully connected and pinged MongoDB at", uri)
	return client, nil
}

// StoreGameReports takes a map of game reports and stores them in the specified MongoDB collection.
// Each key-value pair in the reports map is intended to be a separate document.
// The key (gameID as int) will be used as the _id field in MongoDB for idempotency.
func StoreGameReports(ctx context.Context, collection *mongo.Collection, reports map[int]interface{}) error {
	if collection == nil {
		return fmt.Errorf("MongoDB collection is nil")
	}
	if len(reports) == 0 {
		fmt.Println("No reports to store in MongoDB.")
		return nil
	}

	var operations []mongo.WriteModel

	for gameID, reportData := range reports {
		// We'll use gameID (int) as the _id.
		// Using UpdateOne with Upsert ensures that if we run this multiple times,
		// existing game reports are updated, and new ones are inserted.
		filter := bson.M{"_id": gameID}
		update := bson.M{"$set": reportData} // reportData should be the GameReport struct
		
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		operations = append(operations, model)
	}

	if len(operations) > 0 {
		bulkWriteOptions := options.BulkWrite().SetOrdered(false) // Unordered for potentially better performance
		result, err := collection.BulkWrite(ctx, operations, bulkWriteOptions)
		if err != nil {
			return fmt.Errorf("failed to bulk write game reports to MongoDB: %w", err)
		}
		fmt.Printf("MongoDB BulkWrite: Inserted %d, Updated %d, Upserted %d documents.\n", result.InsertedCount, result.ModifiedCount, result.UpsertedCount)
	} else {
		fmt.Println("No operations to perform for MongoDB storage.")
	}

	return nil
}

// PrintAllStoredGames retrieves all documents from the given collection, sorted by _id, and prints them as JSON.
func PrintAllStoredGames(ctx context.Context, collection *mongo.Collection) error {
	if collection == nil {
		return fmt.Errorf("MongoDB collection is nil")
	}

	fmt.Println("\n--- All Stored Game Reports from MongoDB (Sorted by Game ID) ---")

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", 1}}) // Sort by _id in ascending order

	cursor, err := collection.Find(ctx, bson.D{{}}, findOptions) 
	if err != nil {
		return fmt.Errorf("failed to find documents in MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var foundAny bool
	for cursor.Next(ctx) {
		foundAny = true
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error decoding document from MongoDB: %v", err) 
			continue
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Error marshalling document to JSON: %v", err)
			continue
		}
		fmt.Println(string(jsonData))
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("MongoDB cursor error: %w", err)
	}

	if !foundAny {
		fmt.Println("No game reports found in the collection.")
	}

	return nil
}

// DeleteAllGameReportsFromDB removes all documents from the specified MongoDB collection.
func DeleteAllGameReportsFromDB(ctx context.Context, collection *mongo.Collection) error {
	if collection == nil {
		return fmt.Errorf("MongoDB collection is nil")
	}

	fmt.Printf("\nAttempting to delete all documents from collection '%s'...\n", collection.Name())

	// An empty filter bson.D{{}} matches all documents in the collection.
	result, err := collection.DeleteMany(ctx, bson.D{{}})
	if err != nil {
		return fmt.Errorf("failed to delete documents from collection '%s': %w", collection.Name(), err)
	}

	fmt.Printf("Successfully deleted %d document(s) from collection '%s'.\n", result.DeletedCount, collection.Name())
	return nil
}

// Note: The StoreGameReports function expects 'reports' to be map[string]interface{}
// for flexibility with BSON marshalling, but the values should ideally be structured
// (like reporter.GameReport). The BSON tags on the GameReport struct will guide marshalling.
// We also need to import "go.mongodb.org/mongo-driver/bson" in this file for bson.M
// and ensure that the reporter.GameReport struct is accessible if we were to type `reports` more strictly. 