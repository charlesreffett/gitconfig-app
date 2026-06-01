package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Write serialises cfg back to disk at cfg.FilePath.
//
// # Atomic writes
//
// Writing directly to the target file is dangerous: if the process dies
// mid-write you end up with a half-written (and therefore broken) gitconfig.
// Instead we:
//  1. Write the new content to a temporary file in the same directory.
//  2. Call os.Rename to swap the temp file over the target atomically.
//
// On POSIX systems (macOS, Linux) os.Rename is a single syscall — either the
// old file is there or the new one is; there is no in-between.
func Write(cfg *Config) error {
	dir := filepath.Dir(cfg.FilePath)

	// Create the temp file in the same directory so Rename is always on the
	// same filesystem (cross-device rename would fail).
	tmp, err := os.CreateTemp(dir, ".gitconfig-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	// Guard: if anything goes wrong before the final Rename, delete the temp
	// file so we don't litter the user's home directory.
	ok := false
	defer func() {
		if !ok {
			os.Remove(tmpName)
		}
	}()

	if err := encode(tmp, cfg); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("flush temp file: %w", err)
	}

	// Atomic swap.
	if err := os.Rename(tmpName, cfg.FilePath); err != nil {
		return fmt.Errorf("rename %s → %s: %w", tmpName, cfg.FilePath, err)
	}
	ok = true
	return nil
}

// encode serialises all sections and their entries into the writer.
//
// Output format follows git's canonical style:
//   - Section headers are not indented.
//   - Key=value pairs are indented with a single tab.
//   - Comments and blank lines are emitted verbatim (they carry their original
//     indentation in RawLine).
//   - A blank line is inserted between named sections when the preceding
//     section does not already end with one — this keeps the file readable.
func encode(f *os.File, cfg *Config) error {
	var sb strings.Builder

	for i, sec := range cfg.Sections {
		// Nameless "preamble" sections hold comments that appeared before
		// the first real header; we skip writing a header line for them.
		if sec.Name != "" {
			if sec.Subsection != "" {
				fmt.Fprintf(&sb, "[%s \"%s\"]\n", sec.Name, sec.Subsection)
			} else {
				fmt.Fprintf(&sb, "[%s]\n", sec.Name)
			}
		}

		for _, e := range sec.Entries {
			if e.Key == "" {
				// Comment or blank line — write exactly as we read it.
				sb.WriteString(e.RawLine + "\n")
			} else {
				// Real key=value — canonical tab-indented form.
				fmt.Fprintf(&sb, "\t%s = %s\n", e.Key, e.Value)
			}
		}

		// Insert a blank separator between named sections when the current
		// section does not already end with a blank/comment line.
		if i < len(cfg.Sections)-1 && sec.Name != "" {
			needsBlank := true
			if len(sec.Entries) > 0 {
				last := sec.Entries[len(sec.Entries)-1]
				if strings.TrimSpace(last.RawLine) == "" {
					needsBlank = false
				}
			}
			if needsBlank {
				sb.WriteString("\n")
			}
		}
	}

	_, err := f.WriteString(sb.String())
	return err
}

// ── Helpers for callers ───────────────────────────────────────────────────────

// SetEntry updates the value of an existing key in a section, or appends a
// new entry if the key doesn't exist yet.
//
// This is a convenience used by the UI: instead of the caller iterating
// through entries, they just call SetEntry and let the writer handle it.
func SetEntry(sec *Section, key, value string) {
	for i := range sec.Entries {
		if sec.Entries[i].Key == key {
			sec.Entries[i].Value = value
			sec.Entries[i].RawLine = "" // mark as "needs canonical formatting"
			return
		}
	}
	// Key doesn't exist yet — append it.
	sec.Entries = append(sec.Entries, Entry{Key: key, Value: value})
}

// DeleteEntry removes all entries with the given key from a section.
func DeleteEntry(sec *Section, key string) {
	kept := sec.Entries[:0] // reuse underlying array; avoid allocation
	for _, e := range sec.Entries {
		if e.Key != key {
			kept = append(kept, e)
		}
	}
	sec.Entries = kept
}
