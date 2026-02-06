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

// cmdShop handles /shop command
func (b *Bot) cmdShop(msg *tgbotapi.Message) {
	userID := msg.From.ID
	
	// Get or create shop data
	shopData, err := b.UserService.GetShopData(userID)
	if err != nil || shopData == nil || len(shopData.Characters) == 0 {
		// Initialize new shop
		shopData, err = b.CharacterService.InitializeShop()
		if err != nil {
			reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Shop is empty! Please try again later."))
			b.API.Send(reply)
			return
		}
		_ = b.UserService.UpdateShopData(userID, shopData)
	}
	
	// Check if shop needs reset (24 hours)
	if time.Since(shopData.LastReset) > 24*time.Hour {
		shopData, _ = b.CharacterService.InitializeShop()
		_ = b.UserService.UpdateShopData(userID, shopData)
	}
	
	b.displayShopCharacter(msg.Chat.ID, userID, 0, msg.MessageID)
}

// displayShopCharacter displays a shop character
func (b *Bot) displayShopCharacter(chatID int64, userID int64, index int, replyTo int) {
	shopData, _ := b.UserService.GetShopData(userID)
	if shopData == nil || len(shopData.Characters) == 0 {
		reply := tgbotapi.NewMessage(chatID, utils.ToSmallCaps("⚠️ Shop is empty!"))
		b.API.Send(reply)
		return
	}
	
	if index < 0 {
		index = 0
	}
	if index >= len(shopData.Characters) {
		index = len(shopData.Characters) - 1
	}
	
	// Update current index
	shopData.CurrentIndex = index
	_ = b.UserService.UpdateShopData(userID, shopData)
	
	char := shopData.Characters[index]
	
	// Check if user already owns this character
	owned, _ := b.UserService.HasCharacter(userID, char.ID)
	status := "Available"
	if owned {
		status = "Sold"
	}
	
	// Get owner count
	ownerCount, _ := b.CharacterService.GetCharacterOwnerCount(char.ID)
	
	// Build message
	rarityEmoji := utils.RarityEmojis[char.Rarity]
	rarityName := utils.RarityNames[char.Rarity]
	
	message := fmt.Sprintf(
		"<b>🏪 %s</b>\n\n"+
			"🎭 %s: %s\n"+
			"📺 %s: %s\n"+
			"🆔 %s: %s\n"+
			"✨ %s: %s %s\n"+
			"💸 %s: %s\n"+
			"🛒 %s: %d%%\n"+
			"🏷️ %s: %s\n"+
			"🎴 %s: %d\n"+
			"📋 %s: %s",
		utils.ToSmallCaps(fmt.Sprintf("Character Shop (%d/%d)", index+1, len(shopData.Characters))),
		utils.ToSmallCaps("Name"), utils.ToSmallCaps(char.Name),
		utils.ToSmallCaps("Anime"), utils.ToSmallCaps(char.Anime),
		utils.ToSmallCaps("Id"), char.ID,
		utils.ToSmallCaps("Rarity"), rarityEmoji, rarityName,
		utils.ToSmallCaps("Price"), utils.FormatNumber(char.BasePrice),
		utils.ToSmallCaps("Discount"), char.DiscountPercent,
		utils.ToSmallCaps("Discount Price"), utils.FormatNumber(char.FinalPrice),
		utils.ToSmallCaps("Owner"), ownerCount,
		utils.ToSmallCaps("Stats"), utils.ToSmallCaps(status),
	)
	
	// Build keyboard
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	
	// Purchase button
	if !owned {
		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				utils.ToSmallCaps("💰 Purchase"),
				fmt.Sprintf("shop_purchase:%d:%d", userID, index),
			),
		))
	} else {
		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				utils.ToSmallCaps("❌ Already Owned"),
				"shop_noop",
			),
		))
	}
	
	// Navigation buttons
	var navButtons []tgbotapi.InlineKeyboardButton
	if index > 0 {
		navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
			"⬅️",
			fmt.Sprintf("shop_nav:%d:%d", userID, index-1),
		))
	}
	
	navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("🍃 %s", utils.ToSmallCaps("Refresh")),
		fmt.Sprintf("shop_refresh:%d", userID),
	))
	
	if index < len(shopData.Characters)-1 {
		navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(
			"➡️",
			fmt.Sprintf("shop_nav:%d:%d", userID, index+1),
		))
	}
	
	keyboardRows = append(keyboardRows, navButtons)
	
	// Premium shop button
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("💸 %s", utils.ToSmallCaps("Premium Shop")),
			fmt.Sprintf("shop_premium:%d", userID),
		),
	))
	
	// Close button
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			utils.ToSmallCaps("❌ Close"),
			fmt.Sprintf("shop_close:%d", userID),
		),
	))
	
	if char.ImgURL != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(char.ImgURL))
		photo.Caption = message
		photo.ParseMode = "HTML"
		photo.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		if replyTo > 0 {
			photo.ReplyToMessageID = replyTo
		}
		b.API.Send(photo)
	} else {
		reply := tgbotapi.NewMessage(chatID, message)
		reply.ParseMode = "HTML"
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		if replyTo > 0 {
			reply.ReplyToMessageID = replyTo
		}
		b.API.Send(reply)
	}
}

