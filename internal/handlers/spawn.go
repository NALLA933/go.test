package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// handleMessageCounter handles message counting for character spawning
func (b *Bot) handleMessageCounter(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	
	// Only count in groups
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		return
	}
	
	// Check if user is the same as last
	if lastUser, ok := b.LastUser[chatID]; ok && lastUser.UserID == userID {
		lastUser.Count++
	} else {
		b.LastUser[chatID] = &LastUserInfo{UserID: userID, Count: 1}
	}
	
	// Check if user has sent enough messages
	if lastUser, ok := b.LastUser[chatID]; ok && lastUser.Count >= 5 {
		// Reset count
		lastUser.Count = 0
		
		// Increment message counter
		b.MessageCounters[chatID]++
		
		// Get message frequency for this chat
		freq, _ := b.GroupService.GetMessageFrequency(chatID)
		if freq == 0 {
			freq = 100
		}
		
		// Check if it's time to spawn
		if b.MessageCounters[chatID] >= freq {
			// Reset counter
			b.MessageCounters[chatID] = 0
			
			// Spawn character
			b.spawnCharacter(chatID)
		}
	}
}

// spawnCharacter spawns a character in the chat
func (b *Bot) spawnCharacter(chatID int64) {
	// Get disabled rarities for this chat
	disabledRarities, _ := b.RarityService.GetDisabledRarities(chatID)
	
	// Get locked character IDs
	lockedIDs, _ := b.RarityService.GetLockedCharacterIDs()
	
	// Get random character
	char, err := b.CharacterService.GetRandomCharacter(disabledRarities, lockedIDs)
	if err != nil || char == nil {
		return
	}
	
	// Check if this character was recently sent to this chat
	if sent, ok := b.SentCharacters[chatID]; ok {
		for _, id := range sent {
			if id == char.ID {
				// Try again with a different character
				char, _ = b.CharacterService.GetRandomCharacter(disabledRarities, append(lockedIDs, char.ID))
				if char == nil {
					return
				}
				break
			}
		}
	}
	
	// Track sent character
	b.SentCharacters[chatID] = append(b.SentCharacters[chatID], char.ID)
	if len(b.SentCharacters[chatID]) > 10 {
		b.SentCharacters[chatID] = b.SentCharacters[chatID][1:]
	}
	
	// Store as last character
	b.LastCharacters[chatID] = &LastCharInfo{
		CharacterID: char.ID,
		Name:        char.Name,
		Anime:       char.Anime,
		Rarity:      char.Rarity,
		ImgURL:      char.ImgURL,
	}
	
	// Clear first correct guess
	delete(b.FirstCorrectGuesses, chatID)
	
	// Build spawn message
	rarityDisplay := utils.GetRarityDisplay(char.Rarity)
	
	// Get random message
	spawnMessages := []string{
		"A new character appeared!",
		"Look who just showed up!",
		"A wild character appeared!",
		"Guess who's here!",
		"A character has spawned!",
	}
	spawnMsg := spawnMessages[rand.Intn(len(spawnMessages))]
	
	message := fmt.Sprintf(
		"<b>%s</b>\n\n"+
			"🎭 <b>%s</b> %s\n"+
			"📺 <b>%s</b> %s\n"+
			"⭐ <b>%s</b> %s\n"+
			"🆔 <b>%s</b> <code>%s</code>\n\n"+
			"📝 %s",
		utils.ToSmallCaps(spawnMsg),
		utils.ToSmallCaps("Name:"), utils.ToSmallCaps(char.Name),
		utils.ToSmallCaps("Anime:"), utils.ToSmallCaps(char.Anime),
		utils.ToSmallCaps("Rarity:"), rarityDisplay,
		utils.ToSmallCaps("ID:"), char.ID,
		utils.ToSmallCaps("Use /guess <name> to catch this character!"),
	)
	
	// Send character image
	if char.ImgURL != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(char.ImgURL))
		photo.Caption = message
		photo.ParseMode = "HTML"
		b.API.Send(photo)
	} else {
		reply := tgbotapi.NewMessage(chatID, message)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
	}
}

