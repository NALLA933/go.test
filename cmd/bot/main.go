package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"senpai-waifu-bot/internal/config"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()
	
	// Create bot
	bot, err := handlers.NewBot(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	
	// Start bot in a goroutine
	go bot.Start()
	
	// Wait for interrupt signal
	log.Println("🤖 Bot is running. Press Ctrl+C to stop.")
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("🛑 Shutting down bot...")
}
