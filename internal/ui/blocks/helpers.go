package blocks

// helpers.go provides shared utilities used across multiple block types.

import (
	"fyne.io/fyne/v2"

	"github.com/knightyoshi/git-config-manager/internal/config"
	"github.com/knightyoshi/git-config-manager/internal/schema"
)

// BlockSerializer is implemented by any block type that can be serialised back
// into config.Section structs for writing to disk.
//
// All four block types — SectionBlock, CustomBlock, FileBlock, ConditionalBlock —
// implement this interface. The canvas save logic iterates top-level blocks and
// calls ToSections() on each to reconstruct the config.
type BlockSerializer interface {
	ToSections() []config.Section
}

// CreateBlockFromPayload creates a new fyne.CanvasObject block from a
// DragPayload. This is called by drop targets when a block is successfully dropped.
//
//   - payload  — the drag payload set by the drag source
//   - onChange — called whenever the new block's content changes
//
// Returns nil if the payload kind is unrecognised.
func CreateBlockFromPayload(payload DragPayload, onChange func()) fyne.CanvasObject {
	switch payload.Kind {
	case KindSection:
		if payload.SchemaName != "" {
			if sec, ok := schema.Lookup(payload.SchemaName); ok {
				return NewSectionBlock(sec, "", nil, onChange)
			}
		}
		// Unknown schema name — fall back to custom block.
		return NewCustomBlock(payload.SchemaName, "", nil, onChange)

	case KindFile:
		return NewFileBlock("", onChange)

	case KindConditional:
		return NewConditionalBlock(onChange)
	}
	return nil
}

// pathValue extracts the "path" entry value from a config.Section.
// Returns empty string if no path entry exists.
func pathValue(sec config.Section) string {
	for _, e := range sec.Entries {
		if e.Key == "path" {
			return e.Value
		}
	}
	return ""
}
