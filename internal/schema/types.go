// Package schema defines the known structure of every git config section type.
//
// The schema is the bridge between raw gitconfig key=value pairs and the typed
// visual blocks the UI shows. It answers questions like:
//   - What fields are valid inside a [core] section?
//   - Should "autocrlf" be a text box or a dropdown?
//   - What are the allowed values for "pull.rebase"?
//   - Does [url] take a subsection, and what should the placeholder say?
//
// Nothing in this package touches the file system or the UI — it is pure
// data that both the block widgets (Phase 2) and the field-picker dropdown
// (the [+ Add field] button) read from.
package schema

// FieldType determines which Fyne input widget is rendered for a field.
type FieldType int

const (
	// FieldText is a plain single-line text entry.
	// Used for free-form values like name, email, file paths.
	FieldText FieldType = iota

	// FieldToggle is a boolean on/off switch.
	// Used for fields whose only valid values are true and false.
	FieldToggle

	// FieldDropdown is a select widget with a fixed list of valid values.
	// Used when git defines an explicit set of allowed strings.
	FieldDropdown

	// FieldMultiline is a multi-line text area.
	// Used for values that may contain newlines, such as commit templates.
	FieldMultiline
)

// Field describes a single key inside a config section — its name, how to
// display it, and what input widget it should render as.
type Field struct {
	// Key is the raw git config key, e.g. "autocrlf", "signingkey".
	Key string

	// Label is the human-readable name shown in the block UI, e.g.
	// "Line ending conversion" instead of "autocrlf".
	Label string

	// Type controls which input widget is used.
	Type FieldType

	// Options lists the valid values for FieldDropdown fields.
	// Ignored for other field types.
	Options []string

	// Default is the git default value, shown as a placeholder hint.
	// Empty string means no default (or the default is context-dependent).
	Default string

	// Common marks fields that are frequently used and should appear at the
	// top of the [+ Add field] picker. Uncommon fields are shown further down
	// under a "More…" expander or at least sorted lower.
	Common bool

	// Docs is a one-line description shown as a tooltip on the field label.
	// Taken from git-config(1) or the relevant git documentation.
	Docs string
}

// SectionSchema defines the known structure of one git config section type.
type SectionSchema struct {
	// Name is the canonical section name as it appears in gitconfig,
	// e.g. "user", "core", "includeIf".
	Name string

	// Label is the human-readable display name for the block header.
	Label string

	// Icon is the emoji used in the block library panel and block header.
	Icon string

	// HasSubsection indicates that this section type takes a quoted subsection
	// in its header, e.g. [url "git@github.com"] or [remote "origin"].
	// When true, the block header shows an editable subsection name field.
	HasSubsection bool

	// SubsectionLabel is the label shown next to the subsection field,
	// e.g. "URL pattern", "Remote name", "Filter name".
	SubsectionLabel string

	// SubsectionPlaceholder is the placeholder text in the subsection field.
	SubsectionPlaceholder string

	// Fields lists every known key for this section, in display order.
	// Common fields appear first; uncommon fields follow.
	Fields []Field

	// FreeForm means the section accepts arbitrary user-defined keys rather
	// than (or in addition to) the fixed Fields list. Used for [alias], where
	// every key is a custom command name the user invents.
	FreeForm bool
}
