#!/usr/bin/env python3
"""
World Anvil Export → Obsidian Markdown Converter
=================================================
Reads a full World Anvil export folder and converts all articles
to Obsidian-compatible Markdown, preserving the category hierarchy
as nested folders.

Usage:
    python3 convert.py --input /path/to/World-The_World-151 --output ./obsidian_vault

Defaults:
    --input  : ./input
    --output : ./output
"""

import os
import re
import json
import argparse
from pathlib import Path
from collections import defaultdict


# ── World Anvil markup patterns ──────────────────────────────
WA_MENTIONS   = re.compile(r'@\[([^\]]+)\]\([^:]+:[^)]+\)')
WA_IMG        = re.compile(r'\[img:\d+\]')
WA_CONTAINER  = re.compile(r'\[container[^\]]*\](.*?)\[/container\]', re.DOTALL | re.IGNORECASE)
WA_BLOCK      = re.compile(r'\[[a-z]+[^\]]*\](.*?)\[/[a-z]+\]', re.DOTALL | re.IGNORECASE)
WA_SELF_CLOSE = re.compile(r'\[[^\]]+/\]')
HTML_TAGS     = re.compile(r'<[^>]+>')
MULTI_BLANK   = re.compile(r'\n{3,}')


# ── Section fields per article type ─────────────────────────
SECTION_FIELDS = {
    "article": [
        ("content", None), ("sidebarcontent", "Sidebar"),
        ("footnotes", "Footnotes"), ("authornotes", "Author Notes"),
    ],
    "document": [
        ("content", None), ("sidebarcontent", "Sidebar"),
        ("footnotes", "Footnotes"), ("authornotes", "Author Notes"),
    ],
    "person": [
        ("content", None),
        ("currentstatus", "Current Status"), ("history", "History"),
        ("physique", "Physical Description"), ("bodyFeatures", "Body Features"),
        ("facialFeatures", "Facial Features"),
        ("identifyingCharacteristics", "Identifying Characteristics"),
        ("clothing", "Clothing"), ("items", "Items"),
        ("specialAbilities", "Special Abilities"),
        ("intellectualCharacteristics", "Intellectual Characteristics"),
        ("morality", "Morality"), ("motivation", "Motivation"),
        ("virtues", "Virtues"), ("vices", "Vices"),
        ("likesDislikes", "Likes & Dislikes"),
        ("quirksPersonality", "Personality Quirks"),
        ("quirksPhysical", "Physical Quirks"),
        ("education", "Education"), ("employment", "Employment"),
        ("achievements", "Achievements"), ("failures", "Failures"),
        ("mentalTraumas", "Mental Traumas"), ("relations", "Relations"),
        ("family", "Family"), ("religion", "Religion"),
        ("languages", "Languages"), ("quotes", "Quotes"),
        ("wealth", "Wealth"), ("goals", "Goals"),
        ("hobbies", "Hobbies"), ("speech", "Speech"),
        ("mannerisms", "Mannerisms"), ("socialAptitude", "Social Aptitude"),
        ("sidebarcontent", "Sidebar"), ("footnotes", "Footnotes"),
        ("authornotes", "Author Notes"),
    ],
    "settlement": [
        ("content", None), ("geography", "Geography"),
        ("naturalresources", "Natural Resources"), ("history", "History"),
        ("demographics", "Demographics"), ("government", "Government"),
        ("infrastructure", "Infrastructure"), ("guilds", "Guilds & Factions"),
        ("tourism", "Tourism"), ("industry", "Industry"),
        ("architecture", "Architecture"), ("defences", "Defences"),
        ("pointOfInterest", "Points of Interest"), ("district", "Districts"),
        ("climate", "Climate"), ("florafauna", "Flora & Fauna"),
        ("assets", "Assets"), ("sidebarcontent", "Sidebar"),
        ("footnotes", "Footnotes"), ("authornotes", "Author Notes"),
    ],
    "landmark": [
        ("content", None), ("purpose", "Purpose"), ("design", "Design"),
        ("history", "History"), ("geography", "Geography"),
        ("naturalresources", "Natural Resources"), ("denizens", "Denizens"),
        ("valuables", "Valuables"), ("hazards", "Hazards"),
        ("effects", "Effects"), ("sensory", "Sensory Details"),
        ("properties", "Properties"), ("contents", "Contents"),
        ("pointOfInterest", "Points of Interest"), ("climate", "Climate"),
        ("florafauna", "Flora & Fauna"), ("assets", "Assets"),
        ("sidebarcontent", "Sidebar"), ("footnotes", "Footnotes"),
        ("authornotes", "Author Notes"),
    ],
    "location": [
        ("content", None), ("geography", "Geography"),
        ("naturalresources", "Natural Resources"), ("history", "History"),
        ("demographics", "Demographics"), ("florafauna", "Flora & Fauna"),
        ("climate", "Climate"), ("ecosystem", "Ecosystem"),
        ("ecosystemCycles", "Ecosystem Cycles"),
        ("localizedPhenomena", "Localized Phenomena"),
        ("alterations", "Alterations"), ("assets", "Assets"),
        ("sidebarcontent", "Sidebar"), ("footnotes", "Footnotes"),
        ("authornotes", "Author Notes"),
    ],
    "organization": [
        ("content", None), ("publicAgenda", "Public Agenda"),
        ("history", "History"), ("structure", "Structure"),
        ("origins", "Origins"), ("assets", "Assets"),
        ("demographics", "Demographics"), ("territory", "Territory"),
        ("military", "Military"), ("technology", "Technology"),
        ("foreignrelations", "Foreign Relations"), ("laws", "Laws"),
        ("culture", "Culture"), ("religion", "Religion"),
        ("education", "Education"), ("infrastructure", "Infrastructure"),
        ("imports", "Imports"), ("exports", "Exports"),
        ("agricultureAndIndustry", "Agriculture & Industry"),
        ("tradeAndTransport", "Trade & Transport"),
        ("mythos", "Mythos"), ("intrigue", "Intrigue"),
        ("disbandment", "Disbandment"),
        ("sidebarcontent", "Sidebar"), ("footnotes", "Footnotes"),
        ("authornotes", "Author Notes"),
    ],
    "species": [
        ("content", None), ("anatomy", "Anatomy"),
        ("perception", "Perception & Sensory"), ("genetics", "Genetics"),
        ("ecology", "Ecology"), ("diet", "Diet"),
        ("biologicalCycle", "Biological Cycle"), ("behaviour", "Behaviour"),
        ("civilization", "Civilization & Culture"),
        ("namingTraditions", "Naming Traditions"),
        ("beauty", "Beauty Ideals"), ("genderIdeals", "Gender Ideals"),
        ("courtship", "Courtship Ideals"),
        ("relationship", "Relationship Ideals"),
        ("averageHeight", "Average Height"), ("averageWeight", "Average Weight"),
        ("history", "History"), ("sidebarcontent", "Sidebar"),
        ("footnotes", "Footnotes"), ("authornotes", "Author Notes"),
    ],
    "plot": [
        ("content", None), ("history", "History"),
        ("sidebarcontent", "Sidebar"), ("footnotes", "Footnotes"),
        ("authornotes", "Author Notes"),
    ],
    "report": [
        ("content", None), ("sidebarcontent", "Sidebar"),
        ("footnotes", "Footnotes"), ("authornotes", "Author Notes"),
    ],
}

