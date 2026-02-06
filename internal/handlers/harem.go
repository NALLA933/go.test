package handlers

import (
	"fmt"
	"math"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// cmdHarem handles /harem command
func (b *Bot) cmdHarem(msg *tgbotapi.Message, page int) {
	userID := msg.From.ID
	
	// Get user's sort preference
	rarityFilter, _ := b.SortPrefService.GetUserSortPreference(userID)
	
	// Get user data
	user, err := b.UserService.GetUserByID(userID)
	if err != nil || len(user.Characters) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("You Have Not Guessed any Characters Yet.."))
		b.API.Send(reply)
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
		if rarityFilter != nil {
			filterName := utils.RarityNames[*rarityFilter]
			reply := tgbotapi.NewMessage(msg.Chat.ID, 
				utils.ToSmallCaps(fmt.Sprintf("You Have No Characters of %s!\n\nUse /smode to change filter.", filterName)))
			b.API.Send(reply)
		} else {
			reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("You Have Not Guessed any Characters Yet.."))
			b.API.Send(reply)
		}
		return
	}
	
	// Get unique characters and count duplicates
	charCounts := make(map[string]int)
	uniqueChars := make([]models.UserCharacter, 0)
	seen := make(map[string]bool)
	
	for _, char := range characters {
		charCounts[char.ID]++
		if !seen[char.ID] {
			seen[char.ID] = true
			uniqueChars = append(uniqueChars, char)
		}
	}
	
	// Sort by anime and ID
	sort.Slice(uniqueChars, func(i, j int) bool {
		if uniqueChars[i].Anime != uniqueChars[j].Anime {
			return uniqueChars[i].Anime < uniqueChars[j].Anime
		}
		return uniqueChars[i].ID < uniqueChars[j].ID
	})
	
	// Pagination
	pageSize := 15
	totalPages := int(math.Ceil(float64(len(uniqueChars)) / float64(pageSize)))
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}
	
	startIdx := page * pageSize
	endIdx := startIdx + pageSize
	if endIdx > len(uniqueChars) {
		endIdx = len(uniqueChars)
	}
	
	pageChars := uniqueChars[startIdx:endIdx]
	
	// Get anime counts
	animeSet := make(map[string]bool)
	for _, char := range pageChars {
		animeSet[char.Anime] = true
	}
	animes := make([]string, 0, len(animeSet))
	for anime := range animeSet {
		animes = append(animes, anime)
	}
	
	animeCounts, _ := b.CharacterService.GetAnimeCounts(animes)
	
	// Build message
	headerText := fmt.Sprintf("%s's HAREM - PAGE %d/%d", msg.From.FirstName, page+1, totalPages)
	haremMsg := fmt.Sprintf("<b>%s</b>\n", utils.ToSmallCaps(headerText))
	
	if rarityFilter != nil {
		filterName := utils.RarityNames[*rarityFilter]
		haremMsg += fmt.Sprintf("<b>🔍 %s</b>\n", utils.ToSmallCaps(fmt.Sprintf("Filter: %s (%d/%d)", filterName, len(characters), len(user.Characters))))
	}
	
	haremMsg += "\n"
	
	// Group by anime
	animeGroups := make(map[string][]models.UserCharacter)
	for _, char := range pageChars {
		animeGroups[char.Anime] = append(animeGroups[char.Anime], char)
	}
	
	for anime, chars := range animeGroups {
		totalAnimeChars := animeCounts[anime]
		haremMsg += fmt.Sprintf("<b>𖤍 %s {%d/%d}</b>\n", utils.ToSmallCaps(anime), len(chars), totalAnimeChars)
		haremMsg += fmt.Sprintf("%s\n", utils.ToSmallCaps("--------------------"))
		
		for _, char := range chars {
			rarityEmoji := utils.RarityEmojis[char.Rarity]
			count := charCounts[char.ID]
			haremMsg += fmt.Sprintf("✶ %s [ %s ] %s %s\n",
				char.ID,
				rarityEmoji,
				utils.ToSmallCaps(char.Name),
				utils.ToSmallCaps(fmt.Sprintf("x%d", count)),
			)
		}
		haremMsg += fmt.Sprintf("%s\n\n", utils.ToSmallCaps("--------------------"))
	}
	
	// Build keyboard
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	
	// Collection button
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonSwitch(
			utils.ToSmallCaps(fmt.Sprintf("🔮 See Collection (%d)", len(characters))),
			fmt.Sprintf("collection.%d", userID),
		),
	))
	
	// Cancel/Smode button
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ "+utils.ToSmallCaps("ᴄᴀɴᴄᴇʟ"),
			fmt.Sprintf("open_smode:%d", userID),
		),
	))
	
	// Navigation buttons
	if totalPages > 1 {
		var navButtons []tgbotapi.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
				"⬅️",
				fmt.Sprintf("harem:%d:%d", page-1, userID),
			))
		}
		if page < totalPages-1 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
				"➡️",
				fmt.Sprintf("harem:%d:%d", page+1, userID),
			))
		}
		if len(navButtons) > 0 {
			keyboardRows = append(keyboardRows, navButtons)
		}
	}
	
	// Get photo (first favorite or first character)
	var photoURL string
	if len(user.Favorites) > 0 {
		for _, char := range uniqueChars {
			if char.ID == user.Favorites[0] {
				photoURL = char.ImgURL
				break
			}
		}
	}
	if photoURL == "" && len(uniqueChars) > 0 {
		photoURL = uniqueChars[0].ImgURL
	}
	
	if photoURL != "" {
		photo := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileURL(photoURL))
		photo.Caption = haremMsg
		photo.ParseMode = "HTML"
		photo.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		b.API.Send(photo)
	} else {
		reply := tgbotapi.NewMessage(msg.Chat.ID, haremMsg)
		reply.ParseMode = "HTML"
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		b.API.Send(reply)
	}
}

// cmdSMode handles /smode command
func (b *Bot) cmdSMode(msg *tgbotapi.Message) {
	userID := msg.From.ID
	
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
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, caption)
	reply.ParseMode = "HTML"
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
	b.API.Send(reply)
}
