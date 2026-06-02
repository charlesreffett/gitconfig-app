package blocks

// conditional.go implements ConditionalBlock — the visual card for an
// [includeIf "condition"] section.
//
// A ConditionalBlock:
//   - Holds one or more condition rows (type dropdown + value entry + × button)
//   - Contains exactly one FileBlock (the file that is conditionally loaded)
//   - Acts as a drop target that accepts only FileBlocks
//
// Multiple consecutive [includeIf] sections pointing to the same file are
// combined into a single ConditionalBlock with multiple condition rows when
// loading from disk (see loadConfigToCanvas in app.go).

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/config"
	"github.com/knightyoshi/git-config-manager/internal/schema"
)

// conditionRow stores the two parts of one [includeIf "type:value"] entry.
// This unexported type is used internally within the blocks package.
type conditionRow struct {
	kind  string // e.g. "gitdir", "onbranch"
	value string // e.g. "~/work/**", "feature/*"
}

// ConditionRowExport is the exported version of conditionRow used by app.go
// when building ConditionalBlocks from parsed config sections.
type ConditionRowExport struct {
	Kind  string // e.g. "gitdir", "onbranch"
	Value string // e.g. "~/work/**", "feature/*"
}

// ConditionalBlock is the visual card for one or more [includeIf] sections
// that all point to the same included file.
type ConditionalBlock struct {
	widget.BaseWidget

	base       *blockBase
	conditions []conditionRow // at least one condition
	child      *FileBlock     // the single file this conditional loads; may be nil
	onChange   func()
}

// NewConditionalBlock creates a ConditionalBlock with one empty condition row.
//
//   - onChange — called on any user edit or structural change
func NewConditionalBlock(onChange func()) *ConditionalBlock {
	b := &ConditionalBlock{
		conditions: []conditionRow{{kind: schema.ConditionTypes()[0], value: ""}},
		onChange:   onChange,
	}
	b.ExtendBaseWidget(b)
	b.rebuild()
	return b
}

// NewConditionalBlockFromConditions creates a ConditionalBlock with pre-populated
// conditions and an optional child FileBlock.
//
//   - conditions — slice of (kind, value) pairs
//   - child      — the FileBlock to embed (may be nil)
//   - onChange   — called on any user edit
func NewConditionalBlockFromConditions(conditions []conditionRow, child *FileBlock, onChange func()) *ConditionalBlock {
	b := &ConditionalBlock{
		conditions: conditions,
		child:      child,
		onChange:   onChange,
	}
	b.ExtendBaseWidget(b)
	b.rebuild()
	return b
}

// NewConditionalBlockFromExported creates a ConditionalBlock from exported
// condition rows (used by app.go when loading configs from disk).
//
//   - exported  — condition rows using the exported ConditionRowExport type
//   - child     — the FileBlock to embed (may be nil)
//   - onChange  — called on any user edit
func NewConditionalBlockFromExported(exported []ConditionRowExport, child *FileBlock, onChange func()) *ConditionalBlock {
	conds := make([]conditionRow, len(exported))
	for i, e := range exported {
		conds[i] = conditionRow{kind: e.Kind, value: e.Value}
	}
	return NewConditionalBlockFromConditions(conds, child, onChange)
}

// AddCondition appends a new empty condition row and rebuilds the display.
func (b *ConditionalBlock) AddCondition(kind, value string) {
	b.conditions = append(b.conditions, conditionRow{kind: kind, value: value})
	b.rebuild()
}

// Conditions returns all condition rows as a copy.
func (b *ConditionalBlock) Conditions() []conditionRow {
	out := make([]conditionRow, len(b.conditions))
	copy(out, b.conditions)
	return out
}

// Child returns the FileBlock inside this conditional (may be nil).
func (b *ConditionalBlock) Child() *FileBlock { return b.child }

// AddChild sets the child FileBlock (replacing any existing one).
func (b *ConditionalBlock) AddChild(fb *FileBlock) {
	b.child = fb
	b.rebuild()
	b.Refresh()
}

// SetOnRemove sets the callback for the × button.
func (b *ConditionalBlock) SetOnRemove(fn func()) {
	b.base.onRemove = fn
}

// rebuild re-renders the entire block.
func (b *ConditionalBlock) rebuild() {
	body := b.buildBody()

	if b.base == nil {
		b.base = newBlockBase(KindConditional, "⚡", "Conditional Include", body, func() {
			if b.onChange != nil {
				b.onChange()
			}
		})
	} else {
		b.base = newBlockBase(KindConditional, "⚡", "Conditional Include", body, b.base.onRemove)
	}
}

// buildBody creates the VBox containing condition rows, the [+ Add condition]
// button, a separator, and the child file block (or drop zone).
func (b *ConditionalBlock) buildBody() fyne.CanvasObject {
	var rows []fyne.CanvasObject

	// One row per condition.
	for i := range b.conditions {
		rows = append(rows, b.buildConditionRow(i))
	}

	// [+ Add condition] button.
	addBtn := widget.NewButton("+ Add condition", func() {
		b.conditions = append(b.conditions, conditionRow{
			kind:  schema.ConditionTypes()[0],
			value: "",
		})
		b.rebuild()
		if b.onChange != nil {
			b.onChange()
		}
	})
	rows = append(rows, addBtn, widget.NewSeparator())

	// Child file block or empty drop-zone.
	if b.child != nil {
		rows = append(rows, b.child)
	} else {
		rows = append(rows, b.buildEmptyDropZone())
	}

	return container.NewVBox(rows...)
}

