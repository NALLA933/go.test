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

// handleCommand handles bot commands
func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	command := msg.Command()
	
	switch command {
	case "start":
		b.cmdStart(msg)
	case "ping":
		b.cmdPing(msg)
	case "guess", "protecc", "collect", "grab", "hunt":
		b.cmdGuess(msg)
	case "harem", "collection":
		b.cmdHarem(msg, 0)
	case "bal", "balance":
		b.cmdBalance(msg)
	case "pay":
		b.cmdPay(msg)
	case "fav":
		b.cmdFav(msg)
	case "shop":
		b.cmdShop(msg)
	case "leaderboard":
		b.cmdLeaderboard(msg)
	case "gift":
		b.cmdGift(msg)
	case "trade":
		b.cmdTrade(msg)
	case "sfind", "find":
		b.cmdSFind(msg)
	case "scheck", "s", "check":
		b.cmdSCheck(msg)
	case "smode":
		b.cmdSMode(msg)
	case "sclaim":
		b.cmdSClaim(msg)
	case "claim":
		b.cmdClaim(msg)
	case "credeem":
		b.cmdCRedeem(msg)
	case "redeem":
		b.cmdRedeem(msg)
	case "gen":
		b.cmdGen(msg)
	case "sgen":
		b.cmdSGen(msg)
	case "addbal":
		b.cmdAddBal(msg)
	case "set_on":
		b.cmdSetOn(msg)
	case "set_off":
		b.cmdSetOff(msg)
	case "lock":
		b.cmdLock(msg)
	case "unlock":
		b.cmdUnlock(msg)
	case "locklist":
		b.cmdLockList(msg)
	case "resetshop":
		b.cmdResetShop(msg)
	case "upload":
		b.cmdUpload(msg)
	case "delete":
		b.cmdDelete(msg)
	case "update":
		b.cmdUpdate(msg)
	case "stats":
		b.cmdStats(msg)
	}
}

// cmdStart handles /start command
func (b *Bot) cmdStart(msg *tgbotapi.Message) {
	user := msg.From
	
	// Add to PM users
	_ = b.GroupService.AddPMUser(user.ID, user.UserName, user.FirstName)
	
	// Get random video URL
	var videoURL string
	if len(b.Config.VideoURLs) > 0 {
		videoURL = b.Config.VideoURLs[rand.Intn(len(b.Config.VideoURLs))]
	}
	
	caption := "✨ ᴡᴇʟᴄᴏᴍᴇ ᴛᴏ Sᴇɴᴘᴀɪ Wᴀɪғᴜ Bᴏᴛ ✨\n\nɪ'ᴍ ᴀɴ Sᴇɴᴘᴀɪ ᴄʜᴀʀᴀᴄᴛᴇʀ ᴄᴀᴛᴄʜᴇʀ ʙᴏᴛ ᴅᴇsɪɢɴᴇᴅ ғᴏʀ ᴜʟᴛɪᴍᴀᴛᴇ ᴄᴏʟʟᴇᴄᴛᴏʀs! 🎴"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("✦ ᴀᴅᴅ ᴍᴇ ʙᴀʙʏ", fmt.Sprintf("http://t.me/%s?startgroup=new", b.Config.BotUsername)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("✧ sᴜᴘᴘᴏʀᴛ", fmt.Sprintf("https://t.me/%s", b.Config.SupportChat)),
			tgbotapi.NewInlineKeyboardButtonURL("✧ ᴜᴘᴅᴀᴛᴇs", fmt.Sprintf("https://t.me/%s", b.Config.UpdateChat)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✦ ɢᴜɪᴅᴀɴᴄᴇ", "help"),
		),
	)
	
	if videoURL != "" {
		video := tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FileURL(videoURL))
		video.Caption = caption
		video.ReplyMarkup = keyboard
		video.ParseMode = "HTML"
		b.API.Send(video)
	} else {
		reply := tgbotapi.NewMessage(msg.Chat.ID, caption)
		reply.ReplyMarkup = keyboard
		reply.ParseMode = "HTML"
		b.API.Send(reply)
	}
	
	// Send notification to group if new user
	if msg.Chat.Type == "private" {
		count, _ := b.GroupService.GetPMUsersCount()
		notifText := fmt.Sprintf(
			"#ʙᴏᴛsᴛᴀʀᴛ\n\nʙᴏᴛ sᴛᴀʀᴛᴇᴅ\n\nɴᴀᴍᴇ : <a href='tg://user?id=%d'>%s</a>\nɪᴅ : <code>%d</code>\nᴜsᴇʀɴᴀᴍᴇ : %s\n\nᴛᴏᴛᴀʟ ᴜsᴇʀs : %d",
			user.ID, user.FirstName, user.ID, func() string {
				if user.UserName != "" {
					return "@" + user.UserName
				}
				return "ɴᴏ ᴜsᴇʀɴᴀᴍᴇ"
			}(), count,
		)
		notif := tgbotapi.NewMessage(b.Config.GroupID, notifText)
		notif.ParseMode = "HTML"
		b.API.Send(notif)
	}
}

