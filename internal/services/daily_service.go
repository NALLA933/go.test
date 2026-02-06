package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// DailyService handles daily stats operations
type DailyService struct{}

// NewDailyService creates a new DailyService
func NewDailyService() *DailyService {
	return &DailyService{}
}

// UpdateDailyUserGuess increments daily guess count for a user
func (s *DailyService) UpdateDailyUserGuess(userID int64, username, firstName string) error {
	today := utils.GetISTDate()
	now := time.Now()
	
	_, err := database.DailyUserGuessesCollection.UpdateOne(
		context.Background(),
		bson.M{"date": today, "user_id": userID},
		bson.M{
			"$inc": bson.M{"count": 1},
			"$set": bson.M{
				"username":     username,
				"first_name":   firstName,
				"last_updated": now,
			},
			"$setOnInsert": bson.M{
				"date":   today,
				"user_id": userID,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// UpdateDailyGroupGuess increments daily guess count for a group
func (s *DailyService) UpdateDailyGroupGuess(groupID int64, groupName string) error {
	today := utils.GetISTDate()
	now := time.Now()
	
	_, err := database.DailyGroupGuessesCollection.UpdateOne(
		context.Background(),
		bson.M{"date": today, "group_id": groupID},
		bson.M{
			"$inc": bson.M{"count": 1},
			"$set": bson.M{
				"group_name":   groupName,
				"last_updated": now,
			},
			"$setOnInsert": bson.M{
				"date":    today,
				"group_id": groupID,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetTopDailyUsers gets top users by daily guesses
func (s *DailyService) GetTopDailyUsers(limit int) ([]models.DailyUserGuess, error) {
	today := utils.GetISTDate()
	
	cursor, err := database.DailyUserGuessesCollection.Find(
		context.Background(),
		bson.M{"date": today},
		options.Find().SetSort(bson.M{"count": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var guesses []models.DailyUserGuess
	if err = cursor.All(context.Background(), &guesses); err != nil {
		return nil, err
	}
	return guesses, nil
}

// GetTopDailyGroups gets top groups by daily guesses
func (s *DailyService) GetTopDailyGroups(limit int) ([]models.DailyGroupGuess, error) {
	today := utils.GetISTDate()
	
	cursor, err := database.DailyGroupGuessesCollection.Find(
		context.Background(),
		bson.M{"date": today},
		options.Find().SetSort(bson.M{"count": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var guesses []models.DailyGroupGuess
	if err = cursor.All(context.Background(), &guesses); err != nil {
		return nil, err
	}
	return guesses, nil
}
