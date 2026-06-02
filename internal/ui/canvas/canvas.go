// Package canvas provides the ConfigCanvas widget — the scrollable drag-and-drop
// work area where users arrange config blocks.
//
// The canvas is the central panel in the application layout. Users drag blocks
// from the library panel on the left and drop them onto the canvas to build their
// gitconfig hierarchy. The canvas itself is a drop target: blocks dropped onto
// the canvas background are added at the root level.
package canvas

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/ui/blocks"
)

// ConfigCanvas is the scrollable work area that holds an ordered list of
// top-level blocks. It implements fyne.DropTarget so that blocks can be dropped
// directly onto the canvas background (root level).
type ConfigCanvas struct {
	widget.BaseWidget

	// blockList holds the ordered top-level blocks.
	blockList []fyne.CanvasObject

	// vbox is the inner vertical container that is repopulated on each rebuild().
	vbox *fyne.Container

	// scroll wraps vbox so the content can be scrolled when it overflows.
	scroll *container.Scroll

	// onChange is called whenever the canvas contents change (block added,
	// removed, or modified). Used to mark the file as having unsaved changes.
	onChange func()
}

// NewConfigCanvas creates a new, empty ConfigCanvas.
//
//   - onChange — called on any structural change to the canvas contents
func NewConfigCanvas(onChange func()) *ConfigCanvas {
	c := &ConfigCanvas{onChange: onChange}
	c.ExtendBaseWidget(c)

	c.vbox = container.NewVBox()
	c.scroll = container.NewScroll(c.vbox)
	return c
}

// CreateRenderer implements fyne.Widget. The canvas simply renders its scroll
// container.
func (c *ConfigCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.scroll)
}

// AddBlock appends a block to the canvas at the root level and refreshes.
func (c *ConfigCanvas) AddBlock(b fyne.CanvasObject) {
	// Wire the block's remove function to remove it from c.blockList.
	c.wireRemove(b)
	c.blockList = append(c.blockList, b)
	c.rebuild()
	if c.onChange != nil {
		c.onChange()
	}
}

// RemoveBlock removes the block at index idx from the canvas.
func (c *ConfigCanvas) RemoveBlock(idx int) {
	if idx < 0 || idx >= len(c.blockList) {
		return
	}
	c.blockList = append(c.blockList[:idx], c.blockList[idx+1:]...)
	c.rebuild()
	if c.onChange != nil {
		c.onChange()
	}
}

// Blocks returns the current ordered list of top-level blocks.
func (c *ConfigCanvas) Blocks() []fyne.CanvasObject {
	return c.blockList
}

// Clear removes all blocks from the canvas.
func (c *ConfigCanvas) Clear() {
	c.blockList = nil
	c.rebuild()
}

// rebuild repopulates the vbox with all current blocks plus a root-level drop
// zone at the bottom.
func (c *ConfigCanvas) rebuild() {
	objects := make([]fyne.CanvasObject, 0, len(c.blockList)+1)
	for _, b := range c.blockList {
		objects = append(objects, b)
	}
	// Always show a drop zone at the bottom so there is somewhere to drop
	// even when the canvas is empty.
	objects = append(objects, newRootDropZone(c))
	c.vbox.Objects = objects
	c.vbox.Refresh()
}

// wireRemove sets the onRemove callback on the block (if it is one of our
// known block types) so that removing it from the block list is automatic.
func (c *ConfigCanvas) wireRemove(b fyne.CanvasObject) {
	idx := len(c.blockList) // the block will be at this index after AddBlock

	switch blk := b.(type) {
	case *blocks.SectionBlock:
		blk.SetOnRemove(func() { c.RemoveBlock(idx) })
	case *blocks.CustomBlock:
		blk.SetOnRemove(func() { c.RemoveBlock(idx) })
	case *blocks.FileBlock:
		blk.SetOnRemove(func() { c.RemoveBlock(idx) })
	case *blocks.ConditionalBlock:
		blk.SetOnRemove(func() { c.RemoveBlock(idx) })
	}
}

// ── fyne.DropTarget ──────────────────────────────────────────────────────────

// DraggedOver implements fyne.DropTarget. The canvas background accepts all
// block types (root level), so we don't need to show special highlighting —
// the root drop zone at the bottom handles the visual feedback.
func (c *ConfigCanvas) DraggedOver(_ *fyne.DragEvent) {}

// DraggedOut implements fyne.DropTarget.
func (c *ConfigCanvas) DraggedOut() {}

// Dropped implements fyne.DropTarget. Blocks dropped on the canvas background
// are added at the root level.
func (c *ConfigCanvas) Dropped(e *fyne.DragEvent) {
	if !blocks.Current.Active {
		return
	}
	if blocks.Current.RemoveFromSource != nil {
		blocks.Current.RemoveFromSource()
	}
	newBlock := blocks.CreateBlockFromPayload(blocks.Current, c.onChange)
	blocks.Current = blocks.DragPayload{}
	if newBlock != nil {
		c.AddBlock(newBlock)
	}
}

// ── Root drop zone ───────────────────────────────────────────────────────────

// rootDropZone is the visible drop target shown at the bottom of the canvas.
// It gives users a clear place to drop new blocks at the root level.
type rootDropZone struct {
	widget.BaseWidget
	canvas *ConfigCanvas
	inner  *fyne.Container
}

func newRootDropZone(c *ConfigCanvas) *rootDropZone {
	dz := &rootDropZone{canvas: c}
	dz.ExtendBaseWidget(dz)

	hint := widget.NewLabel("Drop a block here to add it to this config")
	hint.TextStyle = fyne.TextStyle{Italic: true}
	hint.Alignment = fyne.TextAlignCenter
	dz.inner = container.NewStack(hint)
	return dz
}

func (dz *rootDropZone) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(dz.inner)
}

func (dz *rootDropZone) DraggedOver(e *fyne.DragEvent) {
	// Highlight the drop zone background when a block is hovering.
	if blocks.Current.Active {
		_ = color.NRGBA{R: 80, G: 180, B: 255, A: 40} // visual hint colour (unused here)
	}
}

func (dz *rootDropZone) DraggedOut() {}

func (dz *rootDropZone) Dropped(e *fyne.DragEvent) {
	dz.canvas.Dropped(e)
}

func (dz *rootDropZone) MinSize() fyne.Size {
	return fyne.NewSize(0, 60)
}
