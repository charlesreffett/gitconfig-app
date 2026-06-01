package config

import (
	"os"
	"path/filepath"
	"testing"
)

// sampleConfig mirrors the structure from the real .gitconfig so tests stay
// grounded in realistic input.
const sampleConfig = `[user]
	name = KnightYoshi
	email = 6134576+KnightYoshi@users.noreply.github.com

[includeIf "gitdir:~/work/"]
    path = ~/.gitconfig_work

[includeIf "gitdir:~/work/**"]
    path = ~/.gitconfig_work

[core]
	autocrlf = input
[filter "lfs"]
	process = git-lfs filter-process
	required = true
	clean = git-lfs clean -- %f
	smudge = git-lfs smudge -- %f
`

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".gitconfig")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

// ── section count and names ───────────────────────────────────────────────────

func TestParse_SectionCount(t *testing.T) {
	cfg, err := Parse(writeTempConfig(t, sampleConfig))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// user, includeIf×2, core, filter = 5 sections
	if len(cfg.Sections) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(cfg.Sections))
	}
}

func TestParse_UserSection(t *testing.T) {
	cfg, _ := Parse(writeTempConfig(t, sampleConfig))
	user := cfg.Sections[0]

	if user.Name != "user" {
		t.Errorf("section[0].Name = %q; want \"user\"", user.Name)
	}
	if user.Subsection != "" {
		t.Errorf("section[0].Subsection = %q; want empty", user.Subsection)
	}

	// The section has name + email entries plus a trailing blank line.
	kvEntries := keyValueEntries(user)
	if len(kvEntries) != 2 {
		t.Fatalf("user section: expected 2 kv entries, got %d", len(kvEntries))
	}
	if kvEntries[0].Key != "name" || kvEntries[0].Value != "KnightYoshi" {
		t.Errorf("user.name: got key=%q value=%q", kvEntries[0].Key, kvEntries[0].Value)
	}
	if kvEntries[1].Key != "email" {
		t.Errorf("user.email key: got %q", kvEntries[1].Key)
	}
}

func TestParse_IncludeIfSubsection(t *testing.T) {
	cfg, _ := Parse(writeTempConfig(t, sampleConfig))

	inc := cfg.Sections[1]
	if inc.Name != "includeIf" {
		t.Errorf("section[1].Name = %q; want \"includeIf\"", inc.Name)
	}
	if inc.Subsection != "gitdir:~/work/" {
		t.Errorf("section[1].Subsection = %q; want \"gitdir:~/work/\"", inc.Subsection)
	}

	kvs := keyValueEntries(inc)
	if len(kvs) != 1 || kvs[0].Key != "path" || kvs[0].Value != "~/.gitconfig_work" {
		t.Errorf("includeIf path entry: %+v", kvs)
	}
}

func TestParse_FilterSubsection(t *testing.T) {
	cfg, _ := Parse(writeTempConfig(t, sampleConfig))

	flt := cfg.Sections[4]
	if flt.Name != "filter" || flt.Subsection != "lfs" {
		t.Errorf("section[4]: got name=%q sub=%q", flt.Name, flt.Subsection)
	}
	if len(keyValueEntries(flt)) != 4 {
		t.Errorf("filter.lfs: expected 4 kv entries, got %d", len(keyValueEntries(flt)))
	}
}

// ── parseHeader unit tests ────────────────────────────────────────────────────

func TestParseHeader(t *testing.T) {
	cases := []struct {
		line string
		name string
		sub  string
	}{
		{`[user]`, "user", ""},
		{`[core]`, "core", ""},
		{`[includeIf "gitdir:~/work/"]`, "includeIf", "gitdir:~/work/"},
		{`[includeIf "gitdir:~/work/**"]`, "includeIf", "gitdir:~/work/**"},
		{`[filter "lfs"]`, "filter", "lfs"},
		{`[url "git@github.com-work"]`, "url", "git@github.com-work"},
	}
	for _, c := range cases {
		name, sub, err := parseHeader(c.line)
		if err != nil {
			t.Errorf("parseHeader(%q): unexpected error: %v", c.line, err)
			continue
		}
		if name != c.name || sub != c.sub {
			t.Errorf("parseHeader(%q): got (%q,%q); want (%q,%q)",
				c.line, name, sub, c.name, c.sub)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// keyValueEntries returns only the entries that are real key=value pairs
// (i.e. not comment or blank lines).
func keyValueEntries(s Section) []Entry {
	var out []Entry
	for _, e := range s.Entries {
		if e.Key != "" {
			out = append(out, e)
		}
	}
	return out
}
