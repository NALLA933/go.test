package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// cmdSFind handles /sfind command
func (b *Bot) cmdSFind(msg *tgbotapi.Message) {
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🔍 %s</b>\n\n%s <code>/sfind &lt;character_name&gt;</code>\n\n%s",
				utils.ToSmallCaps("SEARCH CHARACTERS"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Search for characters by name or anime.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	query := strings.Join(args[1:], " ")
	
	// Search characters
	chars, err := b.CharacterService.SearchCharacters(query)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Search failed. Please try again."))
		b.API.Send(reply)
		return
	}
	
	if len(chars) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s <b>%s</b>\n\n%s",
				utils.ToSmallCaps("No characters found for:"), query,
				utils.ToSmallCaps("Try a different search term.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Pagination (first page)
	b.showSearchResults(msg.Chat.ID, chars, query, 0, msg.MessageID)
}

// showSearchResults shows search results with pagination
func (b *Bot) showSearchResults(chatID int64, chars []models.Character, query string, page int, replyTo int) {
	pageSize := 10
	totalPages := int(math.Ceil(float64(len(chars)) / float64(pageSize)))
	
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	
	startIdx := page * pageSize
	endIdx := startIdx + pageSize
	if endIdx > len(chars) {
		endIdx = len(chars)
	}
	
	pageChars := chars[startIdx:endIdx]
	
	// Build message
	message := fmt.Sprintf("<b>🔍 %s</b> <code>%s</code>\n<i>%s %d</i>\n\n",
		utils.ToSmallCaps("Search results for:"), query,
		utils.ToSmallCaps("Found"), len(chars))
	
	for _, char := range pageChars {
		rarityEmoji := utils.RarityEmojis[char.Rarity]
		message += fmt.Sprintf("🆔 <code>%s</code> - %s <b>%s</b>\n", 
			char.ID, rarityEmoji, utils.ToSmallCaps(char.Name))
	}
	
	// Build keyboard
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	
	// Navigation buttons
	if totalPages > 1 {
		var navButtons []tgbotapi.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
				"⬅️", fmt.Sprintf("sfind_prev:%s:%d", query, page-1)))
		}
		if page < totalPages-1 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
				"➡️", fmt.Sprintf("sfind_next:%s:%d", query, page+1)))
		}
		keyboardRows = append(keyboardRows, navButtons)
	}
	
	// Close button
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ "+utils.ToSmallCaps("Close"), "close_search"),
	))
	
	reply := tgbotapi.NewMessage(chatID, message)
	reply.ParseMode = "HTML"
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
	if replyTo > 0 {
		reply.ReplyToMessageID = replyTo
	}
	b.API.Send(reply)
}

// cmdSCheck handles /scheck command
func (b *Bot) cmdSCheck(msg *tgbotapi.Message) {
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🔎 %s</b>\n\n%s <code>/scheck &lt;character_id&gt;</code>\n\n%s",
				utils.ToSmallCaps("CHECK CHARACTER"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Get detailed information about a character.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	
	// Get character
	char, err := b.CharacterService.GetCharacterByID(charID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s <code>%s</code>\n\n%s",
				utils.ToSmallCaps("Character not found with ID:"), charID,
				utils.ToSmallCaps("Please check the ID and try again.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Get owner count
	ownerCount, _ := b.CharacterService.GetCharacterOwnerCount(charID)
	
	// Get top grabbers
	topGrabbers, _ := b.CharacterService.GetTopGrabbers(charID, 3)
	
	// Build message
	rarityDisplay := utils.GetRarityDisplay(char.Rarity)
	message := fmt.Sprintf(
		"<b>📋 %s</b>\n\n"+
			"🎭 <b>%s</b> %s\n"+
			"📺 <b>%s</b> %s\n"+
			"🆔 <b>%s</b> <code>%s</code>\n"+
			"⭐ <b>%s</b> %s\n"+
			"👥 <b>%s</b> %d\n\n",
		utils.ToSmallCaps("CHARACTER INFO"),
		utils.ToSmallCaps("Name:"), utils.ToSmallCaps(char.Name),
		utils.ToSmallCaps("Anime:"), utils.ToSmallCaps(char.Anime),
		utils.ToSmallCaps("ID:"), char.ID,
		utils.ToSmallCaps("Rarity:"), rarityDisplay,
		utils.ToSmallCaps("Owners:"), ownerCount,
	)
	
	if len(topGrabbers) > 0 {
		message += fmt.Sprintf("<b>🏆 %s</b>\n", utils.ToSmallCaps("Top Grabbers:"))
		for i, grabber := range topGrabbers {
			name := grabber["first_name"].(string)
			if username, ok := grabber["username"].(string); ok && username != "" {
				name = fmt.Sprintf("@%s", username)
			}
			count := grabber["count"].(int32)
			message += fmt.Sprintf("%d. %s - x%d\n", i+1, name, count)
		}
	}
	
	// Build keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonSwitch(
				utils.ToSmallCaps("📋 See Collection"),
				fmt.Sprintf("collection.%d", msg.From.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ "+utils.ToSmallCaps("Close"),
				"close_check",
			),
		),
	)
	
	if char.ImgURL != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(char.ImgURL))
		photo.Caption = message
		photo.ParseMode = "HTML"
		photo.ReplyMarkup = keyboard
		b.API.Send(photo)
	} else {
		reply := tgbotapi.NewMessage(chatID, message)
		reply.ParseMode = "HTML"
		reply.ReplyMarkup = keyboard
		b.API.Send(reply)
	}
}
