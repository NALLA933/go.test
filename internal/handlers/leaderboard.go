package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/utils"
)

// cmdLeaderboard handles /leaderboard command
func (b *Bot) cmdLeaderboard(msg *tgbotapi.Message) {
	args := strings.Fields(msg.Text)
	
	// Default to global
	boardType := "global"
	if len(args) > 1 {
		boardType = strings.ToLower(args[1])
	}
	
	switch boardType {
	case "global":
		b.showGlobalLeaderboard(msg)
	case "daily":
		b.showDailyLeaderboard(msg)
	case "group":
		b.showGroupLeaderboard(msg)
	case "balance", "bal":
		b.showBalanceLeaderboard(msg)
	default:
		// Show usage
		usage := fmt.Sprintf(
			"<b>📊 %s</b>\n\n"+
				"🌐 <code>/leaderboard global</code> - %s\n"+
				"📅 <code>/leaderboard daily</code> - %s\n"+
				"👥 <code>/leaderboard group</code> - %s\n"+
				"💰 <code>/leaderboard balance</code> - %s",
			utils.ToSmallCaps("LEADERBOARD COMMANDS"),
			utils.ToSmallCaps("Global character rankings"),
			utils.ToSmallCaps("Daily character rankings"),
			utils.ToSmallCaps("Group character rankings"),
			utils.ToSmallCaps("Balance rankings"),
		)
		reply := tgbotapi.NewMessage(msg.Chat.ID, usage)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
	}
}

// showGlobalLeaderboard shows global leaderboard
func (b *Bot) showGlobalLeaderboard(msg *tgbotapi.Message) {
	// Get top users by character count
	users, err := b.UserService.GetTopUsersByCharacters(10)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Could not retrieve leaderboard data."))
		b.API.Send(reply)
		return
	}
	
	if len(users) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("📊 No data available yet!"))
		b.API.Send(reply)
		return
	}
	
	// Build message
	message := fmt.Sprintf("<b>🏆 %s</b>\n\n", utils.ToSmallCaps("GLOBAL LEADERBOARD"))
	
	for i, user := range users {
		charCount := len(user.Characters)
		if charCount == 0 {
			continue
		}
		
		name := user.FirstName
		if user.Username != "" {
			name = fmt.Sprintf("@%s", user.Username)
		}
		
		emoji := "🥉"
		if i == 0 {
			emoji = "🥇"
		} else if i == 1 {
			emoji = "🥈"
		}
		
		message += fmt.Sprintf("%s <b>%d.</b> %s - <b>%d</b> %s\n", 
			emoji, i+1, name, charCount, utils.ToSmallCaps("characters"))
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// showDailyLeaderboard shows daily leaderboard
func (b *Bot) showDailyLeaderboard(msg *tgbotapi.Message) {
	// Get top users by daily guesses
	guesses, err := b.DailyService.GetTopDailyUsers(10)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Could not retrieve daily leaderboard data."))
		b.API.Send(reply)
		return
	}
	
	if len(guesses) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("📊 No daily data available yet!"))
		b.API.Send(reply)
		return
	}
	
	// Build message
	message := fmt.Sprintf("<b>📅 %s</b>\n\n", utils.ToSmallCaps("DAILY LEADERBOARD"))
	
	for i, guess := range guesses {
		name := guess.FirstName
		if guess.Username != "" {
			name = fmt.Sprintf("@%s", guess.Username)
		}
		
		emoji := "🥉"
		if i == 0 {
			emoji = "🥇"
		} else if i == 1 {
			emoji = "🥈"
		}
		
		message += fmt.Sprintf("%s <b>%d.</b> %s - <b>%d</b> %s\n", 
			emoji, i+1, name, guess.Count, utils.ToSmallCaps("guesses"))
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// showGroupLeaderboard shows group leaderboard
func (b *Bot) showGroupLeaderboard(msg *tgbotapi.Message) {
	if msg.Chat.Type == "private" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			utils.ToSmallCaps("⚠️ This command can only be used in groups!"))
		b.API.Send(reply)
		return
	}
	
	// Get top groups
	groups, err := b.GroupService.GetTopGroups(10)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Could not retrieve group leaderboard data."))
		b.API.Send(reply)
		return
	}
	
	if len(groups) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("📊 No group data available yet!"))
		b.API.Send(reply)
		return
	}
	
	// Build message
	message := fmt.Sprintf("<b>👥 %s</b>\n\n", utils.ToSmallCaps("GROUP LEADERBOARD"))
	
	for i, group := range groups {
		emoji := "🥉"
		if i == 0 {
			emoji = "🥇"
		} else if i == 1 {
			emoji = "🥈"
		}
		
		message += fmt.Sprintf("%s <b>%d.</b> %s - <b>%d</b> %s\n", 
			emoji, i+1, group.GroupName, group.Count, utils.ToSmallCaps("guesses"))
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// showBalanceLeaderboard shows balance leaderboard
func (b *Bot) showBalanceLeaderboard(msg *tgbotapi.Message) {
	// Get top users by balance
	users, err := b.UserService.GetTopUsersByBalance(10)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Could not retrieve balance leaderboard data."))
		b.API.Send(reply)
		return
	}
	
	if len(users) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("📊 No balance data available yet!"))
		b.API.Send(reply)
		return
	}
	
	// Build message
	message := fmt.Sprintf("<b>💰 %s</b>\n\n", utils.ToSmallCaps("BALANCE LEADERBOARD"))
	
	for i, user := range users {
		name := user.FirstName
		if user.Username != "" {
			name = fmt.Sprintf("@%s", user.Username)
		}
		
		emoji := "🥉"
		if i == 0 {
			emoji = "🥇"
		} else if i == 1 {
			emoji = "🥈"
		}
		
		message += fmt.Sprintf("%s <b>%d.</b> %s - <b>%s</b> %s\n", 
			emoji, i+1, name, utils.FormatNumber(user.Balance), utils.ToSmallCaps("coins"))
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}
