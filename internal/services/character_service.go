package services

import (
	"context"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/models"
)

// CharacterService handles character-related database operations
type CharacterService struct{}

// NewCharacterService creates a new CharacterService
func NewCharacterService() *CharacterService {
	return &CharacterService{}
}

// GetCharacterByID gets a character by ID
func (s *CharacterService) GetCharacterByID(charID string) (*models.Character, error) {
	var char models.Character
	err := database.CharacterCollection.FindOne(context.Background(), bson.M{"id": charID}).Decode(&char)
	if err != nil {
		return nil, err
	}
	return &char, nil
}

// GetRandomCharacter gets a random character, optionally filtering by excluded rarities and locked IDs
func (s *CharacterService) GetRandomCharacter(excludedRarities []int, lockedIDs []string) (*models.Character, error) {
	filter := bson.M{}
	
	// Exclude disabled rarities
	if len(excludedRarities) > 0 {
		filter["rarity"] = bson.M{"$nin": excludedRarities}
	}
	
	// Exclude locked characters
	if len(lockedIDs) > 0 {
		if _, ok := filter["rarity"]; ok {
			filter["$and"] = []bson.M{
				{"rarity": bson.M{"$nin": excludedRarities}},
				{"id": bson.M{"$nin": lockedIDs}},
			}
			delete(filter, "rarity")
		} else {
			filter["id"] = bson.M{"$nin": lockedIDs}
		}
	}
	
	// Use aggregation with $sample for random selection
	pipeline := []bson.M{
		{"$match": filter},
		{"$sample": bson.M{"size": 1}},
	}
	
	cursor, err := database.CharacterCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	
	if len(chars) == 0 {
		return nil, nil
	}
	
	return &chars[0], nil
}

// GetRandomCharacters gets multiple random characters
func (s *CharacterService) GetRandomCharacters(count int, excludedRarities []int, lockedIDs []string) ([]models.Character, error) {
	filter := bson.M{}
	
	if len(excludedRarities) > 0 {
		filter["rarity"] = bson.M{"$nin": excludedRarities}
	}
	
	if len(lockedIDs) > 0 {
		if _, ok := filter["rarity"]; ok {
			filter["$and"] = []bson.M{
				{"rarity": bson.M{"$nin": excludedRarities}},
				{"id": bson.M{"$nin": lockedIDs}},
			}
			delete(filter, "rarity")
		} else {
			filter["id"] = bson.M{"$nin": lockedIDs}
		}
	}
	
	pipeline := []bson.M{
		{"$match": filter},
		{"$sample": bson.M{"size": count}},
	}
	
	cursor, err := database.CharacterCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	
	return chars, nil
}