PERSON_META = [
    ("firstname", "First Name"), ("lastname", "Last Name"),
    ("honorific", "Honorific"), ("suffix", "Suffix"),
    ("nickname", "Nickname"), ("pronouns", "Pronouns"),
    ("gender", "Gender"), ("age", "Age"),
    ("eyes", "Eyes"), ("hair", "Hair"), ("skin", "Skin"),
    ("height", "Height"), ("weight", "Weight"),
    ("rpgAlignment", "Alignment"), ("deity", "Deity"),
]

SETTLEMENT_META = [
    ("population", "Population"), ("areaSize", "Area Size"),
    ("demonym", "Demonym"), ("alternativename", "Alternative Name"),
    ("constructed", "Constructed"), ("ruined", "Ruined"),
]

ORGANIZATION_META = [
    ("foundingDate", "Founded"), ("dissolutionDate", "Dissolved"),
    ("alternativeNames", "Alternative Names"),
    ("demonym", "Demonym"), ("motto", "Motto"),
]

RELATIONSHIP_FIELDS = [
    ("parent", "Parent"), ("organization", "Organization"),
    ("rulingorganization", "Ruling Organization"),
    ("leader", "Leader"), ("headofstate", "Head of State"),
    ("headofgovernment", "Head of Government"),
    ("capital", "Capital"), ("geographicLocation", "Geographic Location"),
    ("species", "Species"), ("ethnicity", "Ethnicity"),
    ("currentLocation", "Current Location"),
    ("birthplace", "Birthplace"), ("residence", "Residence"),
    ("church", "Church"), ("realm", "Realm"),
    ("familyorganization", "Family Organization"),
    ("vehicle", "Vehicle"), ("rank", "Rank"),
    ("formation", "Formation"), ("statereligion", "State Religion"),
    ("articleParent", "Parent Article"),
]


# ── Helpers ──────────────────────────────────────────────────

def load_json(path):
    try:
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)
    except Exception:
        return None


def safe_filename(name):
    name = re.sub(r'[<>:"/\\|?*]', '', name)
    name = name.strip().replace(' ', '_')
    return name or "untitled"


