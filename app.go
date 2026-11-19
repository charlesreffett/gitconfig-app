package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/knightyoshi/gitconfig-app/internal/config"
	"github.com/knightyoshi/gitconfig-app/internal/schema"
)

// ── App ───────────────────────────────────────────────────────────────────────

// App is the struct bound to the Wails runtime. Every exported method becomes
// callable from the Vue frontend via the auto-generated wailsjs bindings.
type App struct {
	ctx  context.Context
	mu   sync.Mutex
	prfs prefs
}

func NewApp() *App {
	a := &App{}
	a.prfs = a.loadPrefs()
	return a
}

func (a *App) startup(ctx context.Context)  { a.ctx = ctx }
func (a *App) shutdown(ctx context.Context) {}

// ── Preferences ───────────────────────────────────────────────────────────────

type prefs struct {
	RecentRepos []string `json:"recentRepos"`
}

func (a *App) prefsPath() string {
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "gitconfig-app", "prefs.json")
}

func (a *App) loadPrefs() prefs {
	data, err := os.ReadFile(a.prefsPath())
	if err != nil {
		return prefs{}
	}
	var p prefs
	json.Unmarshal(data, &p) //nolint:errcheck
	return p
}

func (a *App) savePrefs(p prefs) {
	path := a.prefsPath()
	os.MkdirAll(filepath.Dir(path), 0o755) //nolint:errcheck
	data, _ := json.MarshalIndent(p, "", "  ")
	os.WriteFile(path, data, 0o600) //nolint:errcheck
}

// ── DTOs ──────────────────────────────────────────────────────────────────────
// These types are serialised as JSON and passed to/from the Vue frontend.

// BlockDTO represents one visual block in the editor.
// The Type field determines which fields are relevant:
//
//	"section"     — SchemaName, Subsection, Entries
//	"custom"      — SchemaName (raw section name), Subsection, Entries
//	"file"        — Path, Children (sections inside the included file)
//	"conditional" — Conditions, File (a nested file block)
type BlockDTO struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	SchemaName string         `json:"schemaName,omitempty"`
	Subsection string         `json:"subsection,omitempty"`
	Entries    []EntryDTO     `json:"entries,omitempty"`
	Path       string         `json:"path,omitempty"`
	Children   []BlockDTO     `json:"children,omitempty"`
	Conditions []ConditionDTO `json:"conditions,omitempty"`
	File       *BlockDTO      `json:"file,omitempty"`
}

// EntryDTO is one key=value pair. RawLine carries the original text for comments
// so round-trips through the editor preserve them.
type EntryDTO struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	RawLine string `json:"rawLine,omitempty"`
}

// ConditionDTO is one row of an [includeIf] block: Kind is the condition type
// (e.g. "gitdir") and Value is the argument (e.g. "~/work/**").
type ConditionDTO struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// FieldDTO describes one field in a known section schema.
type FieldDTO struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Type    string   `json:"type"` // "text" | "toggle" | "dropdown" | "multiline"
	Options []string `json:"options,omitempty"`
	Default string   `json:"default,omitempty"`
	Common  bool     `json:"common"`
	Docs    string   `json:"docs,omitempty"`
}

// SectionSchemaDTO describes a known git config section type and its fields.
type SectionSchemaDTO struct {
	Name                  string     `json:"name"`
	Label                 string     `json:"label"`
	Icon                  string     `json:"icon"`
	HasSubsection         bool       `json:"hasSubsection"`
	SubsectionLabel       string     `json:"subsectionLabel,omitempty"`
	SubsectionPlaceholder string     `json:"subsectionPlaceholder,omitempty"`
	Fields                []FieldDTO `json:"fields"`
	FreeForm              bool       `json:"freeForm"`
}

// ── Bound methods ─────────────────────────────────────────────────────────────

// GetGlobalConfigPath returns the canonical path to ~/.gitconfig.
func (a *App) GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gitconfig"), nil
}

// LoadConfig parses the gitconfig at path and returns a block tree.
// Returns an empty slice (not an error) when the file does not exist yet.
func (a *App) LoadConfig(path string) ([]BlockDTO, error) {
	expanded, err := expandHome(path)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Parse(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			return []BlockDTO{}, nil
		}
		return nil, err
	}
	return sectionsToBlocks(cfg.Sections), nil
}

// SaveConfig converts blocks back to config sections and writes the file
// atomically. It creates the file (and parent dirs) if they don't exist.
func (a *App) SaveConfig(path string, blocks []BlockDTO) error {
	expanded, err := expandHome(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(expanded), 0o755); err != nil {
		return err
	}
	cfg := &config.Config{
		FilePath: expanded,
		Sections: blocksToSections(blocks),
	}
	if err := config.Write(cfg); err != nil {
		return err
	}
	return saveIncludedFiles(blocks)
}

