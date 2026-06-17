package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectTestDB(t *testing.T) *mongo.Database {
	t.Helper()

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auction?authSource=admin"
	}
	mongoDB := os.Getenv("MONGODB_DB")
	if mongoDB == "" {
		mongoDB = "auction_test"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB not reachable: %v", err)
	}

	t.Cleanup(func() { client.Disconnect(context.Background()) })

	return client.Database(mongoDB)
}

func TestCreateAuction_ShouldCloseAutomatically(t *testing.T) {
	db := connectTestDB(t)

	const testDuration = 3 * time.Second

	repo := &AuctionRepository{
		Collection:      db.Collection("auctions"),
		auctionDuration: testDuration,
	}

	auction, internalErr := auction_entity.CreateAuction(
		"Test Notebook",
		"Electronics",
		"High-performance laptop for testing auto-close",
		auction_entity.New,
	)
	if internalErr != nil {
		t.Fatalf("Failed to create auction entity: %v", internalErr)
	}

	if err := repo.CreateAuction(context.Background(), auction); err != nil {
		t.Fatalf("Failed to persist auction: %v", err)
	}

	if auction.Status != auction_entity.Active {
		t.Errorf("Expected initial status Active, got %d", auction.Status)
	}

	t.Logf("Auction %s created with status Active. Waiting %s for auto-close...", auction.Id, testDuration+time.Second)
	time.Sleep(testDuration + time.Second)

	var result AuctionEntityMongo
	filter := bson.M{"_id": auction.Id}
	if err := repo.Collection.FindOne(context.Background(), filter).Decode(&result); err != nil {
		t.Fatalf("Failed to retrieve auction from MongoDB: %v", err)
	}

	if result.Status != auction_entity.Completed {
		t.Errorf("Expected status Completed (%d), got %d", auction_entity.Completed, result.Status)
	}

	t.Logf("Auction %s correctly auto-closed with status Completed", auction.Id)

	t.Cleanup(func() {
		repo.Collection.DeleteOne(context.Background(), filter)
	})
}