// handleChatMemberUpdate handles bot being added/removed from groups
func (b *Bot) handleChatMemberUpdate(update *tgbotapi.ChatMemberUpdated) {
	if update.NewChatMember.User.ID == b.API.Self.ID {
		if update.NewChatMember.Status == "member" || update.NewChatMember.Status == "administrator" {
			// Bot was added to group
			welcomeMsg := fmt.Sprintf(
				"<b>🎉 %s</b>\n\n"+
					"%s\n\n"+
					"📌 %s <code>/guess &lt;name&gt;</code>\n"+
					"📌 %s <code>/collection</code>\n"+
					"📌 %s <code>/help</code>",
				utils.ToSmallCaps("Thanks for adding me!"),
				utils.ToSmallCaps("I'll spawn characters after every 100 messages."),
				utils.ToSmallCaps("Use"),
				utils.ToSmallCaps("View your"),
				utils.ToSmallCaps("For more commands, use"),
			)
			reply := tgbotapi.NewMessage(update.Chat.ID, welcomeMsg)
			reply.ParseMode = "HTML"
			b.API.Send(reply)
		}
	}
}

// handleCallback handles inline keyboard callbacks
func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	data := query.Data
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID
	
	// Answer the callback
	callback := tgbotapi.NewCallback(query.ID, "")
	b.API.Request(callback)
	
	// Handle different callback types
	switch {
	case strings.HasPrefix(data, "harem:"):
		// Harem pagination
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			page, _ := strconv.Atoi(parts[1])
			ownerID, _ := strconv.ParseInt(parts[2], 10, 64)
			if userID == ownerID {
				b.showHaremPage(chatID, messageID, ownerID, page)
			}
		}
		
	case strings.HasPrefix(data, "open_smode:"):
		// Open smode for user
		parts := strings.Split(data, ":")
		if len(parts) == 2 {
			ownerID, _ := strconv.ParseInt(parts[1], 10, 64)
			if userID == ownerID {
				b.showSMode(chatID, messageID, userID)
			}
		}
		
	case strings.HasPrefix(data, "smode_"):
		// Smode selection
		if data == "smode_all" {
			_ = b.SortPrefService.SetUserSortPreference(userID, nil)
			b.showSMode(chatID, messageID, userID)
		} else if data == "smode_cancel" {
			b.deleteMessage(chatID, messageID)
		} else {
			parts := strings.Split(data, "_")
			if len(parts) == 2 {
				rarity, _ := strconv.Atoi(parts[1])
				_ = b.SortPrefService.SetUserSortPreference(userID, &rarity)
				b.showSMode(chatID, messageID, userID)
			}
		}
		
	case strings.HasPrefix(data, "shop_nav:"):
		// Shop navigation
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			ownerID, _ := strconv.ParseInt(parts[1], 10, 64)
			index, _ := strconv.Atoi(parts[2])
			if userID == ownerID {
				b.updateShopMessage(chatID, messageID, ownerID, index)
			}
		}
		
	case strings.HasPrefix(data, "shop_purchase:"):
		// Shop purchase
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			ownerID, _ := strconv.ParseInt(parts[1], 10, 64)
			index, _ := strconv.Atoi(parts[2])
			if userID == ownerID {
				b.processShopPurchase(chatID, ownerID, index)
				b.deleteMessage(chatID, messageID)
			}
		}
		
	case strings.HasPrefix(data, "shop_refresh:"):
		// Shop refresh
		parts := strings.Split(data, ":")
		if len(parts) == 2 {
			ownerID, _ := strconv.ParseInt(parts[1], 10, 64)
			if userID == ownerID {
				b.refreshShop(chatID, ownerID)
				b.deleteMessage(chatID, messageID)
			}
		}
		
	case strings.HasPrefix(data, "shop_close:"):
		// Close shop
		parts := strings.Split(data, ":")
		if len(parts) == 2 {
			ownerID, _ := strconv.ParseInt(parts[1], 10, 64)
			if userID == ownerID {
				b.deleteMessage(chatID, messageID)
			}
		}
		
	case strings.HasPrefix(data, "pay_confirm:"):
		// Confirm payment
		parts := strings.Split(data, ":")
		if len(parts) == 2 {
			token := parts[1]
			b.confirmPayment(chatID, messageID, token, userID)
		}
		
	case strings.HasPrefix(data, "pay_cancel:"):
		// Cancel payment
		parts := strings.Split(data, ":")
		if len(parts) == 2 {
			token := parts[1]
			b.cancelPayment(chatID, messageID, token, userID)
		}
		
	case strings.HasPrefix(data, "accept_trade:"):
		// Accept trade
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			senderID, _ := strconv.ParseInt(parts[1], 10, 64)
			receiverID, _ := strconv.ParseInt(parts[2], 10, 64)
			if userID == receiverID {
				b.acceptTrade(chatID, messageID, senderID, receiverID)
			}
		}
		
	case strings.HasPrefix(data, "decline_trade:"):
		// Decline trade
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			senderID, _ := strconv.ParseInt(parts[1], 10, 64)
			receiverID, _ := strconv.ParseInt(parts[2], 10, 64)
			if userID == receiverID {
				b.declineTrade(chatID, messageID, senderID, receiverID)
			}
		}
		
	case strings.HasPrefix(data, "confirm_gift:"):
		// Confirm gift
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			senderID, _ := strconv.ParseInt(parts[1], 10, 64)
			receiverID, _ := strconv.ParseInt(parts[2], 10, 64)
			if userID == senderID {
				b.confirmGift(chatID, messageID, senderID, receiverID)
			}
		}
		
	case strings.HasPrefix(data, "cancel_gift:"):
		// Cancel gift
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			senderID, _ := strconv.ParseInt(parts[1], 10, 64)
			receiverID, _ := strconv.ParseInt(parts[2], 10, 64)
			if userID == senderID {
				b.cancelGift(chatID, messageID, senderID, receiverID)
			}
		}
		
	case strings.HasPrefix(data, "sfind_prev:") || strings.HasPrefix(data, "sfind_next:"):
		// Search pagination
		parts := strings.SplitN(data, ":", 3)
		if len(parts) == 3 {
			query := parts[1]
			page, _ := strconv.Atoi(parts[2])
			chars, _ := b.CharacterService.SearchCharacters(query)
			b.updateSearchResults(chatID, messageID, chars, query, page)
		}
		
	case data == "close_search" || data == "close_check":
		// Close search/check
		b.deleteMessage(chatID, messageID)
		
	case data == "help":
		// Show help
		b.showHelp(chatID)
	}
}