// cmdResetShop handles /resetshop command (admin only)
func (b *Bot) cmdResetShop(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("Usage: /resetshop <user_id>"))
		b.API.Send(reply)
		return
	}
	
	targetID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Invalid user ID!"))
		b.API.Send(reply)
		return
	}
	
	shopData, _ := b.CharacterService.InitializeShop()
	_ = b.UserService.UpdateShopData(targetID, shopData)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps(fmt.Sprintf("✅ Shop reset successfully for user %d!", targetID)))
	b.API.Send(reply)
}

// processShopPurchase processes a shop purchase
func (b *Bot) processShopPurchase(chatID int64, userID int64, index int) {
	shopData, _ := b.UserService.GetShopData(userID)
	if shopData == nil || index >= len(shopData.Characters) {
		return
	}
	
	char := shopData.Characters[index]
	
	// Check if already owned
	owned, _ := b.UserService.HasCharacter(userID, char.ID)
	if owned {
		b.API.Send(tgbotapi.NewMessage(chatID, utils.ToSmallCaps("⚠️ You already own this character!")))
		return
	}
	
	// Check balance
	balance, _ := b.UserService.GetUserBalance(userID)
	if balance < char.FinalPrice {
		b.API.Send(tgbotapi.NewMessage(chatID, utils.ToSmallCaps(fmt.Sprintf("⚠️ Insufficient balance! Need %s coins", utils.FormatNumber(char.FinalPrice)))))
		return
	}
	
	// Get full character data
	fullChar, err := b.CharacterService.GetCharacterByID(char.ID)
	if err != nil {
		b.API.Send(tgbotapi.NewMessage(chatID, utils.ToSmallCaps("⚠️ Character not found in database!")))
		return
	}
	
	// Deduct balance
	newBalance, _ := b.UserService.UpdateUserBalance(userID, -char.FinalPrice)
	
	// Add character
	userChar := models.UserCharacter{
		ID:     fullChar.ID,
		Name:   fullChar.Name,
		Anime:  fullChar.Anime,
		Rarity: fullChar.Rarity,
		ImgURL: fullChar.ImgURL,
	}
	_ = b.UserService.AddCharacterToUser(userID, userChar)
	
	// Send success message
	successMsg := fmt.Sprintf(
		"<b>✅ %s</b>\n\n"+
			"🎉 %s: %s\n"+
			"💸 %s: %s\n"+
			"💰 %s: %s",
		utils.ToSmallCaps("Purchase Successful!"),
		utils.ToSmallCaps("You got"), utils.ToSmallCaps(char.Name),
		utils.ToSmallCaps("Price"), utils.FormatNumber(char.FinalPrice),
		utils.ToSmallCaps("New Balance"), utils.FormatNumber(newBalance),
	)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("⬅️ %s", utils.ToSmallCaps("Back to Shop")),
				fmt.Sprintf("shop_nav:%d:%d", userID, index),
			),
		),
	)
	
	reply := tgbotapi.NewMessage(chatID, successMsg)
	reply.ParseMode = "HTML"
	reply.ReplyMarkup = keyboard
	b.API.Send(reply)
}

