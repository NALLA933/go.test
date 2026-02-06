package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the bot
type Config struct {
	// Bot Credentials
	BotToken    string
	BotUsername string

	// Telegram API
	APIID   int64
	APIHash string

	// Owner and Sudo
	OwnerID    int64
	SudoUsers  []int64

	// Group IDs
	GroupID         int64
	CharaChannelID  int64

	// Database
	MongoURL string

	// Media
	VideoURLs []string

	// Community Links
	SupportChat string
	UpdateChat  string
}

var (
	// AppConfig is the global configuration instance
	AppConfig *Config
)

// Load loads configuration from environment variables
func Load() *Config {
	// Try to load .env file (optional)
	_ = godotenv.Load()

	config := &Config{
		BotToken:    getEnv("BOT_TOKEN", ""),
		BotUsername: getEnv("BOT_USERNAME", "Senpai_Waifu_Grabbing_Bot"),
		APIHash:     getEnv("API_HASH", ""),
		MongoURL:    getEnv("MONGO_URL", ""),
		SupportChat: getEnv("SUPPORT_CHAT", "THE_DRAGON_SUPPORT"),
		UpdateChat:  getEnv("UPDATE_CHAT", "Senpai_Updates"),
	}

	// Parse integers
	config.APIID = parseInt64(getEnv("API_ID", "0"))
	config.OwnerID = parseInt64(getEnv("OWNER_ID", "0"))
	config.GroupID = parseInt64(getEnv("GROUP_ID", "0"))
	config.CharaChannelID = parseInt64(getEnv("CHARA_CHANNEL_ID", "0"))

	// Parse sudo users
	sudoUsersStr := getEnv("SUDO_USERS", "")
	if sudoUsersStr != "" {
		users := strings.Split(sudoUsersStr, ",")
		for _, u := range users {
			u = strings.TrimSpace(u)
			if u != "" {
				if id, err := strconv.ParseInt(u, 10, 64); err == nil {
					config.SudoUsers = append(config.SudoUsers, id)
				}
			}
		}
	}

	// Add owner to sudo users if not already present
	if config.OwnerID != 0 {
		found := false
		for _, id := range config.SudoUsers {
			if id == config.OwnerID {
				found = true
				break
			}
		}
		if !found {
			config.SudoUsers = append(config.SudoUsers, config.OwnerID)
		}
	}

	// Parse video URLs
	videoURLsStr := getEnv("VIDEO_URL", "")
	if videoURLsStr != "" {
		urls := strings.Split(videoURLsStr, ",")
		for _, url := range urls {
			url = strings.TrimSpace(url)
			if url != "" {
				config.VideoURLs = append(config.VideoURLs, url)
			}
		}
	}

	// Validate critical config
	config.Validate()

	AppConfig = config
	return config
}

// Validate checks if required configuration is present
func (c *Config) Validate() {
	errors := []string{}

	if c.BotToken == "" {
		errors = append(errors, "BOT_TOKEN is required")
	}
	if c.APIID == 0 {
		errors = append(errors, "API_ID is required")
	}
	if c.APIHash == "" {
		errors = append(errors, "API_HASH is required")
	}
	if c.OwnerID == 0 {
		errors = append(errors, "OWNER_ID is required")
	}
	if c.MongoURL == "" {
		errors = append(errors, "MONGO_URL is required")
	}
	if c.GroupID == 0 {
		errors = append(errors, "GROUP_ID is required")
	}
	if c.CharaChannelID == 0 {
		errors = append(errors, "CHARA_CHANNEL_ID is required")
	}

	if len(errors) > 0 {
		log.Println("❌ Configuration Error(s):")
		for _, err := range errors {
			log.Printf("   - %s\n", err)
		}
		log.Println("\n💡 Please set the required environment variables and try again.")
		os.Exit(1)
	}
}

// IsSudo checks if a user is sudo or owner
func (c *Config) IsSudo(userID int64) bool {
	if userID == c.OwnerID {
		return true
	}
	for _, id := range c.SudoUsers {
		if id == userID {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
