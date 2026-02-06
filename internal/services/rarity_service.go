package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/models"
)

// RarityService handles rarity settings and locked characters
type RarityService struct{}

// NewRarityService creates a new RarityService
func NewRarityService() *RarityService {
	return &RarityService{}
}

// GetChatRaritySettings gets rarity settings for a chat
func (s *RarityService) GetChatRaritySettings(chatID int64) (*models.RaritySettings, error) {
	var settings models.RaritySettings
	err := database.RaritySettingsCollection.FindOne(
		context.Background(),
		bson.M{"chat_id": chatID},
	).Decode(&settings)
	
	if err != nil {
		// Create default settings
		settings = models.RaritySettings{
			ChatID:           chatID,
			DisabledRarities: []int{},
		}
		_, _ = database.RaritySettingsCollection.InsertOne(context.Background(), settings)
	}
	
	return &settings, nil
}

// EnableRarity enables a rarity for a chat
func (s *RarityService) EnableRarity(chatID int64, rarity int) error {
	_, err := database.RaritySettingsCollection.UpdateOne(
		context.Background(),
		bson.M{"chat_id": chatID},
		bson.M{"$pull": bson.M{"disabled_rarities": rarity}},
		options.Update().SetUpsert(true),
	)
	return err
}

// DisableRarity disables a rarity for a chat
func (s *RarityService) DisableRarity(chatID int64, rarity int) error {
	_, err := database.RaritySettingsCollection.UpdateOne(
		context.Background(),
		bson.M{"chat_id": chatID},
		bson.M{"$addToSet": bson.M{"disabled_rarities": rarity}},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetDisabledRarities gets list of disabled rarities for a chat
func (s *RarityService) GetDisabledRarities(chatID int64) ([]int, error) {
	settings, err := s.GetChatRaritySettings(chatID)
	if err != nil {
		return []int{}, err
	}
	return settings.DisabledRarities, nil
}

// IsRarityEnabled checks if a rarity is enabled in a chat
func (s *RarityService) IsRarityEnabled(chatID int64, rarity int) (bool, error) {
	settings, err := s.GetChatRaritySettings(chatID)
	if err != nil {
		return true, err
	}
	
	for _, r := range settings.DisabledRarities {
		if r == rarity {
			return false, nil
		}
	}
	return true, nil
}

// LockCharacter locks a character from spawning
func (s *RarityService) LockCharacter(charID, charName string, lockedByID int64, lockedByName, reason string) error {
	lockData := models.LockedCharacter{
		CharacterID:   charID,
		CharacterName: charName,
		LockedByID:    lockedByID,
		LockedByName:  lockedByName,
		Reason:        reason,
		LockedAt:      time.Now(),
	}
	
	_, err := database.LockedCharactersCollection.UpdateOne(
		context.Background(),
		bson.M{"character_id": charID},
		bson.M{"$set": lockData},
		options.Update().SetUpsert(true),
	)
	return err
}

// UnlockCharacter unlocks a character
func (s *RarityService) UnlockCharacter(charID string) error {
	_, err := database.LockedCharactersCollection.DeleteOne(
		context.Background(),
		bson.M{"character_id": charID},
	)
	return err
}

// IsCharacterLocked checks if a character is locked
func (s *RarityService) IsCharacterLocked(charID string) (bool, error) {
	count, err := database.LockedCharactersCollection.CountDocuments(
		context.Background(),
		bson.M{"character_id": charID},
	)
	return count > 0, err
}

// GetLockedCharacters gets all locked characters
func (s *RarityService) GetLockedCharacters() ([]models.LockedCharacter, error) {
	cursor, err := database.LockedCharactersCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var locked []models.LockedCharacter
	if err = cursor.All(context.Background(), &locked); err != nil {
		return nil, err
	}
	return locked, nil
}

// GetLockedCharacterIDs gets all locked character IDs
func (s *RarityService) GetLockedCharacterIDs() ([]string, error) {
	locked, err := s.GetLockedCharacters()
	if err != nil {
		return nil, err
	}
	
	ids := make([]string, len(locked))
	for i, l := range locked {
		ids[i] = l.CharacterID
	}
	return ids, nil
}

// SortPreferenceService handles user sort preferences
type SortPreferenceService struct{}

// NewSortPreferenceService creates a new SortPreferenceService
func NewSortPreferenceService() *SortPreferenceService {
	return &SortPreferenceService{}
}

// GetUserSortPreference gets user's sort preference
func (s *SortPreferenceService) GetUserSortPreference(userID int64) (*int, error) {
	var pref models.SortPreference
	err := database.SortPreferencesCollection.FindOne(
		context.Background(),
		bson.M{"user_id": userID},
	).Decode(&pref)
	
	if err != nil {
		return nil, nil // No preference set
	}
	return pref.RarityFilter, nil
}

// SetUserSortPreference sets user's sort preference
func (s *SortPreferenceService) SetUserSortPreference(userID int64, rarityFilter *int) error {
	_, err := database.SortPreferencesCollection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"rarity_filter": rarityFilter}},
		options.Update().SetUpsert(true),
	)
	return err
}
