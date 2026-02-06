package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/config"
)

var (
	// Client is the MongoDB client
	Client *mongo.Client
	// DB is the database instance
	DB *mongo.Database

	// Collections
	CharacterCollection        *mongo.Collection
	UserCollection             *mongo.Collection
	UserTotalsCollection       *mongo.Collection
	GroupUserTotalsCollection  *mongo.Collection
	TopGlobalGroupsCollection  *mongo.Collection
	PMUsersCollection          *mongo.Collection
	DailyUserGuessesCollection *mongo.Collection
	DailyGroupGuessesCollection *mongo.Collection
	RedeemCodesCollection      *mongo.Collection
	ClaimCodesCollection       *mongo.Collection
	RaritySettingsCollection   *mongo.Collection
	LockedCharactersCollection *mongo.Collection
	SortPreferencesCollection  *mongo.Collection
)

// Connect establishes connection to MongoDB
func Connect(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("✅ Connected to MongoDB!")

	Client = client
	DB = client.Database("Character_catcher")

	// Initialize collections
	CharacterCollection = DB.Collection("anime_characters_lol")
	UserCollection = DB.Collection("user_collection_lmaoooo")
	UserTotalsCollection = DB.Collection("user_totals_lmaoooo")
	GroupUserTotalsCollection = DB.Collection("group_user_totalsssssss")
	TopGlobalGroupsCollection = DB.Collection("top_global_groups")
	PMUsersCollection = DB.Collection("total_pm_users")
	DailyUserGuessesCollection = DB.Collection("daily_user_guesses")
	DailyGroupGuessesCollection = DB.Collection("daily_group_guesses")
	RedeemCodesCollection = DB.Collection("redeem_codes")
	ClaimCodesCollection = DB.Collection("claim_codes")
	RaritySettingsCollection = DB.Collection("rarity_settings")
	LockedCharactersCollection = DB.Collection("locked_characters")
	SortPreferencesCollection = DB.Collection("sort_preferences")

	// Create indexes
	createIndexes()

	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return Client.Disconnect(ctx)
	}
	return nil
}

func createIndexes() {
	ctx := context.Background()

	// User collection indexes
	userIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"balance": -1},
		},
	}
	_, err := UserCollection.Indexes().CreateMany(ctx, userIndexes)
	if err != nil {
		log.Printf("Error creating user indexes: %v", err)
	}

	// Character collection indexes
	charIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"rarity": 1},
		},
	}
	_, err = CharacterCollection.Indexes().CreateMany(ctx, charIndexes)
	if err != nil {
		log.Printf("Error creating character indexes: %v", err)
	}

	// Redeem codes index
	_, err = RedeemCodesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"code": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating redeem code index: %v", err)
	}

	// Claim codes index
	_, err = ClaimCodesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"code": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating claim code index: %v", err)
	}

	// Locked characters index
	_, err = LockedCharactersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"character_id": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating locked characters index: %v", err)
	}

	log.Println("✅ Database indexes created")
}