// buildConditionRow creates one [type dropdown] [value entry] [×] row.
func (b *ConditionalBlock) buildConditionRow(idx int) fyne.CanvasObject {
	cond := &b.conditions[idx]

	// Condition type dropdown.
	kindSel := widget.NewSelect(schema.ConditionTypes(), func(v string) {
		b.conditions[idx].kind = v
		if b.onChange != nil {
			b.onChange()
		}
	})
	kindSel.SetSelected(cond.kind)

	// Condition value entry.
	valEntry := widget.NewEntry()
	valEntry.SetText(cond.value)
	valEntry.SetPlaceHolder("e.g. ~/work/**")
	valEntry.OnChanged = func(v string) {
		b.conditions[idx].value = v
		if b.onChange != nil {
			b.onChange()
		}
	}

	// Delete-row button — disabled when there's only one condition.
	delBtn := widget.NewButton("×", func() {
		if len(b.conditions) <= 1 {
			return // keep at least one condition row
		}
		b.conditions = append(b.conditions[:idx], b.conditions[idx+1:]...)
		b.rebuild()
		if b.onChange != nil {
			b.onChange()
		}
	})

	return container.NewBorder(nil, nil, kindSel, delBtn, valEntry)
}

// buildEmptyDropZone returns a labelled placeholder that accepts FileBlock drops.
func (b *ConditionalBlock) buildEmptyDropZone() fyne.CanvasObject {
	dz := &conditionalDropZone{block: b}
	dz.ExtendBaseWidget(dz)
	return dz
}

// CreateRenderer implements fyne.Widget.
func (b *ConditionalBlock) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.base.inner)
}

// DraggedOver implements fyne.DropTarget. Highlights the block when a FileBlock
// is dragged over it.
func (b *ConditionalBlock) DraggedOver(e *fyne.DragEvent) {
	if Current.Active && ContainerAccepts(KindConditional, Current.Kind) {
		b.base.border.StrokeColor = hoverColor
		b.base.border.StrokeWidth = 2
		b.base.border.FillColor = color.Transparent
		b.base.Refresh()
	}
}

// DraggedOut implements fyne.DropTarget. Removes the highlight.
func (b *ConditionalBlock) DraggedOut() {
	b.base.border.StrokeColor = color.Transparent
	b.base.border.StrokeWidth = 0
	b.base.Refresh()
}

// Dropped implements fyne.DropTarget. Accepts a FileBlock drop and sets it
// as this conditional's child.
func (b *ConditionalBlock) Dropped(e *fyne.DragEvent) {
	if !Current.Active || !ContainerAccepts(KindConditional, Current.Kind) {
		return
	}
	if Current.RemoveFromSource != nil {
		Current.RemoveFromSource()
	}
	// Only FileBlocks can be dropped into a ConditionalBlock.
	if Current.Kind == KindFile {
		b.AddChild(NewFileBlock("", b.onChange))
	}
	Current = DragPayload{}
	b.base.border.StrokeColor = color.Transparent
	b.base.border.StrokeWidth = 0
	b.rebuild()
	b.Refresh()
}

// ToSections serialises this block into config.Section structs.
// Each condition row becomes one [includeIf "type:value"] section.
// All point to the same file (the child FileBlock's filename).
func (b *ConditionalBlock) ToSections() []config.Section {
	var sections []config.Section
	for _, cond := range b.conditions {
		sec := config.Section{
			Name:       "includeIf",
			Subsection: cond.kind + ":" + cond.value,
		}
		if b.child != nil {
			sec.Entries = []config.Entry{{Key: "path", Value: b.child.Filename()}}
		}
		sections = append(sections, sec)
	}
	return sections
}

// ParseCondition splits an [includeIf] subsection string into (kind, value).
// For example "gitdir:~/work/**" → ("gitdir", "~/work/**").
func ParseCondition(subsection string) (kind, value string) {
	idx := strings.Index(subsection, ":")
	if idx < 0 {
		return subsection, ""
	}
	return subsection[:idx], subsection[idx+1:]
}

// ── conditionalDropZone ──────────────────────────────────────────────────────

// conditionalDropZone is the empty-state drop zone shown inside a
// ConditionalBlock when it has no child FileBlock yet.
type conditionalDropZone struct {
	widget.BaseWidget
	block *ConditionalBlock
	inner *fyne.Container
}

func (dz *conditionalDropZone) CreateRenderer() fyne.WidgetRenderer {
	if dz.inner == nil {
		hint := widget.NewLabel("Drop a File block here")
		hint.TextStyle = fyne.TextStyle{Italic: true}
		hint.Alignment = fyne.TextAlignCenter
		dz.inner = container.NewStack(hint)
	}
	return widget.NewSimpleRenderer(dz.inner)
}

func (dz *conditionalDropZone) DraggedOver(e *fyne.DragEvent) {
	dz.block.DraggedOver(e)
}

func (dz *conditionalDropZone) DraggedOut() {
	dz.block.DraggedOut()
}

func (dz *conditionalDropZone) Dropped(e *fyne.DragEvent) {
	dz.block.Dropped(e)
}

func (dz *conditionalDropZone) MinSize() fyne.Size {
	return fyne.NewSize(0, 48)
}
