package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Character represents an anime character in the database
type Character struct {
	ID      string `bson:"id" json:"id"`
	Name    string `bson:"name" json:"name"`
	Anime   string `bson:"anime" json:"anime"`
	Rarity  int    `bson:"rarity" json:"rarity"`
	ImgURL  string `bson:"img_url" json:"img_url"`
}

// UserCharacter represents a character in user's collection
type UserCharacter struct {
	ID     string `bson:"id" json:"id"`
	Name   string `bson:"name" json:"name"`
	Anime  string `bson:"anime" json:"anime"`
	Rarity int    `bson:"rarity" json:"rarity"`
	ImgURL string `bson:"img_url" json:"img_url"`
}

// User represents a user in the database
type User struct {
	ID          int64           `bson:"id" json:"id"`
	Username    string          `bson:"username,omitempty" json:"username,omitempty"`
	FirstName   string          `bson:"first_name" json:"first_name"`
	Characters  []UserCharacter `bson:"characters" json:"characters"`
	Favorites   []string        `bson:"favorites" json:"favorites"`
	Balance     int64           `bson:"balance" json:"balance"`
	LastSClaim  *time.Time      `bson:"last_sclaim,omitempty" json:"last_sclaim,omitempty"`
	LastClaim   *time.Time      `bson:"last_claim,omitempty" json:"last_claim,omitempty"`
	ShopData    *ShopData       `bson:"shop_data,omitempty" json:"shop_data,omitempty"`
}

// ShopData represents user's shop data
type ShopData struct {
	Characters   []ShopCharacter `bson:"characters" json:"characters"`
	LastReset    time.Time       `bson:"last_reset" json:"last_reset"`
	RefreshUsed  bool            `bson:"refresh_used" json:"refresh_used"`
	CurrentIndex int             `bson:"current_index" json:"current_index"`
}

// ShopCharacter represents a character in the shop
type ShopCharacter struct {
	ID              string `bson:"id" json:"id"`
	Name            string `bson:"name" json:"name"`
	Anime           string `bson:"anime" json:"anime"`
	Rarity          int    `bson:"rarity" json:"rarity"`
	ImgURL          string `bson:"img_url" json:"img_url"`
	BasePrice       int64  `bson:"base_price" json:"base_price"`
	DiscountPercent int    `bson:"discount_percent" json:"discount_percent"`
	FinalPrice      int64  `bson:"final_price" json:"final_price"`
}

// GroupUserTotal represents user stats in a group
type GroupUserTotal struct {
	UserID    int64  `bson:"user_id" json:"user_id"`
	GroupID   int64  `bson:"group_id" json:"group_id"`
	Username  string `bson:"username,omitempty" json:"username,omitempty"`
	FirstName string `bson:"first_name" json:"first_name"`
	Count     int64  `bson:"count" json:"count"`
}

// TopGlobalGroup represents a group's global stats
type TopGlobalGroup struct {
	GroupID   int64  `bson:"group_id" json:"group_id"`
	GroupName string `bson:"group_name" json:"group_name"`
	Count     int64  `bson:"count" json:"count"`
}

// UserTotal represents message count settings for a chat
type UserTotal struct {
	ChatID            int64 `bson:"chat_id" json:"chat_id"`
	MessageFrequency  int   `bson:"message_frequency" json:"message_frequency"`
}

// PMUser represents a user who started the bot
type PMUser struct {
	ID        int64      `bson:"_id" json:"id"`
	Username  string     `bson:"username,omitempty" json:"username,omitempty"`
	FirstName string     `bson:"first_name" json:"first_name"`
	StartedAt *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
}

// DailyUserGuess represents daily guess stats for a user
type DailyUserGuess struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Date       string             `bson:"date" json:"date"`
	UserID     int64              `bson:"user_id" json:"user_id"`
	Username   string             `bson:"username" json:"username"`
	FirstName  string             `bson:"first_name" json:"first_name"`
	Count      int64              `bson:"count" json:"count"`
	LastUpdated time.Time         `bson:"last_updated" json:"last_updated"`
}

// DailyGroupGuess represents daily guess stats for a group
type DailyGroupGuess struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Date       string             `bson:"date" json:"date"`
	GroupID    int64              `bson:"group_id" json:"group_id"`
	GroupName  string             `bson:"group_name" json:"group_name"`
	Count      int64              `bson:"count" json:"count"`
	LastUpdated time.Time         `bson:"last_updated" json:"last_updated"`
}

// RedeemCode represents a redeem code
type RedeemCode struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Code         string             `bson:"code" json:"code"`
	Type         string             `bson:"type" json:"type"` // "coin" or "character"
	Amount       int64              `bson:"amount,omitempty" json:"amount,omitempty"`
	CharacterID  string             `bson:"character_id,omitempty" json:"character_id,omitempty"`
	MaxUses      int                `bson:"max_uses" json:"max_uses"`
	UsedBy       []int64            `bson:"used_by" json:"used_by"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	CreatedBy    int64              `bson:"created_by" json:"created_by"`
}

// ClaimCode represents a claim code for daily coins
type ClaimCode struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Code        string             `bson:"code" json:"code"`
	UserID      int64              `bson:"user_id" json:"user_id"`
	Amount      int64              `bson:"amount" json:"amount"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	IsRedeemed  bool               `bson:"is_redeemed" json:"is_redeemed"`
	RedeemedAt  *time.Time         `bson:"redeemed_at,omitempty" json:"redeemed_at,omitempty"`
}

// RaritySettings represents rarity settings for a chat
type RaritySettings struct {
	ChatID           int64 `bson:"chat_id" json:"chat_id"`
	DisabledRarities []int `bson:"disabled_rarities" json:"disabled_rarities"`
}

// LockedCharacter represents a locked character
type LockedCharacter struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CharacterID    string             `bson:"character_id" json:"character_id"`
	CharacterName  string             `bson:"character_name" json:"character_name"`
	LockedByID     int64              `bson:"locked_by_id" json:"locked_by_id"`
	LockedByName   string             `bson:"locked_by_name" json:"locked_by_name"`
	Reason         string             `bson:"reason" json:"reason"`
	LockedAt       time.Time          `bson:"locked_at" json:"locked_at"`
}

// SortPreference represents user's sort preference for harem
type SortPreference struct {
	UserID       int64 `bson:"user_id" json:"user_id"`
	RarityFilter *int  `bson:"rarity_filter" json:"rarity_filter"`
}

// PendingPayment represents a pending payment transaction
type PendingPayment struct {
	Token     string    `json:"token"`
	SenderID  int64     `json:"sender_id"`
	TargetID  int64     `json:"target_id"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	ChatID    int64     `json:"chat_id"`
	MessageID int       `json:"message_id,omitempty"`
}

// PendingTrade represents a pending trade
type PendingTrade struct {
	SenderID         int64     `json:"sender_id"`
	ReceiverID       int64     `json:"receiver_id"`
	SenderCharID     string    `json:"sender_char_id"`
	ReceiverCharID   string    `json:"receiver_char_id"`
	Timestamp        time.Time `json:"timestamp"`
}

// PendingGift represents a pending gift
type PendingGift struct {
	SenderID           int64     `json:"sender_id"`
	ReceiverID         int64     `json:"receiver_id"`
	Character          UserCharacter `json:"character"`
	ReceiverUsername   string    `json:"receiver_username"`
	ReceiverFirstName  string    `json:"receiver_first_name"`
	Timestamp          time.Time `json:"timestamp"`
}
