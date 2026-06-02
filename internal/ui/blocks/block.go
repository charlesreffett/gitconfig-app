package blocks

// block.go defines blockBase — the shared visual shell that every block type
// embeds. It renders a coloured rounded card with:
//
//   - A drag handle button (≡) on the left side of the header
//   - An icon + bold title label in the centre of the header
//   - A remove button (×) on the right side of the header
//   - A horizontal separator between the header and block body
//   - The body content (unique to each block type) below the separator
//
// Colour coding makes blocks identifiable at a glance:
//   - Blue   → ConditionalBlock (wraps an included file with conditions)
//   - Purple → FileBlock        (a config file boundary)
//   - Theme  → SectionBlock     (a regular config section)

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// hoverColor is the bright blue highlight drawn as a border around a block
// when a dragged item is held over it and the drop would be accepted.
var hoverColor = color.NRGBA{R: 80, G: 180, B: 255, A: 255}

// blockBase is the visual shell shared by all block types. It is not exported
// because external code works through ConditionalBlock, FileBlock, etc.
//
// Embedding widget.BaseWidget is the standard Fyne pattern for custom widgets.
// It provides the Refresh/Resize/Show/Hide plumbing.
type blockBase struct {
	widget.BaseWidget

	// kind identifies this block for drag-and-drop acceptance checks.
	kind BlockKind

	// bg is the background rectangle. Its fill colour changes per block kind.
	bg *canvas.Rectangle

	// border is drawn on top of bg when a compatible block is dragged over this
	// container. Normally its StrokeWidth is 0 (invisible).
	border *canvas.Rectangle

	// inner is the full composited content: stack(bg, border, vbox(header, sep, body)).
	inner *fyne.Container

	// onRemove is called when the user presses the × button.
	// Each block type sets this to remove itself from its parent container.
	onRemove func()
}

// newBlockBase constructs the shared card shell for a block.
//
//   - kind  — used for colour selection and drag acceptance rules
//   - icon  — emoji displayed in the header, e.g. "🔧"
//   - label — human-readable title in the header, e.g. "Core"
//   - body  — the block-type-specific content (fields, children, etc.)
//   - onRemove — called when the × button is pressed
func newBlockBase(kind BlockKind, icon, label string, body fyne.CanvasObject, onRemove func()) *blockBase {
	b := &blockBase{
		kind:     kind,
		onRemove: onRemove,
	}
	b.ExtendBaseWidget(b)

	// ── Background ──────────────────────────────────────────────────────────
	b.bg = canvas.NewRectangle(bgColor(kind))
	b.bg.CornerRadius = 6

	// ── Hover border ────────────────────────────────────────────────────────
	// StrokeWidth 0 means invisible; set to 2 when a drag enters this container.
	b.border = canvas.NewRectangle(color.Transparent)
	b.border.StrokeColor = color.Transparent
	b.border.StrokeWidth = 0
	b.border.CornerRadius = 6
	b.border.FillColor = color.Transparent

	// ── Drag handle ─────────────────────────────────────────────────────────
	// The ≡ button is the drag initiator. LowImportance makes it look subtle.
	handle := &dragHandle{base: b}
	handle.ExtendBaseWidget(handle)

	// ── Title label ─────────────────────────────────────────────────────────
	titleText := icon + " " + label
	titleLabel := widget.NewLabel(titleText)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// ── Remove button ───────────────────────────────────────────────────────
	removeBtn := widget.NewButton("×", func() {
		if b.onRemove != nil {
			b.onRemove()
		}
	})
	removeBtn.Importance = widget.LowImportance

	// ── Header: drag handle | title | remove button ──────────────────────
	header := container.NewBorder(nil, nil, handle, removeBtn, titleLabel)

	// ── Full inner layout: bg + border behind the visible content ───────────
	// container.NewStack layers its children so each fills the same area.
	// bg and border are drawn first (behind), then the content VBox on top.
	b.inner = container.NewStack(
		b.bg,
		b.border,
		container.NewVBox(
			header,
			widget.NewSeparator(),
			body,
		),
	)

	return b
}

// CreateRenderer implements fyne.Widget. We use NewSimpleRenderer because
// blockBase delegates all layout work to b.inner.
func (b *blockBase) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.inner)
}

// bgColor returns the background fill colour for a given block kind.
func bgColor(kind BlockKind) color.Color {
	switch kind {
	case KindConditional:
		// Dark blue — signals "conditional logic"
		return color.NRGBA{R: 30, G: 60, B: 100, A: 220}
	case KindFile:
		// Dark purple — signals "file boundary"
		return color.NRGBA{R: 60, G: 40, B: 90, A: 220}
	case KindSection:
		// Adapts to the current Fyne theme (light or dark mode)
		return theme.OverlayBackgroundColor()
	default:
		// KindRoot — the root canvas uses the window background
		return color.Transparent
	}
}

// ── dragHandle ───────────────────────────────────────────────────────────────

// dragHandle is the ≡ button in the block header. It implements fyne.Draggable
// to initiate drag operations by populating blocks.Current.
type dragHandle struct {
	widget.BaseWidget
	base *blockBase
}

// CreateRenderer renders the drag handle as a simple ≡ label button.
func (h *dragHandle) CreateRenderer() fyne.WidgetRenderer {
	btn := widget.NewButton("≡", nil)
	btn.Importance = widget.LowImportance
	return widget.NewSimpleRenderer(btn)
}

// Dragged is called by Fyne each time the mouse moves during a drag gesture
// that started on this widget. We only need to set the payload on the first
// event; subsequent events are no-ops.
func (h *dragHandle) Dragged(e *fyne.DragEvent) {
	// Populate the global drag payload so drop targets can inspect it.
	Current = DragPayload{
		Active:           true,
		Kind:             h.base.kind,
		RemoveFromSource: h.base.onRemove,
	}
}

// DragEnd is called when the user releases the mouse button, ending the drag.
func (h *dragHandle) DragEnd() {
	Current.Active = false
}

// MinSize returns a compact size for the handle button.
func (h *dragHandle) MinSize() fyne.Size {
	return fyne.NewSize(30, 30)
}