// showHaremPage shows a specific harem page
func (b *Bot) showHaremPage(chatID int64, messageID int, userID int64, page int) {
	// Get user's sort preference
	rarityFilter, _ := b.SortPrefService.GetUserSortPreference(userID)
	
	// Get user data
	user, err := b.UserService.GetUserByID(userID)
	if err != nil || len(user.Characters) == 0 {
		return
	}
	
	// Filter characters if needed
	characters := user.Characters
	if rarityFilter != nil {
		var filtered []models.UserCharacter
		for _, char := range characters {
			if char.Rarity == *rarityFilter {
				filtered = append(filtered, char)
			}
		}
		characters = filtered
	}
	
	if len(characters) == 0 {
		return
	}
	
	// This would update the harem message - simplified for now
	// In a full implementation, we'd rebuild the message and keyboard
}

// showSMode shows the smode selection
func (b *Bot) showSMode(chatID int64, messageID int, userID int64) {
	// Get current preference
	rarityFilter, _ := b.SortPrefService.GetUserSortPreference(userID)
	
	var currentText string
	if rarityFilter == nil {
		currentText = "🍃 " + utils.ToSmallCaps("default")
	} else {
		currentText = utils.RarityNames[*rarityFilter]
	}
	
	caption := fmt.Sprintf(
		"<b>✨ %s</b>\n\n🎯 %s <b>%s</b>\n",
		utils.ToSmallCaps("SMODE"),
		utils.ToSmallCaps("Current Model:"),
		currentText,
	)
	
	// Build keyboard
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	
	// All rarities button
	allText := utils.ToSmallCaps("🍃 default")
	if rarityFilter == nil {
		allText += " ✓"
	}
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(allText, "smode_all"),
	))
	
	// Rarity buttons (3 per row)
	var currentRow []tgbotapi.InlineKeyboardButton
	for i := 1; i <= 15; i++ {
		btnText := utils.RarityNames[i]
		if rarityFilter != nil && *rarityFilter == i {
			btnText += " ✓"
		}
		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(
			btnText,
			fmt.Sprintf("smode_%d", i),
		))
		
		if len(currentRow) == 3 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		keyboardRows = append(keyboardRows, currentRow)
	}
	
	// Cancel button
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ "+utils.ToSmallCaps("Cancel"), "smode_cancel"),
	))
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, caption)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
	b.API.Send(edit)
}

