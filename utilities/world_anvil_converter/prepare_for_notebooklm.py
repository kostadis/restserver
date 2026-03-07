#!/usr/bin/env python3
"""
prepare_for_notebooklm.py
=========================
Copies all converted Markdown files into a single flat folder
so you can select them all at once for NotebookLLM upload.

Also prints a summary of what's included.

Usage:
    python3 prepare_for_notebooklm.py --input ./output --flat ./notebooklm_upload
"""

import os
import re
import shutil
import argparse
from pathlib import Path
from collections import defaultdict


def safe_flat_name(original_path, root):
    """
    Convert a nested path like:
      World_Atlas/The_Northwest/Icewind_Dale/Korkolohk.md
    into a flat filename:
      World_Atlas__The_Northwest__Icewind_Dale__Korkolohk.md
    so there are no collisions in the flat folder.
    """
    rel = Path(original_path).relative_to(root)
    parts = list(rel.parts)
    return "__".join(parts)


def prepare(input_dir, flat_dir):
    input_path = Path(input_dir)
    flat_path  = Path(flat_dir)

    if not input_path.exists():
        print(f"❌ Input folder not found: {input_dir}")
        return

    md_files = sorted(input_path.rglob("*.md"))

    if not md_files:
        print(f"⚠️  No .md files found in {input_dir}")
        return

    flat_path.mkdir(parents=True, exist_ok=True)

    print("=" * 55)
    print("  Prepare for NotebookLLM Upload")
    print("=" * 55)
    print(f"\n📂 Source : {input_path.resolve()}")
    print(f"📁 Output : {flat_path.resolve()}")
    print(f"\n📄 Copying {len(md_files)} files...\n")

    copied     = 0
    collisions = defaultdict(int)

    for md_file in md_files:
        flat_name = safe_flat_name(md_file, input_path)
        dest      = flat_path / flat_name

        # Handle any unlikely name collisions
        if dest.exists():
            collisions[flat_name] += 1
            stem, suffix = flat_name.rsplit(".", 1)
            flat_name = f"{stem}_{collisions[flat_name]}.{suffix}"
            dest = flat_path / flat_name

        shutil.copy2(md_file, dest)
        copied += 1

    print(f"✅ {copied} files copied to:\n")
    print(f"   {flat_path.resolve()}\n")
    print("─" * 55)
    print("\n📋 File list:\n")

    for md_file in sorted(flat_path.glob("*.md")):
        size_kb = md_file.stat().st_size / 1024
        print(f"   {md_file.name:<70}  {size_kb:.1f} KB")

    print(f"\n─" + "─" * 54)
    print(f"\n💡 To upload to NotebookLLM:")
    print(f"   1. Open NotebookLLM and create/open your notebook")
    print(f"   2. Click 'Add Source'")
    print(f"   3. Navigate to:  {flat_path.resolve()}")
    print(f"   4. Press Ctrl+A to select all files, then upload\n")


def main():
    parser = argparse.ArgumentParser(description="Flatten Obsidian output for NotebookLLM upload")
    parser.add_argument("--input", default="output",            help="Converted Markdown folder (default: ./output)")
    parser.add_argument("--flat",  default="notebooklm_upload", help="Flat output folder (default: ./notebooklm_upload)")
    args = parser.parse_args()

    prepare(args.input, args.flat)


if __name__ == "__main__":
    main()