// cmdPing handles /ping command
func (b *Bot) cmdPing(msg *tgbotapi.Message) {
	// Check if user is sudo
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ ᴛʜɪs ᴄᴏᴍᴍᴀɴᴅ ɪs ʀᴇsᴛʀɪᴄᴛᴇᴅ ᴛᴏ sᴜᴅᴏ ᴜsᴇʀs ᴏɴʟʏ.")
		b.API.Send(reply)
		return
	}
	
	start := time.Now()
	sentMsg, _ := b.API.Send(tgbotapi.NewMessage(msg.Chat.ID, "🏓 ᴘᴏɴɢ!"))
	latency := time.Since(start).Milliseconds()
	
	status := "ғᴀɪʀ"
	if latency < 100 {
		status = "ᴇxᴄᴇʟʟᴇɴᴛ"
	} else if latency < 300 {
		status = "ɢᴏᴏᴅ"
	}
	
	edit := tgbotapi.NewEditMessageText(
		msg.Chat.ID,
		sentMsg.MessageID,
		fmt.Sprintf("🏓 **ᴘᴏɴɢ!**\n📊 ʟᴀᴛᴇɴᴄʏ: `%dᴍs`\n⚡ sᴛᴀᴛᴜs: %s", latency, status),
	)
	edit.ParseMode = "Markdown"
	b.API.Send(edit)
}

// cmdGuess handles /guess command
func (b *Bot) cmdGuess(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	
	// Check if there's a character to guess
	lastChar, exists := b.LastCharacters[chatID]
	if !exists {
		return
	}
	
	// Check if already guessed
	if _, guessed := b.FirstCorrectGuesses[chatID]; guessed {
		reply := tgbotapi.NewMessage(chatID, utils.ToSmallCaps("❌ Already guessed by someone. Try next time."))
		b.API.Send(reply)
		return
	}
	
	// Get guess text
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(chatID, "Please provide a guess, e.g. /guess Alice")
		b.API.Send(reply)
		return
	}
	
	guessText := strings.ToLower(strings.Join(args[1:], " "))
	
	// Check for invalid characters
	if strings.Contains(guessText, "()") || strings.Contains(guessText, "&") {
		reply := tgbotapi.NewMessage(chatID, utils.ToSmallCaps("You can't use these characters in your guess."))
		b.API.Send(reply)
		return
	}
	
	// Check guess
	nameParts := strings.Fields(strings.ToLower(lastChar.Name))
	guessParts := strings.Fields(guessText)
	
	correct := false
	if strings.EqualFold(strings.Join(nameParts, " "), strings.Join(guessParts, " ")) {
		correct = true
	} else {
		for _, part := range nameParts {
			if part == guessText {
				correct = true
				break
			}
		}
	}
	
	if correct {
		// Mark as guessed
		b.FirstCorrectGuesses[chatID] = userID
		
		// Update user info
		user, _ := b.UserService.GetOrCreateUser(userID, msg.From.UserName, msg.From.FirstName)
		_ = user
		
		// Add balance
		_, _ = b.UserService.UpdateUserBalance(userID, 100)
		
		// Add character to user
		userChar := models.UserCharacter{
			ID:     lastChar.CharacterID,
			Name:   lastChar.Name,
			Anime:  lastChar.Anime,
			Rarity: lastChar.Rarity,
			ImgURL: lastChar.ImgURL,
		}
		_ = b.UserService.AddCharacterToUser(userID, userChar)
		
		// Update group stats
		_ = b.GroupService.UpdateGroupUserTotal(userID, chatID, msg.From.UserName, msg.From.FirstName)
		_ = b.GroupService.UpdateTopGlobalGroup(chatID, msg.Chat.Title)
		
		// Update daily stats
		_ = b.DailyService.UpdateDailyUserGuess(userID, msg.From.UserName, msg.From.FirstName)
		if msg.Chat.Type == "group" || msg.Chat.Type == "supergroup" {
			_ = b.DailyService.UpdateDailyGroupGuess(chatID, msg.Chat.Title)
		}
		
		// Send congratulations
		coinMsg := tgbotapi.NewMessage(chatID, utils.ToSmallCaps("✨ ᴄᴏɴɢʀᴀᴛᴜʟᴀᴛɪᴏɴꜱ 🎉  ʏᴏᴜ ɢᴜᴇꜱꜱᴇᴅ ɪᴛ ʀɪɢʜᴛ! ᴀꜱ ᴀ ʀᴇᴡᴀʀᴅ, 100 ᴄᴏɪɴꜱ ʜᴀᴠᴇ ʙᴇᴇɴ ᴀᴅᴅᴇᴅ ᴛᴏ ʏᴏᴜʀ ʙᴀʟᴀɴᴄᴇ.."))
		coinMsg.ParseMode = "HTML"
		sentCoinMsg, _ := b.API.Send(coinMsg)
		
		// Set reaction if possible (requires additional API call)
		_ = sentCoinMsg
		
		// Send character details
		rarityDisplay := utils.GetRarityDisplay(lastChar.Rarity)
		detailsText := fmt.Sprintf(
			"✨ ᴄᴏɴɢʀᴀᴛᴜʟᴀᴛɪᴏɴꜱ 🎊 %s ᴛʜɪꜱ ᴄʜᴀʀᴀᴄᴛᴇʀ ʜᴀꜱ ʙᴇᴇɴ ᴀᴅᴅᴇᴅ ᴛᴏ ʏᴏᴜʀ.\n\n"+
				"👤 ɴᴀᴍᴇ: %s\n"+
				"🎬 ᴀɴɪᴍᴇ: %s\n"+
				"✨ ʀᴀʀɪᴛʏ: %s\n"+
				"🆔 ɪᴅ: %s\n\n"+
				"✅ ꜱᴜᴄᴄᴇꜱꜱ ꜰᴜʟʟ ᴀᴅᴅ ʜᴀʀᴇᴍ.",
			msg.From.FirstName,
			lastChar.Name,
			lastChar.Anime,
			rarityDisplay,
			lastChar.CharacterID,
		)
		
		detailsMsg := tgbotapi.NewMessage(chatID, utils.ToSmallCaps(detailsText))
		detailsMsg.ParseMode = "HTML"
		detailsMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonSwitch("ꜱᴇᴇ ʜᴀʀᴇᴍ", fmt.Sprintf("collection.%d", userID)),
			),
		)
		b.API.Send(detailsMsg)
	} else {
		reply := tgbotapi.NewMessage(chatID, utils.ToSmallCaps("Please write the correct character name. ❌"))
		b.API.Send(reply)
	}
}

