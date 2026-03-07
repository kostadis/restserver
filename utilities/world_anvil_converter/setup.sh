#!/bin/bash
# setup.sh — Set up the virtual environment (run once)

set -e

echo "=================================="
echo "  WA → Obsidian Converter Setup"
echo "=================================="

if ! command -v python3 &> /dev/null; then
    echo "Installing Python3..."
    sudo apt update && sudo apt install python3 python3-venv -y
fi

if ! python3 -m venv --help &> /dev/null; then
    sudo apt install python3-venv -y
fi

if [ -d "venv" ]; then
    echo "⚠️  Removing existing venv..."
    rm -rf venv
fi

echo "🔧 Creating virtual environment..."
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip --quiet
echo "✅ Setup complete — no extra packages needed (uses Python stdlib only)"

echo ""
echo "=================================="
echo "  Ready!"
echo "=================================="
echo ""
echo "Usage:"
echo "  1. Put your JSON files in the ./input folder"
echo "  2. Run: bash run.sh"
echo "  3. Find your .md files in ./output"
echo ""