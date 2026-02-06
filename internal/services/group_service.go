package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/models"
)

// GroupService handles group-related database operations
type GroupService struct{}

// NewGroupService creates a new GroupService
func NewGroupService() *GroupService {
	return &GroupService{}
}

// UpdateGroupUserTotal updates or creates group user total
func (s *GroupService) UpdateGroupUserTotal(userID, groupID int64, username, firstName string) error {
	result := database.GroupUserTotalsCollection.FindOneAndUpdate(
		context.Background(),
		bson.M{"user_id": userID, "group_id": groupID},
		bson.M{
			"$set": bson.M{
				"username":   username,
				"first_name": firstName,
			},
			"$inc": bson.M{"count": 1},
		},
		options.FindOneAndUpdate().SetUpsert(true),
	)
	
	// If document didn't exist, it was created
	if result.Err() != nil {
		// Insert new document with count 1
		_, err := database.GroupUserTotalsCollection.InsertOne(context.Background(), models.GroupUserTotal{
			UserID:    userID,
			GroupID:   groupID,
			Username:  username,
			FirstName: firstName,
			Count:     1,
		})
		return err
	}
	
	return nil
}

// UpdateTopGlobalGroup updates or creates top global group entry
func (s *GroupService) UpdateTopGlobalGroup(groupID int64, groupName string) error {
	_, err := database.TopGlobalGroupsCollection.UpdateOne(
		context.Background(),
		bson.M{"group_id": groupID},
		bson.M{
			"$set": bson.M{"group_name": groupName},
			"$inc": bson.M{"count": 1},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetTopGroups gets top groups by guess count
func (s *GroupService) GetTopGroups(limit int) ([]models.TopGlobalGroup, error) {
	cursor, err := database.TopGlobalGroupsCollection.Find(
		context.Background(),
		bson.M{},
		options.Find().SetSort(bson.M{"count": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var groups []models.TopGlobalGroup
	if err = cursor.All(context.Background(), &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GetGroupUserTotals gets top users in a group
func (s *GroupService) GetGroupUserTotals(groupID int64, limit int) ([]models.GroupUserTotal, error) {
	cursor, err := database.GroupUserTotalsCollection.Find(
		context.Background(),
		bson.M{"group_id": groupID},
		options.Find().SetSort(bson.M{"count": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var totals []models.GroupUserTotal
	if err = cursor.All(context.Background(), &totals); err != nil {
		return nil, err
	}
	return totals, nil
}

// GetMessageFrequency gets message frequency for a chat
func (s *GroupService) GetMessageFrequency(chatID int64) (int, error) {
	var total models.UserTotal
	err := database.UserTotalsCollection.FindOne(
		context.Background(),
		bson.M{"chat_id": chatID},
	).Decode(&total)
	
	if err != nil {
		// Return default frequency
		return 100, nil
	}
	
	if total.MessageFrequency == 0 {
		return 100, nil
	}
	return total.MessageFrequency, nil
}

// SetMessageFrequency sets message frequency for a chat
func (s *GroupService) SetMessageFrequency(chatID int64, frequency int) error {
	_, err := database.UserTotalsCollection.UpdateOne(
		context.Background(),
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"message_frequency": frequency}},
		options.Update().SetUpsert(true),
	)
	return err
}

// AddPMUser adds a user to PM users collection
func (s *GroupService) AddPMUser(userID int64, username, firstName string) error {
	_, err := database.PMUsersCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"username":   username,
				"first_name": firstName,
			},
			"$setOnInsert": bson.M{"started_at": time.Now()},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetPMUsersCount gets total PM users count
func (s *GroupService) GetPMUsersCount() (int64, error) {
	return database.PMUsersCollection.CountDocuments(context.Background(), bson.M{})
}
