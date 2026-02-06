package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// cmdTrade handles /trade command
func (b *Bot) cmdTrade(msg *tgbotapi.Message) {
	senderID := msg.From.ID
	
	// Check cooldown
	if nextAllowed, ok := b.TradeCooldowns[senderID]; ok && time.Now().Before(nextAllowed) {
		remaining := int(time.Until(nextAllowed).Seconds())
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("⏰ Please wait %d seconds before initiating another trade!", remaining))
		b.API.Send(reply)
		return
	}
	
	// Must reply to a message
	if msg.ReplyToMessage == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			"❌ You need to reply to a user's message to trade a character!\n\nUsage: /trade <your_character_id> <their_character_id>")
		b.API.Send(reply)
		return
	}
	
	receiverID := msg.ReplyToMessage.From.ID
	receiverMention := fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", receiverID, msg.ReplyToMessage.From.FirstName)
	
	if senderID == receiverID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ You can't trade a character with yourself!")
		b.API.Send(reply)
		return
	}
	
	// Parse arguments
	args := strings.Fields(msg.Text)
	if len(args) < 3 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			"❌ Invalid format! Usage:\n/trade <your_character_id> <their_character_id>\n\nReply to the user's message you want to trade with.")
		b.API.Send(reply)
		return
	}
	
	senderCharID := args[1]
	receiverCharID := args[2]
	
	// Check if sender has the character
	sender, err := b.UserService.GetUserByID(senderID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ You don't have any characters yet!")
		b.API.Send(reply)
		return
	}
	
	senderHasChar := false
	var senderChar models.UserCharacter
	for _, char := range sender.Characters {
		if char.ID == senderCharID {
			senderHasChar = true
			senderChar = char
			break
		}
	}
	
	if !senderHasChar {
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("❌ You don't have a character with ID %s!", senderCharID))
		b.API.Send(reply)
		return
	}
	
	// Check if receiver has the character
	receiver, err := b.UserService.GetUserByID(receiverID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("❌ %s doesn't have any characters yet!", receiverMention))
		b.API.Send(reply)
		return
	}
	
	receiverHasChar := false
	var receiverChar models.UserCharacter
	for _, char := range receiver.Characters {
		if char.ID == receiverCharID {
			receiverHasChar = true
			receiverChar = char
			break
		}
	}
	
	if !receiverHasChar {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s doesn't have a character with ID %s!", receiverMention, receiverCharID))
		b.API.Send(reply)
		return
	}
	
	// Check for existing pending trade
	tradeKey := fmt.Sprintf("%d:%d", senderID, receiverID)
	if _, exists := b.PendingTrades[tradeKey]; exists {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ You already have a pending trade with this user!")
		b.API.Send(reply)
		return
	}
	
	// Store pending trade
	b.PendingTrades[tradeKey] = &PendingTradeInfo{
		SenderID:       senderID,
		ReceiverID:     receiverID,
		SenderCharID:   senderCharID,
		ReceiverCharID: receiverCharID,
		Timestamp:      time.Now(),
	}
	
	// Send trade request
	tradeMsg := fmt.Sprintf(
		"🔄 **Trade Request**\n\n"+
			"**%s** wants to trade:\n"+
			"**%s**\n⭐ Rarity: %s\n📺 Anime: %s\n\n"+
			"For your:\n"+
			"**%s**\n⭐ Rarity: %s\n📺 Anime: %s\n\n"+
			"%s, please confirm or decline this trade.",
		msg.From.FirstName,
		senderChar.Name, utils.GetRarityDisplay(senderChar.Rarity), senderChar.Anime,
		receiverChar.Name, utils.GetRarityDisplay(receiverChar.Rarity), receiverChar.Anime,
		receiverMention,
	)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Accept Trade", fmt.Sprintf("accept_trade:%d:%d", senderID, receiverID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Decline Trade", fmt.Sprintf("decline_trade:%d:%d", senderID, receiverID)),
		),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, tradeMsg)
	reply.ParseMode = "Markdown"
	reply.ReplyMarkup = keyboard
	b.API.Send(reply)
	
	// Set cooldown
	b.TradeCooldowns[senderID] = time.Now().Add(60 * time.Second)
}

