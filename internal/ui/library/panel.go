package library

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/schema"
	"github.com/knightyoshi/git-config-manager/internal/ui/blocks"
)

// Panel is the left-side block library widget. It shows two sections:
//
//  1. "Boundaries" — ConditionalBlock and FileBlock templates
//  2. "Sections"   — one entry per schema in schema.All
//
// Each entry can be clicked (to add a block via keyboard/mouse) or dragged
// onto the canvas or a container block.
type Panel struct {
	widget.BaseWidget
	content *fyne.Container
}

// NewPanel builds the library panel.
//
//   - onAdd — called with (kind, schemaName) when an item is tapped or dropped
func NewPanel(onAdd func(kind blocks.BlockKind, schemaName string)) *Panel {
	p := &Panel{}
	p.ExtendBaseWidget(p)

	// ── Boundaries section ──────────────────────────────────────────────────
	boundaryHeader := makeSectionHeader("Boundaries")

	conditionalItem := newLibraryItem(
		"⚡ Conditional Include",
		blocks.KindConditional, "",
		onAdd,
	)
	fileItem := newLibraryItem(
		"📄 File",
		blocks.KindFile, "",
		onAdd,
	)

	// ── Sections section ───────────────────────────────────────────────────
	sectionHeader := makeSectionHeader("Sections")

	var sectionItems []fyne.CanvasObject
	for _, s := range schema.All {
		// Capture loop variable: without this, all closures would reference
		// the same (final) schema s after the loop ends.
		s := s
		item := newLibraryItem(s.Icon+" "+s.Label, blocks.KindSection, s.Name, onAdd)
		sectionItems = append(sectionItems, item)
	}

	// ── Assemble all items ──────────────────────────────────────────────────
	all := []fyne.CanvasObject{
		boundaryHeader,
		conditionalItem,
		fileItem,
		sectionHeader,
	}
	all = append(all, sectionItems...)

	scroll := container.NewScroll(container.NewVBox(all...))
	p.content = container.NewStack(scroll)
	return p
}

// CreateRenderer implements fyne.Widget.
func (p *Panel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.content)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// makeSectionHeader creates a bold label used as a visual group header inside
// the library panel (e.g. "Boundaries", "Sections").
func makeSectionHeader(title string) fyne.CanvasObject {
	lbl := widget.NewLabel(title)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(lbl, widget.NewSeparator())
}

// newItemContainer wraps the given label in a padded container that looks like
// a clickable library row.
func newItemContainer(lbl *widget.Label) *fyne.Container {
	return container.NewStack(lbl)
}