// refreshShop refreshes the shop for a user
func (b *Bot) refreshShop(chatID int64, userID int64) {
	shopData, _ := b.UserService.GetShopData(userID)
	if shopData != nil && shopData.RefreshUsed {
		b.API.Send(tgbotapi.NewMessage(chatID, utils.ToSmallCaps("⚠️ You have reached daily limit of 1 refresh!")))
		return
	}
	
	// Check balance
	balance, _ := b.UserService.GetUserBalance(userID)
	refreshCost := int64(20000)
	if balance < refreshCost {
		b.API.Send(tgbotapi.NewMessage(chatID, utils.ToSmallCaps(fmt.Sprintf("⚠️ Insufficient balance! Need %s coins", utils.FormatNumber(refreshCost)))))
		return
	}
	
	// Deduct refresh cost
	_, _ = b.UserService.UpdateUserBalance(userID, -refreshCost)
	
	// Generate new shop
	newShopData, _ := b.CharacterService.RefreshShop()
	newShopData.RefreshUsed = true
	_ = b.UserService.UpdateShopData(userID, newShopData)
	
	b.API.Send(tgbotapi.NewMessage(chatID, utils.ToSmallCaps(fmt.Sprintf("✅ Shop refreshed! Cost: %s coins", utils.FormatNumber(refreshCost)))))
	b.displayShopCharacter(chatID, userID, 0, 0)
}

// cmdSClaim handles /sclaim command
func (b *Bot) cmdSClaim(msg *tgbotapi.Message) {
	userID := msg.From.ID
	
	// Check cooldown
	canClaim, remaining, _ := b.UserService.CanSClaim(userID)
	if !canClaim {
		hours := int(remaining.Hours())
		minutes := int(remaining.Minutes()) % 60
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf(
			"<b>⏰ %s</b>\n\n⏳ %s <b>%dh %dm</b>\n\n💡 %s",
			utils.ToSmallCaps("COOLDOWN ACTIVE"),
			utils.ToSmallCaps("You can use /sclaim again in:"),
			hours, minutes,
			utils.ToSmallCaps("Come back later!"),
		))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Get random character from allowed rarities (2, 3, 4)
	allowedRarities := []int{2, 3, 4}
	chars, _ := b.CharacterService.GetRandomCharactersByRarities(allowedRarities, 1)
	if len(chars) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("❌ No characters available at the moment!"))
		b.API.Send(reply)
		return
	}
	
	char := chars[0]
	
	// Add character to user
	userChar := models.UserCharacter{
		ID:     char.ID,
		Name:   char.Name,
		Anime:  char.Anime,
		Rarity: char.Rarity,
		ImgURL: char.ImgURL,
	}
	_ = b.UserService.AddCharacterToUser(userID, userChar)
	
	// Update last claim time
	_ = b.UserService.UpdateLastSClaim(userID)
	
	// Send message
	rarityDisplay := utils.GetRarityDisplay(char.Rarity)
	message := fmt.Sprintf(
		"<b>🎉 %s</b>\n\n"+
			"🎴 <b>%s</b> %s\n"+
			"📺 <b>%s</b> %s\n"+
			"⭐ <b>%s</b> %s\n"+
			"🆔 <b>%s</b> %s\n\n"+
			"✅ %s",
		utils.ToSmallCaps("CONGRATULATIONS!"),
		utils.ToSmallCaps("Character:"), char.Name,
		utils.ToSmallCaps("Anime:"), char.Anime,
		utils.ToSmallCaps("Rarity:"), rarityDisplay,
		utils.ToSmallCaps("ID:"), char.ID,
		utils.ToSmallCaps("Character has been added to your collection!"),
	)
	
	if char.ImgURL != "" {
		photo := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileURL(char.ImgURL))
		photo.Caption = message
		photo.ParseMode = "HTML"
		b.API.Send(photo)
	} else {
		reply := tgbotapi.NewMessage(msg.Chat.ID, message)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
	}
}

