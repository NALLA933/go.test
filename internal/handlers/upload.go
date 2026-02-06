package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"senpai-waifu-bot/internal/database"
	"senpai-waifu-bot/internal/utils"
)

// Rarity levels for upload
var uploadRarityNames = map[int]string{
	1:  "⚪ ᴄᴏᴍᴍᴏɴ",
	2:  "🔵 ʀᴀʀᴇ",
	3:  "🟡 ʟᴇɢᴇɴᴅᴀʀʏ",
	4:  "💮 ꜱᴘᴇᴄɪᴀʟ",
	5:  "👹 ᴀɴᴄɪᴇɴᴛ",
	6:  "🎐 ᴄᴇʟᴇꜱᴛɪᴀʟ",
	7:  "🔮 ᴇᴘɪᴄ",
	8:  "🪐 ᴄᴏꜱᴍɪᴄ",
	9:  "⚰️ ɴɪɢʜᴛᴍᴀʀᴇ",
	10: "🌬️ ꜰʀᴏꜱᴛʙᴏʀɴ",
	11: "💝 ᴠᴀʟᴇɴᴛɪɴᴇ",
	12: "🌸 ꜱᴘʀɪɴɢ",
	13: "🏖️ ᴛʀᴏᴘɪᴄᴀʟ",
	14: "🍭 ᴋᴀᴡᴀɪɪ",
	15: "🧬 ʜʏʙʀɪᴅ",
}

// Upload format text
var uploadFormatText = `❌ Wrong format!

<b>Usage:</b> Reply to an image with:
<code>/upload character-name anime-name rarity-number</code>

<b>Example:</b>
<code>/upload naruto-uzumaki naruto 3</code>

<b>Available Rarities:</b>
1 - ⚪ ᴄᴏᴍᴍᴏɴ
2 - 🔵 ʀᴀʀᴇ
3 - 🟡 ʟᴇɢᴇɴᴅᴀʀʏ
4 - 💮 ꜱᴘᴇᴄɪᴀʟ
5 - 👹 ᴀɴᴄɪᴇɴᴛ
6 - 🎐 ᴄᴇʟᴇꜱᴛɪᴀʟ
7 - 🔮 ᴇᴘɪᴄ
8 - 🪐 ᴄᴏꜱᴍɪᴄ
9 - ⚰️ ɴɪɢʜᴛᴍᴀʀᴇ
10 - 🌬️ ꜰʀᴏꜱᴛʙᴏʀɴ
11 - 💝 ᴠᴀʟᴇɴᴛɪɴᴇ
12 - 🌸 ꜱᴘʀɪɴɢ
13 - 🏖️ ᴛʀᴏᴘɪᴄᴀʟ
14 - 🍭 ᴋᴀᴡᴀɪɪ
15 - 🧬 ʜʏʙʀɪᴅ`

// ImageUploader handles image uploads to various hosting services
type ImageUploader struct {
	ImgBBKey string
	Client   *http.Client
}

// NewImageUploader creates a new image uploader
func NewImageUploader() *ImageUploader {
	return &ImageUploader{
		ImgBBKey: "6d52008ec9026912f9f50c8ca96a09c3", // Default key
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UploadToImgBB uploads image to ImgBB
func (u *ImageUploader) UploadToImgBB(imageData []byte) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	// Add image field
	fw, err := w.CreateFormFile("image", "image.jpg")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, bytes.NewReader(imageData)); err != nil {
		return "", err
	}
	
	// Add key field
	fw, _ = w.CreateFormField("key")
	fw.Write([]byte(u.ImgBBKey))
	
	w.Close()
	
	req, err := http.NewRequest("POST", "https://api.imgbb.com/1/upload", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	resp, err := u.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("imgbb upload failed: %d", resp.StatusCode)
	}
	
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	if !result.Success {
		return "", fmt.Errorf("imgbb upload failed")
	}
	
	return result.Data.URL, nil
}

