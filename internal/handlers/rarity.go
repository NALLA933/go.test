package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/utils"
)

// cmdSetOn handles /set_on command (enable rarity)
func (b *Bot) cmdSetOn(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>✅ %s</b>\n\n%s <code>/set_on &lt;rarity_number&gt;</code>\n\n%s",
				utils.ToSmallCaps("ENABLE RARITY"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Enable a rarity for spawning in this chat.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	rarity, err := strconv.Atoi(args[1])
	if err != nil || rarity < 1 || rarity > 15 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			utils.ToSmallCaps("❌ Invalid rarity! Please use a number between 1 and 15."))
		b.API.Send(reply)
		return
	}
	
	// Enable rarity for this chat
	_ = b.RarityService.EnableRarity(msg.Chat.ID, rarity)
	
	rarityDisplay := utils.GetRarityDisplay(rarity)
	reply := tgbotapi.NewMessage(msg.Chat.ID, 
		fmt.Sprintf("✅ %s %s %s",
			utils.ToSmallCaps("Rarity"),
			rarityDisplay,
			utils.ToSmallCaps("has been enabled for this chat!")))
	b.API.Send(reply)
}

// cmdSetOff handles /set_off command (disable rarity)
func (b *Bot) cmdSetOff(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>❌ %s</b>\n\n%s <code>/set_off &lt;rarity_number&gt;</code>\n\n%s",
				utils.ToSmallCaps("DISABLE RARITY"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Disable a rarity from spawning in this chat.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	rarity, err := strconv.Atoi(args[1])
	if err != nil || rarity < 1 || rarity > 15 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			utils.ToSmallCaps("❌ Invalid rarity! Please use a number between 1 and 15."))
		b.API.Send(reply)
		return
	}
	
	// Disable rarity for this chat
	_ = b.RarityService.DisableRarity(msg.Chat.ID, rarity)
	
	rarityDisplay := utils.GetRarityDisplay(rarity)
	reply := tgbotapi.NewMessage(msg.Chat.ID, 
		fmt.Sprintf("❌ %s %s %s",
			utils.ToSmallCaps("Rarity"),
			rarityDisplay,
			utils.ToSmallCaps("has been disabled for this chat!")))
	b.API.Send(reply)
}

// cmdLock handles /lock command (lock character from spawning)
func (b *Bot) cmdLock(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🔒 %s</b>\n\n%s <code>/lock &lt;character_id&gt; [reason]</code>\n\n%s",
				utils.ToSmallCaps("LOCK CHARACTER"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Lock a character from spawning.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	reason := "No reason provided"
	if len(args) >= 3 {
		reason = strings.Join(args[2:], " ")
	}
	
	// Get character
	char, err := b.CharacterService.GetCharacterByID(charID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s <code>%s</code>",
				utils.ToSmallCaps("Character not found with ID:"), charID))
		b.API.Send(reply)
		return
	}
	
	// Lock character
	_ = b.RarityService.LockCharacter(charID, char.Name, msg.From.ID, msg.From.FirstName, reason)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, 
		fmt.Sprintf("🔒 <b>%s</b> %s\n\n%s <code>%s</code>\n%s %s",
			utils.ToSmallCaps("Character Locked"),
			utils.ToSmallCaps(char.Name),
			utils.ToSmallCaps("ID:"),
			charID,
			utils.ToSmallCaps("Reason:"),
			reason))
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdUnlock handles /unlock command (unlock character)
func (b *Bot) cmdUnlock(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🔓 %s</b>\n\n%s <code>/unlock &lt;character_id&gt;</code>\n\n%s",
				utils.ToSmallCaps("UNLOCK CHARACTER"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Unlock a character to allow spawning.")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	
	// Check if locked
	isLocked, _ := b.RarityService.IsCharacterLocked(charID)
	if !isLocked {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s <code>%s</code> %s",
				utils.ToSmallCaps("Character"), charID, utils.ToSmallCaps("is not locked!")))
		b.API.Send(reply)
		return
	}
	
	// Unlock character
	_ = b.RarityService.UnlockCharacter(charID)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, 
		fmt.Sprintf("🔓 <b>%s</b> %s",
			utils.ToSmallCaps("Character unlocked:"),
			charID))
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdLockList handles /locklist command (show locked characters)
func (b *Bot) cmdLockList(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	// Get locked characters
	locked, err := b.RarityService.GetLockedCharacters()
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ Could not retrieve locked characters!"))
		b.API.Send(reply)
		return
	}
	
	if len(locked) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("📋 No characters are currently locked."))
		b.API.Send(reply)
		return
	}
	
	// Build message
	message := fmt.Sprintf("<b>🔒 %s</b>\n\n", utils.ToSmallCaps("LOCKED CHARACTERS"))
	
	for _, lock := range locked {
		message += fmt.Sprintf("🆔 <code>%s</code> - %s\n", lock.CharacterID, utils.ToSmallCaps(lock.CharacterName))
		message += fmt.Sprintf("   %s %s\n", utils.ToSmallCaps("Locked by:"), lock.LockedByName)
		message += fmt.Sprintf("   %s %s\n\n", utils.ToSmallCaps("Reason:"), lock.Reason)
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}
