# VPS Deployment Guide 🖥️

Complete guide to deploy Senpai Waifu Bot on a VPS (Ubuntu/Debian).

---

## 📋 Requirements

- VPS with Ubuntu 20.04/22.04 or Debian 11/12
- Minimum 1GB RAM, 1 CPU Core
- 10GB Storage
- Root or sudo access

---

## 🚀 Quick Deploy (One-Liner)

```bash
curl -sSL https://raw.githubusercontent.com/yourusername/senpai-waifu-bot-go/main/deploy.sh | bash
```

---

## 📖 Step-by-Step Deployment

### Step 1: Connect to VPS

```bash
ssh root@your-vps-ip
```

### Step 2: Update System

```bash
apt update && apt upgrade -y
```

### Step 3: Install Required Packages

```bash
apt install -y git curl wget nano build-essential
```

### Step 4: Install Go

```bash
# Download Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz

# Extract to /usr/local
tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### Step 5: Install Docker (Optional but Recommended)

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Docker Compose
apt install -y docker-compose

# Start Docker
systemctl start docker
systemctl enable docker

# Add user to docker group (optional)
usermod -aG docker $USER
```

### Step 6: Create Bot Directory

```bash
mkdir -p /opt/senpai-bot
cd /opt/senpai-bot
```

### Step 7: Upload Bot Files

**Option A: Using SCP from local machine**
```bash
# On your local machine
scp senpai-waifu-bot-go.tar.gz root@your-vps-ip:/opt/senpai-bot/
```

**Option B: Using wget/curl**
```bash
cd /opt/senpai-bot
wget https://your-server.com/senpai-waifu-bot-go.tar.gz
```

**Option C: Clone from GitHub**
```bash
cd /opt
git clone https://github.com/yourusername/senpai-waifu-bot-go.git
mv senpai-waifu-bot-go senpai-bot
cd senpai-bot
```

### Step 8: Extract Files

```bash
cd /opt/senpai-bot
tar -xzf senpai-waifu-bot-go.tar.gz
mv senpai-waifu-bot-go/* .
mv senpai-waifu-bot-go/.* . 2>/dev/null || true
rmdir senpai-waifu-bot-go
```

### Step 9: Configure Environment

```bash
# Create .env file
nano /opt/senpai-bot/.env
```

**Add your configuration:**
```bash
BOT_TOKEN=your_bot_token_here
BOT_USERNAME=your_bot_username
API_ID=your_api_id
API_HASH=your_api_hash
OWNER_ID=your_telegram_user_id
GROUP_ID=-1001234567890
CHARA_CHANNEL_ID=-1009876543210
MONGO_URL=mongodb+srv://username:password@cluster.mongodb.net/database
VIDEO_URL=https://example.com/video1.mp4
SUPPORT_CHAT=your_support_chat
UPDATE_CHAT=your_update_chat
```

**Save:** Press `Ctrl+X`, then `Y`, then `Enter`

### Step 10: Deploy (Choose One Method)

---

## 🐳 Method 1: Docker Deployment (Recommended)

```bash
cd /opt/senpai-bot

# Build and run
docker-compose up -d --build

# Check logs
docker-compose logs -f

# Stop bot
docker-compose down

# Restart bot
docker-compose restart

# Update bot after code changes
docker-compose up -d --build --force-recreate
```

---

## ⚙️ Method 2: Systemd Service (Native Go)

### Step 1: Build the Bot

```bash
cd /opt/senpai-bot

# Download dependencies
go mod download

# Build binary
go build -o bot ./cmd/bot

# Make executable
chmod +x bot
```

### Step 2: Create Systemd Service

```bash
nano /etc/systemd/system/senpai-bot.service
```

**Add this content:**
```ini
[Unit]
Description=Senpai Waifu Bot
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/senpai-bot
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin"
EnvironmentFile=/opt/senpai-bot/.env
ExecStart=/opt/senpai-bot/bot
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=senpai-bot

[Install]
WantedBy=multi-user.target
```

### Step 3: Start Service

```bash
# Reload systemd
systemctl daemon-reload

# Enable service (start on boot)
systemctl enable senpai-bot

# Start service
systemctl start senpai-bot

# Check status
systemctl status senpai-bot
```

### Step 4: Manage Service

```bash
# Start
systemctl start senpai-bot

# Stop
systemctl stop senpai-bot

# Restart
systemctl restart senpai-bot

# View logs
journalctl -u senpai-bot -f

# View all logs
journalctl -u senpai-bot --no-pager

# Clear logs
journalctl --rotate
journalctl --vacuum-time=1d
```