// saveIncludedFiles writes the (editable) children of every file/conditional
// block back to the file they were loaded from. Blocks with no children are
// skipped so a missing or unreadable include target is never created or wiped.
func saveIncludedFiles(blocks []BlockDTO) error {
	for _, b := range blocks {
		if (b.Type != "file" && b.Type != "conditional") || b.Path == "" || len(b.Children) == 0 {
			continue
		}
		expanded, err := expandHome(b.Path)
		if err != nil {
			return err
		}
		existing, err := config.Parse(expanded)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			existing = &config.Config{}
		}
		cfg := &config.Config{
			FilePath: expanded,
			Sections: mergeIncludedSections(existing.Sections, blocksToSections(b.Children)),
		}
		if err := config.Write(cfg); err != nil {
			return err
		}
	}
	return nil
}

// mergeIncludedSections rebuilds an included file from edited sections while
// keeping the original preamble and in-section comments. loadIncludedFileBlocks
// maps each named section to one child block in order, so sections are matched
// by position.
func mergeIncludedSections(old, edited []config.Section) []config.Section {
	var out []config.Section
	var named []config.Section
	for _, s := range old {
		if s.Name == "" {
			out = append(out, s) // preamble comments
		} else {
			named = append(named, s)
		}
	}
	for i, sec := range edited {
		if i < len(named) && strings.EqualFold(named[i].Name, sec.Name) {
			sec.Entries = mergeEntries(named[i].Entries, sec.Entries)
		}
		out = append(out, sec)
	}
	return out
}

// mergeEntries keeps comment/blank lines from old in place and substitutes
// keyed entries positionally from edited. Extra edited entries are appended
// before any trailing blank lines; surplus old keyed entries are dropped.
func mergeEntries(old, edited []config.Entry) []config.Entry {
	var out []config.Entry
	k := 0
	for _, e := range old {
		if e.Key == "" {
			out = append(out, e)
			continue
		}
		if k < len(edited) {
			out = append(out, edited[k])
			k++
		}
	}
	if k < len(edited) {
		// Insert remaining entries before trailing blank lines.
		cut := len(out)
		for cut > 0 && out[cut-1].Key == "" && strings.TrimSpace(out[cut-1].RawLine) == "" {
			cut--
		}
		tail := append([]config.Entry{}, out[cut:]...)
		out = append(append(out[:cut], edited[k:]...), tail...)
	}
	return out
}

// GetSchema returns the full list of known section schemas.
func (a *App) GetSchema() []SectionSchemaDTO {
	out := make([]SectionSchemaDTO, 0, len(schema.All))
	for _, s := range schema.All {
		out = append(out, schemaToDTO(s))
	}
	return out
}

// GetConditionTypes returns the valid condition type strings for [includeIf].
func (a *App) GetConditionTypes() []string {
	return schema.ConditionTypes()
}

// SelectProjectFolder opens a native folder-picker dialog. Returns "" if the
// user cancels without selecting a folder.
func (a *App) SelectProjectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Git Repository",
	})
}

// ExpandHome expands a leading ~ in path to the user's home directory.
func (a *App) ExpandHome(path string) string {
	expanded, err := expandHome(path)
	if err != nil {
		return path
	}
	return expanded
}

// GetRecentRepos returns the persisted list of recently-opened repo paths.
func (a *App) GetRecentRepos() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.prfs.RecentRepos == nil {
		return []string{}
	}
	return a.prfs.RecentRepos
}

// AddRecentRepo prepends repoPath to the recent repos list (max 5, deduped).
func (a *App) AddRecentRepo(repoPath string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var filtered []string
	for _, p := range a.prfs.RecentRepos {
		if p != repoPath {
			filtered = append(filtered, p)
		}
	}
	updated := append([]string{repoPath}, filtered...)
	if len(updated) > 5 {
		updated = updated[:5]
	}
	a.prfs.RecentRepos = updated
	a.savePrefs(a.prfs)
}

// ── Config ↔ Block conversion ─────────────────────────────────────────────────

// sectionsToBlocks converts parsed config sections into the block tree that the
// frontend renders. This mirrors the loadConfigToCanvas logic from the old UI.
func sectionsToBlocks(sections []config.Section) []BlockDTO {
	var blocks []BlockDTO
	i := 0
	for i < len(sections) {
		sec := sections[i]

		switch sec.Name {
		case "includeIf":
			// Collect consecutive [includeIf] sections that all point to the same
			// file and merge them into one ConditionalBlock with multiple condition rows.
			path := pathValue(sec)
			k, v := parseCondition(sec.Subsection)
			conds := []ConditionDTO{{Kind: k, Value: v}}

			j := i + 1
			for j < len(sections) &&
				sections[j].Name == "includeIf" &&
				pathValue(sections[j]) == path {
				k2, v2 := parseCondition(sections[j].Subsection)
				conds = append(conds, ConditionDTO{Kind: k2, Value: v2})
				j++
			}

			blocks = append(blocks, BlockDTO{
				ID:         newID(),
				Type:       "conditional",
				Conditions: conds,
				Path:       path,
				Children:   loadIncludedFileBlocks(path),
			})
			i = j

		case "include":
			path := pathValue(sec)
			blocks = append(blocks, BlockDTO{
				ID:       newID(),
				Type:     "file",
				Path:     path,
				Children: loadIncludedFileBlocks(path),
			})
			i++

		default:
			if sec.Name == "" {
				i++ // skip preamble comment sections
				continue
			}
			if _, ok := schema.Lookup(sec.Name); ok {
				blocks = append(blocks, BlockDTO{
					ID:         newID(),
					Type:       "section",
					SchemaName: sec.Name,
					Subsection: sec.Subsection,
					Entries:    entriesToDTOs(sec.Entries),
				})
			} else {
				blocks = append(blocks, BlockDTO{
					ID:         newID(),
					Type:       "custom",
					SchemaName: sec.Name,
					Subsection: sec.Subsection,
					Entries:    entriesToDTOs(sec.Entries),
				})
			}
			i++
		}
	}
	return blocks
}