def format_date(date_obj):
    if not date_obj or not isinstance(date_obj, dict):
        return None
    raw = date_obj.get("date", "")
    return raw[:10] if raw else None


def get_linked_title(obj):
    if not obj or not isinstance(obj, dict):
        return None
    return obj.get("title")


def clean_wa_content(text):
    if not text:
        return ""
    text = WA_MENTIONS.sub(r'[[\1]]', text)
    text = WA_IMG.sub('`[image]`', text)
    text = WA_CONTAINER.sub(r'\1', text)
    text = WA_BLOCK.sub(r'\1', text)
    text = WA_SELF_CLOSE.sub('', text)
    text = HTML_TAGS.sub('', text)
    text = text.replace('\r\n', '\n').replace('\r', '\n')
    text = MULTI_BLANK.sub('\n\n', text)
    return text.strip()


# ── Category tree ────────────────────────────────────────────

def build_category_tree(categories_dir):
    """
    Load all category JSONs and return a mapping of
    article_id -> nested folder Path based on the real hierarchy.
    e.g. World Atlas / The Northwest / Icewind Dale
    """
    cat_files = list(Path(categories_dir).glob("*.json"))
    cats = []
    for f in cat_files:
        data = load_json(f)
        if data:
            cats.append(data)

    cat_by_id = {c["id"]: c for c in cats if "id" in c}

    path_memo = {}

    def get_path(cat_id):
        if cat_id in path_memo:
            return path_memo[cat_id]
        cat = cat_by_id.get(cat_id)
        if not cat:
            return Path("Uncategorised")
        parent = cat.get("parent")
        pid = parent.get("id") if isinstance(parent, dict) else None
        title = safe_filename(cat.get("title", "Unknown"))
        if pid and pid in cat_by_id:
            path = get_path(pid) / title
        else:
            path = Path(title)
        path_memo[cat_id] = path
        return path

    # Map article_id → folder path
    article_to_path = {}
    for cat in cats:
        cat_id   = cat.get("id")
        articles = cat.get("articles") or []
        folder   = get_path(cat_id)
        for art in articles:
            aid = art.get("id") if isinstance(art, dict) else None
            if aid:
                article_to_path[aid] = folder

    return article_to_path


# ── Markdown builders ────────────────────────────────────────

def build_frontmatter(article, category_path=None):
    lines = ["---"]
    title = article.get("title", "Untitled")
    lines.append(f'title: "{title}"')

    entity_class = article.get("entityClass", "")
    if entity_class:
        lines.append(f"type: {entity_class}")

    if category_path:
        cat_str = str(category_path).replace(os.sep, " / ")
        lines.append(f'category: "{cat_str}"')
    else:
        cat = get_linked_title(article.get("category"))
        if cat:
            lines.append(f'category: "{cat}"')

    tags_raw = article.get("tags") or ""
    if tags_raw and tags_raw.strip():
        tag_list = [t.strip() for t in re.split(r'[,;]', tags_raw) if t.strip()]
        if tag_list:
            lines.append("tags:")
            for tag in tag_list:
                lines.append(f"  - {tag}")

    state = article.get("state")
    if state:
        lines.append(f"status: {state}")
    if article.get("isWip"):
        lines.append("wip: true")

    author = get_linked_title(article.get("author"))
    if author:
        lines.append(f"author: {author}")

    created = format_date(article.get("creationDate"))
    if created:
        lines.append(f"created: {created}")

    updated = format_date(article.get("updateDate"))
    if updated:
        lines.append(f"updated: {updated}")

    wa_url = article.get("url")
    if wa_url:
        lines.append(f'source: "{wa_url}"')

    lines.append("---")
    return "\n".join(lines)


def build_meta_table(article, template_type):
    if template_type == "person":
        fields = PERSON_META
    elif template_type in ("settlement", "landmark", "location"):
        fields = SETTLEMENT_META
    elif template_type == "organization":
        fields = ORGANIZATION_META
    else:
        return ""

    rows = []
    for key, label in fields:
        val = article.get(key)
        if val and isinstance(val, dict):
            val = get_linked_title(val)
        if val:
            rows.append(f"| **{label}** | {val} |")

    if not rows:
        return ""
    return "\n## Details\n\n| Field | Value |\n|---|---|\n" + "\n".join(rows)


def build_relationships(article):
    links = []
    for key, label in RELATIONSHIP_FIELDS:
        val = article.get(key)
        title = get_linked_title(val)
        if title:
            links.append(f"- **{label}:** [[{title}]]")
    if not links:
        return ""
    return "\n## Relationships\n\n" + "\n".join(links)


