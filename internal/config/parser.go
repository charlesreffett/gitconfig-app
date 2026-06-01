// Package config handles reading and writing Git's INI-style config files.
//
// # Git config format primer
//
// A gitconfig file looks like this:
//
//	[user]
//	    name = KnightYoshi
//	    email = alice@example.com
//
//	[includeIf "gitdir:~/work/"]
//	    path = ~/.gitconfig_work
//
// It is a variant of the classic INI format:
//   - A [section header] opens a new block.
//   - An optional "subsection" follows the name in quotes: [name "subsection"].
//   - Key=value pairs belong to the section above them.
//   - Lines starting with # or ; are comments.
//   - Blank lines are allowed anywhere.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Parse reads the gitconfig file at path and returns an in-memory Config.
//
// Every line is preserved — including comments and blank lines — so that a
// round-trip through Write produces output that is structurally identical to
// the input (modulo intentional edits).
func Parse(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	cfg := &Config{FilePath: path}

	// current points to the Section we are currently adding entries to.
	// It starts nil — if there are comments/blank lines before the first
	// section header we place them in a nameless "preamble" section.
	var current *Section

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		// ── blank line or comment ─────────────────────────────────────────────
		case trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";"):
			if current == nil {
				// Comments/blanks before the first [section] go into a nameless
				// preamble so they are preserved on write.
				cfg.Sections = append(cfg.Sections, Section{})
				current = &cfg.Sections[len(cfg.Sections)-1]
			}
			// Key is empty → Writer knows to emit RawLine verbatim.
			current.Entries = append(current.Entries, Entry{RawLine: line})

		// ── section header: [name] or [name "subsection"] ────────────────────
		case strings.HasPrefix(trimmed, "["):
			name, sub, err := parseHeader(trimmed)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}
			cfg.Sections = append(cfg.Sections, Section{Name: name, Subsection: sub})
			current = &cfg.Sections[len(cfg.Sections)-1]

		// ── key = value ───────────────────────────────────────────────────────
		default:
			if current == nil {
				// key=value before any section header — technically invalid git
				// config but we handle it gracefully rather than crashing.
				cfg.Sections = append(cfg.Sections, Section{})
				current = &cfg.Sections[len(cfg.Sections)-1]
			}
			key, value, err := parseKeyValue(trimmed)
			if err != nil {
				// Unknown line — preserve verbatim rather than discarding.
				current.Entries = append(current.Entries, Entry{RawLine: line})
				continue
			}
			current.Entries = append(current.Entries, Entry{
				Key:     key,
				Value:   value,
				RawLine: line, // keep original so we can detect if anything changed
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return cfg, nil
}

// parseHeader extracts the name and optional subsection from a header like
// [user] or [includeIf "gitdir:~/work/"].
func parseHeader(line string) (name, subsection string, err error) {
	// Strip the surrounding [ and ].
	inner := strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(line), "]"), "[")
	inner = strings.TrimSpace(inner)

	// If there is no quote, the whole inner string is the section name.
	quoteIdx := strings.Index(inner, "\"")
	if quoteIdx == -1 {
		return strings.TrimSpace(inner), "", nil
	}

	// Everything before the quote is the name; the quoted part is the subsection.
	name = strings.TrimSpace(inner[:quoteIdx])
	quoted := inner[quoteIdx:]
	if len(quoted) < 2 || !strings.HasPrefix(quoted, "\"") || !strings.HasSuffix(quoted, "\"") {
		return "", "", fmt.Errorf("malformed section header %q", line)
	}
	subsection = quoted[1 : len(quoted)-1]
	return name, subsection, nil
}

// parseKeyValue splits a "key = value" line into its two parts.
// Both leading/trailing whitespace around the = are stripped.
func parseKeyValue(line string) (key, value string, err error) {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return "", "", fmt.Errorf("no '=' in line %q", line)
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", fmt.Errorf("empty key in line %q", line)
	}
	return key, value, nil
}
