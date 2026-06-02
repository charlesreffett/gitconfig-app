// Package library provides the left-side block library panel where users can
// drag or click blocks to add them to the config canvas.
//
// The library shows two groups:
//  1. Boundaries — Conditional Include and File blocks
//  2. Sections   — one item per schema in schema.All (22 known git config sections)
package library

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/ui/blocks"
)

// libraryItem is a draggable and tappable entry in the block library.
// Clicking it calls onAdd (so the user can add a block without needing to
// drag). Dragging it populates blocks.Current so drop targets can react.
type libraryItem struct {
	widget.BaseWidget

	kind       blocks.BlockKind
	schemaName string
	label      *widget.Label
	inner      *fyne.Container

	// onAdd is called when the item is tapped (clicked). It receives the block
	// kind and schema name so the caller can add the block to the active canvas.
	onAdd func(blocks.BlockKind, string)
}

// newLibraryItem creates a single library panel entry.
//
//   - displayLabel — text shown in the panel (e.g. "🔧 Core")
//   - kind         — the BlockKind this item creates
//   - schemaName   — schema registry name, or "" for boundary items
//   - onAdd        — called when the item is tapped
func newLibraryItem(displayLabel string, kind blocks.BlockKind, schemaName string, onAdd func(blocks.BlockKind, string)) *libraryItem {
	i := &libraryItem{
		kind:       kind,
		schemaName: schemaName,
		onAdd:      onAdd,
	}
	i.ExtendBaseWidget(i)

	i.label = widget.NewLabel(displayLabel)
	i.inner = newItemContainer(i.label)
	return i
}

// CreateRenderer implements fyne.Widget.
func (i *libraryItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(i.inner)
}

// Tapped implements fyne.Tappable. Called when the user single-clicks the item.
func (i *libraryItem) Tapped(_ *fyne.PointEvent) {
	if i.onAdd != nil {
		i.onAdd(i.kind, i.schemaName)
	}
}

// Dragged implements fyne.Draggable. Populates blocks.Current so that any
// drop target the user drags over can inspect the payload.
func (i *libraryItem) Dragged(_ *fyne.DragEvent) {
	blocks.Current = blocks.DragPayload{
		Active:     true,
		Kind:       i.kind,
		SchemaName: i.schemaName,
		// RemoveFromSource is nil for library items — they are templates, not
		// objects that need to be removed from anywhere.
	}
}

// DragEnd implements fyne.Draggable. Clears the Active flag when the drag ends.
func (i *libraryItem) DragEnd() {
	blocks.Current.Active = false
}
