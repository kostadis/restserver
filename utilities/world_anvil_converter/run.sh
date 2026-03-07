#!/bin/bash
# run.sh — Convert all JSON files in ./input to Markdown in ./output

set -e

if [ ! -d "venv" ]; then
    echo "⚠️  Virtual environment not found. Running setup first..."
    bash setup.sh
fi

source venv/bin/activate

# Allow custom input/output via args, default to ./input and ./output
INPUT=${1:-input}
OUTPUT=${2:-output}

mkdir -p "$INPUT" "$OUTPUT"

python3 convert.py --input "$INPUT" --output "$OUTPUT"