// UploadToTelegraph uploads image to Telegraph
func (u *ImageUploader) UploadToTelegraph(imageData []byte) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	fw, err := w.CreateFormFile("file", "image.jpg")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, bytes.NewReader(imageData)); err != nil {
		return "", err
	}
	w.Close()
	
	req, err := http.NewRequest("POST", "https://telegra.ph/upload", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	resp, err := u.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("telegraph upload failed: %d", resp.StatusCode)
	}
	
	var result []struct {
		Src string `json:"src"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	if len(result) == 0 {
		return "", fmt.Errorf("telegraph upload failed")
	}
	
	return "https://telegra.ph" + result[0].Src, nil
}

// UploadToCatbox uploads image to Catbox
func (u *ImageUploader) UploadToCatbox(imageData []byte) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	// Add reqtype field
	fw, _ := w.CreateFormField("reqtype")
	fw.Write([]byte("fileupload"))
	
	// Add file field
	fw, err := w.CreateFormFile("fileToUpload", "image.jpg")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, bytes.NewReader(imageData)); err != nil {
		return "", err
	}
	w.Close()
	
	req, err := http.NewRequest("POST", "https://catbox.moe/user/api.php", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	resp, err := u.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	url := strings.TrimSpace(string(body))
	if !strings.HasPrefix(url, "http") {
		return "", fmt.Errorf("catbox upload failed")
	}
	
	return url, nil
}

// UploadWithFailover tries multiple upload services
func (u *ImageUploader) UploadWithFailover(imageData []byte) (string, error) {
	services := []func([]byte) (string, error){
		u.UploadToImgBB,
		u.UploadToTelegraph,
		u.UploadToCatbox,
	}
	
	var lastErr error
	for _, service := range services {
		url, err := service(imageData)
		if err == nil && url != "" {
			return url, nil
		}
		lastErr = err
	}
	
	return "", lastErr
}

// GetNextSequenceNumber gets the next character ID
func (b *Bot) GetNextSequenceNumber() (string, error) {
	// Get total count and add 1
	count, err := b.CharacterService.GetCharacterCount()
	if err != nil {
		return "", err
	}
	
	// Format as 3-digit number (001, 010, 100)
	return fmt.Sprintf("%03d", count+1), nil
}

// cmdUpload handles /upload command
func (b *Bot) cmdUpload(msg *tgbotapi.Message) {
	// Check if user is admin
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ You do not have permission to use this command.")
		b.API.Send(reply)
		return
	}
	
	// Check if replying to a message
	if msg.ReplyToMessage == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, uploadFormatText)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Check if replied message has a photo
	if len(msg.ReplyToMessage.Photo) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ The replied message must contain an image!")
		b.API.Send(reply)
		return
	}
	
	// Parse arguments
	args := strings.Fields(msg.Text)
	if len(args) != 4 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, uploadFormatText)
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Progress message
	progressMsg, _ := b.API.Send(tgbotapi.NewMessage(msg.Chat.ID, "⏳ <b>Starting upload process...</b>"))
	progressMsg.ParseMode = "HTML"
	
	// Parse arguments
	characterName := strings.Title(strings.ReplaceAll(args[1], "-", " "))
	animeName := strings.Title(strings.ReplaceAll(args[2], "-", " "))
	
	rarityNum, err := strconv.Atoi(args[3])
	if err != nil || rarityNum < 1 || rarityNum > 15 {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Rarity must be a number between 1-15."))
		return
	}
	
	rarityName := uploadRarityNames[rarityNum]
	
	// Step 1: Download image
	b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "📥 <b>Downloading image...</b>"))
	
	// Get the largest photo
	photo := msg.ReplyToMessage.Photo[len(msg.ReplyToMessage.Photo)-1]
	file, err := b.API.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to download image!"))
		return
	}
	
	// Download image data
	imageURL := file.Link(b.API.Token)
	resp, err := http.Get(imageURL)
	if err != nil {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to download image!"))
		return
	}
	defer resp.Body.Close()
	
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to read image data!"))
		return
	}
	
	// Check image size (10MB limit)
	if len(imageData) > 10*1024*1024 {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Image too large! Max size: 10MB"))
		return
	}
	
	// Step 2: Upload to hosting
	b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "☁️ <b>Uploading to cloud storage...</b>\n<i>This may take a few seconds...</i>"))
	
	uploader := NewImageUploader()
	imgURL, err := uploader.UploadWithFailover(imageData)
	if err != nil {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to upload image. All hosting services failed.\nPlease try again later."))
		return
	}
	
	// Step 3: Generate ID and save to database
	b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "💾 <b>Saving to database...</b>"))
	
	charID, err := b.GetNextSequenceNumber()
	if err != nil {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to generate character ID!"))
		return
	}
	
	// Create character document
	character := map[string]interface{}{
		"id":            charID,
		"name":          characterName,
		"anime":         animeName,
		"rarity":        rarityNum,
		"img_url":       imgURL,
		"created_at":    time.Now(),
		"added_by":      msg.From.ID,
		"added_by_name": msg.From.FirstName,
	}
	
	// Step 4: Post to channel
	caption := fmt.Sprintf(
		"<b>🎴 Character:</b> %s\n"+
			"<b>📺 Anime:</b> %s\n"+
			"<b>⭐ Rarity:</b> %s\n"+
			"<b>🆔 ID:</b> <code>%s</code>\n\n"+
			"<b>👤 Added by:</b> <a href=\"tg://user?id=%d\">%s</a>\n"+
			"<b>📅 Date:</b> %s",
		characterName,
		animeName,
		rarityName,
		charID,
		msg.From.ID,
		msg.From.FirstName,
		time.Now().Format("2006-01-02 15:04"),
	)
	
	channelMsg := tgbotapi.NewPhoto(b.Config.CharaChannelID, tgbotapi.FileURL(imgURL))
	channelMsg.Caption = caption
	channelMsg.ParseMode = "HTML"
	
	sentMsg, err := b.API.Send(channelMsg)
	if err != nil {
		// Fallback: send image directly
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "⚠️ <b>URL failed, sending image directly...</b>"))
		
		channelMsg := tgbotapi.NewPhoto(b.Config.CharaChannelID, tgbotapi.FileBytes{Bytes: imageData})
		channelMsg.Caption = caption
		channelMsg.ParseMode = "HTML"
		
		sentMsg, err = b.API.Send(channelMsg)
		if err != nil {
			b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to post to channel!"))
			return
		}
	}
	
	character["message_id"] = sentMsg.MessageID
	
	// Step 5: Save to database
	_, err = database.CharacterCollection.InsertOne(nil, character)
	if err != nil {
		b.API.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, progressMsg.MessageID, "❌ Failed to save to database!"))
		return
	}
	
	// Delete progress message
	b.deleteMessage(msg.Chat.ID, progressMsg.MessageID)
	
	// Success message
	channelUsername := b.Config.CharaChannelID
	if strings.HasPrefix(strconv.FormatInt(channelUsername, 10), "-100") {
		channelUsername, _ = strconv.ParseInt(strconv.FormatInt(channelUsername, 10)[4:], 10, 64)
	}
	
	successMsg := fmt.Sprintf(
		"✅ <b>Character Added Successfully!</b>\n\n"+
			"🆔 ID: <code>%s</code>\n"+
			"👤 Name: %s\n"+
			"📺 Anime: %s\n"+
			"⭐ Rarity: %s\n"+
			"🔗 <a href=\"%s\">Image Link</a>\n\n"+
			"<b>View in channel:</b> <a href=\"https://t.me/c/%d/%d\">Click here</a>",
		charID,
		characterName,
		animeName,
		rarityName,
		imgURL,
		channelUsername,
		sentMsg.MessageID,
	)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, successMsg)
	reply.ParseMode = "HTML"
	reply.DisableWebPagePreview = true
	b.API.Send(reply)
}

// cmdDelete handles /delete command
func (b *Bot) cmdDelete(msg *tgbotapi.Message) {
	// Check if user is admin
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ You do not have permission to use this command.")
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) != 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ <b>Incorrect format!</b>\n\n"+
				"<b>Usage:</b> <code>/delete ID</code>\n"+
				"<b>Example:</b> <code>/delete 042</code>")
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	
	// Find character
	char, err := b.CharacterService.GetCharacterByID(charID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("❌ Character with ID <code>%s</code> not found.", charID))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Delete from channel if message exists
	// (Would need to store message_id in character document)
	
	// Delete from database
	_, err = database.CharacterCollection.DeleteOne(nil, map[string]interface{}{"id": charID})
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to delete character!")
		b.API.Send(reply)
		return
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf(
		"✅ <b>Character Deleted!</b>\n\n"+
			"🆔 ID: <code>%s</code>\n"+
			"👤 Was: %s\n"+
			"📺 Anime: %s",
		charID, char.Name, char.Anime))
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}

// cmdUpdate handles /update command
func (b *Bot) cmdUpdate(msg *tgbotapi.Message) {
	// Check if user is admin
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ You do not have permission to use this command.")
		b.API.Send(reply)
		return
	}
	
	args := strings.Fields(msg.Text)
	if len(args) < 4 {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ <b>Incorrect format!</b>\n\n"+
				"<b>Usage:</b> <code>/update ID field new_value</code>\n\n"+
				"<b>Fields:</b> name, anime, rarity, img_url\n\n"+
				"<b>Examples:</b>\n"+
				"<code>/update 042 name Naruto-Uzumaki</code>\n"+
				"<code>/update 042 rarity 5</code>\n"+
				"<code>/update 042 img_url https://example.com/image.jpg</code>")
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	charID := args[1]
	field := args[2]
	newValue := strings.Join(args[3:], " ")
	
	validFields := map[string]bool{"img_url": true, "name": true, "anime": true, "rarity": true}
	if !validFields[field] {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Invalid field. Use one of: img_url, name, anime, rarity")
		b.API.Send(reply)
		return
	}
	
	// Find character
	char, err := b.CharacterService.GetCharacterByID(charID)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("❌ Character with ID <code>%s</code> not found.", charID))
		reply.ParseMode = "HTML"
		b.API.Send(reply)
		return
	}
	
	// Process new value
	var processedValue interface{}
	if field == "name" || field == "anime" {
		processedValue = strings.Title(strings.ReplaceAll(newValue, "-", " "))
	} else if field == "rarity" {
		rarityNum, err := strconv.Atoi(newValue)
		if err != nil || rarityNum < 1 || rarityNum > 15 {
			reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Rarity must be a number between 1-15.")
			b.API.Send(reply)
			return
		}
		processedValue = rarityNum
	} else {
		processedValue = newValue
	}
	
	// Update database
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			field:        processedValue,
			"updated_at": time.Now(),
			"updated_by": msg.From.ID,
		},
	}
	
	_, err = database.CharacterCollection.UpdateOne(nil, map[string]interface{}{"id": charID}, update)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to update character!")
		b.API.Send(reply)
		return
	}
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf(
		"✅ <b>Updated Successfully!</b>\n\n"+
			"🆔 ID: <code>%s</code>\n"+
			"🔄 Field: <code>%s</code>\n"+
			"✨ New Value: <code>%v</code>",
		charID, field, processedValue))
	reply.ParseMode = "HTML"
	b.API.Send(reply)
	
	_ = char // Use char variable
}

// cmdStats handles /stats command
func (b *Bot) cmdStats(msg *tgbotapi.Message) {
	// Check if user is admin
	if !b.Config.IsSudo(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ You do not have permission to use this command.")
		b.API.Send(reply)
		return
	}
	
	// Get total count
	total, err := b.CharacterService.GetCharacterCount()
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Error fetching stats!")
		b.API.Send(reply)
		return
	}
	
	// Build message
	text := fmt.Sprintf("📊 <b>Database Statistics</b>\n\n")
	text += fmt.Sprintf("📦 <b>Total Characters:</b> <code>%d</code>\n", total)
	
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "HTML"
	b.API.Send(reply)
}
