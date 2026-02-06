package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/models"
)

// UserService handles user-related database operations
type UserService struct{}

// NewUserService creates a new UserService
func NewUserService() *UserService {
	return &UserService{}
}

// GetUserByID gets a user by their ID
func (s *UserService) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	err := database.UserCollection.FindOne(context.Background(), bson.M{"id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetOrCreateUser gets a user or creates if not exists
func (s *UserService) GetOrCreateUser(userID int64, username, firstName string) (*models.User, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		// Create new user
		newUser := &models.User{
			ID:         userID,
			Username:   username,
			FirstName:  firstName,
			Characters: []models.UserCharacter{},
			Favorites:  []string{},
			Balance:    0,
		}
		_, err = database.UserCollection.InsertOne(context.Background(), newUser)
		if err != nil {
			return nil, err
		}
		return newUser, nil
	}
	
	// Update username/firstname if changed
	update := bson.M{}
	if user.Username != username && username != "" {
		update["username"] = username
	}
	if user.FirstName != firstName && firstName != "" {
		update["first_name"] = firstName
	}
	
	if len(update) > 0 {
		_, _ = database.UserCollection.UpdateOne(
			context.Background(),
			bson.M{"id": userID},
			bson.M{"$set": update},
		)
	}
	
	return user, nil
}

// UpdateUserBalance updates a user's balance
func (s *UserService) UpdateUserBalance(userID int64, amount int64) (int64, error) {
	result := database.UserCollection.FindOneAndUpdate(
		context.Background(),
		bson.M{"id": userID},
		bson.M{"$inc": bson.M{"balance": amount}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	
	var user models.User
	if err := result.Decode(&user); err != nil {
		return 0, err
	}
	return user.Balance, nil
}

// GetUserBalance gets a user's balance
func (s *UserService) GetUserBalance(userID int64) (int64, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return 0, nil // Return 0 if user doesn't exist
	}
	return user.Balance, nil
}

// AddCharacterToUser adds a character to user's collection
func (s *UserService) AddCharacterToUser(userID int64, char models.UserCharacter) error {
	_, err := database.UserCollection.UpdateOne(
		context.Background(),
		bson.M{"id": userID},
		bson.M{
			"$push": bson.M{"characters": char},
			"$setOnInsert": bson.M{"id": userID, "balance": 0},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// RemoveCharacterFromUser removes a character from user's collection
func (s *UserService) RemoveCharacterFromUser(userID int64, charID string) error {
	_, err := database.UserCollection.UpdateOne(
		context.Background(),
		bson.M{"id": userID},
		bson.M{"$pull": bson.M{"characters": bson.M{"id": charID}}},
	)
	return err
}

// HasCharacter checks if user has a character
func (s *UserService) HasCharacter(userID int64, charID string) (bool, error) {
	count, err := database.UserCollection.CountDocuments(
		context.Background(),
		bson.M{"id": userID, "characters.id": charID},
	)
	return count > 0, err
}

// AddToFavorites adds a character to user's favorites
func (s *UserService) AddToFavorites(userID int64, charID string) error {
	_, err := database.UserCollection.UpdateOne(
		context.Background(),
		bson.M{"id": userID},
		bson.M{"$addToSet": bson.M{"favorites": charID}},
	)
	return err
}

// GetUserCharactersCount gets the count of user's characters
func (s *UserService) GetUserCharactersCount(userID int64) (int, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	return len(user.Characters), nil
}

// GetTopUsersByBalance gets top users by balance
func (s *UserService) GetTopUsersByBalance(limit int) ([]models.User, error) {
	cursor, err := database.UserCollection.Find(
		context.Background(),
		bson.M{},
		options.Find().SetSort(bson.M{"balance": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var users []models.User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetTopUsersByCharacters gets top users by character count
func (s *UserService) GetTopUsersByCharacters(limit int) ([]models.User, error) {
	pipeline := []bson.M{
		{"$project": bson.M{
			"id":         1,
			"username":   1,
			"first_name": 1,
			"characters": 1,
			"charCount":  bson.M{"$size": bson.M{"$ifNull": []interface{}{"$characters", []interface{}{}}}},
		}},
		{"$sort": bson.M{"charCount": -1}},
		{"$limit": limit},
	}
	
	cursor, err := database.UserCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var users []models.User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateLastSClaim updates user's last sclaim time
func (s *UserService) UpdateLastSClaim(userID int64) error {
	now := time.Now()
	_, err := database.UserCollection.UpdateOne(
		context.Background(),
		bson.M{"id": userID},
		bson.M{"$set": bson.M{"last_sclaim": now}},
		options.Update().SetUpsert(true),
	)
	return err
}

// UpdateLastClaim updates user's last claim time
func (s *UserService) UpdateLastClaim(userID int64) error {
	now := time.Now()
	_, err := database.UserCollection.UpdateOne(
		context.Background(),
		bson.M{"id": userID},
		bson.M{"$set": bson.M{"last_claim": now}},
		options.Update().SetUpsert(true),
	)
	return err
}

// CanSClaim checks if user can use sclaim (24h cooldown)
func (s *UserService) CanSClaim(userID int64) (bool, time.Duration, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return true, 0, nil // New user can claim
	}
	
	if user.LastSClaim == nil {
		return true, 0, nil
	}
	
	timeSince := time.Since(*user.LastSClaim)
	if timeSince >= 24*time.Hour {
		return true, 0, nil
	}
	
	remaining := 24*time.Hour - timeSince
	return false, remaining, nil
}

// CanClaim checks if user can use claim (24h cooldown)
func (s *UserService) CanClaim(userID int64) (bool, time.Duration, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return true, 0, nil // New user can claim
	}
	
	if user.LastClaim == nil {
		return true, 0, nil
	}
	
	timeSince := time.Since(*user.LastClaim)
	if timeSince >= 24*time.Hour {
		return true, 0, nil
	}
	
	remaining := 24*time.Hour - timeSince
	return false, remaining, nil
}

// GetShopData gets user's shop data
func (s *UserService) GetShopData(userID int64) (*models.ShopData, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return user.ShopData, nil
}

// UpdateShopData updates user's shop data
func (s *UserService) UpdateShopData(userID int64, shopData *models.ShopData) error {
	_, err := database.UserCollection.UpdateOne(
		context.Background(),
		bson.M{"id": userID},
		bson.M{"$set": bson.M{"shop_data": shopData}},
	)
	return err
}
