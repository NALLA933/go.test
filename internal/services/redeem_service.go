package services

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// RedeemService handles redeem code operations
type RedeemService struct{}

// NewRedeemService creates a new RedeemService
func NewRedeemService() *RedeemService {
	return &RedeemService{}
}

// CreateCoinCode creates a coin redeem code
func (s *RedeemService) CreateCoinCode(amount int64, maxUses int, createdBy int64) (string, error) {
	for i := 0; i < 10; i++ {
		code := utils.GenerateUniqueCode()
		
		doc := models.RedeemCode{
			Code:      strings.ToLower(code),
			Type:      "coin",
			Amount:    amount,
			MaxUses:   maxUses,
			UsedBy:    []int64{},
			IsActive:  true,
			CreatedBy: createdBy,
		}
		
		_, err := database.RedeemCodesCollection.InsertOne(context.Background(), doc)
		if err == nil {
			return code, nil
		}
		// If duplicate key error, try again
	}
	return "", nil
}

// CreateCharacterCode creates a character redeem code
func (s *RedeemService) CreateCharacterCode(charID string, maxUses int, createdBy int64) (string, error) {
	for i := 0; i < 10; i++ {
		code := utils.GenerateUniqueCode()
		
		doc := models.RedeemCode{
			Code:        strings.ToLower(code),
			Type:        "character",
			CharacterID: charID,
			MaxUses:     maxUses,
			UsedBy:      []int64{},
			IsActive:    true,
			CreatedBy:   createdBy,
		}
		
		_, err := database.RedeemCodesCollection.InsertOne(context.Background(), doc)
		if err == nil {
			return code, nil
		}
	}
	return "", nil
}

// RedeemCode redeems a code for a user
func (s *RedeemService) RedeemCode(code string, userID int64) (*models.RedeemCode, error) {
	code = strings.ToLower(code)
	
	// Find and update atomically
	result := database.RedeemCodesCollection.FindOneAndUpdate(
		context.Background(),
		bson.M{
			"code":      code,
			"is_active": true,
			"used_by":   bson.M{"$ne": userID},
			"$expr":     bson.M{"$lt": []interface{}{bson.M{"$size": "$used_by"}, "$max_uses"}},
		},
		bson.M{"$push": bson.M{"used_by": userID}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	
	var redeemCode models.RedeemCode
	if err := result.Decode(&redeemCode); err != nil {
		return nil, err
	}
	
	// Check if max uses reached and deactivate
	if len(redeemCode.UsedBy) >= redeemCode.MaxUses {
		_, _ = database.RedeemCodesCollection.UpdateOne(
			context.Background(),
			bson.M{"code": code},
			bson.M{"$set": bson.M{"is_active": false}},
		)
	}
	
	return &redeemCode, nil
}

// GetRedeemCode gets a redeem code by code string
func (s *RedeemService) GetRedeemCode(code string) (*models.RedeemCode, error) {
	code = strings.ToLower(code)
	
	var redeemCode models.RedeemCode
	err := database.RedeemCodesCollection.FindOne(
		context.Background(),
		bson.M{"code": code},
	).Decode(&redeemCode)
	
	if err != nil {
		return nil, err
	}
	return &redeemCode, nil
}

// HasUserRedeemed checks if a user has already redeemed a code
func (s *RedeemService) HasUserRedeemed(code string, userID int64) (bool, error) {
	code = strings.ToLower(code)
	
	count, err := database.RedeemCodesCollection.CountDocuments(
		context.Background(),
		bson.M{"code": code, "used_by": userID},
	)
	return count > 0, err
}

// ClaimCodeService handles claim code operations
type ClaimCodeService struct{}

// NewClaimCodeService creates a new ClaimCodeService
func NewClaimCodeService() *ClaimCodeService {
	return &ClaimCodeService{}
}

// CreateClaimCode creates a claim code for daily coins
func (s *ClaimCodeService) CreateClaimCode(userID int64, amount int64) (string, error) {
	for i := 0; i < 10; i++ {
		code := utils.GenerateCoinCode()
		
		doc := models.ClaimCode{
			Code:       code,
			UserID:     userID,
			Amount:     amount,
			CreatedAt:  time.Now(),
			IsRedeemed: false,
		}
		
		_, err := database.ClaimCodesCollection.InsertOne(context.Background(), doc)
		if err == nil {
			return code, nil
		}
	}
	return "", nil
}

// GetClaimCode gets a claim code
func (s *ClaimCodeService) GetClaimCode(code string) (*models.ClaimCode, error) {
	var claimCode models.ClaimCode
	err := database.ClaimCodesCollection.FindOne(
		context.Background(),
		bson.M{"code": code},
	).Decode(&claimCode)
	
	if err != nil {
		return nil, err
	}
	return &claimCode, nil
}

// RedeemClaimCode redeems a claim code
func (s *ClaimCodeService) RedeemClaimCode(code string, userID int64) (*models.ClaimCode, error) {
	now := time.Now()
	
	result := database.ClaimCodesCollection.FindOneAndUpdate(
		context.Background(),
		bson.M{
			"code":        code,
			"user_id":     userID,
			"is_redeemed": false,
		},
		bson.M{
			"$set": bson.M{
				"is_redeemed": true,
				"redeemed_at": now,
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	
	var claimCode models.ClaimCode
	if err := result.Decode(&claimCode); err != nil {
		return nil, err
	}
	return &claimCode, nil
}

// IsClaimCodeExpired checks if a claim code is expired (24 hours)
func (s *ClaimCodeService) IsClaimCodeExpired(code *models.ClaimCode) bool {
	return time.Since(code.CreatedAt) > 24*time.Hour
}
