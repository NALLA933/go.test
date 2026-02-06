package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/utils"
)

// cmdPay handles /pay command
func (b *Bot) cmdPay(msg *tgbotapi.Message) {
	senderID := msg.From.ID
	
	// Check cooldown
	if nextAllowed, ok := b.PaymentCooldowns[senderID]; ok && time.Now().Before(nextAllowed) {
		remaining := int(time.Until(nextAllowed).Seconds())
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			utils.ToSmallCaps(fmt.Sprintf("⏱️ ʏᴏᴜ ᴍᴜsᴛ ᴡᴀɪᴛ %ds ʙᴇғᴏʀᴇ sᴛᴀʀᴛɪɴɢ ᴀɴᴏᴛʜᴇʀ ᴘᴀʏᴍᴇɴᴛ.", remaining)))
		b.API.Send(reply)
		return
	}
	
	var targetID int64
	var amount int64
	
	// Parse arguments
	args := strings.Fields(msg.Text)
	
	if msg.ReplyToMessage != nil && len(args) == 2 {
		// Reply to message with amount
		targetID = msg.ReplyToMessage.From.ID
		amount, _ = strconv.ParseInt(args[1], 10, 64)
	} else if len(args) >= 3 {
		// Username/ID and amount
		rawTarget := args[1]
		if strings.HasPrefix(rawTarget, "@") {
			// For now, we don't have username lookup implemented
			// Would need to store username -> ID mapping
			reply := tgbotapi.NewMessage(msg.Chat.ID, 
				utils.ToSmallCaps("✘ ᴄᴏᴜʟᴅ ɴᴏᴛ ʀᴇsᴏʟᴠᴇ ᴛᴀʀɢᴇᴛ ᴜsᴇʀ. ᴜsᴇ ᴜsᴇʀ ɪᴅ ᴏʀ ʀᴇᴘʟʏ ᴛᴏ ᴛʜᴇɪʀ ᴍᴇssᴀɢᴇ."))
			b.API.Send(reply)
			return
		}
		targetID, _ = strconv.ParseInt(rawTarget, 10, 64)
		amount, _ = strconv.ParseInt(args[2], 10, 64)
	} else {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("Usage: /pay <amount> (reply to user) or /pay <user_id> <amount>"))
		b.API.Send(reply)
		return
	}
	
	if targetID == 0 || amount <= 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("✘ ɪɴᴠᴀʟɪᴅ ᴀʀɢᴜᴍᴇɴᴛs."))
		b.API.Send(reply)
		return
	}
	
	if targetID == senderID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, utils.ToSmallCaps("✓ ʏᴏᴜ ᴄᴀɴɴᴏᴛ ᴘᴀʏ ʏᴏᴜʀsᴇʟғ."))
		b.API.Send(reply)
		return
	}
	
	// Check sender balance
	balance, _ := b.UserService.GetUserBalance(senderID)
	if balance < amount {
		reply := tgbotapi.NewMessage(msg.Chat.ID, 
			utils.ToSmallCaps(fmt.Sprintf("✘ ʏᴏᴜ ᴅᴏɴ'ᴛ ʜᴀᴠᴇ ᴇɴᴏᴜɢʜ ᴄᴏɪɴs. ʏᴏᴜʀ ʙᴀʟᴀɴᴄᴇ: %s", utils.FormatNumber(balance))))
		b.API.Send(reply)
		return
	}
	
	// Generate token
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	
	// Store pending payment
	b.PendingPayments[token] = &PendingPaymentInfo{
		Token:     token,
		SenderID:  senderID,
		TargetID:  targetID,
		Amount:    amount,
		CreatedAt: time.Now(),
		ChatID:    msg.Chat.ID,
	}
	
	// Get target info
	targetChat, err := b.API.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: targetID}})
	targetName := fmt.Sprintf("User %d", targetID)
	if err == nil {
		targetName = targetChat.FirstName
	}
	
	// Send confirmation message
	text := fmt.Sprintf(
		"❗ <b>ᴘᴀʏᴍᴇɴᴛ ᴄᴏɴғɪʀᴍᴀᴛɪᴏɴ</b>\n\n"+
			"sᴇɴᴅᴇʀ: <a href='tg://user?id=%d'>%s</a>\n"+
			"ʀᴇᴄɪᴘɪᴇɴᴛ: <a href='tg://user?id=%d'>%s</a>\n"+
			"ᴀᴍᴏᴜɴᴛ: <b>%s</b> ᴄᴏɪɴs\n\n"+
			"ᴀʀᴇ ʏᴏᴜ sᴜʀᴇ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ ᴘʀᴏᴄᴇᴇᴅ?",
		senderID, msg.From.FirstName,
		targetID, targetName,
		utils.FormatNumber(amount),
	)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✓ ᴄᴏɴғɪʀᴍ", fmt.Sprintf("pay_confirm:%s", token)),
			tgbotapi.NewInlineKeyboardButtonData("✘ ᴄᴀɴᴄᴇʟ", fmt.Sprintf("pay_cancel:%s", token)),
		),
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "HTML"
	reply.ReplyMarkup = keyboard
	sentMsg, _ := b.API.Send(reply)
	
	// Store message ID for editing later
	if payment, ok := b.PendingPayments[token]; ok {
		payment.MessageID = sentMsg.MessageID
	}
}