// GetCharactersByRarity gets characters by rarity
func (s *CharacterService) GetCharactersByRarity(rarity int) ([]models.Character, error) {
	cursor, err := database.CharacterCollection.Find(context.Background(), bson.M{"rarity": rarity})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

// SearchCharacters searches characters by name
func (s *CharacterService) SearchCharacters(query string) ([]models.Character, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"anime": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	
	cursor, err := database.CharacterCollection.Find(
		context.Background(),
		filter,
		options.Find().SetProjection(bson.M{
			"id":     1,
			"name":   1,
			"anime":  1,
			"rarity": 1,
			"img_url": 1,
		}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

// GetCharacterCount gets the total count of characters
func (s *CharacterService) GetCharacterCount() (int64, error) {
	return database.CharacterCollection.CountDocuments(context.Background(), bson.M{})
}

// GetCharacterOwnerCount gets how many users own a specific character
func (s *CharacterService) GetCharacterOwnerCount(charID string) (int64, error) {
	return database.UserCollection.CountDocuments(
		context.Background(),
		bson.M{"characters.id": charID},
	)
}

// GetTopGrabbers gets top users who own a specific character
func (s *CharacterService) GetTopGrabbers(charID string, limit int) ([]map[string]interface{}, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"characters.id": charID}},
		{"$project": bson.M{
			"id":         1,
			"username":   1,
			"first_name": 1,
			"count": bson.M{
				"$size": bson.M{
					"$filter": bson.M{
						"input": "$characters",
						"as":    "char",
						"cond":  bson.M{"$eq": []interface{}{"$$char.id", charID}},
					},
				},
			},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": limit},
	}
	
	cursor, err := database.UserCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var results []map[string]interface{}
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetAllCharacters gets all characters (use with caution on large datasets)
func (s *CharacterService) GetAllCharacters() ([]models.Character, error) {
	cursor, err := database.CharacterCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

// GetCharactersByRarities gets characters by multiple rarities
func (s *CharacterService) GetCharactersByRarities(rarities []int) ([]models.Character, error) {
	cursor, err := database.CharacterCollection.Find(
		context.Background(),
		bson.M{"rarity": bson.M{"$in": rarities}},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

// GetRandomCharactersByRarities gets random characters from specified rarities
func (s *CharacterService) GetRandomCharactersByRarities(rarities []int, count int) ([]models.Character, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"rarity": bson.M{"$in": rarities}}},
		{"$sample": bson.M{"size": count}},
	}
	
	cursor, err := database.CharacterCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var chars []models.Character
	if err = cursor.All(context.Background(), &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

// ToUserCharacter converts a Character to UserCharacter
func (s *CharacterService) ToUserCharacter(char *models.Character) models.UserCharacter {
	return models.UserCharacter{
		ID:     char.ID,
		Name:   char.Name,
		Anime:  char.Anime,
		Rarity: char.Rarity,
		ImgURL: char.ImgURL,
	}
}

// GetAnimeCounts gets count of characters per anime
func (s *CharacterService) GetAnimeCounts(animes []string) (map[string]int64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"anime": bson.M{"$in": animes}}},
		{"$group": bson.M{
			"_id":   "$anime",
			"count": bson.M{"$sum": 1},
		}},
	}
	
	cursor, err := database.CharacterCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}
	
	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.ID] = r.Count
	}
	return counts, nil
}

// ShopRarities defines which rarities are available in the shop
var ShopRarities = []int{4, 5, 6, 14} // Special, Ancient, Celestial, Kawaii

// ShopPriceRanges defines price ranges for each rarity
var ShopPriceRanges = map[int][2]int64{
	4:  {400000, 500000}, // Special
	5:  {600000, 700000}, // Ancient
	6:  {650000, 750000}, // Celestial
	14: {450000, 550000}, // Kawaii
}

// GenerateShopCharacter generates a shop character with pricing
func (s *CharacterService) GenerateShopCharacter(char models.Character) models.ShopCharacter {
	priceRange := ShopPriceRanges[char.Rarity]
	basePrice := rand.Int63n(priceRange[1]-priceRange[0]) + priceRange[0]
	discountPercent := rand.Intn(11) + 5 // 5-15%
	discountAmount := basePrice * int64(discountPercent) / 100
	finalPrice := basePrice - discountAmount
	
	return models.ShopCharacter{
		ID:              char.ID,
		Name:            char.Name,
		Anime:           char.Anime,
		Rarity:          char.Rarity,
		ImgURL:          char.ImgURL,
		BasePrice:       basePrice,
		DiscountPercent: discountPercent,
		FinalPrice:      finalPrice,
	}
}

// InitializeShop generates initial shop data for a user
func (s *CharacterService) InitializeShop() (*models.ShopData, error) {
	chars, err := s.GetRandomCharactersByRarities(ShopRarities, 3)
	if err != nil {
		return nil, err
	}
	
	shopChars := make([]models.ShopCharacter, len(chars))
	for i, char := range chars {
		shopChars[i] = s.GenerateShopCharacter(char)
	}
	
	return &models.ShopData{
		Characters:   shopChars,
		LastReset:    time.Now(),
		RefreshUsed:  false,
		CurrentIndex: 0,
	}, nil
}

// RefreshShop refreshes the shop with new characters
func (s *CharacterService) RefreshShop() (*models.ShopData, error) {
	return s.InitializeShop()
}