// cmdClaim handles /claim command (daily coin code)
func (b *Bot) cmdClaim(msg *tgbotapi.Message) {
	userID := msg.From.ID
	
	// Check cooldown
	canClaim, remaining, _ := b.UserService.CanClaim(userID)
	if !canClaim {
		hours := int(remaining.Hours())
		minutes := int(remaining.Minutes()) % 60
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf(
			"<b>⏰ %s</b>\n\n⏳ %s <b>%dh %dm</b>\n\n💡 %s",
			utils.ToSmallCaps("COOLDOWN ACTIVE"),
			utils.ToSmallCaps("You can use /claim again in:"),
			hours, minutes,
			utils.ToSmallCaps("Come back later!"),
		))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Generate coin amount and code
	coinAmount := rand.Int63n(2001) + 1000 // 1000-3000
	code, _ := b.ClaimCodeService.CreateClaimCode(userID, coinAmount)
	
	// Update last claim time
	_ = b.UserService.UpdateLastClaim(userID)
	
	// Send message
	message := fmt.Sprintf(
		"<b>💰 %s</b>\n\n"+
			"🎟️ <b>%s</b> <code>%s</code>\n"+
			"💎 <b>%s</b> %s %s\n\n"+
			"📌 %s <code>/credeem %s</code> %s\n"+
			"⏰ %s",
		utils.ToSmallCaps("COIN CODE GENERATED!"),
		utils.ToSmallCaps("Your Code:"), code,
		utils.ToSmallCaps("Amount:"), utils.FormatNumber(coinAmount), utils.ToSmallCaps("coins"),
		utils.ToSmallCaps("Use"), code, utils.ToSmallCaps("to claim your coins!"),
		utils.ToSmallCaps("Valid for 24 hours"),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdCRedeem handles /credeem command
func (b *Bot) cmdCRedeem(msg *tgbotapi.Message) {
	userID := msg.From.ID
	args := strings.Fields(msg.Text)
	
	if len(args) < 2 {
		message := fmt.Sprintf(
			"<b>🎁 %s</b>\n\n📝 %s <code>/credeem &lt;CODE&gt;</code>\n\n💡 %s",
			utils.ToSmallCaps("REDEEM CODE"),
			utils.ToSmallCaps("Usage:"),
			utils.ToSmallCaps("Redeem your coin codes to add coins to your balance!"),
		)
		reply := tgbotapi.NewMessage(msg.Chat.ID, message)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	code := strings.ToUpper(args[1])
	
	// Get claim code
	claimCode, err := b.ClaimCodeService.GetClaimCode(code)
	if err != nil || claimCode.UserID != userID {
		message := fmt.Sprintf(
			"<b>❌ %s</b>\n\n⚠️ %s\n\n💡 %s",
			utils.ToSmallCaps("INVALID CODE"),
			utils.ToSmallCaps("This code does not exist or does not belong to you."),
			utils.ToSmallCaps("Use /claim to generate a new code!"),
		)
		reply := tgbotapi.NewMessage(msg.Chat.ID, message)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Check if already redeemed
	if claimCode.IsRedeemed {
		message := fmt.Sprintf(
			"<b>❌ %s</b>\n\n⚠️ %s\n\n💡 %s",
			utils.ToSmallCaps("CODE ALREADY REDEEMED"),
			utils.ToSmallCaps("This code has already been used."),
			utils.ToSmallCaps("Use /claim to generate a new code!"),
		)
		reply := tgbotapi.NewMessage(msg.Chat.ID, message)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Check if expired
	if b.ClaimCodeService.IsClaimCodeExpired(claimCode) {
		message := fmt.Sprintf(
			"<b>❌ %s</b>\n\n⚠️ %s\n\n💡 %s",
			utils.ToSmallCaps("CODE EXPIRED"),
			utils.ToSmallCaps("This code has expired (24 hours limit)."),
			utils.ToSmallCaps("Use /claim to generate a new code!"),
		)
		reply := tgbotapi.NewMessage(msg.Chat.ID, message)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Redeem code
	_, _ = b.ClaimCodeService.RedeemClaimCode(code, userID)
	
	// Add coins
	newBalance, _ := b.UserService.UpdateUserBalance(userID, claimCode.Amount)
	
	// Send success message
	message := fmt.Sprintf(
		"<b>✅ %s</b>\n\n"+
			"💰 <b>%s</b> %s\n"+
			"💎 <b>%s</b> %s %s\n\n"+
			"🎉 %s",
		utils.ToSmallCaps("CODE REDEEMED SUCCESSFULLY!"),
		utils.ToSmallCaps("Coins Added:"), utils.FormatNumber(claimCode.Amount),
		utils.ToSmallCaps("New Balance:"), utils.FormatNumber(newBalance), utils.ToSmallCaps("coins"),
		utils.ToSmallCaps("Enjoy your coins!"),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}
