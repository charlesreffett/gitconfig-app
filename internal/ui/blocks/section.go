package blocks

// section.go implements SectionBlock — the visual card for a single git config
// section like [user], [core], [remote "origin"], etc.
//
// The block header shows the section icon and name. For sections with
// subsections (HasSubsection == true), the header also shows an editable
// subsection field (e.g. the remote name for [remote "origin"]).
//
// The body renders:
//   - One FieldRow per currently-present key=value entry
//   - A [+ Add field] Select widget listing fields not yet present
//   - For FreeForm sections (e.g. [alias]): key+value Entry rows instead

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/config"
	"github.com/knightyoshi/git-config-manager/internal/schema"
)

// SectionEntry holds one key=value pair within a section block.
type SectionEntry struct {
	Key        string
	Value      string
	Superseded bool // true when the same key is overridden in the project config
}

// SectionBlock is the visual card widget for a known git config section.
// It embeds blockBase for the shared card chrome (drag handle, header, border).
type SectionBlock struct {
	widget.BaseWidget

	base       *blockBase
	sec        schema.SectionSchema // schema describing valid fields for this section
	subsection string               // e.g. "origin" for [remote "origin"]
	entries    []SectionEntry       // the live key=value data being edited
	onChange   func()               // called when the user makes any change
}

// NewSectionBlock creates a SectionBlock pre-populated with existing entries.
//
//   - sec        — the schema for this section type
//   - subsection — subsection name (empty if none)
//   - cfgEntries — initial entries loaded from the config file
//   - onChange   — called on any user edit so the parent can mark unsaved changes
func NewSectionBlock(sec schema.SectionSchema, subsection string, cfgEntries []config.Entry, onChange func()) *SectionBlock {
	b := &SectionBlock{
		sec:        sec,
		subsection: subsection,
		onChange:   onChange,
	}
	b.ExtendBaseWidget(b)

	// Convert config.Entry slice to SectionEntry slice (adds Superseded field).
	for _, e := range cfgEntries {
		if e.Key != "" { // skip comment/blank lines
			b.entries = append(b.entries, SectionEntry{Key: e.Key, Value: e.Value})
		}
	}

	b.rebuild()
	return b
}

// rebuild re-renders the entire block. It is called after any structural change
// (adding or removing a field row). Individual value edits do NOT call rebuild
// — they update b.entries in place via OnChanged closures.
func (b *SectionBlock) rebuild() {
	// Build the body content for this section.
	body := b.buildBody()

	// Determine the header label.
	// For sections with subsections, append a small Entry field to the header.
	label := b.sec.Label
	headerBody := fyne.CanvasObject(widget.NewLabel("")) // placeholder
	if b.sec.HasSubsection {
		subEntry := widget.NewEntry()
		subEntry.SetText(b.subsection)
		subEntry.SetPlaceHolder(b.sec.SubsectionPlaceholder)
		subEntry.OnChanged = func(v string) {
			b.subsection = v
			if b.onChange != nil {
				b.onChange()
			}
		}
		// Wrap the entry in a labelled row so it fits in the header.
		headerBody = container.NewBorder(nil, nil,
			widget.NewLabel(b.sec.SubsectionLabel+":"), nil,
			subEntry,
		)
	}

	// If we already have a base, rebuild its inner container.
	// Otherwise, create it fresh.
	if b.base == nil {
		var bodyWithSub fyne.CanvasObject
		if b.sec.HasSubsection {
			bodyWithSub = container.NewVBox(headerBody, body)
		} else {
			bodyWithSub = body
		}
		b.base = newBlockBase(KindSection, b.sec.Icon, label, bodyWithSub, func() {
			if b.onChange != nil {
				b.onChange()
			}
		})
	} else {
		// Rebuild inner by creating a fresh body and replacing the inner container.
		var bodyWithSub fyne.CanvasObject
		if b.sec.HasSubsection {
			bodyWithSub = container.NewVBox(headerBody, body)
		} else {
			bodyWithSub = body
		}
		b.base = newBlockBase(KindSection, b.sec.Icon, label, bodyWithSub, b.base.onRemove)
	}
}

// buildBody constructs the VBox containing field rows + the [+ Add field] picker.
func (b *SectionBlock) buildBody() fyne.CanvasObject {
	var rows []fyne.CanvasObject

	if b.sec.FreeForm {
		// FreeForm sections (like [alias]) allow arbitrary key names.
		rows = b.buildFreeFormRows()
	} else {
		// Known sections render typed field widgets.
		rows = b.buildKnownRows()
	}

	// [+ Add field] picker at the bottom.
	addPicker := b.buildAddPicker()
	rows = append(rows, widget.NewSeparator(), addPicker)

	return container.NewVBox(rows...)
}

// buildKnownRows creates one FieldRow per entry in b.entries using the schema.
func (b *SectionBlock) buildKnownRows() []fyne.CanvasObject {
	var rows []fyne.CanvasObject
	for i := range b.entries {
		idx := i // capture loop variable
		entry := &b.entries[idx]

		// Look up the schema field to get the correct input widget type.
		field, ok := schema.FieldByKey(b.sec, entry.Key)
		if !ok {
			// Key not in schema — render as a plain text row.
			field = schema.Field{
				Key:   entry.Key,
				Label: entry.Key,
				Type:  schema.FieldText,
			}
		}

		row := BuildKnownFieldRow(
			field,
			entry.Value,
			func(v string) {
				b.entries[idx].Value = v
				if b.onChange != nil {
					b.onChange()
				}
			},
			func() {
				// Remove this entry and rebuild.
				b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
				b.rebuild()
				if b.onChange != nil {
					b.onChange()
				}
			},
		)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		placeholder := widget.NewLabel("No fields yet — use [+ Add field] below.")
		placeholder.TextStyle = fyne.TextStyle{Italic: true}
		rows = append(rows, placeholder)
	}
	return rows
}

