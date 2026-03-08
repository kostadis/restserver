#!/bin/bash
# =============================================================================
# 5etools WSL Setup Script
# Installs and configures 5etools to run as a persistent service under WSL
# Run this script inside your WSL terminal (Ubuntu recommended)
# =============================================================================

set -e  # Exit on any error

# --- Configuration -----------------------------------------------------------
INSTALL_DIR="$HOME/5etools"          # Where 5etools source will be cloned
IMAGES_DIR="$INSTALL_DIR/img"        # Images subfolder (inside source repo)
PORT=5050                            # Port to serve 5etools on
SERVICE_NAME="5etools"               # systemd service name
LOG_FILE="$HOME/5etools-install.log" # Log file for this install

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log()  { echo -e "${GREEN}[✔]${NC} $1" | tee -a "$LOG_FILE"; }
warn() { echo -e "${YELLOW}[!]${NC} $1" | tee -a "$LOG_FILE"; }
info() { echo -e "${BLUE}[i]${NC} $1" | tee -a "$LOG_FILE"; }
fail() { echo -e "${RED}[✘]${NC} $1" | tee -a "$LOG_FILE"; exit 1; }

echo "" | tee "$LOG_FILE"
echo "============================================================" | tee -a "$LOG_FILE"
echo "  5etools WSL Installer — $(date)"                             | tee -a "$LOG_FILE"
echo "============================================================" | tee -a "$LOG_FILE"
echo ""

# --- Step 0: Preflight checks ------------------------------------------------
info "Checking prerequisites..."

# Must be running in WSL
if ! grep -qi microsoft /proc/version 2>/dev/null; then
    warn "This script is designed for WSL. Continuing anyway..."
fi

# Must NOT be root
if [[ "$EUID" -eq 0 ]]; then
    fail "Do not run this script as root. Run as your normal WSL user."
fi

# Check internet
if ! curl -s --max-time 5 https://github.com > /dev/null; then
    fail "No internet connection detected. Please connect and retry."
fi

# --- Step 1: Install system dependencies -------------------------------------
info "Updating apt and installing git, curl, build tools..."
sudo apt-get update -qq
sudo apt-get install -y -qq git curl build-essential

log "System dependencies installed."

# --- Step 2: Install Node.js (via NodeSource) --------------------------------
if command -v node &>/dev/null; then
    NODE_VER=$(node --version)
    log "Node.js already installed: $NODE_VER"
else
    info "Installing Node.js 20.x (LTS)..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - >> "$LOG_FILE" 2>&1
    sudo apt-get install -y nodejs >> "$LOG_FILE" 2>&1
    log "Node.js installed: $(node --version)"
fi

if ! command -v npm &>/dev/null; then
    fail "npm not found after Node.js install. Something went wrong."
fi
log "npm version: $(npm --version)"

# --- Step 3: Clone 5etools source --------------------------------------------
if [[ -d "$INSTALL_DIR/.git" ]]; then
    warn "5etools already cloned at $INSTALL_DIR — skipping clone."
else
    info "Cloning 5etools source (this may take several minutes)..."
    # Retry loop — large repos can time out on first attempt
    for attempt in 1 2 3; do
        git clone https://github.com/5etools-mirror-3/5etools-src.git "$INSTALL_DIR" \
            >> "$LOG_FILE" 2>&1 && break
        warn "Clone attempt $attempt failed. Retrying..."
        sleep 5
    done
    [[ -d "$INSTALL_DIR/.git" ]] || fail "Failed to clone 5etools after 3 attempts."
    log "5etools source cloned to $INSTALL_DIR"
fi

# --- Step 4: (Optional) Clone images -----------------------------------------
echo ""
read -r -p "$(echo -e "${YELLOW}[?]${NC} Download 5etools images? (~5 GB, takes a long time) [y/N]: ")" DOWNLOAD_IMAGES
if [[ "$DOWNLOAD_IMAGES" =~ ^[Yy]$ ]]; then
    if [[ -d "$IMAGES_DIR/.git" ]]; then
        warn "Images already cloned at $IMAGES_DIR — skipping."
    else
        info "Cloning images (this will take a LONG time)..."
        for attempt in 1 2 3; do
            git clone https://github.com/5etools-mirror-3/5etools-img.git "$IMAGES_DIR" \
                >> "$LOG_FILE" 2>&1 && break
            warn "Image clone attempt $attempt failed. Retrying..."
            sleep 5
        done
        [[ -d "$IMAGES_DIR/.git" ]] || warn "Image clone failed — site will work but some art may be missing."
        log "Images cloned."
    fi