// confirmPayment confirms a payment
func (b *Bot) confirmPayment(chatID int64, messageID int, token string, userID int64) {
	payment, ok := b.PendingPayments[token]
	if !ok {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			utils.ToSmallCaps("✖️ ᴛʜɪs ᴘᴀʏᴍᴇɴᴛ ʀᴇǫᴜᴇsᴛ ʜᴀs ᴇxᴘɪʀᴇᴅ ᴏʀ ɪs ɪɴᴠᴀʟɪᴅ."))
		b.API.Send(edit)
		return
	}
	
	// Only sender can confirm
	if userID != payment.SenderID {
		b.API.Send(tgbotapi.NewCallback(tgbotapi.CallbackConfig{Text: "ᴏɴʟʏ ᴛʜᴇ ᴘᴀʏᴍᴇɴᴛ ɪɴɪᴛɪᴀᴛᴏʀ ᴄᴀɴ ᴄᴏɴғɪʀᴍ ᴏʀ ᴄᴀɴᴄᴇʟ ᴛʜɪs ᴘᴀʏᴍᴇɴᴛ.", ShowAlert: true}))
		return
	}
	
	// Check expiry
	if time.Since(payment.CreatedAt) > 5*time.Minute {
		delete(b.PendingPayments, token)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			utils.ToSmallCaps("⏱️ ᴛʜɪs ᴘᴀʏᴍᴇɴᴛ ʀᴇǫᴜᴇsᴛ ʜᴀs ᴇxᴘɪʀᴇᴅ."))
		b.API.Send(edit)
		return
	}
	
	// Check cooldown again
	if nextAllowed, ok := b.PaymentCooldowns[payment.SenderID]; ok && time.Now().Before(nextAllowed) {
		remaining := int(time.Until(nextAllowed).Seconds())
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			utils.ToSmallCaps(fmt.Sprintf("⏱️ ʏᴏᴜ ᴍᴜsᴛ ᴡᴀɪᴛ %ds ʙᴇғᴏʀᴇ ᴍᴀᴋɪɴɢ ᴀɴᴏᴛʜᴇʀ ᴘᴀʏᴍᴇɴᴛ.", remaining)))
		b.API.Send(edit)
		delete(b.PendingPayments, token)
		return
	}
	
	// Perform transfer
	// First deduct from sender
	senderBalance, err := b.UserService.UpdateUserBalance(payment.SenderID, -payment.Amount)
	if err != nil || senderBalance < 0 {
		// Rollback - add back to sender
		_, _ = b.UserService.UpdateUserBalance(payment.SenderID, payment.Amount)
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			utils.ToSmallCaps("✘ ᴛʀᴀɴsᴀᴄᴛɪᴏɴ ғᴀɪʟᴇᴅ: ɪɴsᴜғғɪᴄɪᴇɴᴛ ғᴜɴᴅs ᴏʀ ɪɴᴛᴇʀɴᴀʟ ᴇʀʀᴏʀ."))
		b.API.Send(edit)
		delete(b.PendingPayments, token)
		return
	}
	
	// Add to receiver
	_, _ = b.UserService.UpdateUserBalance(payment.TargetID, payment.Amount)
	
	// Set cooldown
	b.PaymentCooldowns[payment.SenderID] = time.Now().Add(60 * time.Second)
	
	// Get names
	senderChat, _ := b.API.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: payment.SenderID}})
	targetChat, _ := b.API.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: payment.TargetID}})
	
	senderName := fmt.Sprintf("User %d", payment.SenderID)
	targetName := fmt.Sprintf("User %d", payment.TargetID)
	
	if senderChat != nil {
		senderName = senderChat.FirstName
	}
	if targetChat != nil {
		targetName = targetChat.FirstName
	}
	
	// Edit message to show success
	text := fmt.Sprintf(
		"✓ <b>ᴘᴀʏᴍᴇɴᴛ sᴜᴄᴄᴇssғᴜʟ</b>\n\n"+
			"ꜱᴇɴᴅᴇʀ: <a href='tg://user?id=%d'>%s</a>\n"+
			"ʀᴇᴄɪᴘɪᴇɴᴛ: <a href='tg://user?id=%d'>%s</a>\n"+
			"ᴀᴍᴏᴜɴᴛ: <b>%s</b> ᴄᴏɪɴs\n\n"+
			"ɴᴇxᴛ ᴘᴀʏᴍᴇɴᴛ ᴀʟʟᴏᴡᴇᴅ ᴀғᴛᴇʀ 60 ꜱᴇᴄᴏɴᴅꜱ.",
		payment.SenderID, senderName,
		payment.TargetID, targetName,
		utils.FormatNumber(payment.Amount),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "HTML"
	b.API.Send(edit)
	
	delete(b.PendingPayments, token)
}

// cancelPayment cancels a payment
func (b *Bot) cancelPayment(chatID int64, messageID int, token string, userID int64) {
	payment, ok := b.PendingPayments[token]
	if !ok {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, 
			utils.ToSmallCaps("✖️ ᴛʜɪs ᴘᴀʏᴍᴇɴᴛ ʀᴇǫᴜᴇsᴛ ʜᴀs ᴇxᴘɪʀᴇᴅ ᴏʀ ɪs ɪɴᴠᴀʟɪᴅ."))
		b.API.Send(edit)
		return
	}
	
	// Only sender can cancel
	if userID != payment.SenderID {
		b.API.Send(tgbotapi.NewCallback(tgbotapi.CallbackConfig{Text: "ᴏɴʟʏ ᴛʜᴇ ᴘᴀʏᴍᴇɴᴛ ɪɴɪᴛɪᴀᴛᴏʀ ᴄᴀɴ ᴄᴏɴғɪʀᴍ ᴏʀ ᴄᴀɴᴄᴇʟ ᴛʜɪs ᴘᴀʏᴍᴇɴᴛ.", ShowAlert: true}))
		return
	}
	
	delete(b.PendingPayments, token)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, utils.ToSmallCaps("✘ ᴘᴀʏᴍᴇɴᴛ ᴄᴀɴᴄᴇʟʟᴇᴅ ʙʏ sᴇɴᴅᴇʀ."))
	b.API.Send(edit)
}