// buildFreeFormRows creates editable key+value rows for FreeForm sections.
func (b *SectionBlock) buildFreeFormRows() []fyne.CanvasObject {
	var rows []fyne.CanvasObject
	for i := range b.entries {
		idx := i
		entry := &b.entries[idx]

		row := BuildFreeFormRow(
			entry.Key,
			entry.Value,
			func(k string) {
				b.entries[idx].Key = k
				if b.onChange != nil {
					b.onChange()
				}
			},
			func(v string) {
				b.entries[idx].Value = v
				if b.onChange != nil {
					b.onChange()
				}
			},
			func() {
				b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
				b.rebuild()
				if b.onChange != nil {
					b.onChange()
				}
			},
		)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		placeholder := widget.NewLabel("No entries yet — use [+ Add entry] below.")
		placeholder.TextStyle = fyne.TextStyle{Italic: true}
		rows = append(rows, placeholder)
	}
	return rows
}

// buildAddPicker returns the [+ Add field] select widget that lists fields
// not yet present in the block.
func (b *SectionBlock) buildAddPicker() fyne.CanvasObject {
	if b.sec.FreeForm {
		// FreeForm: just a plain [+ Add entry] button.
		return widget.NewButton("+ Add entry", func() {
			b.entries = append(b.entries, SectionEntry{Key: "", Value: ""})
			b.rebuild()
			if b.onChange != nil {
				b.onChange()
			}
		})
	}

	// Collect keys already present so we can exclude them from the picker.
	present := make(map[string]bool, len(b.entries))
	for _, e := range b.entries {
		present[e.Key] = true
	}

	// Build the options list: common fields first, then a divider, then uncommon.
	var options []string
	for _, f := range schema.CommonFields(b.sec) {
		if !present[f.Key] {
			options = append(options, f.Label+" ("+f.Key+")")
		}
	}
	uncommon := schema.UncommonFields(b.sec)
	hasUncommon := false
	for _, f := range uncommon {
		if !present[f.Key] {
			if !hasUncommon {
				options = append(options, "── more ──")
				hasUncommon = true
			}
			options = append(options, f.Label+" ("+f.Key+")")
		}
	}

	if len(options) == 0 {
		// All fields already added.
		lbl := widget.NewLabel("All known fields added.")
		lbl.TextStyle = fyne.TextStyle{Italic: true}
		return lbl
	}

	picker := widget.NewSelect(options, func(selected string) {})
	picker.PlaceHolder = "+ Add field…"

	picker.OnChanged = func(selected string) {
		if selected == "" || selected == "── more ──" {
			return
		}
		// Extract the key from "Label (key)" format.
		key := extractKey(selected)
		if key == "" {
			return
		}
		b.entries = append(b.entries, SectionEntry{Key: key, Value: ""})
		b.rebuild()
		if b.onChange != nil {
			b.onChange()
		}
	}

	return picker
}

// extractKey parses the key out of a "Label (key)" picker option string.
// Returns empty string if parsing fails.
func extractKey(option string) string {
	// Format is "Human Label (rawKey)" — find the last '(' and ')'.
	start := -1
	for i := len(option) - 1; i >= 0; i-- {
		if option[i] == ')' {
			continue
		}
		if option[i] == '(' {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(option) - 1
	// Trim from start+1 to last ')'.
	for end >= 0 && option[end] != ')' {
		end--
	}
	if end <= start {
		return ""
	}
	return option[start+1 : end]
}

// CreateRenderer implements fyne.Widget by delegating to the base card.
func (b *SectionBlock) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.base.inner)
}

// SetOnRemove sets the callback invoked when the × button is pressed.
// Call this after attaching the block to its parent container.
func (b *SectionBlock) SetOnRemove(fn func()) {
	b.base.onRemove = fn
}

// ToSections serialises this block back into config.Section structs for saving.
func (b *SectionBlock) ToSections() []config.Section {
	sec := config.Section{
		Name:       b.sec.Name,
		Subsection: b.subsection,
	}
	for _, e := range b.entries {
		if e.Key != "" {
			sec.Entries = append(sec.Entries, config.Entry{
				Key:   e.Key,
				Value: e.Value,
			})
		}
	}
	return []config.Section{sec}
}

// Entries returns the current slice of section entries (for superseded-key
// highlighting from the app-level comparison logic).
func (b *SectionBlock) Entries() []SectionEntry {
	return b.entries
}

// MarkSuperseded marks the entry at idx as superseded (overridden by project
// config). The block is refreshed to re-render with a dimmed style.
func (b *SectionBlock) MarkSuperseded(key string, superseded bool) {
	for i := range b.entries {
		if b.entries[i].Key == key {
			b.entries[i].Superseded = superseded
		}
	}
	b.rebuild()
	b.Refresh()
}

// SectionName returns the schema section name (e.g. "core", "user").
func (b *SectionBlock) SectionName() string { return b.sec.Name }

// Subsection returns the current subsection value.
func (b *SectionBlock) Subsection() string { return b.subsection }