// cmdGift handles /gift command
func (b *Bot) cmdGift(msg *tgbotapi.Message) {
	senderID := msg.From.ID
	
	// Check cooldown
	if nextAllowed, ok := b.GiftCooldowns[senderID]; ok && time.Now().Before(nextAllowed) {
		remaining := int(time.Until(nextAllowed).Seconds())
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("⏰ Please wait %d seconds before gifting another character!", remaining))
		b.API.Send(reply)
		return
	}
	
	// Must reply to a message
	if msg.ReplyToMessage == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			"❌ You need to reply to a user's message to gift a character!\n\nUsage: /gift <character_id>")
		b.API.Send(reply)
		return
	}
	
	receiverID := msg.ReplyToMessage.From.ID
	receiverMention := fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", receiverID, msg.ReplyToMessage.From.FirstName)
	
	if senderID == receiverID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ You can't gift a character to yourself!")
		b.API.Send(reply)
		return
	}
	
	// Parse arguments
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			"❌ Invalid format! Usage:\n/gift <character_id>\n\nReply to the user's message you want to gift to.")
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	
	// Check if sender has the character
	sender, err := b.UserService.GetUserByID(senderID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ You don't have any characters yet!")
		b.API.Send(reply)
		return
	}
	
	var giftChar models.UserCharacter
	charFound := false
	for _, char := range sender.Characters {
		if char.ID == charID {
			giftChar = char
			charFound = true
			break
		}
	}
	
	if !charFound {
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("❌ You don't have a character with ID %s!", charID))
		b.API.Send(reply)
		return
	}
	
	// Check for existing pending gift
	giftKey := fmt.Sprintf("%d:%d", senderID, receiverID)
	if _, exists := b.PendingGifts[giftKey]; exists {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ You already have a pending gift for this user!")
		b.API.Send(reply)
		return
	}
	
	// Store pending gift
	b.PendingGifts[giftKey] = &PendingGiftInfo{
		SenderID:          senderID,
		ReceiverID:        receiverID,
		CharacterID:       charID,
		CharacterName:     giftChar.Name,
		ReceiverUsername:  msg.ReplyToMessage.From.UserName,
		ReceiverFirstName: msg.ReplyToMessage.From.FirstName,
		Timestamp:         time.Now(),
	}
	
	// Format gift card
	rarityDisplay := utils.GetRarityDisplay(giftChar.Rarity)
	giftCard := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━\n"+
			"🎁 %s\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"✨ %s   : **%s**\n"+
			"🎬 %s  : **%s**\n"+
			"🆔 %s     : `%s`\n"+
			"⭐ %s : %s\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"💎 %s **%s**",
		utils.ToSmallCaps("gift card"),
		utils.ToSmallCaps("name"), utils.ToSmallCaps(giftChar.Name),
		utils.ToSmallCaps("anime"), utils.ToSmallCaps(giftChar.Anime),
		utils.ToSmallCaps("id"), giftChar.ID,
		utils.ToSmallCaps("rarity"), rarityDisplay,
		utils.ToSmallCaps("premium gift from"), msg.From.FirstName,
	)
	
	giftMsg := fmt.Sprintf(
		"%s\n\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"Are you sure you want to gift this to %s?\n\n"+
			"⏰ %s",
		giftCard,
		receiverMention,
		utils.ToSmallCaps("you have 30 seconds to confirm"),
	)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Confirm Gift", fmt.Sprintf("confirm_gift:%d:%d", senderID, receiverID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel Gift", fmt.Sprintf("cancel_gift:%d:%d", senderID, receiverID)),
		),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, giftMsg)
	reply.ParseMode = "Markdown"
	reply.ReplyMarkup = keyboard
	b.API.Send(reply)
	
	// Set cooldown
	b.GiftCooldowns[senderID] = time.Now().Add(30 * time.Second)
}

// acceptTrade accepts a trade
func (b *Bot) acceptTrade(chatID int64, messageID int, senderID, receiverID int64) {
	tradeKey := fmt.Sprintf("%d:%d", senderID, receiverID)
	trade, ok := b.PendingTrades[tradeKey]
	if !ok {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ This trade has expired or doesn't exist!")
		b.API.Send(edit)
		return
	}
	
	// Check expiry (5 minutes)
	if time.Since(trade.Timestamp) > 5*time.Minute {
		delete(b.PendingTrades, tradeKey)
		delete(b.TradeCooldowns, senderID)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ This trade request has expired!")
		b.API.Send(edit)
		return
	}
	
	// Get both users' data
	sender, err1 := b.UserService.GetUserByID(senderID)
	receiver, err2 := b.UserService.GetUserByID(receiverID)
	
	if err1 != nil || err2 != nil {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Trade failed! Could not retrieve user data.")
		b.API.Send(edit)
		return
	}
	
	// Verify both still have their characters
	var senderChar, receiverChar models.UserCharacter
	senderHasChar := false
	receiverHasChar := false
	
	for _, char := range sender.Characters {
		if char.ID == trade.SenderCharID {
			senderChar = char
			senderHasChar = true
			break
		}
	}
	
	for _, char := range receiver.Characters {
		if char.ID == trade.ReceiverCharID {
			receiverChar = char
			receiverHasChar = true
			break
		}
	}
	
	if !senderHasChar {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Trade failed! The sender's character no longer exists.")
		b.API.Send(edit)
		delete(b.PendingTrades, tradeKey)
		return
	}
	
	if !receiverHasChar {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Trade failed! Your character no longer exists.")
		b.API.Send(edit)
		delete(b.PendingTrades, tradeKey)
		return
	}
	
	// Perform the trade
	// Remove sender's character
	_ = b.UserService.RemoveCharacterFromUser(senderID, trade.SenderCharID)
	// Remove receiver's character
	_ = b.UserService.RemoveCharacterFromUser(receiverID, trade.ReceiverCharID)
	// Add receiver's character to sender
	_ = b.UserService.AddCharacterToUser(senderID, receiverChar)
	// Add sender's character to receiver
	_ = b.UserService.AddCharacterToUser(receiverID, senderChar)
	
	// Clean up
	delete(b.PendingTrades, tradeKey)
	
	// Send success message
	successMsg := fmt.Sprintf(
		"✅ **Trade Successful!**\n\n"+
			"**%s** received:\n"+
			"**%s**\n⭐ Rarity: %s\n📺 Anime: %s\n\n"+
			"**%s** received:\n"+
			"**%s**\n⭐ Rarity: %s\n📺 Anime: %s",
		sender.FirstName,
		receiverChar.Name, utils.GetRarityDisplay(receiverChar.Rarity), receiverChar.Anime,
		receiver.FirstName,
		senderChar.Name, utils.GetRarityDisplay(senderChar.Rarity), senderChar.Anime,
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, successMsg)
	edit.ParseMode = "Markdown"
	b.API.Send(edit)
}

