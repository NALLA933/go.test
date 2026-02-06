package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// SmallCapsMap maps regular letters to small caps Unicode
var SmallCapsMap = map[rune]rune{
	'a': 'ᴀ', 'b': 'ʙ', 'c': 'ᴄ', 'd': 'ᴅ', 'e': 'ᴇ', 'f': 'ꜰ', 'g': 'ɢ',
	'h': 'ʜ', 'i': 'ɪ', 'j': 'ᴊ', 'k': 'ᴋ', 'l': 'ʟ', 'm': 'ᴍ', 'n': 'ɴ',
	'o': 'ᴏ', 'p': 'ᴘ', 'q': 'ǫ', 'r': 'ʀ', 's': 'ꜱ', 't': 'ᴛ', 'u': 'ᴜ',
	'v': 'ᴠ', 'w': 'ᴡ', 'x': 'x', 'y': 'ʏ', 'z': 'ᴢ',
	'A': 'ᴀ', 'B': 'ʙ', 'C': 'ᴄ', 'D': 'ᴅ', 'E': 'ᴇ', 'F': 'ꜰ', 'G': 'ɢ',
	'H': 'ʜ', 'I': 'ɪ', 'J': 'ᴊ', 'K': 'ᴋ', 'L': 'ʟ', 'M': 'ᴍ', 'N': 'ɴ',
	'O': 'ᴏ', 'P': 'ᴘ', 'Q': 'ǫ', 'R': 'ʀ', 'S': 'ꜱ', 'T': 'ᴛ', 'U': 'ᴜ',
	'V': 'ᴠ', 'W': 'ᴡ', 'X': 'x', 'Y': 'ʏ', 'Z': 'ᴢ',
}

// ToSmallCaps converts text to small caps Unicode
func ToSmallCaps(text string) string {
	var result strings.Builder
	for _, char := range text {
		if small, ok := SmallCapsMap[char]; ok {
			result.WriteRune(small)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// RarityMap maps rarity numbers to display names
var RarityMap = map[int]string{
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

// RarityEmojis maps rarity numbers to emojis only
var RarityEmojis = map[int]string{
	1:  "⚪", 2: "🔵", 3: "🟡", 4: "💮", 5: "👹",
	6:  "🎐", 7: "🔮", 8: "🪐", 9: "⚰️", 10: "🌬️",
	11: "💝", 12: "🌸", 13: "🏖️", 14: "🍭", 15: "🧬",
}

// RarityNames maps rarity numbers to names
var RarityNames = map[int]string{
	1:  "ᴄᴏᴍᴍᴏɴ",
	2:  "ʀᴀʀᴇ",
	3:  "ʟᴇɢᴇɴᴅᴀʀʏ",
	4:  "ꜱᴘᴇᴄɪᴀʟ",
	5:  "ᴀɴᴄɪᴇɴᴛ",
	6:  "ᴄᴇʟᴇꜱᴛɪᴀʟ",
	7:  "ᴇᴘɪᴄ",
	8:  "ᴄᴏꜱᴍɪᴄ",
	9:  "ɴɪɢʜᴛᴍᴀʀᴇ",
	10: "ꜰʀᴏꜱᴛʙᴏʀɴ",
	11: "ᴠᴀʟᴇɴᴛɪɴᴇ",
	12: "ꜱᴘʀɪɴɢ",
	13: "ᴛʀᴏᴘɪᴄᴀʟ",
	14: "ᴋᴀᴡᴀɪɪ",
	15: "ʜʏʙʀɪᴅ",
}

// GetRarityDisplay returns the display string for a rarity
func GetRarityDisplay(rarity int) string {
	if display, ok := RarityMap[rarity]; ok {
		return display
	}
	return fmt.Sprintf("⚪ ᴜɴᴋɴᴏᴡɴ (%d)", rarity)
}

// GetRarityFromString parses rarity from various formats
func GetRarityFromString(rarityValue interface{}) int {
	switch v := rarityValue.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case string:
		rarityStr := strings.TrimSpace(strings.ToLower(v))
		
		// Try parsing as number
		if num, err := parseInt(rarityStr); err == nil {
			return num
		}
		
		// Emoji to int mapping
		emojiToInt := map[string]int{
			"⚪": 1, "🔵": 2, "🟡": 3, "💮": 4, "👹": 5,
			"🎐": 6, "🔮": 7, "🪐": 8, "⚰️": 9, "🌬️": 10,
			"💝": 11, "🌸": 12, "🏖️": 13, "🍭": 14, "🧬": 15,
		}
		
		for emoji, num := range emojiToInt {
			if strings.Contains(rarityStr, emoji) {
				return num
			}
		}
		
		// Name to int mapping
		nameToInt := map[string]int{
			"common": 1, "rare": 2, "legendary": 3, "special": 4, "ancient": 5,
			"celestial": 6, "epic": 7, "cosmic": 8, "nightmare": 9, "frostborn": 10,
			"valentine": 11, "spring": 12, "tropical": 13, "kawaii": 14, "hybrid": 15,
		}
		
		if num, ok := nameToInt[rarityStr]; ok {
			return num
		}
	}
	return 1
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// ParseInt64 parses a string to int64
func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseInt parses a string to int
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// GenerateRandomCode generates a random code of specified length
func GenerateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

// GenerateUniqueCode generates a unique redeem code
func GenerateUniqueCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "sanpai-" + hex.EncodeToString(b)[:8]
}

// GenerateCoinCode generates a coin code
func GenerateCoinCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 8)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return "COIN-" + string(result)
}

// EscapeMarkdown escapes markdown characters
func EscapeMarkdown(text string) string {
	chars := []string{"\\", "*", "_", "`", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range chars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// FormatNumber formats a number with commas
func FormatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	
	str := fmt.Sprintf("%d", n)
	var result strings.Builder
	count := 0
	
	for i := len(str) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteByte(str[i])
		count++
	}
	
	// Reverse the result
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// GetISTDate returns current date in IST timezone
func GetISTDate() string {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return time.Now().In(loc).Format("2006-01-02")
}

// GetISTNow returns current time in IST timezone
func GetISTNow() time.Time {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return time.Now().In(loc)
}

// ContainsInt checks if an int slice contains a value
func ContainsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// ContainsInt64 checks if an int64 slice contains a value
func ContainsInt64(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// ContainsString checks if a string slice contains a value
func ContainsString(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// RemoveString removes a string from a slice
func RemoveString(slice []string, val string) []string {
	var result []string
	for _, v := range slice {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}

// RemoveInt removes an int from a slice
func RemoveInt(slice []int, val int) []int {
	var result []int
	for _, v := range slice {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}
