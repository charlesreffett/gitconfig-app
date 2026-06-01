package ui

// editor.go builds the right-side content panel for a selected section.
//
// Two kinds of editor exist:
//
//  1. Regular editor  — for sections like [user], [core], [filter "lfs"].
//     Shows a scrollable list of editable key/value rows plus an Add Entry
//     button.
//
//  2. Include editor  — for [include] and [includeIf "gitdir:…"] sections.
//     Shows the condition and path with an "Open linked file" button that
//     navigates into the referenced config.

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/config"
)

// ── Regular section editor ────────────────────────────────────────────────────

// buildRegularEditor returns a VBox of key/value rows for sec.
//
// onChange is called whenever the entry list changes structurally (a row is
// added or deleted) so the parent can rebuild the editor.  Value edits are
// handled inline via OnChanged and do NOT trigger a full rebuild.
func buildRegularEditor(sec *config.Section, onChange func()) fyne.CanvasObject {
	// Title row.
	header := widget.NewLabel(formatHeader(sec))
	header.TextStyle = fyne.TextStyle{Bold: true}

	sep := widget.NewSeparator()

	// Build one row per key=value entry (skip blank/comment entries).
	rows := buildEntryRows(sec, onChange)

	// "Add entry" button at the bottom.
	addBtn := widget.NewButtonWithIcon("Add entry", theme.ContentAddIcon(), func() {
		showAddEntryDialog(sec, onChange)
	})

	// container.NewVBox stacks children vertically with natural (minimum) height.
	// It grows as items are added — that is why we wrap it in a container.NewScroll
	// in app.go: so a long section doesn't clip.
	return container.NewVBox(
		append([]fyne.CanvasObject{header, sep}, append(rows, addBtn)...)...,
	)
}

// buildEntryRows returns one row widget per key=value entry in sec.
// Comment/blank entries are skipped in the visual list (they are preserved in
// memory and will be written out on Save).
func buildEntryRows(sec *config.Section, onChange func()) []fyne.CanvasObject {
	var rows []fyne.CanvasObject

	for i := range sec.Entries {
		// We must capture i by value in the closure — if we captured the loop
		// variable directly, all closures would share the same (final) value.
		//
		// This is a classic Go gotcha: loop variables are reused each iteration,
		// so any closure that references them sees the last value after the loop
		// finishes. The fix is to create a local copy inside the loop body.
		idx := i
		e := &sec.Entries[idx]

		if e.Key == "" {
			// Comment or blank line — not shown in the editor, but kept in memory.
			continue
		}

		row := buildEntryRow(sec, idx, e, onChange)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		empty := widget.NewLabel("No entries yet. Use Add entry to create one.")
		empty.TextStyle = fyne.TextStyle{Italic: true}
		rows = append(rows, empty)
	}

	return rows
}

// buildEntryRow builds a single three-column row:
//
//	[ key label ] [ value entry ] [ delete button ]
//
// The key is displayed as a read-only label. Editable keys would risk creating
// invalid git config (git is case-sensitive for some keys). The user can
// always delete and re-add an entry to rename a key.
func buildEntryRow(sec *config.Section, idx int, e *config.Entry, onChange func()) fyne.CanvasObject {
	keyLabel := widget.NewLabel(e.Key)
	keyLabel.TextStyle = fyne.TextStyle{Monospace: true}

	valueEntry := widget.NewEntry()
	valueEntry.SetText(e.Value)
	// OnChanged fires every time the text changes (each keystroke).
	// We write directly into the in-memory Entry so the data is always current.
	// The Save button commits to disk.
	valueEntry.OnChanged = func(newVal string) {
		sec.Entries[idx].Value = newVal
	}

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		// Remove the entry at idx from the slice.
		//
		// append(a[:i], a[i+1:]...) is the idiomatic Go way to delete element i.
		// It shifts everything after idx one position left in-place.
		sec.Entries = append(sec.Entries[:idx], sec.Entries[idx+1:]...)
		onChange() // trigger rebuild so the deleted row disappears
	})
	deleteBtn.Importance = widget.DangerImportance

	// container.NewGridWithColumns creates a grid with a fixed column count.
	// Each column gets an equal share of horizontal space.
	// We give the key 25% of the space and the value the rest by nesting grids:
	//   outer 4-col grid: [key] [value] [value] [delete]
	//
	// A simpler alternative is a 3-column grid with weighted sizing, but Fyne's
	// grid doesn't support weights — equal columns only.  We use a Border
	// container instead: key on the left at natural width, delete on the right
	// at natural width, and value stretches to fill the middle.
	return container.NewBorder(nil, nil,
		container.NewHBox(keyLabel), // fixed-width left
		deleteBtn,                   // fixed-width right
		valueEntry,                  // stretches to fill centre
	)
}