else
    info "Skipping image download. You can run this script again later and choose Y."
fi

# --- Step 5: Install npm dependencies ----------------------------------------
info "Running npm ci in $INSTALL_DIR ..."
cd "$INSTALL_DIR"
npm ci >> "$LOG_FILE" 2>&1
log "npm dependencies installed."

# --- Step 6: Build service worker (optional cache) ---------------------------
echo ""
read -r -p "$(echo -e "${YELLOW}[?]${NC} Build service worker for faster local caching? [Y/n]: ")" BUILD_SW
if [[ ! "$BUILD_SW" =~ ^[Nn]$ ]]; then
    info "Building production service worker..."
    npm run build:sw:prod >> "$LOG_FILE" 2>&1
    log "Service worker built."
else
    info "Skipping service worker build."
fi

# --- Step 7: Write the systemd service file ----------------------------------
info "Creating systemd service: $SERVICE_NAME ..."

# WSL 2 supports systemd if enabled in /etc/wsl.conf
SYSTEMD_ENABLED=false
if systemctl is-system-running &>/dev/null 2>&1; then
    SYSTEMD_ENABLED=true
fi

SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=5etools local web server
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$(which npm) run serve:dev
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

log "Service file written to $SERVICE_FILE"

# --- Step 8: Enable and start the service ------------------------------------
if $SYSTEMD_ENABLED; then
    info "Enabling and starting $SERVICE_NAME via systemd..."
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    sudo systemctl start "$SERVICE_NAME"
    sleep 2
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log "Service is running!"
    else
        warn "Service may not have started. Check: sudo journalctl -u $SERVICE_NAME -n 30"
    fi
else
    warn "systemd does not appear to be active in this WSL instance."
    warn "To enable systemd in WSL 2, add the following to /etc/wsl.conf:"
    warn "  [boot]"
    warn "  systemd=true"
    warn "Then restart WSL: run  wsl --shutdown  from PowerShell, then reopen WSL."
    echo ""
    warn "For now, you can start 5etools manually by running:"
    warn "  cd $INSTALL_DIR && npm run serve:dev"
fi

# --- Step 9: Write a convenience update script --------------------------------
UPDATE_SCRIPT="$HOME/update-5etools.sh"
cat > "$UPDATE_SCRIPT" <<'UPDATESCRIPT'
#!/bin/bash
# 5etools updater — run this whenever a new version is released
set -e
INSTALL_DIR="$HOME/5etools"
cd "$INSTALL_DIR"

echo "[i] Pulling latest 5etools source..."
git pull

if [[ -d "img/.git" ]]; then
    echo "[i] Pulling latest images..."
    cd img && git pull && cd ..
fi

echo "[i] Updating npm dependencies..."
npm ci

# Rebuild service worker if it was previously built
if [[ -f "sw.js" ]]; then
    echo "[i] Rebuilding service worker..."
    npm run build:sw:prod
fi

# Restart the service if systemd is running
if systemctl is-system-running &>/dev/null 2>&1; then
    echo "[i] Restarting 5etools service..."
    sudo systemctl restart 5etools
    echo "[✔] Done! 5etools updated and restarted."
else
    echo "[✔] Done! Restart 5etools manually: cd $INSTALL_DIR && npm run serve:dev"
fi
UPDATESCRIPT

chmod +x "$UPDATE_SCRIPT"
log "Update script written to $UPDATE_SCRIPT"

# --- Done --------------------------------------------------------------------
echo ""
echo "============================================================"
log "Installation complete!"
echo "============================================================"
echo ""
echo -e "  ${GREEN}5etools URL:${NC}  http://localhost:${PORT}/index.html"
echo ""
echo -e "  ${BLUE}Useful commands:${NC}"
echo -e "    Start service:    sudo systemctl start $SERVICE_NAME"
echo -e "    Stop service:     sudo systemctl stop $SERVICE_NAME"
echo -e "    Restart service:  sudo systemctl restart $SERVICE_NAME"
echo -e "    View logs:        sudo journalctl -u $SERVICE_NAME -f"
echo -e "    Manual start:     cd $INSTALL_DIR && npm run serve:dev"
echo -e "    Update site:      ~/update-5etools.sh"
echo ""
echo -e "  ${YELLOW}Note:${NC} If systemd is not yet enabled in WSL, see the warning above."
echo ""