package config

import (
	"os"
	"strings"
	"testing"
)

// TestRoundTrip checks that Parse → Write preserves all the structural data
// that matters: section names, subsections, keys, and values.
func TestRoundTrip(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)

	cfg, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := Write(cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after Write: %v", err)
	}
	out := string(written)

	mustContain := []string{
		"[user]",
		"name = KnightYoshi",
		"email = 6134576+KnightYoshi@users.noreply.github.com",
		`[includeIf "gitdir:~/work/"]`,
		"path = ~/.gitconfig_work",
		"[core]",
		"autocrlf = input",
		`[filter "lfs"]`,
		"process = git-lfs filter-process",
		"required = true",
	}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Errorf("round-trip output missing %q\n\nFull output:\n%s", want, out)
		}
	}
}

// TestRoundTrip_SectionCount ensures no sections are gained or lost.
func TestRoundTrip_SectionCount(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)

	cfg, _ := Parse(path)
	before := len(cfg.Sections)

	_ = Write(cfg)
	cfg2, _ := Parse(path)

	if len(cfg2.Sections) != before {
		t.Errorf("section count changed: %d → %d", before, len(cfg2.Sections))
	}
}

// TestWrite_AddEntry verifies that a newly appended entry appears after Write.
func TestWrite_AddEntry(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)

	cfg, _ := Parse(path)
	// Add signingkey to [user].
	cfg.Sections[0].Entries = append(cfg.Sections[0].Entries,
		Entry{Key: "signingkey", Value: "DEADBEEF"},
	)
	_ = Write(cfg)

	written, _ := os.ReadFile(path)
	if !strings.Contains(string(written), "signingkey = DEADBEEF") {
		t.Errorf("added entry not found in written file:\n%s", written)
	}
}

// TestSetEntry_Existing checks that SetEntry updates a value in-place.
func TestSetEntry_Existing(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)
	cfg, _ := Parse(path)

	SetEntry(&cfg.Sections[0], "name", "NewName")
	_ = Write(cfg)

	written, _ := os.ReadFile(path)
	if !strings.Contains(string(written), "name = NewName") {
		t.Errorf("updated name not found:\n%s", written)
	}
	// Old value must be gone.
	if strings.Contains(string(written), "name = KnightYoshi") {
		t.Errorf("old name still present:\n%s", written)
	}
}

// TestSetEntry_New checks that SetEntry appends a brand-new key.
func TestSetEntry_New(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)
	cfg, _ := Parse(path)

	SetEntry(&cfg.Sections[0], "gpgsign", "true")
	_ = Write(cfg)

	written, _ := os.ReadFile(path)
	if !strings.Contains(string(written), "gpgsign = true") {
		t.Errorf("new key not found:\n%s", written)
	}
}

// TestDeleteEntry verifies that DeleteEntry removes a key completely.
func TestDeleteEntry(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)
	cfg, _ := Parse(path)

	DeleteEntry(&cfg.Sections[0], "email")
	_ = Write(cfg)

	written, _ := os.ReadFile(path)
	if strings.Contains(string(written), "email") {
		t.Errorf("deleted key 'email' still present:\n%s", written)
	}
}

// TestWrite_Atomic checks that the temp file is cleaned up on success
// (i.e. it is renamed to the target, leaving no .tmp file behind).
func TestWrite_Atomic(t *testing.T) {
	path := writeTempConfig(t, sampleConfig)
	cfg, _ := Parse(path)

	dir := t.TempDir()
	// Point cfg at a path inside our controlled temp dir.
	cfg.FilePath = dir + "/.gitconfig"
	_ = os.WriteFile(cfg.FilePath, []byte(sampleConfig), 0600)

	_ = Write(cfg)

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}