// declineTrade declines a trade
func (b *Bot) declineTrade(chatID int64, messageID int, senderID, receiverID int64) {
	tradeKey := fmt.Sprintf("%d:%d", senderID, receiverID)
	delete(b.PendingTrades, tradeKey)
	delete(b.TradeCooldowns, senderID)
	
	receiver, _ := b.UserService.GetUserByID(receiverID)
	receiverName := fmt.Sprintf("User %d", receiverID)
	if receiver != nil {
		receiverName = receiver.FirstName
	}
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, 
		fmt.Sprintf("❌ **Trade Declined**\n\n%s has declined the trade.", receiverName))
	edit.ParseMode = "Markdown"
	b.API.Send(edit)
}

// confirmGift confirms a gift
func (b *Bot) confirmGift(chatID int64, messageID int, senderID, receiverID int64) {
	giftKey := fmt.Sprintf("%d:%d", senderID, receiverID)
	gift, ok := b.PendingGifts[giftKey]
	if !ok {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ This gift has expired or doesn't exist!")
		b.API.Send(edit)
		return
	}
	
	// Check expiry (30 seconds)
	if time.Since(gift.Timestamp) > 30*time.Second {
		delete(b.PendingGifts, giftKey)
		delete(b.GiftCooldowns, senderID)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			"❌ This gift request has expired!\n\nYou can now send a new gift.")
		b.API.Send(edit)
		return
	}
	
	// Get sender's data
	sender, err := b.UserService.GetUserByID(senderID)
	if err != nil {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Gift failed! Could not retrieve user data.")
		b.API.Send(edit)
		return
	}
	
	// Verify sender still has the character
	var giftChar models.UserCharacter
	charFound := false
	for _, char := range sender.Characters {
		if char.ID == gift.CharacterID {
			giftChar = char
			charFound = true
			break
		}
	}
	
	if !charFound {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			"❌ Gift failed! The character no longer exists in your collection.")
		b.API.Send(edit)
		delete(b.PendingGifts, giftKey)
		return
	}
	
	// Check receiver inventory size
	receiverCharCount, _ := b.UserService.GetUserCharactersCount(receiverID)
	if receiverCharCount >= 5000 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "❌ Gift failed! Receiver's inventory is full.")
		b.API.Send(edit)
		delete(b.PendingGifts, giftKey)
		return
	}
	
	// Perform the gift
	_ = b.UserService.RemoveCharacterFromUser(senderID, gift.CharacterID)
	_ = b.UserService.AddCharacterToUser(receiverID, giftChar)
	
	// Clean up
	delete(b.PendingGifts, giftKey)
	
	// Send success message
	successMsg := fmt.Sprintf(
		"🎉 **%s**\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"💝 **%s** %s\n"+
			"%s **%s**\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"✨ %s",
		utils.ToSmallCaps("gift successful"),
		utils.ToSmallCaps(giftChar.Name), utils.ToSmallCaps("has been sent"),
		utils.ToSmallCaps("to"), gift.ReceiverFirstName,
		utils.ToSmallCaps("thank you for being generous"),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, successMsg)
	edit.ParseMode = "Markdown"
	b.API.Send(edit)
}

// cancelGift cancels a gift
func (b *Bot) cancelGift(chatID int64, messageID int, senderID, receiverID int64) {
	giftKey := fmt.Sprintf("%d:%d", senderID, receiverID)
	delete(b.PendingGifts, giftKey)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, 
		"❌ **Gift Cancelled**\n\nThe gift has been cancelled.")
	edit.ParseMode = "Markdown"
	b.API.Send(edit)
}
