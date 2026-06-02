// Package blocks provides the visual block widgets used in the drag-and-drop
// config editor. Each block type corresponds to a structural element of a
// gitconfig file:
//
//   - SectionBlock    — a single [section] like [user] or [core]
//   - FileBlock       — an included config file ([include] path)
//   - ConditionalBlock — a conditional include ([includeIf "..."])
//
// Drag state is managed via the package-level [Current] variable because Fyne's
// drag API has no built-in payload mechanism. The source sets [Current] in its
// Dragged() callback, and the drop target reads it in Dropped().
package blocks

// BlockKind identifies which structural category a block belongs to.
// The drag-and-drop acceptance rules depend on this type.
type BlockKind int

const (
	// KindSection represents a leaf config block like [user], [core], [alias].
	// Sections contain key=value entries and cannot hold child blocks.
	KindSection BlockKind = iota

	// KindFile represents a config file boundary — a block created from an
	// [include] directive. Files can contain sections, other files, and
	// conditional blocks.
	KindFile

	// KindConditional represents an [includeIf "..."] block. It contains
	// exactly one FileBlock (the file that is conditionally loaded) and one
	// or more condition rows.
	KindConditional

	// KindRoot is the implicit top-level canvas. It is never dragged; it only
	// acts as a drop target for any block type.
	KindRoot
)

// ContainerAccepts reports whether a container block of kind container is
// allowed to accept a dropped child block of kind child.
//
// Rules (from the visual editor design spec):
//   - Sections are leaf nodes — they accept nothing.
//   - Conditionals cannot be nested inside another Conditional.
//   - Root and File containers accept any block type.
func ContainerAccepts(container, child BlockKind) bool {
	// Sections are leaf nodes — they can never be containers.
	if container == KindSection {
		return false
	}
	// Conditionals cannot nest inside another Conditional to avoid
	// ambiguous evaluation order.
	if child == KindConditional && container == KindConditional {
		return false
	}
	return true
}

// DragPayload holds everything the drop target needs to complete a drop
// operation. Because Fyne has no built-in drag payload, we use a package-level
// variable that the source writes and the target reads.
type DragPayload struct {
	// Active is true while a drag gesture is in progress.
	Active bool

	// Kind identifies what type of block is being dragged.
	Kind BlockKind

	// SchemaName is the schema.SectionSchema.Name for library section drags
	// (e.g. "core", "user"). Empty for boundary (file/conditional) drags.
	SchemaName string

	// RemoveFromSource, if non-nil, is called by the drop target after a
	// successful drop. It removes the block from its original location,
	// implementing move semantics (as opposed to copy).
	RemoveFromSource func()
}

// Current holds the drag state for the block that is currently being dragged.
// The zero value represents "no drag in progress" (Active == false).
// Only one drag can be active at a time.
var Current DragPayload
