# Git Config Manager — Visual Editor Design Plan

## Overview

A desktop application for managing `.gitconfig` files through a visual drag-and-drop
block editor. The goal is to make git config discoverable and editable without needing
to know the syntax, directive names, or file format — while still teaching the user
the real structure rather than hiding it.

Built in Go using the Fyne v2 GUI framework.

---

## UI Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Block Library (left)  │  Global Config (~/.gitconfig)  │  Project Config   │
│  ────────────────────  │  ──────────────────────────────│  ─────────────────│
│  Boundaries            │                                │  [Open Project…]  │
│    ⚡ Conditional      │  📁 Root (~/.gitconfig)        │  [recent repos]   │
│    📄 File             │  ┌──────────────────────────┐  │                   │
│                        │  │ 👤 User            [≡][×]│  │                   │
│  Sections              │  │  name  → KnightYoshi     │  │                   │
│    👤 User             │  │  email → 6134576+…       │  │                   │
│    🔧 Core             │  │  [+ Add field ▾]         │  │                   │
│    🔗 URL              │  └──────────────────────────┘  │                   │
│    📦 Filter           │  ┌──────────────────────────┐  │                   │
│    🏷️  Alias           │  │ ⚡ Conditional     [≡][×]│  │                   │
│    🌐 HTTP             │  │  When [gitdir▾] [~/work/]│  │                   │
│    ⬆️  Push            │  │  When [gitdir▾] [~/w/**] │  │                   │
│    ⬇️  Pull            │  │  [+ condition]           │  │                   │
│    🔄 Fetch            │  │  ┌────────────────────┐  │  │                   │
│    🌿 Branch           │  │  │ 📄 ~/.gitconfig_w  │  │  │                   │
│    🔀 Merge            │  │  │ ┌────────────────┐ │  │  │                   │
│    📊 Diff             │  │  │ │ 👤 User  [≡][×]│ │  │  │                   │
│    📝 Commit           │  │  │ │  name → …      │ │  │  │                   │
│    🏠 Remote           │  │  │ └────────────────┘ │  │  │                   │
│    🔁 Rebase           │  │  └────────────────────┘  │  │                   │
│    🎨 Color            │  └──────────────────────────┘  │                   │
│    🏷️  Tag             │                                │                   │
│    🔑 Credential       │  [+ Add Section to Root]       │                   │
│    📦 Submodule        │                                │                   │
│    ⚙️  Custom          │                                │                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Block Types

### Conditional Boundary (`[includeIf "condition"]`)

- Container block — accepts **File boundary only** as a child drop target
- Has one or more condition rows (OR semantics)
- Each condition row: `[condition type ▾]` `[value text field]`
- Condition types (dropdown): `gitdir`, `gitdir/i`, `onbranch`, `hasconfig`, `worktree`
- Multiple rows → multiple `[includeIf "..."]` sections written on save, all pointing
  to the same file
- `[+ Add condition]` button appends another row

Serialises as:
```
[includeIf "gitdir:~/work/"]
    path = ~/.gitconfig_work
[includeIf "gitdir:~/work/**"]
    path = ~/.gitconfig_work
```

### File Boundary (`[include]` / `[includeIf]` target)

- Container block — accepts **Sections, Conditional boundary, File boundary** as children
- Has a filename text field on the block header (e.g. `~/.gitconfig_work`)
- When at root level with no conditional parent → serialises as `[include path=...]`
  (unconditional, always loaded)
- When inside a conditional boundary → becomes the `path = ...` value of that
  `[includeIf]` section, with content written to the named file

Labelled differently based on context:
```
📄 Always load: ~/.gitconfig_users    ← unconditional [include]
📄 ~/.gitconfig_work                  ← inside ⚡, target of [includeIf]
```

### Section Blocks

- Leaf containers — accept no child blocks, only contain typed field rows
- Each section type has a known field schema (see Schema section below)
- Fields shown with the appropriate input widget (text, toggle, dropdown)
- `[+ Add field ▾]` shows a picker listing only valid fields for that section type
- Sections with subsections (`url`, `filter`, `remote`, `branch`, `credential`,
  `submodule`) have an editable subsection name field on the block header

### Custom Block (fallback)

- Used for any section name not in the known schema
- Shows section name (editable) and a plain key=value list
- `[+ Add entry]` adds a free-text key + value row
- Preserves unknown sections from existing configs without data loss

---

## Drag-and-Drop Rules

| Block being dragged | Valid drop targets |
|---|---|
| Section | Root config, File boundary, Conditional boundary |
| File boundary | Root config, Conditional boundary, File boundary |
| Conditional boundary | Root config, File boundary |

- Valid drop targets **glow** when a drag enters them
- Invalid drop targets stay dim — Fyne simply doesn't implement `DropTarget` on them
  for the dragged block type, so no drop is possible
- Blocks can be reordered within the same container via a `≡` drag handle
- Cross-panel drag (global ↔ project) = **move** semantics (removed from source,
  added to destination)
- To copy a block, use cmd+C / cmd+V

---

## Two-Panel Config Layout

### Global Panel

- Always shows `~/.gitconfig` as the implicit root container
- Loads `~/.gitconfig` on startup and reconstructs the visual block tree from it
- Included files (`[include]` / `[includeIf]`) are shown as nested File boundary
  blocks; their content is loaded and displayed inline

### Project Panel

- Starts empty with an **"Open Project…"** button and a recent repos list (last 5)
- Selecting a repo reads `.git/config` from `<repo>/.git/config`
- `.git/config` is shown as its own root container, independent of the global tree
- This file is **not** referenced by `~/.gitconfig` — git loads it automatically
  based on the current directory

### Superseded Key Highlighting

- When a key appears in both global and project configs, the global key is **grayed
  out** to indicate it is overridden
- Granularity is **key-level** (not section-level): only the specific overridden key
  is dimmed, not the whole section
- Example: if project sets `user.email` but not `user.name`, only `email` is grayed
  in the global `[user]` block

---

## Comments

- Comment lines (`# ...`, `; ...`) and blank lines are preserved in memory and written
  back to disk on save — the parser stores them as `Entry{Key: "", RawLine: "..."}` 
- Comments are **not shown** in the visual editor
- This means a commented-out line survives round-trip edits without appearing in the UI

---

## Known Section Schemas (`internal/schema/`)

Each section type in the schema defines:
- Its canonical name (e.g. `"user"`)
- Whether it takes a subsection (e.g. `url "pattern"`)
- Its fields, each with:
  - Key name
  - Display label
  - Input type: `text | toggle | dropdown | multiline`
  - For dropdowns: the list of valid values
  - Whether the field is commonly used (affects display order)

Sections covered by the schema:

| Section | Subsection? | Notable fields |
|---|---|---|
| `user` | No | name, email, signingkey |
| `core` | No | autocrlf (dropdown), editor, fileMode (toggle), ignorecase (toggle), … |
| `url` | Yes (target URL) | insteadOf, pushInsteadOf |
| `filter` | Yes (filter name) | process, required (toggle), clean, smudge |
| `alias` | No | free key=value (custom command names) |
| `http` | No | proxy, sslVerify (toggle), cookieFile, … |
| `push` | No | default (dropdown), followTags (toggle), … |
| `pull` | No | rebase (dropdown: true/false/merges), ff (dropdown) |
| `fetch` | No | prune (toggle), pruneTags (toggle), … |
| `branch` | Yes (branch name) | remote, merge, rebase (dropdown) |
| `merge` | No | tool, conflictstyle (dropdown), ff (dropdown) |
| `diff` | No | tool, algorithm (dropdown), renames (toggle) |
| `commit` | No | template, gpgSign (toggle), verbose (toggle) |
| `remote` | Yes (remote name) | url, fetch, pushurl, tagOpt (dropdown) |
| `rebase` | No | autoSquash (toggle), autoStash (toggle), … |
| `color` | No | ui (dropdown: true/false/always/auto) |
| `tag` | No | gpgSign (toggle), sort |
| `credential` | Yes (URL pattern) | helper, username |
| `submodule` | Yes (name) | path, url, update (dropdown), branch |
| `init` | No | defaultBranch |
| `gpg` | No | format (dropdown: openpgp/x509/ssh), program |

Anything not in this list → **Custom block**.

---

## Implementation Phases

### Phase 1 — `internal/schema/`
Define the field schema for all known sections. This is the data that drives which
fields the `[+ Add field]` picker shows and which input widget each field renders as.

Files: `internal/schema/schema.go`, `internal/schema/sections.go`

### Phase 2 — Block widgets (`internal/ui/blocks/`)
Custom Fyne widgets for each block type. Each widget:
- Renders its own header (title, drag handle `≡`, remove button `×`)
- Implements `fyne.DropTarget` for the block types it accepts
- Implements `fyne.Draggable` via its `≡` handle for reordering
- Emits callbacks when its content changes (for dirty-state tracking)

Files: `blocks/conditional.go`, `blocks/file.go`, `blocks/section.go`,
`blocks/custom.go`, `blocks/field_row.go`

### Phase 3 — Canvas (`internal/ui/canvas/`)
The scrollable root container. Manages:
- The ordered list of top-level blocks
- Drop-zone insertion between existing blocks (above/below midpoint detection)
- Visual feedback (highlight valid targets, dim invalid)
- Undo/redo stack (stretch goal)

Files: `canvas/canvas.go`, `canvas/drop_zone.go`

### Phase 4 — Block library panel (`internal/ui/library/`)
Left-side panel with draggable block templates, grouped into Boundaries and Sections.
Dragging from the library creates a new block instance on the canvas.

Files: `library/panel.go`, `library/item.go`

### Phase 5 — Global config panel wiring
- On startup, parse `~/.gitconfig`
- Walk the parsed `Config` struct and reconstruct a visual block tree:
  - `[includeIf]` sections → ConditionalBlock containing a FileBlock
  - `[include]` sections → FileBlock at root level
  - Everything else → SectionBlock or CustomBlock
- Save button walks the visual tree and serialises back to disk
  (writing each FileBlock's contents to the named file atomically)

### Phase 6 — Project config panel wiring
- "Open Project…" button → folder picker → read `<selection>/.git/config`
- Recent repos persisted to Fyne preferences storage
- After loading, compare project keys against global keys → gray out superseded
  global keys
- Project save writes only to `.git/config`; never touches global files

---

## What Is Kept From the Current Implementation

- `internal/config/types.go` — Config / Section / Entry types (unchanged)
- `internal/config/parser.go` — unchanged  
- `internal/config/writer.go` — unchanged (small addition: `WriteToPath` helper
  for writing a Config to an explicit path, used by file boundaries)
- All 12 tests — unchanged
- `main.go` — minor update to call the new UI entry point

Everything in `internal/ui/` is replaced.

---

## Out of Scope (for now)

- System config (`/etc/gitconfig`) — not shown, not edited
- Undo/redo — deferred
- Syntax validation (e.g. warning if `user.email` is not a valid email address)
- Git LFS / credential helper installation flows
