#!/bin/bash

# Senpai Waifu Bot - VPS Deployment Script
# Usage: curl -sSL https://your-server.com/deploy.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BOT_DIR="/opt/senpai-bot"
GO_VERSION="1.21.6"
SERVICE_NAME="senpai-bot"

# Logging
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root. Use: sudo bash deploy.sh"
    fi
}

# Update system
update_system() {
    log "Updating system packages..."
    apt update && apt upgrade -y
    success "System updated"
}

# Install dependencies
install_deps() {
    log "Installing dependencies..."
    apt install -y git curl wget nano build-essential ufw fail2ban
    success "Dependencies installed"
}

# Install Go
install_go() {
    if command -v go &> /dev/null; then
        INSTALLED_GO=$(go version | awk '{print $3}')
        log "Go already installed: $INSTALLED_GO"
        return
    fi

    log "Installing Go $GO_VERSION..."
    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
    echo 'export GOPATH=$HOME/go' >> /etc/profile.d/go.sh
    export PATH=$PATH:/usr/local/go/bin
    
    success "Go $GO_VERSION installed"
}

# Install Docker
install_docker() {
    if command -v docker &> /dev/null; then
        log "Docker already installed"
        return
    fi

    log "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    apt install -y docker-compose
    systemctl start docker
    systemctl enable docker
    success "Docker installed"
}

# Setup firewall
setup_firewall() {
    log "Configuring firewall..."
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow 22/tcp
    ufw --force enable
    success "Firewall configured"
}

# Create bot directory
setup_directory() {
    log "Setting up bot directory..."
    mkdir -p "$BOT_DIR"
    cd "$BOT_DIR"
    success "Directory created at $BOT_DIR"
}

# Download bot files
download_bot() {
    log "Downloading bot files..."
    cd "$BOT_DIR"
    
    # Check if files already exist
    if [[ -f "$BOT_DIR/go.mod" ]]; then
        warning "Bot files already exist. Updating..."
        git pull origin main 2>/dev/null || true
    else
        # Download from GitHub or your server
        log "Please upload the bot files to $BOT_DIR"
        log "Or clone from GitHub:"
        log "  git clone https://github.com/yourusername/senpai-waifu-bot-go.git ."
        
        # Create placeholder
        mkdir -p cmd/bot internal
    fi
    
    success "Bot files ready"
}

# Create environment file
setup_env() {
    if [[ -f "$BOT_DIR/.env" ]]; then
        warning ".env file already exists. Skipping..."
        return
    fi

    log "Creating environment file..."
    cat > "$BOT_DIR/.env" << 'EOF'
# Bot Configuration
BOT_TOKEN=your_bot_token_here
BOT_USERNAME=your_bot_username

# Telegram API
API_ID=your_api_id
API_HASH=your_api_hash

# Owner and Sudo
OWNER_ID=your_telegram_user_id
SUDO_USERS=

# Group IDs
GROUP_ID=-1001234567890
CHARA_CHANNEL_ID=-1009876543210

# MongoDB
MONGO_URL=mongodb+srv://username:password@cluster.mongodb.net/database

# Media
VIDEO_URL=

# Community Links
SUPPORT_CHAT=your_support_chat
UPDATE_CHAT=your_update_chat
EOF

    success ".env file created"
    warning "Please edit $BOT_DIR/.env with your actual values!"
}

# Build bot
build_bot() {
    log "Building bot..."
    cd "$BOT_DIR"
    
    # Check if source exists
    if [[ ! -f "$BOT_DIR/cmd/bot/main.go" ]]; then
        error "Bot source code not found in $BOT_DIR"
    fi
    
    # Download dependencies and build
    /usr/local/go/bin/go mod download
    /usr/local/go/bin/go build -o bot ./cmd/bot
    chmod +x bot
    
    success "Bot built successfully"
}

# Create systemd service
create_service() {
    log "Creating systemd service..."
    
    cat > "/etc/systemd/system/${SERVICE_NAME}.service" << EOF
[Unit]
Description=Senpai Waifu Bot
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${BOT_DIR}
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin"
EnvironmentFile=${BOT_DIR}/.env
ExecStart=${BOT_DIR}/bot
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    success "Systemd service created"
}

# Start bot
start_bot() {
    log "Starting bot..."
    systemctl start "$SERVICE_NAME"
    sleep 2
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        success "Bot started successfully!"
    else
        error "Bot failed to start. Check logs: journalctl -u $SERVICE_NAME -f"
    fi
}

# Show status
show_status() {
    echo ""
    echo "========================================"
    echo "  Senpai Waifu Bot - Deployment Status"
    echo "========================================"
    echo ""
    
    echo -e "${BLUE}Service Status:${NC}"
    systemctl status "$SERVICE_NAME" --no-pager | head -5
    
    echo ""
    echo -e "${BLUE}Installation Directory:${NC} $BOT_DIR"
    echo -e "${BLUE}Config File:${NC} $BOT_DIR/.env"
    echo -e "${BLUE}Service Name:${NC} $SERVICE_NAME"
    
    echo ""
    echo -e "${GREEN}Useful Commands:${NC}"
    echo "  Start:    systemctl start $SERVICE_NAME"
    echo "  Stop:     systemctl stop $SERVICE_NAME"
    echo "  Restart:  systemctl restart $SERVICE_NAME"
    echo "  Status:   systemctl status $SERVICE_NAME"
    echo "  Logs:     journalctl -u $SERVICE_NAME -f"
    
    echo ""
    echo -e "${YELLOW}IMPORTANT:${NC} Edit $BOT_DIR/.env with your actual values before using the bot!"
    echo ""
}

# Main function
main() {
    echo "========================================"
    echo "  Senpai Waifu Bot - VPS Deployer"
    echo "========================================"
    echo ""
    
    check_root
    update_system
    install_deps
    install_go
    install_docker
    setup_firewall
    setup_directory
    download_bot
    setup_env
    build_bot
    create_service
    start_bot
    show_status
    
    success "Deployment completed!"
}

# Run main function
main "$@"
