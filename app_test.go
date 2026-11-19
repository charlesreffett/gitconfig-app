package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knightyoshi/gitconfig-app/internal/config"
	"github.com/knightyoshi/gitconfig-app/internal/schema"
)

// ── sectionsToBlocks ──────────────────────────────────────────────────────────

func TestSectionsToBlocks_Empty(t *testing.T) {
	if got := sectionsToBlocks(nil, ""); len(got) != 0 {
		t.Fatalf("expected 0 blocks, got %d", len(got))
	}
}

func TestSectionsToBlocks_KnownSection(t *testing.T) {
	sections := []config.Section{
		{Name: "user", Entries: []config.Entry{
			{Key: "name", Value: "Alice"},
			{Key: "email", Value: "alice@example.com"},
		}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != "section" {
		t.Errorf("Type = %q; want \"section\"", b.Type)
	}
	if b.SchemaName != "user" {
		t.Errorf("SchemaName = %q; want \"user\"", b.SchemaName)
	}
	if len(b.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(b.Entries))
	}
	if b.ID == "" {
		t.Error("ID must not be empty")
	}
}

func TestSectionsToBlocks_UnknownSection(t *testing.T) {
	sections := []config.Section{
		{Name: "myplugin", Entries: []config.Entry{{Key: "opt", Value: "val"}}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != "custom" {
		t.Errorf("Type = %q; want \"custom\"", blocks[0].Type)
	}
	if blocks[0].SchemaName != "myplugin" {
		t.Errorf("SchemaName = %q; want \"myplugin\"", blocks[0].SchemaName)
	}
}

func TestSectionsToBlocks_Subsection(t *testing.T) {
	sections := []config.Section{
		{Name: "branch", Subsection: "main", Entries: []config.Entry{
			{Key: "rebase", Value: "true"},
		}},
	}
	blocks := sectionsToBlocks(sections, "")
	if blocks[0].Subsection != "main" {
		t.Errorf("Subsection = %q; want \"main\"", blocks[0].Subsection)
	}
}

func TestSectionsToBlocks_Include(t *testing.T) {
	sections := []config.Section{
		{Name: "include", Entries: []config.Entry{{Key: "path", Value: "~/.gitconfig_extra"}}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != "file" {
		t.Errorf("Type = %q; want \"file\"", b.Type)
	}
	if b.Path != "~/.gitconfig_extra" {
		t.Errorf("Path = %q; want \"~/.gitconfig_extra\"", b.Path)
	}
}

func TestSectionsToBlocks_IncludeIfMerged(t *testing.T) {
	// Two consecutive [includeIf] sections pointing to the same file must merge
	// into one ConditionalBlock with two condition rows.
	sections := []config.Section{
		{Name: "includeIf", Subsection: "gitdir:~/work/", Entries: []config.Entry{{Key: "path", Value: "~/.gitconfig_work"}}},
		{Name: "includeIf", Subsection: "gitdir:~/work/**", Entries: []config.Entry{{Key: "path", Value: "~/.gitconfig_work"}}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 merged ConditionalBlock, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != "conditional" {
		t.Errorf("Type = %q; want \"conditional\"", b.Type)
	}
	if len(b.Conditions) != 2 {
		t.Fatalf("Conditions len = %d; want 2", len(b.Conditions))
	}
	if b.Conditions[0].Kind != "gitdir" || b.Conditions[0].Value != "~/work/" {
		t.Errorf("Conditions[0] = %+v", b.Conditions[0])
	}
	if b.Conditions[1].Kind != "gitdir" || b.Conditions[1].Value != "~/work/**" {
		t.Errorf("Conditions[1] = %+v", b.Conditions[1])
	}
	if b.Path != "~/.gitconfig_work" {
		t.Errorf("Path = %q; want \"~/.gitconfig_work\"", b.Path)
	}
}

func TestSectionsToBlocks_IncludeIfNotMergedDifferentPaths(t *testing.T) {
	// Two [includeIf] sections pointing to different files must NOT be merged.
	sections := []config.Section{
		{Name: "includeIf", Subsection: "gitdir:~/work/", Entries: []config.Entry{{Key: "path", Value: "~/.gitconfig_work"}}},
		{Name: "includeIf", Subsection: "gitdir:~/home/", Entries: []config.Entry{{Key: "path", Value: "~/.gitconfig_home"}}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks for different paths, got %d", len(blocks))
	}
	if blocks[0].Path != "~/.gitconfig_work" || blocks[1].Path != "~/.gitconfig_home" {
		t.Errorf("paths: %q, %q", blocks[0].Path, blocks[1].Path)
	}
}

func TestSectionsToBlocks_SkipsPreamble(t *testing.T) {
	// Sections with empty Name (preamble comment sections) are ignored.
	sections := []config.Section{
		{Name: "", Entries: []config.Entry{{RawLine: "# top comment"}}},
		{Name: "user", Entries: []config.Entry{{Key: "name", Value: "Bob"}}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block (preamble skipped), got %d", len(blocks))
	}
	if blocks[0].SchemaName != "user" {
		t.Errorf("SchemaName = %q; want \"user\"", blocks[0].SchemaName)
	}
}

func TestSectionsToBlocks_SkipsCommentEntries(t *testing.T) {
	// Entries with empty Key (comments/blank lines) must not appear in BlockDTO.Entries.
	sections := []config.Section{
		{Name: "core", Entries: []config.Entry{
			{RawLine: "# a comment"},
			{Key: "autocrlf", Value: "input"},
		}},
	}
	blocks := sectionsToBlocks(sections, "")
	if len(blocks[0].Entries) != 1 {
		t.Errorf("expected 1 kv entry, got %d", len(blocks[0].Entries))
	}
	if blocks[0].Entries[0].Key != "autocrlf" {
		t.Errorf("entry key = %q; want \"autocrlf\"", blocks[0].Entries[0].Key)
	}
}

// ── blocksToSections ──────────────────────────────────────────────────────────

func TestBlocksToSections_SectionBlock(t *testing.T) {
	blocks := []BlockDTO{
		{ID: "1", Type: "section", SchemaName: "user", Entries: []EntryDTO{
			{Key: "name", Value: "Alice"},
			{Key: "email", Value: "alice@example.com"},
		}},
	}
	sections := blocksToSections(blocks)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	s := sections[0]
	if s.Name != "user" {
		t.Errorf("Name = %q; want \"user\"", s.Name)
	}
	if len(s.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(s.Entries))
	}
}

func TestBlocksToSections_CustomBlock(t *testing.T) {
	blocks := []BlockDTO{
		{ID: "1", Type: "custom", SchemaName: "myplugin", Entries: []EntryDTO{
			{Key: "x", Value: "y"},
		}},
	}
	sections := blocksToSections(blocks)
	if len(sections) != 1 || sections[0].Name != "myplugin" {
		t.Errorf("expected section \"myplugin\", got %+v", sections)
	}
}

func TestBlocksToSections_SubsectionPreserved(t *testing.T) {
	blocks := []BlockDTO{
		{ID: "1", Type: "section", SchemaName: "branch", Subsection: "main",
			Entries: []EntryDTO{{Key: "rebase", Value: "true"}}},
	}
	sections := blocksToSections(blocks)
	if sections[0].Subsection != "main" {
		t.Errorf("Subsection = %q; want \"main\"", sections[0].Subsection)
	}
}

func TestBlocksToSections_FileBlock(t *testing.T) {
	blocks := []BlockDTO{
		{ID: "1", Type: "file", Path: "~/.gitconfig_extra"},
	}
	sections := blocksToSections(blocks)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	s := sections[0]
	if s.Name != "include" {
		t.Errorf("Name = %q; want \"include\"", s.Name)
	}
	if len(s.Entries) != 1 || s.Entries[0].Key != "path" || s.Entries[0].Value != "~/.gitconfig_extra" {
		t.Errorf("entries = %+v", s.Entries)
	}
}

func TestBlocksToSections_FileBlock_EmptyPathSkipped(t *testing.T) {
	blocks := []BlockDTO{{ID: "1", Type: "file", Path: ""}}
	if got := blocksToSections(blocks); len(got) != 0 {
		t.Errorf("expected no sections for empty path, got %d", len(got))
	}
}

func TestBlocksToSections_ConditionalBlock_ExpandsConditions(t *testing.T) {
	// Each condition row must produce one [includeIf] section with the same path.
	blocks := []BlockDTO{
		{
			ID:   "1",
			Type: "conditional",
			Path: "~/.gitconfig_work",
			Conditions: []ConditionDTO{
				{Kind: "gitdir", Value: "~/work/"},
				{Kind: "gitdir", Value: "~/work/**"},
			},
		},
	}
	sections := blocksToSections(blocks)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	for i, sec := range sections {
		if sec.Name != "includeIf" {
			t.Errorf("[%d] Name = %q; want \"includeIf\"", i, sec.Name)
		}
		if len(sec.Entries) != 1 || sec.Entries[0].Value != "~/.gitconfig_work" {
			t.Errorf("[%d] entries = %+v", i, sec.Entries)
		}
	}
	if sections[0].Subsection != "gitdir:~/work/" {
		t.Errorf("Subsection[0] = %q; want \"gitdir:~/work/\"", sections[0].Subsection)
	}
	if sections[1].Subsection != "gitdir:~/work/**" {
		t.Errorf("Subsection[1] = %q; want \"gitdir:~/work/**\"", sections[1].Subsection)
	}
}

func TestBlocksToSections_SkipsEmptyKeyEntries(t *testing.T) {
	blocks := []BlockDTO{
		{ID: "1", Type: "section", SchemaName: "core", Entries: []EntryDTO{
			{Key: "", Value: "should be skipped"},
			{Key: "bare", Value: "false"},
		}},
	}
	sections := blocksToSections(blocks)
	if len(sections[0].Entries) != 1 || sections[0].Entries[0].Key != "bare" {
		t.Errorf("entries = %+v", sections[0].Entries)
	}
}

// ── Round-trip ────────────────────────────────────────────────────────────────

func TestRoundTrip_SectionsToBlocksAndBack(t *testing.T) {
	original := []config.Section{
		{Name: "user", Entries: []config.Entry{
			{Key: "name", Value: "Alice"},
			{Key: "email", Value: "alice@example.com"},
		}},
		{Name: "core", Entries: []config.Entry{
			{Key: "autocrlf", Value: "input"},
		}},
		{Name: "includeIf", Subsection: "gitdir:~/work/",
			Entries: []config.Entry{{Key: "path", Value: "~/.work"}}},
		{Name: "include",
			Entries: []config.Entry{{Key: "path", Value: "~/.extra"}}},
	}

	blocks := sectionsToBlocks(original, "")
	recovered := blocksToSections(blocks)

	if len(recovered) != len(original) {
		t.Fatalf("section count: got %d, want %d", len(recovered), len(original))
	}
	for i := range original {
		if recovered[i].Name != original[i].Name {
			t.Errorf("[%d] Name: got %q, want %q", i, recovered[i].Name, original[i].Name)
		}
		if recovered[i].Subsection != original[i].Subsection {
			t.Errorf("[%d] Subsection: got %q, want %q", i, recovered[i].Subsection, original[i].Subsection)
		}
	}
}

// ── parseCondition ────────────────────────────────────────────────────────────

func TestParseCondition(t *testing.T) {
	cases := []struct {
		input string
		kind  string
		value string
	}{
		{"gitdir:~/work/", "gitdir", "~/work/"},
		{"onbranch:main", "onbranch", "main"},
		{"gitdir/i:~/Work/", "gitdir/i", "~/Work/"},
		// colon in value — only the first colon splits kind from value
		{"hasconfig:remote.*.url:git@github.com:**", "hasconfig", "remote.*.url:git@github.com:**"},
		// no colon → value is empty
		{"nocolon", "nocolon", ""},
	}
	for _, c := range cases {
		k, v := parseCondition(c.input)
		if k != c.kind || v != c.value {
			t.Errorf("parseCondition(%q): got (%q, %q); want (%q, %q)",
				c.input, k, v, c.kind, c.value)
		}
	}
}

// ── expandHome ────────────────────────────────────────────────────────────────

func TestExpandHome_AbsolutePath(t *testing.T) {
	got, err := expandHome("/absolute/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/absolute/path" {
		t.Errorf("expected path unchanged, got %q", got)
	}
}

func TestExpandHome_Tilde(t *testing.T) {
	got, err := expandHome("~/.gitconfig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "~/.gitconfig" {
		t.Error("expected ~ to be expanded")
	}
	if !strings.HasSuffix(got, "/.gitconfig") {
		t.Errorf("expected path ending in /.gitconfig, got %q", got)
	}
	if strings.HasPrefix(got, "~") {
		t.Errorf("expected ~ replaced, got %q", got)
	}
}

func TestExpandHome_NoTilde(t *testing.T) {
	got, err := expandHome("relative/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "relative/path" {
		t.Errorf("expected path unchanged, got %q", got)
	}
}

// ── schemaToDTO ───────────────────────────────────────────────────────────────

func TestSchemaToDTO_FieldTypeMapping(t *testing.T) {
	s := schema.SectionSchema{
		Name:  "test",
		Label: "Test",
		Fields: []schema.Field{
			{Key: "a", Type: schema.FieldText},
			{Key: "b", Type: schema.FieldToggle},
			{Key: "c", Type: schema.FieldDropdown, Options: []string{"x", "y", "z"}},
			{Key: "d", Type: schema.FieldMultiline},
		},
	}
	dto := schemaToDTO(s)

	wantTypes := []string{"text", "toggle", "dropdown", "multiline"}
	for i, f := range dto.Fields {
		if f.Type != wantTypes[i] {
			t.Errorf("Fields[%d].Type = %q; want %q", i, f.Type, wantTypes[i])
		}
	}
	if len(dto.Fields[2].Options) != 3 {
		t.Errorf("dropdown Options: got %v", dto.Fields[2].Options)
	}
}

func TestSchemaToDTO_PreservesMetadata(t *testing.T) {
	s := schema.SectionSchema{
		Name:                  "remote",
		Label:                 "Remote",
		Icon:                  "🌐",
		HasSubsection:         true,
		SubsectionLabel:       "Remote name",
		SubsectionPlaceholder: "origin",
		FreeForm:              false,
	}
	dto := schemaToDTO(s)

	if dto.Name != "remote" || dto.Label != "Remote" || dto.Icon != "🌐" {
		t.Errorf("metadata mismatch: %+v", dto)
	}
	if !dto.HasSubsection {
		t.Error("HasSubsection should be true")
	}
	if dto.SubsectionLabel != "Remote name" || dto.SubsectionPlaceholder != "origin" {
		t.Errorf("subsection labels: %q, %q", dto.SubsectionLabel, dto.SubsectionPlaceholder)
	}
}

// ── AddRecentRepo ─────────────────────────────────────────────────────────────

func TestAddRecentRepo_Prepends(t *testing.T) {
	a := &App{}
	a.prfs.RecentRepos = []string{"/existing/repo"}
	a.AddRecentRepo("/new/repo")

	repos := a.GetRecentRepos()
	if repos[0] != "/new/repo" {
		t.Errorf("expected /new/repo at index 0, got %q", repos[0])
	}
	if repos[1] != "/existing/repo" {
		t.Errorf("expected /existing/repo at index 1, got %q", repos[1])
	}
}

func TestAddRecentRepo_DeduplicatesExisting(t *testing.T) {
	a := &App{}
	a.prfs.RecentRepos = []string{"/a", "/b", "/c"}
	a.AddRecentRepo("/b")

	repos := a.GetRecentRepos()
	if repos[0] != "/b" {
		t.Errorf("expected /b at front, got %q", repos[0])
	}
	if len(repos) != 3 {
		t.Errorf("expected 3 repos after dedup, got %d: %v", len(repos), repos)
	}
}

func TestAddRecentRepo_CapsAtFive(t *testing.T) {
	a := &App{}
	a.prfs.RecentRepos = []string{"/a", "/b", "/c", "/d", "/e"}
	a.AddRecentRepo("/f")

	repos := a.GetRecentRepos()
	if len(repos) != 5 {
		t.Errorf("expected 5 repos, got %d", len(repos))
	}
	if repos[0] != "/f" {
		t.Errorf("expected /f at front, got %q", repos[0])
	}
	// "/e" should have been evicted
	for _, r := range repos {
		if r == "/e" {
			t.Error("expected /e to be evicted, but it's still present")
		}
	}
}

func TestAddRecentRepo_EmptyList(t *testing.T) {
	a := &App{}
	a.AddRecentRepo("/first/repo")

	repos := a.GetRecentRepos()
	if len(repos) != 1 || repos[0] != "/first/repo" {
		t.Errorf("unexpected repos: %v", repos)
	}
}

// ── GetConditionTypes ─────────────────────────────────────────────────────────

func TestGetConditionTypes_ContainsStandardTypes(t *testing.T) {
	a := &App{}
	types := a.GetConditionTypes()
	if len(types) == 0 {
		t.Fatal("expected non-empty condition types")
	}
	required := []string{"gitdir", "gitdir/i", "onbranch"}
	found := make(map[string]bool, len(types))
	for _, ct := range types {
		found[ct] = true
	}
	for _, r := range required {
		if !found[r] {
			t.Errorf("missing condition type %q in %v", r, types)
		}
	}
}

// ── SaveConfig: included file children ───────────────────────────────────────

func TestSaveConfig_WritesIncludedFileChildren(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "gitconfig")
	incPath := filepath.Join(dir, "gitconfig_work")

	inc := "# work identity\n[user]\n\tname = bob\n\t# keep me\n\temail = old@example.com\n"
	if err := os.WriteFile(incPath, []byte(inc), 0o644); err != nil {
		t.Fatal(err)
	}
	mainCfg := "[includeIf \"gitdir:~/work/\"]\n\tpath = " + incPath + "\n"
	if err := os.WriteFile(mainPath, []byte(mainCfg), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &App{}
	blocks, err := a.LoadConfig(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || len(blocks[0].Children) != 1 {
		t.Fatalf("unexpected blocks: %+v", blocks)
	}
	for i, e := range blocks[0].Children[0].Entries {
		if e.Key == "email" {
			blocks[0].Children[0].Entries[i].Value = "new@example.com"
		}
	}
	if err := a.SaveConfig(mainPath, blocks); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(incPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"email = new@example.com", "# work identity", "# keep me", "name = bob"} {
		if !strings.Contains(s, want) {
			t.Errorf("included file missing %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "old@example.com") {
		t.Errorf("old email still present:\n%s", s)
	}
}

func TestSaveConfig_SkipsIncludeWithoutChildren(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "gitconfig")
	incPath := filepath.Join(dir, "missing")
	a := &App{}
	err := a.SaveConfig(mainPath, []BlockDTO{{ID: "1", Type: "file", Path: incPath}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(incPath); !os.IsNotExist(err) {
		t.Errorf("included file should not be created")
	}
}

// ── Relative include paths ───────────────────────────────────────────────────

func TestResolveIncludePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	cases := []struct{ raw, base, want string }{
		{"config/.gitconfig_work", "/home/me", "/home/me/config/.gitconfig_work"},
		{"../shared.inc", "/repo/.git", "/repo/shared.inc"},
		{"/abs/file", "/home/me", "/abs/file"},
		{"~/configs/work", "/elsewhere", filepath.Join(home, "configs/work")},
		{"relative/path", "", "relative/path"},
	}
	for _, c := range cases {
		got, err := resolveIncludePath(c.raw, c.base)
		if err != nil {
			t.Fatalf("resolveIncludePath(%q, %q): %v", c.raw, c.base, err)
		}
		if got != c.want {
			t.Errorf("resolveIncludePath(%q, %q) = %q; want %q", c.raw, c.base, got, c.want)
		}
	}
}

func TestLoadAndSave_RelativeIncludePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(dir, "gitconfig")
	incPath := filepath.Join(dir, "config", ".gitconfig_work")
	if err := os.WriteFile(incPath, []byte("[user]\n\temail = old@example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainCfg := "[includeIf \"gitdir:~/work/\"]\n\tpath = config/.gitconfig_work\n"
	if err := os.WriteFile(mainPath, []byte(mainCfg), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run from a different cwd to prove resolution doesn't depend on it.
	wd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(wd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	a := &App{}
	blocks, err := a.LoadConfig(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || len(blocks[0].Children) != 1 {
		t.Fatalf("relative include not loaded: %+v", blocks)
	}
	blocks[0].Children[0].Entries[0].Value = "new@example.com"
	if err := a.SaveConfig(mainPath, blocks); err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(incPath)
	if !strings.Contains(string(got), "email = new@example.com") {
		t.Errorf("included file not updated:\n%s", got)
	}
	mainGot, _ := os.ReadFile(mainPath)
	if !strings.Contains(string(mainGot), "path = config/.gitconfig_work") {
		t.Errorf("main config should keep the relative path as written:\n%s", mainGot)
	}
}