// updateShopMessage updates the shop message
func (b *Bot) updateShopMessage(chatID int64, messageID int, userID int64, index int) {
	// This would update the shop message - simplified for now
	// In a full implementation, we'd rebuild the message and keyboard
}

// updateSearchResults updates search results
func (b *Bot) updateSearchResults(chatID int64, messageID int, chars []models.Character, query string, page int) {
	// This would update the search results - simplified for now
}

// deleteMessage deletes a message
func (b *Bot) deleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.API.Request(deleteMsg)
}

// showHelp shows help message
func (b *Bot) showHelp(chatID int64) {
	helpText := fmt.Sprintf(
		"<b>📚 %s</b>\n\n"+
			"<b>%s</b>\n"+
			"• <code>/guess &lt;name&gt;</code> - %s\n"+
			"• <code>/collection</code> - %s\n"+
			"• <code>/balance</code> - %s\n"+
			"• <code>/pay &lt;amount&gt;</code> - %s\n"+
			"• <code>/shop</code> - %s\n"+
			"• <code>/gift &lt;id&gt;</code> - %s\n"+
			"• <code>/trade &lt;your_id&gt; &lt;their_id&gt;</code> - %s\n"+
			"• <code>/sclaim</code> - %s\n"+
			"• <code>/claim</code> - %s\n"+
			"• <code>/redeem &lt;code&gt;</code> - %s\n"+
			"• <code>/leaderboard</code> - %s\n"+
			"• <code>/sfind &lt;name&gt;</code> - %s\n"+
			"• <code>/scheck &lt;id&gt;</code> - %s\n"+
			"• <code>/smode</code> - %s\n"+
			"• <code>/fav &lt;id&gt;</code> - %s",
		utils.ToSmallCaps("HELP MENU"),
		utils.ToSmallCaps("Commands:"),
		utils.ToSmallCaps("Guess the character name"),
		utils.ToSmallCaps("View your collection"),
		utils.ToSmallCaps("Check your balance"),
		utils.ToSmallCaps("Send coins to another user"),
		utils.ToSmallCaps("Browse the character shop"),
		utils.ToSmallCaps("Gift a character to someone"),
		utils.ToSmallCaps("Trade characters with someone"),
		utils.ToSmallCaps("Claim a free character (24h cooldown)"),
		utils.ToSmallCaps("Generate a coin code (24h cooldown)"),
		utils.ToSmallCaps("Redeem a code for rewards"),
		utils.ToSmallCaps("View leaderboards"),
		utils.ToSmallCaps("Search for characters"),
		utils.ToSmallCaps("Check character details"),
		utils.ToSmallCaps("Change collection filter"),
		utils.ToSmallCaps("Add character to favorites"),
	)
	
	reply := tgbotapi.NewMessage(chatID, helpText)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}