// showAddEntryDialog opens a small form for a new key=value pair.
func showAddEntryDialog(sec *config.Section, onChange func()) {
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("e.g. name, email, signingkey")

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("e.g. KnightYoshi")

	items := []*widget.FormItem{
		{Text: "Key", Widget: keyEntry},
		{Text: "Value", Widget: valueEntry},
	}

	// We need the Fyne window to attach the dialog to; we grab it from the
	// canvas of any existing widget. In a larger app you'd pass the window
	// explicitly — here the canvas approach is the simplest path.
	w := fyne.CurrentApp().Driver().AllWindows()[0]

	dialog.ShowForm("Add Entry", "Add", "Cancel", items, func(confirmed bool) {
		if !confirmed {
			return
		}
		key := strings.TrimSpace(keyEntry.Text)
		if key == "" {
			return
		}
		value := strings.TrimSpace(valueEntry.Text)
		sec.Entries = append(sec.Entries, config.Entry{Key: key, Value: value})
		onChange()
	}, w)
}

// ── Include section editor ────────────────────────────────────────────────────

// buildIncludeEditor returns a panel for [include] and [includeIf] sections.
//
// onOpen is called with the resolved path when the user clicks "Open linked file".
func buildIncludeEditor(sec *config.Section, onOpen func(path string)) fyne.CanvasObject {
	header := widget.NewLabel(formatHeader(sec))
	header.TextStyle = fyne.TextStyle{Bold: true}

	sep := widget.NewSeparator()

	rows := []fyne.CanvasObject{header, sep}

	// Show the condition (the subsection) if present.
	if sec.Subsection != "" {
		condLabel := widget.NewLabel("Condition:")
		condLabel.TextStyle = fyne.TextStyle{Bold: true}
		condValue := widget.NewLabel(sec.Subsection)
		condValue.TextStyle = fyne.TextStyle{Monospace: true}
		rows = append(rows,
			container.NewBorder(nil, nil, condLabel, nil, condValue),
		)
	}

	// Find the path entry — this is the linked config file.
	linkedPath := ""
	for _, e := range sec.Entries {
		if e.Key == "path" {
			linkedPath = e.Value
			break
		}
	}

	pathLabel := widget.NewLabel("Path:")
	pathLabel.TextStyle = fyne.TextStyle{Bold: true}

	pathEntry := widget.NewEntry()
	pathEntry.SetText(linkedPath)
	// Update the in-memory entry when the path is edited.
	pathEntry.OnChanged = func(newVal string) {
		for i := range sec.Entries {
			if sec.Entries[i].Key == "path" {
				sec.Entries[i].Value = newVal
				linkedPath = newVal
				return
			}
		}
		// No path entry yet — add one.
		sec.Entries = append(sec.Entries, config.Entry{Key: "path", Value: newVal})
		linkedPath = newVal
	}

	openBtn := widget.NewButtonWithIcon("Open linked file", theme.FolderOpenIcon(), func() {
		if linkedPath == "" {
			return
		}
		onOpen(linkedPath)
	})
	openBtn.Importance = widget.HighImportance

	rows = append(rows,
		container.NewBorder(nil, nil, pathLabel, nil, pathEntry),
		openBtn,
	)

	// Info note explaining what includeIf does — helpful for users new to gitconfig.
	note := widget.NewRichTextFromMarkdown(
		"**How conditional includes work**\n\n" +
			"Git loads the linked config file only when the `gitdir:` " +
			"pattern matches the current repository's `.git` directory. " +
			"This lets you use a different identity (name, email, SSH key) " +
			"in work repositories without touching your personal settings.",
	)
	note.Wrapping = fyne.TextWrapWord

	rows = append(rows, widget.NewSeparator(), note)

	return container.NewVBox(rows...)
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// formatHeader returns the section header string as it appears in the file,
// e.g. `[user]` or `[includeIf "gitdir:~/work/"]`.
func formatHeader(sec *config.Section) string {
	if sec.Subsection == "" {
		return "[" + sec.Name + "]"
	}
	return `[` + sec.Name + ` "` + sec.Subsection + `"]`
}
