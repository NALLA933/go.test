package handlers

import (
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/config"
	"senpai-waifu-bot/internal/services"
)

// Bot represents the Telegram bot
type Bot struct {
	API                *tgbotapi.BotAPI
	Config             *config.Config
	UserService        *services.UserService
	CharacterService   *services.CharacterService
	GroupService       *services.GroupService
	DailyService       *services.DailyService
	RedeemService      *services.RedeemService
	ClaimCodeService   *services.ClaimCodeService
	RarityService      *services.RarityService
	SortPrefService    *services.SortPreferenceService
	
	// In-memory state
	MessageCounters    map[int64]int
	LastCharacters     map[int64]*LastCharInfo
	SentCharacters     map[int64][]string
	FirstCorrectGuesses map[int64]int64
	LastUser           map[int64]*LastUserInfo
	WarnedUsers        map[int64]time.Time
	ChatLocks          sync.Map
	
	// Payment state
	PendingPayments    map[string]*PendingPaymentInfo
	PaymentCooldowns   map[int64]time.Time
	
	// Trade/Gift state
	PendingTrades      map[string]*PendingTradeInfo
	PendingGifts       map[string]*PendingGiftInfo
	TradeCooldowns     map[int64]time.Time
	GiftCooldowns      map[int64]time.Time
}

// LastCharInfo stores info about the last spawned character in a chat
type LastCharInfo struct {
	CharacterID string
	Name        string
	Anime       string
	Rarity      int
	ImgURL      string
}

// LastUserInfo stores info about the last user who sent a message in a chat
type LastUserInfo struct {
	UserID int64
	Count  int
}

// PendingPaymentInfo stores pending payment info
type PendingPaymentInfo struct {
	Token     string
	SenderID  int64
	TargetID  int64
	Amount    int64
	CreatedAt time.Time
	ChatID    int64
	MessageID int
}

// PendingTradeInfo stores pending trade info
type PendingTradeInfo struct {
	SenderID       int64
	ReceiverID     int64
	SenderCharID   string
	ReceiverCharID string
	Timestamp      time.Time
}

// PendingGiftInfo stores pending gift info
type PendingGiftInfo struct {
	SenderID          int64
	ReceiverID        int64
	CharacterID       string
	CharacterName     string
	ReceiverUsername  string
	ReceiverFirstName string
	Timestamp         time.Time
}

// NewBot creates a new Bot instance
func NewBot(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}
	
	api.Debug = false
	log.Printf("✅ Authorized on account %s", api.Self.UserName)
	
	bot := &Bot{
		API:                 api,
		Config:              cfg,
		UserService:         services.NewUserService(),
		CharacterService:    services.NewCharacterService(),
		GroupService:        services.NewGroupService(),
		DailyService:        services.NewDailyService(),
		RedeemService:       services.NewRedeemService(),
		ClaimCodeService:    services.NewClaimCodeService(),
		RarityService:       services.NewRarityService(),
		SortPrefService:     services.NewSortPreferenceService(),
		MessageCounters:     make(map[int64]int),
		LastCharacters:      make(map[int64]*LastCharInfo),
		SentCharacters:      make(map[int64][]string),
		FirstCorrectGuesses: make(map[int64]int64),
		LastUser:            make(map[int64]*LastUserInfo),
		WarnedUsers:         make(map[int64]time.Time),
		PendingPayments:     make(map[string]*PendingPaymentInfo),
		PaymentCooldowns:    make(map[int64]time.Time),
		PendingTrades:       make(map[string]*PendingTradeInfo),
		PendingGifts:        make(map[string]*PendingGiftInfo),
		TradeCooldowns:      make(map[int64]time.Time),
		GiftCooldowns:       make(map[int64]time.Time),
	}
	
	// Start cleanup goroutine
	go bot.cleanupRoutine()
	
	return bot, nil
}

// Start starts the bot
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	updates := b.API.GetUpdatesChan(u)
	
	log.Println("🤖 Bot is running...")
	
	for update := range updates {
		go b.handleUpdate(update)
	}
}

// handleUpdate handles a single update
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()
	
	// Handle callback queries
	if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
		return
	}
	
	// Handle messages
	if update.Message == nil {
		return
	}
	
	// Handle chat member updates (bot added/removed from groups)
	if update.MyChatMember != nil {
		b.handleChatMemberUpdate(update.MyChatMember)
		return
	}
	
	// Count messages for character spawning
	if update.Message.Chat != nil && update.Message.From != nil {
		b.handleMessageCounter(update.Message)
	}
	
	// Handle commands
	if update.Message.IsCommand() {
		b.handleCommand(update.Message)
	}
}

// cleanupRoutine periodically cleans up expired data
func (b *Bot) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		
		// Clean up warned users (after 10 minutes)
		for userID, warnedAt := range b.WarnedUsers {
			if now.Sub(warnedAt) > 10*time.Minute {
				delete(b.WarnedUsers, userID)
			}
		}
		
		// Clean up pending payments (after 5 minutes)
		for token, payment := range b.PendingPayments {
			if now.Sub(payment.CreatedAt) > 5*time.Minute {
				delete(b.PendingPayments, token)
			}
		}
		
		// Clean up pending trades (after 5 minutes)
		for key, trade := range b.PendingTrades {
			if now.Sub(trade.Timestamp) > 5*time.Minute {
				delete(b.PendingTrades, key)
				delete(b.TradeCooldowns, trade.SenderID)
			}
		}
		
		// Clean up pending gifts (after 30 seconds)
		for key, gift := range b.PendingGifts {
			if now.Sub(gift.Timestamp) > 30*time.Second {
				delete(b.PendingGifts, key)
				delete(b.GiftCooldowns, gift.SenderID)
			}
		}
	}
}

// getChatLock gets or creates a lock for a chat
func (b *Bot) getChatLock(chatID int64) *sync.Mutex {
	lock, _ := b.ChatLocks.LoadOrStore(chatID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
