package blocks

// custom.go implements CustomBlock — the visual card for config sections whose
// name does not appear in the schema registry (e.g. tool-specific sections like
// [difftool "vimdiff"] that git-config-manager doesn't have a schema for).
//
// CustomBlock renders all entries as free-form key=value rows because we don't
// know what fields are valid. It also displays the raw section name so the user
// can recognise what they're editing.

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/config"
)

// CustomBlock is the visual card for an unknown (not in schema) config section.
// All entries are rendered as editable key=value pairs.
type CustomBlock struct {
	widget.BaseWidget

	base       *blockBase
	name       string         // raw section name, e.g. "difftool"
	subsection string         // e.g. "vimdiff" for [difftool "vimdiff"]
	entries    []SectionEntry // current key=value rows
	onChange   func()
}

// NewCustomBlock creates a CustomBlock for an unknown section.
//
//   - name       — the raw section name from the gitconfig file
//   - subsection — subsection string (empty if none)
//   - cfgEntries — initial entries loaded from the config file
//   - onChange   — called on any user edit
func NewCustomBlock(name, subsection string, cfgEntries []config.Entry, onChange func()) *CustomBlock {
	b := &CustomBlock{
		name:       name,
		subsection: subsection,
		onChange:   onChange,
	}
	b.ExtendBaseWidget(b)

	// Convert config.Entry → SectionEntry (adds Superseded field).
	for _, e := range cfgEntries {
		if e.Key != "" {
			b.entries = append(b.entries, SectionEntry{Key: e.Key, Value: e.Value})
		}
	}

	b.rebuild()
	return b
}

// rebuild regenerates the block's inner content. Called after structural
// changes (add/remove row). Value edits are handled inline via OnChanged.
func (b *CustomBlock) rebuild() {
	body := b.buildBody()

	// Display header: show subsection in the label if present.
	label := b.name
	if b.subsection != "" {
		label = b.name + " · " + b.subsection
	}

	if b.base == nil {
		b.base = newBlockBase(KindSection, "⚙", label, body, func() {
			if b.onChange != nil {
				b.onChange()
			}
		})
	} else {
		b.base = newBlockBase(KindSection, "⚙", label, body, b.base.onRemove)
	}
}

// buildBody builds the VBox with free-form key=value rows + [+ Add entry] button.
func (b *CustomBlock) buildBody() fyne.CanvasObject {
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
		placeholder := widget.NewLabel("No entries yet.")
		placeholder.TextStyle = fyne.TextStyle{Italic: true}
		rows = append(rows, placeholder)
	}

	addBtn := widget.NewButton("+ Add entry", func() {
		b.entries = append(b.entries, SectionEntry{Key: "", Value: ""})
		b.rebuild()
		if b.onChange != nil {
			b.onChange()
		}
	})
	rows = append(rows, widget.NewSeparator(), addBtn)

	return container.NewVBox(rows...)
}

// CreateRenderer implements fyne.Widget.
func (b *CustomBlock) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.base.inner)
}

// SetOnRemove sets the callback for the × button.
func (b *CustomBlock) SetOnRemove(fn func()) {
	b.base.onRemove = fn
}

// ToSections serialises the block into config.Section structs.
func (b *CustomBlock) ToSections() []config.Section {
	sec := config.Section{
		Name:       b.name,
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

// SectionName returns the raw section name.
func (b *CustomBlock) SectionName() string { return b.name }

// Subsection returns the current subsection value.
func (b *CustomBlock) Subsection() string { return b.subsection }
