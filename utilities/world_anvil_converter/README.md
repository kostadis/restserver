# World Anvil → Obsidian Converter

Converts World Anvil JSON article exports into clean Obsidian-compatible Markdown files, organised into folders by category.

---

## Setup (once)

```bash
bash setup.sh
```

> No extra Python packages required — uses the standard library only.

---

## Usage

1. Copy all your World Anvil JSON files into the `input/` folder
2. Run the converter:

```bash
bash run.sh
```

3. Find your Markdown files in `output/`, organised by category:

```
output/
  Player_Characters/
    Felkur.md
    Daien.md
  Organizations/
    La_Resistance.md
  Icewind_Dale/
    Korkolohk.md
  The_Northwest/
    Hightop_Peak.md
  ...
```

### Custom folders

```bash
bash run.sh /path/to/my/json /path/to/my/vault
```

---

## What gets converted

Each Markdown file includes:

- **YAML frontmatter** — title, type, category, tags, status, author, dates, source URL
- **Obsidian internal links** — `@[Karl](person:id)` becomes `[[Karl]]`
- **Cover & portrait images** — linked inline (non-default only)
- **All content sections** — mapped per article type (Person, Settlement, Organization, Landmark, etc.)
- **Metadata table** — structured fields like population, alignment, pronouns, founding date, etc.
- **Relationships** — linked entities (parent, leader, organization, location, etc.) as `[[WikiLinks]]`
- **World Anvil source link** — footer link back to the original article

---

## Supported article types

| Type           | Sections rendered                                            |
| -------------- | ------------------------------------------------------------ |
| Person         | Overview, history, physical description, personality, relationships, and 30+ more fields |
| Settlement     | Overview, geography, government, history, demographics, and more |
| Landmark       | Overview, purpose, design, hazards, denizens, and more       |
| Organization   | Overview, history, structure, military, foreign relations, and more |
| Article        | Content, sidebar, footnotes, author notes                    |
| Any other type | All non-null text fields rendered automatically              |

---

## Importing into Obsidian

1. Open Obsidian
2. Open your vault (or create a new one pointing to the `output/` folder)
3. All articles appear as linked notes — `[[WikiLinks]]` will resolve automatically if the referenced article exists in the vault

---

## Tips for lore generation in NotebookLLM

If you also want to use these in NotebookLLM, you can upload the `.md` files directly — NotebookLLM accepts Markdown. No need to convert to PDF separately.