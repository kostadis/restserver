#!/bin/bash
# =============================================================================
# 5etools Unpacker — restores 5etools from a portable archive on a new machine
# Run this on your DESTINATION WSL machine
# Usage:  ./unpack-5etools.sh [/path/to/5etools-portable-*.tar.gz]
# =============================================================================

set -e

INSTALL_DIR="$HOME/5etools"
PORT=5050
SERVICE_NAME="5etools"

# Colors
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; RED='\033[0;31m'; NC='\033[0m'
log()  { echo -e "${GREEN}[✔]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
info() { echo -e "${BLUE}[i]${NC} $1"; }
fail() { echo -e "${RED}[✘]${NC} $1"; exit 1; }

echo ""
echo "============================================================"
echo "  5etools Unpacker"
echo "============================================================"
echo ""

# --- Locate the archive ------------------------------------------------------
if [[ -n "$1" ]]; then
    ARCHIVE="$1"
else
    # Auto-search USB drives then home dir
    ARCHIVE=$(find /mnt /home/"$USER" -maxdepth 3 -name "5etools-portable-*.tar.gz" \
        2>/dev/null | sort | tail -1)
    if [[ -n "$ARCHIVE" ]]; then
        info "Auto-detected archive: $ARCHIVE"
    else
        fail "No archive found. Pass the path as an argument:\n  ./unpack-5etools.sh /mnt/e/5etools-portable-*.tar.gz"
    fi
fi

[[ -f "$ARCHIVE" ]] || fail "Archive not found: $ARCHIVE"

# --- Guard against overwriting -----------------------------------------------
if [[ -d "$INSTALL_DIR" ]]; then
    warn "5etools directory already exists at $INSTALL_DIR"
    read -r -p "    Overwrite it? [y/N]: " OW
    [[ "$OW" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }
    rm -rf "$INSTALL_DIR"
    log "Removed existing installation."
fi

# --- Install dependencies (git, Node.js) -------------------------------------
info "Installing prerequisites (git, Node.js)..."

sudo apt-get update -qq
sudo apt-get install -y -qq git curl build-essential

if ! command -v node &>/dev/null; then
    info "Installing Node.js 20.x..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - > /dev/null 2>&1
    sudo apt-get install -y nodejs > /dev/null 2>&1
fi
log "Node.js $(node --version) / npm $(npm --version)"

# --- Extract archive ---------------------------------------------------------
info "Extracting archive to $HOME ..."
tar -xzf "$ARCHIVE" -C "$HOME"

[[ -d "$INSTALL_DIR" ]] || fail "Extraction complete but $INSTALL_DIR not found. \
Check that the archive contains a folder named '5etools'."
log "Extracted to $INSTALL_DIR"

# --- Rebuild node_modules ----------------------------------------------------
info "Running npm ci to rebuild dependencies..."
cd "$INSTALL_DIR"
npm ci
log "Dependencies installed."

# --- Rebuild service worker if sw source files exist ------------------------
if [[ -f "node/build-sw.mjs" ]]; then
    info "Rebuilding service worker..."
    npm run build:sw:prod
    log "Service worker rebuilt."
fi

# --- Write systemd service ---------------------------------------------------
info "Creating systemd service: $SERVICE_NAME ..."
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

log "Service file written."

# --- Enable and start --------------------------------------------------------
SYSTEMD_ENABLED=false
if systemctl is-system-running &>/dev/null 2>&1; then
    SYSTEMD_ENABLED=true
fi

if $SYSTEMD_ENABLED; then
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    sudo systemctl start "$SERVICE_NAME"
    sleep 2
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log "Service started successfully."
    else
        warn "Service may not have started. Check: sudo journalctl -u $SERVICE_NAME -n 30"
    fi
else
    warn "systemd is not active. To enable it, add to /etc/wsl.conf:"
    warn "  [boot]"
    warn "  systemd=true"
    warn "Then from PowerShell run:  wsl --shutdown  and reopen WSL."
    warn "Manual start:  cd $INSTALL_DIR && npm run serve:dev"
fi

# --- Write update script -----------------------------------------------------
cat > "$HOME/update-5etools.sh" <<'UPDATESCRIPT'
#!/bin/bash
set -e
INSTALL_DIR="$HOME/5etools"
cd "$INSTALL_DIR"
echo "[i] Pulling latest source..."
git pull
[[ -d "img/.git" ]] && { echo "[i] Pulling images..."; cd img && git pull && cd ..; }
echo "[i] Updating dependencies..."
npm ci
[[ -f "sw.js" ]] && { echo "[i] Rebuilding service worker..."; npm run build:sw:prod; }
if systemctl is-system-running &>/dev/null 2>&1; then
    sudo systemctl restart 5etools
    echo "[✔] Updated and restarted."
else
    echo "[✔] Done. Restart manually: cd $INSTALL_DIR && npm run serve:dev"
fi
UPDATESCRIPT
chmod +x "$HOME/update-5etools.sh"

# --- Done --------------------------------------------------------------------
echo ""
echo "============================================================"
log "Unpack complete!"
echo "============================================================"
echo ""
echo -e "  ${GREEN}5etools URL:${NC}  http://localhost:${PORT}/index.html"
echo ""
echo -e "  ${BLUE}Useful commands:${NC}"
echo -e "    Start:    sudo systemctl start $SERVICE_NAME"
echo -e "    Stop:     sudo systemctl stop $SERVICE_NAME"
echo -e "    Logs:     sudo journalctl -u $SERVICE_NAME -f"
echo -e "    Manual:   cd $INSTALL_DIR && npm run serve:dev"
echo -e "    Update:   ~/update-5etools.sh"
echo ""
