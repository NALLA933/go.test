# Senpai Waifu Bot - Go Edition 🎴

A high-performance Telegram bot for collecting anime characters, written in Go for maximum speed and efficiency.

## Features ⚡

- **Character Guessing Game** - Characters spawn after every X messages
- **Harem/Collection System** - Collect and manage your favorite characters
- **Balance & Shop** - Earn coins and buy characters from the shop
- **Gift & Trade** - Exchange characters with other users
- **Leaderboards** - Daily, global, and balance rankings
- **Redeem Codes** - Admin-generated codes for coins and characters
- **Rarity Management** - Enable/disable rarities per chat
- **Character Locking** - Lock characters from spawning
- **Fast Response** - Go-powered for 100-200ms ping times

## Tech Stack 🚀

- **Language**: Go 1.21+
- **Bot Framework**: go-telegram-bot-api
- **Database**: MongoDB with official Go driver
- **Deployment**: Docker support included

## Commands 📋

### User Commands
- `/start` - Start the bot
- `/guess <name>` - Guess the character name
- `/collection` or `/harem` - View your collection
- `/balance` - Check your coin balance
- `/pay <amount>` - Send coins to another user
- `/shop` - Browse the character shop
- `/gift <character_id>` - Gift a character
- `/trade <your_id> <their_id>` - Trade characters
- `/sclaim` - Claim a free character (24h cooldown)
- `/claim` - Generate a coin code (24h cooldown)
- `/redeem <code>` - Redeem a code
- `/leaderboard [global|daily|group|balance]` - View rankings
- `/sfind <name>` - Search for characters
- `/scheck <id>` - Check character details
- `/smode` - Change collection filter
- `/fav <id>` - Add character to favorites

### Admin Commands
- `/ping` - Check bot latency
- `/gen <amount> [max_uses]` - Generate coin code
- `/sgen <char_id> [max_uses]` - Generate character code
- `/addbal <user_id> <amount>` - Add balance to user
- `/set_on <rarity>` - Enable rarity
- `/set_off <rarity>` - Disable rarity
- `/lock <char_id> [reason]` - Lock character
- `/unlock <char_id>` - Unlock character
- `/locklist` - List locked characters
- `/resetshop <user_id>` - Reset user's shop

## Setup 🛠️

### Prerequisites
- Go 1.21 or higher
- MongoDB database
- Telegram Bot Token

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/senpai-waifu-bot-go.git
cd senpai-waifu-bot-go
```

2. Copy the environment file:
```bash
cp .env.example .env
```

3. Edit `.env` with your configuration:
```bash
BOT_TOKEN=your_bot_token
BOT_USERNAME=your_bot_username
API_ID=your_api_id
API_HASH=your_api_hash
OWNER_ID=your_telegram_user_id
GROUP_ID=-1001234567890
CHARA_CHANNEL_ID=-1009876543210
MONGO_URL=mongodb+srv://...
```

4. Run the bot:
```bash
go run ./cmd/bot
```

### Docker Deployment

1. Build the Docker image:
```bash
docker build -t senpai-waifu-bot .
```

2. Run the container:
```bash
docker run -d --env-file .env --name senpai-bot senpai-waifu-bot
```

## Performance 💨

This bot is built with Go for maximum performance:
- **Fast Response Times**: 100-200ms average ping
- **Concurrent Handling**: Goroutines for parallel processing
- **Efficient Memory Usage**: Go's garbage collector
- **Low Latency**: Direct MongoDB connection pooling

## Database Schema 📊

The bot uses the following MongoDB collections:
- `anime_characters_lol` - Character data
- `user_collection_lmaoooo` - User collections
- `user_totals_lmaoooo` - Message frequency settings
- `group_user_totalsssssss` - Group user stats
- `top_global_groups` - Global group rankings
- `daily_user_guesses` - Daily user stats
- `daily_group_guesses` - Daily group stats
- `redeem_codes` - Redeem codes
- `claim_codes` - Claim codes
- `rarity_settings` - Chat rarity settings
- `locked_characters` - Locked characters

## Contributing 🤝

Contributions are welcome! Please feel free to submit a Pull Request.

## License 📄

This project is licensed under the MIT License.

## Credits 🙏

- Original Python bot by the Senpai team
- Go conversion for performance optimization
