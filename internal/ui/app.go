// Package ui implements the Fyne desktop interface for Git Config Manager.
//
// # How Fyne works (short primer)
//
// Fyne is a Go GUI framework. You build a UI by composing "widgets" (Button,
// Label, Entry, List…) inside "containers" (VBox, HBox, HSplit, Border…).
// Containers arrange their children according to a layout rule.
//
// Key rules:
//  1. All UI code runs on the main goroutine. Fyne enforces this.
//  2. After changing a widget's data, call widget.Refresh() to repaint.
//  3. app.NewWindow(...).ShowAndRun() is the event loop — it blocks until the
//     window is closed.
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/knightyoshi/git-config-manager/internal/config"
)

// ── Entry point ───────────────────────────────────────────────────────────────

// Run loads the gitconfig at mainConfigPath and opens the application window.
// It blocks until the user closes the window.
func Run(mainConfigPath string) {
	// app.NewWithID creates the Fyne application. The ID is a reverse-domain
	// string used by the OS to identify the app (for preferences, dock, etc.)
	a := app.NewWithID("com.knightyoshi.gitconfigmanager")

	w := a.NewWindow("Git Config Manager")
	w.Resize(fyne.NewSize(960, 640))

	u := newAppUI(w)
	if err := u.pushConfig(mainConfigPath); err != nil {
		// If we can't read the config, show an error dialog before starting.
		// dialog.ShowError is non-blocking — it creates a modal over the window.
		dialog.ShowError(fmt.Errorf("could not load %s:\n%w", mainConfigPath, err), w)
	}

	// SetContent replaces the window's root widget. Must be called before
	// ShowAndRun so the first frame is painted correctly.
	w.SetContent(u.buildLayout())

	// ShowAndRun makes the window visible and starts the Fyne event loop.
	// Nothing after this line runs until the window is closed.
	w.ShowAndRun()
}

// ── Application state ─────────────────────────────────────────────────────────

// appUI owns all mutable state and the widget references we need to update
// programmatically (i.e., after user actions that change the data model).
//
// In Go, structs are the primary way to bundle related data. This one acts
// like a "view model" — the single source of truth for what the window shows.
type appUI struct {
	window fyne.Window

	// configStack implements breadcrumb navigation.
	//
	// [0] is always the root ~/.gitconfig.
	// When the user clicks "Open linked file" on an includeIf section, the
	// linked config is parsed and pushed here.
	// The Back button pops the stack.
	configStack []*config.Config

	// selectedDisplay is the index inside visibleSections() the user has
	// clicked in the sidebar. -1 means nothing is selected.
	selectedDisplay int

	// These widget pointers let us refresh specific parts of the UI without
	// rebuilding the whole layout from scratch.
	sectionList  *widget.List    // sidebar list
	editorHolder *fyne.Container // wraps the right-side editor; we swap its contents
	pathLabel    *widget.Label   // shows the current file path in the toolbar
	statusLabel  *widget.Label   // bottom status bar
	backBtn      *widget.Button  // enabled only when configStack has >1 entry
}

func newAppUI(w fyne.Window) *appUI {
	return &appUI{
		window:          w,
		selectedDisplay: -1,
	}
}

// activeConfig returns the config at the top of the navigation stack, or nil.
func (u *appUI) activeConfig() *config.Config {
	if len(u.configStack) == 0 {
		return nil
	}
	return u.configStack[len(u.configStack)-1]
}

// pushConfig parses the file at path and pushes the result onto the stack.
func (u *appUI) pushConfig(path string) error {
	// Expand ~ manually — os.Open does not expand shell shortcuts.
	expanded, err := expandHome(path)
	if err != nil {
		return err
	}
	cfg, err := config.Parse(expanded)
	if err != nil {
		return err
	}
	u.configStack = append(u.configStack, cfg)
	u.selectedDisplay = -1
	return nil
}

// popConfig navigates back to the parent config.
func (u *appUI) popConfig() {
	if len(u.configStack) <= 1 {
		return
	}
	u.configStack = u.configStack[:len(u.configStack)-1]
	u.selectedDisplay = -1
}

// visibleSections returns the real section indices (into activeConfig.Sections)
// of sections that should appear in the sidebar. We skip nameless "preamble"
// sections — those are just comment holders with no header.
func (u *appUI) visibleSections() []int {
	cfg := u.activeConfig()
	if cfg == nil {
		return nil
	}
	var out []int
	for i, s := range cfg.Sections {
		if s.Name != "" {
			out = append(out, i)
		}
	}
	return out
}

// ── Layout ────────────────────────────────────────────────────────────────────

