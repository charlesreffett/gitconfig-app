package blocks

// field_row.go builds the visual row for a single key=value entry inside a
// SectionBlock or CustomBlock.
//
// Layout:
//
//	[ key label (bold) ] [ input widget ]  [ × delete ]
//
// The input widget type is chosen based on the schema.FieldType:
//   - FieldText      → single-line Entry
//   - FieldMultiline → multi-line Entry
//   - FieldToggle    → Check box
//   - FieldDropdown  → Select widget

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/schema"
)

// BuildKnownFieldRow constructs a row widget for a field that exists in the
// section schema.
//
//   - field     — the schema definition (type, options, default, docs)
//   - value     — the current value string
//   - onChanged — called with the new string value whenever the input changes
//   - onDelete  — called when the user presses ×
func BuildKnownFieldRow(field schema.Field, value string, onChanged func(string), onDelete func()) fyne.CanvasObject {
	// Key label on the left — bold monospace to look like a config key.
	keyLabel := widget.NewLabelWithStyle(
		field.Label,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Monospace: true},
	)

	// Build the appropriate input widget for this field's type.
	inputWidget := buildInputWidget(field, value, onChanged)

	// Delete button on the right.
	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), onDelete)
	deleteBtn.Importance = widget.DangerImportance

	// Border layout: key label fixed-width on the left, delete button fixed-width
	// on the right, input widget stretches to fill the centre.
	return container.NewBorder(nil, nil, keyLabel, deleteBtn, inputWidget)
}

// BuildFreeFormRow constructs a row for a custom (free-form) key=value entry
// where the key is also user-editable (used by CustomBlock and [alias]).
//
//   - key, value — initial text values
//   - onKeyChanged   — called when the key entry changes
//   - onValueChanged — called when the value entry changes
//   - onDelete       — called when × is pressed
func BuildFreeFormRow(key, value string, onKeyChanged, onValueChanged func(string), onDelete func()) fyne.CanvasObject {
	keyEntry := widget.NewEntry()
	keyEntry.SetText(key)
	keyEntry.SetPlaceHolder("key")
	keyEntry.OnChanged = onKeyChanged

	valueEntry := widget.NewEntry()
	valueEntry.SetText(value)
	valueEntry.SetPlaceHolder("value")
	valueEntry.OnChanged = onValueChanged

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), onDelete)
	deleteBtn.Importance = widget.DangerImportance

	// Key entry takes ~30% of width, value entry fills the rest.
	// We use a split approach: key on the left in a fixed-width box,
	// value stretches in the centre, delete button on the far right.
	keyBox := container.NewGridWithColumns(1, keyEntry)
	keyBox.Resize(fyne.NewSize(120, keyEntry.MinSize().Height))

	return container.NewBorder(nil, nil, keyBox, deleteBtn, valueEntry)
}

// buildInputWidget creates the appropriate Fyne input widget for field.Type.
func buildInputWidget(field schema.Field, value string, onChanged func(string)) fyne.CanvasObject {
	switch field.Type {
	case schema.FieldToggle:
		// Parse the current value as a boolean for the initial check state.
		checked, _ := strconv.ParseBool(value)
		check := widget.NewCheck("", func(v bool) {
			onChanged(strconv.FormatBool(v))
		})
		check.Checked = checked
		return check

	case schema.FieldDropdown:
		sel := widget.NewSelect(field.Options, func(v string) {
			onChanged(v)
		})
		// Set the current value if it's in the options list; otherwise leave unset.
		for _, opt := range field.Options {
			if opt == value {
				sel.SetSelected(value)
				break
			}
		}
		// Show default as placeholder text when there is one.
		if field.Default != "" {
			sel.PlaceHolder = "(default: " + field.Default + ")"
		}
		return sel

	case schema.FieldMultiline:
		entry := widget.NewMultiLineEntry()
		entry.SetText(value)
		entry.SetPlaceHolder(field.Default)
		entry.OnChanged = onChanged
		return entry

	default: // FieldText and any unknown type → plain single-line entry
		entry := widget.NewEntry()
		entry.SetText(value)
		entry.SetPlaceHolder(field.Default)
		entry.OnChanged = onChanged
		return entry
	}
}