// blocksToSections converts the block tree back to config.Section structs for
// writing. This is the inverse of sectionsToBlocks.
func blocksToSections(blocks []BlockDTO) []config.Section {
	var sections []config.Section
	for _, b := range blocks {
		switch b.Type {
		case "section", "custom":
			sec := config.Section{
				Name:       b.SchemaName,
				Subsection: b.Subsection,
			}
			for _, e := range b.Entries {
				if e.Key != "" {
					sec.Entries = append(sec.Entries, config.Entry{
						Key:     e.Key,
						Value:   e.Value,
						RawLine: e.RawLine,
					})
				}
			}
			sections = append(sections, sec)

		case "file":
			if b.Path != "" {
				sections = append(sections, config.Section{
					Name:    "include",
					Entries: []config.Entry{{Key: "path", Value: b.Path}},
				})
			}
			// Children live in the included file — written by saveIncludedFiles.

		case "conditional":
			// Each condition row becomes one [includeIf] section pointing to the
			// same path.
			filePath := b.Path
			for _, cond := range b.Conditions {
				sec := config.Section{
					Name:       "includeIf",
					Subsection: cond.Kind + ":" + cond.Value,
				}
				if filePath != "" {
					sec.Entries = []config.Entry{{Key: "path", Value: filePath}}
				}
				sections = append(sections, sec)
			}
		}
	}
	return sections
}

// loadIncludedFileBlocks parses rawPath and returns its sections as blocks.
// Errors (missing file, parse failure) are silently ignored — the file may
// not exist yet or be temporarily inaccessible.
func loadIncludedFileBlocks(rawPath string) []BlockDTO {
	if rawPath == "" {
		return nil
	}
	expanded, err := expandHome(rawPath)
	if err != nil {
		return nil
	}
	subCfg, err := config.Parse(expanded)
	if err != nil {
		return nil
	}
	var blocks []BlockDTO
	for _, sec := range subCfg.Sections {
		if sec.Name == "" {
			continue
		}
		if _, ok := schema.Lookup(sec.Name); ok {
			blocks = append(blocks, BlockDTO{
				ID:         newID(),
				Type:       "section",
				SchemaName: sec.Name,
				Subsection: sec.Subsection,
				Entries:    entriesToDTOs(sec.Entries),
			})
		} else {
			blocks = append(blocks, BlockDTO{
				ID:         newID(),
				Type:       "custom",
				SchemaName: sec.Name,
				Subsection: sec.Subsection,
				Entries:    entriesToDTOs(sec.Entries),
			})
		}
	}
	return blocks
}

// entriesToDTOs converts config entries, skipping comment/blank lines.
func entriesToDTOs(entries []config.Entry) []EntryDTO {
	var out []EntryDTO
	for _, e := range entries {
		if e.Key != "" {
			out = append(out, EntryDTO{Key: e.Key, Value: e.Value, RawLine: e.RawLine})
		}
	}
	return out
}

// schemaToDTO converts a schema.SectionSchema to its wire-format DTO.
func schemaToDTO(s schema.SectionSchema) SectionSchemaDTO {
	dto := SectionSchemaDTO{
		Name:                  s.Name,
		Label:                 s.Label,
		Icon:                  s.Icon,
		HasSubsection:         s.HasSubsection,
		SubsectionLabel:       s.SubsectionLabel,
		SubsectionPlaceholder: s.SubsectionPlaceholder,
		FreeForm:              s.FreeForm,
		Fields:                make([]FieldDTO, 0, len(s.Fields)),
	}
	for _, f := range s.Fields {
		fd := FieldDTO{
			Key:     f.Key,
			Label:   f.Label,
			Common:  f.Common,
			Default: f.Default,
			Docs:    f.Docs,
		}
		switch f.Type {
		case schema.FieldToggle:
			fd.Type = "toggle"
		case schema.FieldDropdown:
			fd.Type = "dropdown"
			fd.Options = f.Options
		case schema.FieldMultiline:
			fd.Type = "multiline"
		default:
			fd.Type = "text"
		}
		dto.Fields = append(dto.Fields, fd)
	}
	return dto
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func pathValue(sec config.Section) string {
	for _, e := range sec.Entries {
		if e.Key == "path" {
			return e.Value
		}
	}
	return ""
}

func parseCondition(subsection string) (kind, value string) {
	idx := strings.Index(subsection, ":")
	if idx < 0 {
		return subsection, ""
	}
	return subsection[:idx], subsection[idx+1:]
}

func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