---

## 📊 Method 3: PM2 Process Manager

### Step 1: Install Node.js & PM2

```bash
# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y nodejs

# Install PM2
npm install -g pm2
```

### Step 2: Build Bot

```bash
cd /opt/senpai-bot
go build -o bot ./cmd/bot
```

### Step 3: Create PM2 Config

```bash
nano /opt/senpai-bot/ecosystem.config.js
```

**Add:**
```javascript
module.exports = {
  apps: [{
    name: 'senpai-bot',
    script: './bot',
    cwd: '/opt/senpai-bot',
    env: {
      NODE_ENV: 'production'
    },
    env_file: '/opt/senpai-bot/.env',
    autorestart: true,
    max_restarts: 10,
    min_uptime: '10s',
    log_file: '/opt/senpai-bot/logs/combined.log',
    out_file: '/opt/senpai-bot/logs/out.log',
    error_file: '/opt/senpai-bot/logs/error.log',
    time: true
  }]
};
```

### Step 4: Start with PM2

```bash
# Create logs directory
mkdir -p /opt/senpai-bot/logs

# Start bot
pm2 start ecosystem.config.js

# Save PM2 config
pm2 save

# Setup startup script
pm2 startup systemd
```

### Step 5: PM2 Commands

```bash
# Start
pm2 start senpai-bot

# Stop
pm2 stop senpai-bot

# Restart
pm2 restart senpai-bot

# View logs
pm2 logs senpai-bot

# Monitor
pm2 monit

# List processes
pm2 list
```

---

## 🔧 Troubleshooting

### Bot Not Starting

```bash
# Check logs
journalctl -u senpai-bot -f

# Or for Docker
docker-compose logs -f

# Check if binary exists
ls -la /opt/senpai-bot/bot

# Test run manually
cd /opt/senpai-bot
./bot
```

### Permission Denied

```bash
chmod +x /opt/senpai-bot/bot
chown -R root:root /opt/senpai-bot
```

### MongoDB Connection Failed

```bash
# Test MongoDB connection
mongo "your-mongodb-url" --eval "db.adminCommand('ping')"

# Check network
ping your-mongodb-host
```

### High Memory Usage

```bash
# Check memory usage
free -h

# Check bot memory
ps aux | grep bot

# Limit memory in systemd (add to service file)
# MemoryLimit=512M
```

---

## 🔄 Update Bot

### Docker Method

```bash
cd /opt/senpai-bot

# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose up -d --build
```

### Systemd Method

```bash
cd /opt/senpai-bot

# Stop service
systemctl stop senpai-bot

# Pull latest code
git pull origin main

# Rebuild
go build -o bot ./cmd/bot

# Start service
systemctl start senpai-bot

# Check status
systemctl status senpai-bot
```

---

## 📈 Monitoring

### Check Bot Status

```bash
# Systemd
systemctl is-active senpai-bot

# Docker
docker-compose ps

# PM2
pm2 status
```

### View Logs

```bash
# Systemd
journalctl -u senpai-bot -f -n 100

# Docker
docker-compose logs -f --tail=100

# PM2
pm2 logs senpai-bot --lines 100
```

### Resource Usage

```bash
# CPU & Memory
top -p $(pgrep -d',' bot)

# Disk usage
df -h

# Network
netstat -tulpn | grep bot
```

---

## 🛡️ Security

### Firewall (UFW)

```bash
# Install UFW
apt install -y ufw

# Allow SSH
ufw allow 22/tcp

# Enable firewall
ufw enable

# Check status
ufw status
```

### Fail2Ban

```bash
# Install
apt install -y fail2ban

# Start
systemctl enable fail2ban
systemctl start fail2ban
```

---

## 💡 Performance Tips

1. **Use SSD storage** for faster I/O
2. **Deploy close to Telegram servers** (EU recommended)
3. **Use MongoDB Atlas** for managed database
4. **Enable swap** for low memory VPS:
   ```bash
   fallocate -l 2G /swapfile
   chmod 600 /swapfile
   mkswap /swapfile
   swapon /swapfile
   echo '/swapfile none swap sw 0 0' >> /etc/fstab
   ```

---

## 📞 Support

If you face any issues:
1. Check logs: `journalctl -u senpai-bot -f`
2. Verify `.env` configuration
3. Test MongoDB connection
4. Check firewall settings

---

**Your bot should now be running on your VPS!** 🎉
