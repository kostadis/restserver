#!/usr/bin/env python3
"""
inspect_export.py — Inspect a World Anvil export folder structure
Usage: python3 inspect_export.py /path/to/World-The_World-151
"""

import os
import json
import sys
from pathlib import Path
from collections import defaultdict


def load_json(path):
    try:
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)
    except Exception as e:
        return None


def inspect(root):
    root = Path(root)
    if not root.exists():
        print(f"❌ Path not found: {root}")
        sys.exit(1)

    print("=" * 55)
    print(f"  World Anvil Export Inspector")
    print(f"  {root.resolve()}")
    print("=" * 55)

    # ── Top-level structure ──────────────────────────────────
    print("\n📁 Top-level contents:\n")
    for item in sorted(root.iterdir()):
        if item.is_dir():
            count = sum(1 for _ in item.rglob("*.json"))
            print(f"   [DIR]  {item.name:<30} {count} JSON file(s)")
        elif item.suffix in (".json", ".html", ".css"):
            size = item.stat().st_size
            print(f"   [FILE] {item.name:<30} {size} bytes")

    # ── Article types in articles/ ───────────────────────────
    articles_dir = root / "articles"
    if articles_dir.exists():
        print("\n\n📄 Entity types found in articles/:\n")
        type_counts = defaultdict(int)
        for f in sorted(articles_dir.glob("*.json")):
            # Prefix is everything before the last hyphen-segment
            parts = f.stem.rsplit("-", 1)
            prefix = parts[0] if len(parts) == 2 else f.stem
            # Extract entity type (first word)
            entity_type = prefix.split("-")[0]
            type_counts[entity_type] += 1

        for etype, count in sorted(type_counts.items()):
            print(f"   {etype:<30} {count} file(s)")

    # ── Categories ───────────────────────────────────────────
    categories_dir = root / "categories"
    if categories_dir.exists():
        cat_files = list(categories_dir.glob("*.json"))
        print(f"\n\n🗂️  Categories ({len(cat_files)} found):\n")
        categories = []
        for f in sorted(cat_files):
            data = load_json(f)
            if data:
                categories.append(data)

        # Try to build hierarchy
        cat_by_id = {c.get("id"): c for c in categories}
        roots = []
        for cat in categories:
            parent = cat.get("parent")
            parent_id = parent.get("id") if isinstance(parent, dict) else None
            if not parent_id or parent_id not in cat_by_id:
                roots.append(cat)

        def print_tree(cat, indent=0):
            title = cat.get("title", "?")
            cid   = cat.get("id", "")[:8]
            print(f"   {'  ' * indent}{'└─ ' if indent else ''}{title}  [{cid}...]")
            # Find children
            for c in categories:
                parent = c.get("parent")
                pid = parent.get("id") if isinstance(parent, dict) else None
                if pid == cat.get("id"):
                    print_tree(c, indent + 1)

        for r in sorted(roots, key=lambda x: x.get("title", "")):
            print_tree(r)

    # ── Maps ─────────────────────────────────────────────────
    maps_dir = root / "maps"
    if maps_dir.exists():
        map_files = list(maps_dir.rglob("*.json"))
        print(f"\n\n🗺️  Maps ({len(map_files)} found):\n")
        for f in sorted(map_files)[:10]:
            data = load_json(f)
            title = data.get("title", f.stem) if data else f.stem
            print(f"   {title}")
        if len(map_files) > 10:
            print(f"   ... and {len(map_files) - 10} more")

    # ── Secrets ──────────────────────────────────────────────
    secrets_dir = root / "secrets"
    if secrets_dir.exists():
        count = sum(1 for _ in secrets_dir.glob("*.json"))
        print(f"\n\n🔒 Secrets: {count} file(s)")

    # ── Notebooks ────────────────────────────────────────────
    notebooks_dir = root / "notebooks"
    if notebooks_dir.exists():
        count = sum(1 for _ in notebooks_dir.rglob("*.json"))
        print(f"\n📓 Notebooks: {count} file(s)")

    # ── Sample category JSON ─────────────────────────────────
    if categories_dir.exists():
        cat_files = list(categories_dir.glob("*.json"))
        if cat_files:
            print(f"\n\n🔍 Sample category file ({cat_files[0].name}):\n")
            data = load_json(cat_files[0])
            if data:
                # Print just the keys and non-null values
                for k, v in data.items():
                    if v is not None and v != "" and v != []:
                        if isinstance(v, dict):
                            print(f"   {k}: {{title: {v.get('title', '?')}, id: {str(v.get('id',''))[:8]}...}}")
                        else:
                            print(f"   {k}: {str(v)[:80]}")

    print("\n" + "=" * 55 + "\n")


if __name__ == "__main__":
    path = sys.argv[1] if len(sys.argv) > 1 else "."
    inspect(path)