def build_cover_image(article):
    cover = article.get("cover")
    if not cover or not isinstance(cover, dict):
        return ""
    url   = cover.get("url", "")
    title = cover.get("title", "")
    if "WorldCover_Default" in url or title == "Default Cover":
        return ""
    return f"\n![{title}]({url})\n" if url else ""


def build_portrait(article):
    portrait = article.get("portrait")
    if not portrait or not isinstance(portrait, dict):
        return ""
    url   = portrait.get("url", "")
    title = portrait.get("title", article.get("title", "Portrait"))
    return f"\n![{title}]({url})\n" if url else ""


def build_content_sections(article, template_type):
    field_list = SECTION_FIELDS.get(template_type, [
        ("content", None), ("history", "History"),
        ("sidebarcontent", "Sidebar"), ("footnotes", "Footnotes"),
        ("authornotes", "Author Notes"),
    ])

    sections = []
    for key, label in field_list:
        raw = article.get(key)
        if not raw or not str(raw).strip():
            continue
        text = clean_wa_content(str(raw))
        if not text:
            continue
        sections.append(f"\n{text}" if label is None else f"\n## {label}\n\n{text}")

    return "\n".join(sections)


def article_to_markdown(article, category_path=None):
    title         = article.get("title", "Untitled")
    template_type = (article.get("templateType") or article.get("entityClass") or "article").lower()

    parts = []
    parts.append(build_frontmatter(article, category_path))
    parts.append(f"\n# {title}")

    subheading = article.get("subheading") or article.get("pronunciation")
    if subheading:
        parts.append(f"\n*{subheading.strip()}*\n")

    if template_type == "person":
        parts.append(build_portrait(article))
    parts.append(build_cover_image(article))

    excerpt = article.get("excerpt") or article.get("snippet")
    if excerpt:
        cleaned = clean_wa_content(excerpt)
        if cleaned:
            parts.append(f"\n> {cleaned}\n")

    parts.append(build_meta_table(article, template_type))
    parts.append(build_relationships(article))
    parts.append(build_content_sections(article, template_type))

    wa_url = article.get("url")
    if wa_url:
        parts.append(f"\n\n---\n[View on World Anvil]({wa_url})")

    return "\n".join(p for p in parts if p)


# ── Main ─────────────────────────────────────────────────────

def convert_all(input_dir, output_dir):
    input_path  = Path(input_dir)
    output_path = Path(output_dir)

    articles_dir   = input_path / "articles"
    categories_dir = input_path / "categories"

    if articles_dir.exists():
        json_files = list(articles_dir.glob("*.json"))
        print(f"📂 Detected full World Anvil export folder")
    else:
        json_files = list(input_path.glob("*.json"))
        print(f"📂 Detected flat JSON folder (no category hierarchy)")
        categories_dir = None

    if not json_files:
        print(f"⚠️  No JSON files found")
        return

    print(f"🔍 Found {len(json_files)} article files")

    article_to_path = {}
    if categories_dir and categories_dir.exists():
        print(f"🗂️  Building category hierarchy...")
        article_to_path = build_category_tree(categories_dir)
        print(f"   Mapped {len(article_to_path)} articles to category paths")

    converted   = 0
    skipped     = 0
    path_counts = defaultdict(int)

    for json_file in sorted(json_files):
        data = load_json(json_file)
        if not data or not data.get("success"):
            skipped += 1
            continue

        article_id    = data.get("id")
        title         = data.get("title", "Untitled")
        category_path = article_to_path.get(article_id)

        if category_path is None:
            # Fallback: use entityClass as folder
            category_path = Path(data.get("entityClass", "Uncategorised"))

        out_dir  = output_path / category_path
        out_dir.mkdir(parents=True, exist_ok=True)
        out_file = out_dir / (safe_filename(title) + ".md")

        md = article_to_markdown(data, category_path)
        with open(out_file, "w", encoding="utf-8") as f:
            f.write(md)

        path_counts[str(category_path)] += 1
        converted += 1

    print(f"\n✅ Converted {converted} articles ({skipped} skipped)\n")
    print(f"📁 Folder structure:\n")
    for path, count in sorted(path_counts.items()):
        depth  = path.count(os.sep)
        indent = "   " + "  " * depth
        label  = path.split(os.sep)[-1]
        print(f"{indent}{label:<40} {count} article(s)")

    print(f"\n📂 Output: {output_path.resolve()}")


def main():
    parser = argparse.ArgumentParser(description="Convert World Anvil JSON export to Obsidian Markdown")
    parser.add_argument("--input",  default="input",  help="WA export root folder (default: ./input)")
    parser.add_argument("--output", default="output", help="Output vault folder (default: ./output)")
    args = parser.parse_args()

    print("=" * 55)
    print("  World Anvil → Obsidian Converter")
    print("=" * 55 + "\n")

    convert_all(args.input, args.output)


if __name__ == "__main__":
    main()