// cmdBalance handles /balance command
func (b *Bot) cmdBalance(msg *tgbotapi.Message) {
	targetID := msg.From.ID
	targetName := msg.From.FirstName
	
	// Check if replying to someone
	if msg.ReplyToMessage != nil {
		targetID = msg.ReplyToMessage.From.ID
		targetName = msg.ReplyToMessage.From.FirstName
	} else {
		// Check for username or ID argument
		args := strings.Fields(msg.Text)
		if len(args) > 1 {
			arg := args[1]
			if strings.HasPrefix(arg, "@") {
				// Try to get user by username - this would require storing usernames
				// For now, just show own balance
			} else if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
				targetID = id
				targetName = fmt.Sprintf("User %d", id)
			}
		}
	}
	
	balance, _ := b.UserService.GetUserBalance(targetID)
	
	replyText := fmt.Sprintf("💰 <b>%s</b>'s %s: <b>%s</b> ᴄᴏɪɴs",
		targetName,
		utils.ToSmallCaps("Balance"),
		utils.FormatNumber(balance),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdFav handles /fav command
func (b *Bot) cmdFav(msg *tgbotapi.Message) {
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("Please provide a character id: /fav <id>"))
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	userID := msg.From.ID
	
	// Check if user has this character
	hasChar, _ := b.UserService.HasCharacter(userID, charID)
	if !hasChar {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("That character is not in your collection."))
		b.API.Send(reply)
		return
	}
	
	// Add to favorites
	_ = b.UserService.AddToFavorites(userID, charID)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps(fmt.Sprintf("Character has been added to your favorites.")))
	b.API.Send(reply)
}

// cmdAddBal handles /addbal command (admin only)
func (b *Bot) cmdAddBal(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("✘ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ."))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 3 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("Usage: /addbal <user_id> <amount>"))
		b.API.Send(reply)
		return
	}
	
	targetID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("✘ ɪɴᴠᴀʟɪᴅ ᴜsᴇʀ ɪᴅ."))
		b.API.Send(reply)
		return
	}
	
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("✘ ɪɴᴠᴀʟɪᴅ ᴀᴍᴏᴜɴᴛ."))
		b.API.Send(reply)
		return
	}
	
	newBalance, _ := b.UserService.UpdateUserBalance(targetID, amount)
	
	replyText := fmt.Sprintf("✓ ᴜᴘᴅᴀᴛᴇᴅ ʙᴀʟᴀɴᴄᴇ ғᴏʀ <a href='tg://user?id=%d'>ᴜsᴇʀ</a>: <b>%s</b>",
		targetID,
		utils.FormatNumber(newBalance),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}