// buildLayout assembles the complete window content.
//
// The layout hierarchy is:
//
//	Border(top=toolbar, bottom=statusBar, center=HSplit)
//	  HSplit
//	    left:  Border(top=listTitle, center=sectionList)
//	    right: editorHolder  (swapped dynamically on selection)
func (u *appUI) buildLayout() fyne.CanvasObject {
	// ── sidebar (left) ────────────────────────────────────────────────────
	u.sectionList = u.buildSectionList()

	listTitle := widget.NewLabel("Sections")
	listTitle.TextStyle = fyne.TextStyle{Bold: true}

	// container.NewBorder places items at the cardinal edges and stretches the
	// center to fill remaining space. Signature: (top, bottom, left, right, center...)
	leftPanel := container.NewBorder(
		container.NewVBox(listTitle, widget.NewSeparator()),
		nil, nil, nil,
		u.sectionList,
	)

	// ── editor (right) ────────────────────────────────────────────────────
	// container.NewStack stacks its children so each fills the whole area.
	// We use it as a "slot" — when a section is selected we replace its sole
	// child with a freshly built editor panel.
	placeholder := widget.NewLabel("← Select a section to edit it")
	placeholder.Alignment = fyne.TextAlignCenter
	u.editorHolder = container.NewStack(placeholder)

	// ── split ─────────────────────────────────────────────────────────────
	split := container.NewHSplit(leftPanel, u.editorHolder)
	split.SetOffset(0.28) // 28% left, 72% right

	// ── toolbar (top) ─────────────────────────────────────────────────────
	toolbar := u.buildToolbar()

	// ── status bar (bottom) ───────────────────────────────────────────────
	u.statusLabel = widget.NewLabel("")
	statusBar := container.NewHBox(u.statusLabel)

	return container.NewBorder(toolbar, statusBar, nil, nil, split)
}

// buildToolbar creates the top bar: back button, file path label, and action buttons.
func (u *appUI) buildToolbar() fyne.CanvasObject {
	// Back button — only useful when we have drilled into a nested config.
	u.backBtn = widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		u.popConfig()
		u.refreshAfterNavigation()
	})
	u.backBtn.Disable() // disabled initially; enabled when stack depth > 1

	// Path label shows which file we are currently editing.
	pathText := ""
	if cfg := u.activeConfig(); cfg != nil {
		pathText = cfg.FilePath
	}
	u.pathLabel = widget.NewLabel(pathText)
	u.pathLabel.Truncation = fyne.TextTruncateEllipsis

	// Save button — writes the active config to disk atomically.
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		u.save()
	})

	// Add Section button — opens a dialog to create a new [section].
	addSecBtn := widget.NewButtonWithIcon("Add Section", theme.ContentAddIcon(), func() {
		u.showAddSectionDialog()
	})

	// container.NewHBox stacks its children horizontally with natural sizing.
	left := container.NewHBox(u.backBtn, u.pathLabel)
	right := container.NewHBox(addSecBtn, saveBtn)
	return container.NewBorder(nil, widget.NewSeparator(), nil, right, left)
}

// buildSectionList creates the sidebar widget.List.
//
// widget.List is Fyne's virtualised list: it creates a small fixed number of
// row widgets ("templates") and reuses them as the user scrolls.  You drive
// it with three functions:
//   - Length() int              — how many rows in total
//   - CreateItem() CanvasObject — return a fresh template widget
//   - UpdateItem(id, obj)       — fill the template for row `id`
func (u *appUI) buildSectionList() *widget.List {
	list := widget.NewList(
		func() int {
			return len(u.visibleSections())
		},
		func() fyne.CanvasObject {
			// The template is a plain Label. Fyne reuses these across rows.
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label) // type-assert the template back to *Label
			vis := u.visibleSections()
			if id >= len(vis) {
				return
			}
			sec := u.activeConfig().Sections[vis[id]]
			label.SetText(sectionLabel(sec))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		u.selectedDisplay = id
		vis := u.visibleSections()
		if id >= len(vis) {
			return
		}
		sectionIdx := vis[id]
		u.showSection(sectionIdx)
	}

	return list
}

// ── Dynamic editor swapping ───────────────────────────────────────────────────

// showSection rebuilds the right-side editor for the section at sectionIdx
// (an index into activeConfig().Sections).
func (u *appUI) showSection(sectionIdx int) {
	cfg := u.activeConfig()
	if cfg == nil || sectionIdx >= len(cfg.Sections) {
		return
	}
	sec := &cfg.Sections[sectionIdx]

	var editor fyne.CanvasObject
	if isIncludeSection(sec) {
		editor = buildIncludeEditor(sec, func(linkedPath string) {
			u.openLinkedConfig(linkedPath)
		})
	} else {
		editor = buildRegularEditor(sec, func() {
			// Called whenever the editor makes a structural change (add/delete entry)
			// that requires the editor itself to be rebuilt.
			u.showSection(sectionIdx)
		})
	}

	// Wrap in a scroll view so long sections don't clip.
	scrolled := container.NewScroll(editor)
	u.editorHolder.Objects = []fyne.CanvasObject{scrolled}
	u.editorHolder.Refresh()
}

