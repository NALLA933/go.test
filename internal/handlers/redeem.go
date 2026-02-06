package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/models"
	"senpai-waifu-bot/internal/utils"
)

// cmdRedeem handles /redeem command
func (b *Bot) cmdRedeem(msg *tgbotapi.Message) {
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🎁 %s</b>\n\n%s <code>/redeem &lt;code&gt;</code>\n\n%s",
				utils.ToSmallCaps("REDEEM CODE"),
				utils.ToSmallCaps("Usage:"),
				utils.ToSmallCaps("Redeem a code to get coins or characters!")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	code := strings.ToLower(args[1])
	userID := msg.From.ID
	
	// Get redeem code
	redeemCode, err := b.RedeemService.GetRedeemCode(code)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s\n\n%s",
				utils.ToSmallCaps("Invalid or expired code!"),
				utils.ToSmallCaps("Please check the code and try again.")))
		b.API.Send(reply)
		return
	}
	
	// Check if user already redeemed
	alreadyRedeemed, _ := b.RedeemService.HasUserRedeemed(code, userID)
	if alreadyRedeemed {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s\n\n%s",
				utils.ToSmallCaps("You have already redeemed this code!"),
				utils.ToSmallCaps("Each code can only be used once per user.")))
		b.API.Send(reply)
		return
	}
	
	// Check if max uses reached
	if len(redeemCode.UsedBy) >= redeemCode.MaxUses {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s\n\n%s",
				utils.ToSmallCaps("This code has reached its maximum uses!"),
				utils.ToSmallCaps("The code is no longer valid.")))
		b.API.Send(reply)
		return
	}
	
	// Redeem the code
	redeemCode, err = b.RedeemService.RedeemCode(code, userID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s\n\n%s",
				utils.ToSmallCaps("Failed to redeem code!"),
				utils.ToSmallCaps("Please try again later.")))
		b.API.Send(reply)
		return
	}
	
	// Process reward
	var rewardMsg string
	switch redeemCode.Type {
	case "coin":
		newBalance, _ := b.UserService.UpdateUserBalance(userID, redeemCode.Amount)
		rewardMsg = fmt.Sprintf(
			"<b>✅ %s</b>\n\n"+
				"💰 <b>%s</b> %s %s\n"+
				"💎 <b>%s</b> %s %s",
			utils.ToSmallCaps("CODE REDEEMED!"),
			utils.ToSmallCaps("You received"), utils.FormatNumber(redeemCode.Amount), utils.ToSmallCaps("coins"),
			utils.ToSmallCaps("New balance:"), utils.FormatNumber(newBalance), utils.ToSmallCaps("coins"),
		)
		
	case "character":
		char, err := b.CharacterService.GetCharacterByID(redeemCode.CharacterID)
		if err != nil {
			rewardMsg = fmt.Sprintf(
				"<b>❌ %s</b>\n\n%s",
				utils.ToSmallCaps("ERROR"),
				utils.ToSmallCaps("Character not found in database!"))
		} else {
			userChar := models.UserCharacter{
				ID:     char.ID,
				Name:   char.Name,
				Anime:  char.Anime,
				Rarity: char.Rarity,
				ImgURL: char.ImgURL,
			}
			_ = b.UserService.AddCharacterToUser(userID, userChar)
			rarityDisplay := utils.GetRarityDisplay(char.Rarity)
			rewardMsg = fmt.Sprintf(
				"<b>✅ %s</b>\n\n"+
					"🎴 <b>%s</b> %s\n"+
					"📺 <b>%s</b> %s\n"+
					"⭐ <b>%s</b> %s\n\n"+
					"✨ %s",
				utils.ToSmallCaps("CODE REDEEMED!"),
				utils.ToSmallCaps("You received"), utils.ToSmallCaps(char.Name),
				utils.ToSmallCaps("Anime:"), utils.ToSmallCaps(char.Anime),
				utils.ToSmallCaps("Rarity:"), rarityDisplay,
				utils.ToSmallCaps("Character has been added to your collection!"),
			)
		}
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, rewardMsg)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdGen handles /gen command (admin - generate coin code)
func (b *Bot) cmdGen(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🎫 %s</b>\n\n%s <code>/gen &lt;amount&gt; [max_uses]</code>",
				utils.ToSmallCaps("GENERATE COIN CODE"),
				utils.ToSmallCaps("Usage:")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	amount, err := utils.ParseInt64(args[1])
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("❌ Invalid amount!"))
		b.API.Send(reply)
		return
	}
	
	maxUses := 1
	if len(args) >= 3 {
		maxUses, _ = utils.ParseInt(args[2])
		if maxUses < 1 {
			maxUses = 1
		}
	}
	
	code, err := b.RedeemService.CreateCoinCode(amount, maxUses, msg.From.ID)
	if err != nil || code == "" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("❌ Failed to generate code!"))
		b.API.Send(reply)
		return
	}
	
	message := fmt.Sprintf(
		"<b>✅ %s</b>\n\n"+
			"🎟️ <b>%s</b> <code>%s</code>\n"+
			"💰 <b>%s</b> %s %s\n"+
			"👥 <b>%s</b> %d\n\n"+
			"📌 %s",
		utils.ToSmallCaps("COIN CODE GENERATED!"),
		utils.ToSmallCaps("Code:"), code,
		utils.ToSmallCaps("Amount:"), utils.FormatNumber(amount), utils.ToSmallCaps("coins"),
		utils.ToSmallCaps("Max Uses:"), maxUses,
		utils.ToSmallCaps("Share this code with users to redeem!"),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdSGen handles /sgen command (admin - generate character code)
func (b *Bot) cmdSGen(msg *tgbotapi.Message) {
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("⚠️ You are not authorized!"))
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("<b>🎴 %s</b>\n\n%s <code>/sgen &lt;character_id&gt; [max_uses]</code>",
				utils.ToSmallCaps("GENERATE CHARACTER CODE"),
				utils.ToSmallCaps("Usage:")))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	
	// Verify character exists
	char, err := b.CharacterService.GetCharacterByID(charID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			fmt.Sprintf("❌ %s <code>%s</code>",
				utils.ToSmallCaps("Character not found with ID:"), charID))
		b.API.Send(reply)
		return
	}
	
	maxUses := 1
	if len(args) >= 3 {
		maxUses, _ = utils.ParseInt(args[2])
		if maxUses < 1 {
			maxUses = 1
		}
	}
	
	code, err := b.RedeemService.CreateCharacterCode(charID, maxUses, msg.From.ID)
	if err != nil || code == "" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("❌ Failed to generate code!"))
		b.API.Send(reply)
		return
	}
	
	rarityDisplay := utils.GetRarityDisplay(char.Rarity)
	message := fmt.Sprintf(
		"<b>✅ %s</b>\n\n"+
			"🎟️ <b>%s</b> <code>%s</code>\n"+
			"🎴 <b>%s</b> %s\n"+
			"📺 <b>%s</b> %s\n"+
			"⭐ <b>%s</b> %s\n"+
			"👥 <b>%s</b> %d\n\n"+
			"📌 %s",
		utils.ToSmallCaps("CHARACTER CODE GENERATED!"),
		utils.ToSmallCaps("Code:"), code,
		utils.ToSmallCaps("Character:"), utils.ToSmallCaps(char.Name),
		utils.ToSmallCaps("Anime:"), utils.ToSmallCaps(char.Anime),
		utils.ToSmallCaps("Rarity:"), rarityDisplay,
		utils.ToSmallCaps("Max Uses:"), maxUses,
		utils.ToSmallCaps("Share this code with users to redeem!"),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, message)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// Helper functions for parsing
func (u *utils) ParseInt64(s string) (int64, error) {
	return utils.ParseInt64(s)
}

func (u *utils) ParseInt(s string) (int, error) {
	return utils.ParseInt(s)
}
