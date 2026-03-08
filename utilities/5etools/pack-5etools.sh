#!/bin/bash
# =============================================================================
# 5etools Packer — creates a portable zip for transfer to another machine
# Run this on your SOURCE WSL machine
# Usage:  ./pack-5etools.sh [/path/to/usb/drive]
# =============================================================================

set -e

INSTALL_DIR="$HOME/5etools"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ARCHIVE_NAME="5etools-portable-${TIMESTAMP}.tar.gz"

# --- Determine output destination -------------------------------------------
if [[ -n "$1" ]]; then
    OUTPUT_DIR="$1"
else
    # Try to auto-detect a mounted USB drive (common WSL mount points)
    DETECTED=$(find /mnt -maxdepth 1 -mindepth 1 -type d \
        ! -name 'c' ! -name 'wsl' ! -name 'wslg' ! -name 'wsl2' \
        ! -name 'user-data' 2>/dev/null | head -1)
    if [[ -n "$DETECTED" ]]; then
        OUTPUT_DIR="$DETECTED"
        echo "[i] Auto-detected drive: $OUTPUT_DIR"
    else
        OUTPUT_DIR="$HOME"
        echo "[!] No USB drive detected — saving to $HOME instead."
        echo "    Re-run with path argument:  ./pack-5etools.sh /mnt/e"
    fi
fi

OUTPUT_PATH="${OUTPUT_DIR}/${ARCHIVE_NAME}"

# --- Validate source ---------------------------------------------------------
if [[ ! -d "$INSTALL_DIR" ]]; then
    echo "[✘] 5etools not found at $INSTALL_DIR"
    echo "    Run the installer first, or set INSTALL_DIR at the top of this script."
    exit 1
fi

if [[ ! -w "$OUTPUT_DIR" ]]; then
    echo "[✘] Cannot write to $OUTPUT_DIR — check that the USB drive is mounted and writable."
    exit 1
fi

# --- Show what will be included ----------------------------------------------
echo ""
echo "============================================================"
echo "  5etools Packer"
echo "============================================================"
echo "  Source:      $INSTALL_DIR"
echo "  Output:      $OUTPUT_PATH"
echo ""

# Calculate approximate size (excluding node_modules and .git to save space)
SRC_SIZE=$(du -sh --exclude="$INSTALL_DIR/node_modules" \
                   --exclude="$INSTALL_DIR/.git" \
                   "$INSTALL_DIR" 2>/dev/null | cut -f1)
echo "  Approx size (excl. node_modules & .git): $SRC_SIZE"
echo ""
echo "  node_modules will be EXCLUDED (npm ci will rebuild them on the target)."
echo "  .git history will be EXCLUDED (saves space)."
if [[ -d "$INSTALL_DIR/img" ]]; then
    echo "  Images (img/) WILL be included."
else
    echo "  Images (img/) not present — not included."
fi
echo ""
read -r -p "Proceed? [Y/n]: " CONFIRM
[[ "$CONFIRM" =~ ^[Nn]$ ]] && echo "Aborted." && exit 0

# --- Create archive ----------------------------------------------------------
echo ""
echo "[i] Creating archive (this may take a few minutes)..."

tar -czf "$OUTPUT_PATH" \
    --exclude="$INSTALL_DIR/node_modules" \
    --exclude="$INSTALL_DIR/.git" \
    --exclude="$INSTALL_DIR/img/.git" \
    -C "$(dirname "$INSTALL_DIR")" \
    "$(basename "$INSTALL_DIR")"

SIZE=$(du -sh "$OUTPUT_PATH" | cut -f1)
echo ""
echo "[✔] Archive created: $OUTPUT_PATH"
echo "[✔] Archive size:    $SIZE"
echo ""
echo "Transfer this file to your USB drive (if not already there),"
echo "then run  unpack-5etools.sh  on the target machine."
echo ""
