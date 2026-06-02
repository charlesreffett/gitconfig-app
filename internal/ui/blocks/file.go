package blocks

// file.go implements FileBlock — the visual card representing a config file
// boundary. A FileBlock is created from an [include] section or as the child
// of a ConditionalBlock.
//
// The header shows an editable filename field so the user can change which
// file is included. The body shows an ordered list of child blocks (sections,
// other files, or conditional blocks) plus a drop zone at the bottom.
//
// Drop target rules: FileBlock accepts any block kind (sections, files,
// conditionals) because a config file can contain all of them.

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/config"
)

// FileBlock is the visual card for a config file boundary (an [include] path).
// It acts as a container for child blocks (sections, other files, conditionals).
type FileBlock struct {
	widget.BaseWidget

	base     *blockBase
	filename *widget.Entry      // editable filename in the header area
	children []fyne.CanvasObject // child block widgets, in display order
	onChange func()
}

// NewFileBlock creates a FileBlock with the given initial filename.
//
//   - filename — the file path (may be empty for a "new file" placeholder)
//   - onChange — called on any user edit or structural change
func NewFileBlock(filename string, onChange func()) *FileBlock {
	b := &FileBlock{
		onChange: onChange,
	}
	b.ExtendBaseWidget(b)

	// The filename entry lives in the body area (not the title label) so the
	// user can type a path. The block header title is fixed to "📄 File".
	b.filename = widget.NewEntry()
	b.filename.SetText(filename)
	b.filename.SetPlaceHolder("~/path/to/config")
	b.filename.OnChanged = func(v string) {
		if b.onChange != nil {
			b.onChange()
		}
	}

	b.rebuild()
	return b
}

// rebuild re-renders the block's inner content. Called after structural changes
// (adding/removing child blocks).
func (b *FileBlock) rebuild() {
	body := b.buildBody()

	if b.base == nil {
		b.base = newBlockBase(KindFile, "📄", "File", body, func() {
			if b.onChange != nil {
				b.onChange()
			}
		})
	} else {
		b.base = newBlockBase(KindFile, "📄", "File", body, b.base.onRemove)
	}
}

// buildBody creates the VBox: filename entry, child blocks, and the drop zone.
func (b *FileBlock) buildBody() fyne.CanvasObject {
	// Filename row at the top of the body.
	fileRow := container.NewBorder(nil, nil,
		widget.NewLabel("Path:"), nil,
		b.filename,
	)

	var rows []fyne.CanvasObject
	rows = append(rows, fileRow, widget.NewSeparator())

	// Render each child block.
	for _, child := range b.children {
		rows = append(rows, child)
	}

	// Drop zone — a dashed-border area at the bottom that accepts dropped blocks.
	dropZone := newDropZone(b)
	rows = append(rows, dropZone)

	return container.NewVBox(rows...)
}

// AddChild appends a child block to this FileBlock and rebuilds the display.
func (b *FileBlock) AddChild(child fyne.CanvasObject) {
	b.children = append(b.children, child)
	b.rebuild()
	b.Refresh()
}

// RemoveChild removes the child at index idx and rebuilds.
func (b *FileBlock) RemoveChild(idx int) {
	if idx < 0 || idx >= len(b.children) {
		return
	}
	b.children = append(b.children[:idx], b.children[idx+1:]...)
	b.rebuild()
	b.Refresh()
}

// Filename returns the current filename text.
func (b *FileBlock) Filename() string {
	return b.filename.Text
}

// SetOnRemove sets the callback invoked when the × button is pressed.
func (b *FileBlock) SetOnRemove(fn func()) {
	b.base.onRemove = fn
}

// ToSections serialises this FileBlock. A FileBlock itself maps to an
// [include] section. Its children are serialised and returned inline.
//
// Note: when a FileBlock is a child of a ConditionalBlock, the ConditionalBlock
// handles the [includeIf] serialisation separately via its own ToSections().
func (b *FileBlock) ToSections() []config.Section {
	// The [include] section itself.
	var sections []config.Section
	if b.filename.Text != "" {
		sections = append(sections, config.Section{
			Name:    "include",
			Entries: []config.Entry{{Key: "path", Value: b.filename.Text}},
		})
	}
	// Also serialise child blocks (they belong to the included file, so we
	// can't embed them in this config — skip them here; the included file
	// is written separately on save).
	return sections
}

// CreateRenderer implements fyne.Widget.
func (b *FileBlock) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.base.inner)
}

// DraggedOver implements fyne.DropTarget. Highlights the border when a
// compatible block is dragged over this FileBlock.
func (b *FileBlock) DraggedOver(e *fyne.DragEvent) {
	if Current.Active && ContainerAccepts(KindFile, Current.Kind) {
		b.base.border.StrokeColor = hoverColor
		b.base.border.StrokeWidth = 2
		b.base.border.FillColor = color.Transparent
		b.base.Refresh()
	}
}

// DraggedOut implements fyne.DropTarget. Removes the highlight.
func (b *FileBlock) DraggedOut() {
	b.base.border.StrokeColor = color.Transparent
	b.base.border.StrokeWidth = 0
	b.base.Refresh()
}

// Dropped implements fyne.DropTarget. Creates a new child block from the
// drag payload and appends it to this FileBlock.
func (b *FileBlock) Dropped(e *fyne.DragEvent) {
	if !Current.Active || !ContainerAccepts(KindFile, Current.Kind) {
		return
	}
	// Move semantics: remove from the original location first.
	if Current.RemoveFromSource != nil {
		Current.RemoveFromSource()
	}
	// Create the appropriate block based on what was dragged and add it.
	if child := CreateBlockFromPayload(Current, b.onChange); child != nil {
		b.AddChild(child)
	}
	Current = DragPayload{}
	b.base.border.StrokeColor = color.Transparent
	b.base.border.StrokeWidth = 0
	b.rebuild()
	b.Refresh()
}

// ── Drop zone ────────────────────────────────────────────────────────────────

// dropZone is a small visible area at the bottom of a container block.
// It provides an explicit hit target for drops and shows a hint label.
type dropZone struct {
	widget.BaseWidget
	parent *FileBlock
	inner  *fyne.Container
}

func newDropZone(parent *FileBlock) *dropZone {
	dz := &dropZone{parent: parent}
	dz.ExtendBaseWidget(dz)

	hint := widget.NewLabel("Drop block here")
	hint.TextStyle = fyne.TextStyle{Italic: true}
	hint.Alignment = fyne.TextAlignCenter

	dz.inner = container.NewStack(hint)
	return dz
}

func (dz *dropZone) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(dz.inner)
}

func (dz *dropZone) DraggedOver(e *fyne.DragEvent) {
	dz.parent.DraggedOver(e)
}

func (dz *dropZone) DraggedOut() {
	dz.parent.DraggedOut()
}

func (dz *dropZone) Dropped(e *fyne.DragEvent) {
	dz.parent.Dropped(e)
}

func (dz *dropZone) MinSize() fyne.Size {
	return fyne.NewSize(0, 32)
}