// ── Navigation helpers ────────────────────────────────────────────────────────

// openLinkedConfig resolves a path (possibly relative or ~/…) to an absolute
// path, parses that gitconfig, and navigates into it.
func (u *appUI) openLinkedConfig(rawPath string) {
	if err := u.pushConfig(rawPath); err != nil {
		dialog.ShowError(fmt.Errorf("open %s: %w", rawPath, err), u.window)
		return
	}
	u.refreshAfterNavigation()
}

// refreshAfterNavigation repaints the sidebar, toolbar, and clears the editor
// after the config stack changes (push or pop).
func (u *appUI) refreshAfterNavigation() {
	// Update path label.
	if cfg := u.activeConfig(); cfg != nil {
		u.pathLabel.SetText(cfg.FilePath)
	}

	// Enable/disable back button.
	if len(u.configStack) > 1 {
		u.backBtn.Enable()
	} else {
		u.backBtn.Disable()
	}

	// Reset selection and repaint the list.
	u.selectedDisplay = -1
	u.sectionList.UnselectAll()
	u.sectionList.Refresh()

	// Clear the editor pane.
	placeholder := widget.NewLabel("← Select a section to edit it")
	placeholder.Alignment = fyne.TextAlignCenter
	u.editorHolder.Objects = []fyne.CanvasObject{placeholder}
	u.editorHolder.Refresh()

	u.status("")
}

// ── Save ──────────────────────────────────────────────────────────────────────

// save writes the currently active config to disk.
func (u *appUI) save() {
	cfg := u.activeConfig()
	if cfg == nil {
		u.status("Nothing to save.")
		return
	}
	if err := config.Write(cfg); err != nil {
		dialog.ShowError(fmt.Errorf("save failed: %w", err), u.window)
		u.status("❌ Save failed.")
		return
	}
	u.status("✓ Saved " + cfg.FilePath)
}

// ── Add Section dialog ────────────────────────────────────────────────────────

// showAddSectionDialog opens a form where the user names a new section.
func (u *appUI) showAddSectionDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("e.g. alias, push, pull")

	subEntry := widget.NewEntry()
	subEntry.SetPlaceHolder(`e.g. lfs  (leave blank for none)`)

	items := []*widget.FormItem{
		{Text: "Section name", Widget: nameEntry},
		{Text: "Subsection (optional)", Widget: subEntry},
	}

	// dialog.ShowForm shows a modal form dialog with OK/Cancel buttons.
	dialog.ShowForm("Add Section", "Add", "Cancel", items, func(confirmed bool) {
		if !confirmed {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			u.status("Section name cannot be empty.")
			return
		}
		sub := strings.TrimSpace(subEntry.Text)

		cfg := u.activeConfig()
		cfg.Sections = append(cfg.Sections, config.Section{
			Name:       name,
			Subsection: sub,
		})

		// Repaint the sidebar to include the new section.
		u.sectionList.Refresh()

		// Auto-select the new section so the user sees its (empty) editor.
		vis := u.visibleSections()
		newDisplayIdx := len(vis) - 1
		u.sectionList.Select(newDisplayIdx)
		u.status(fmt.Sprintf("Added [%s].", sectionLabel(config.Section{Name: name, Subsection: sub})))
	}, u.window)
}

// ── Status bar ────────────────────────────────────────────────────────────────

func (u *appUI) status(msg string) {
	u.statusLabel.SetText(msg)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// sectionLabel formats a config.Section for display in the sidebar.
// Examples: "user", "core", "filter · lfs", "includeIf · gitdir:~/work/"
func sectionLabel(s config.Section) string {
	if s.Subsection == "" {
		return s.Name
	}
	return s.Name + " · " + s.Subsection
}

// isIncludeSection reports whether a section is [include] or [includeIf …].
// These sections are handled differently in the editor because their purpose
// is to link to another file, not to hold key=value configuration.
func isIncludeSection(s *config.Section) bool {
	return s.Name == "include" || s.Name == "includeIf"
}

// expandHome replaces a leading ~ with the user's home directory.
// Go's standard library does not do this automatically; only shells do.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand home: %w", err)
	}
	return filepath.Join(home, path[1:]), nil
}
