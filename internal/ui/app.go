// Package ui is the top-level package for the Git Config Manager desktop UI.
//
// # Architecture overview
//
// The application shows three panels side by side:
//
//	[Library (200pt)] | [Global Config Canvas] | [Project Config Canvas]
//
// The library panel (left) lets users drag or click block templates to add them
// to a config canvas. Each canvas is an independent scrollable area that holds
// an ordered list of block widgets representing one gitconfig file's structure.
//
// Blocks are visual cards:
//   - SectionBlock     — [user], [core], [alias], etc.
//   - FileBlock        — an included config file ([include] path)
//   - ConditionalBlock — a conditional include ([includeIf "gitdir:…"])
//   - CustomBlock      — unknown sections not in the schema registry
//
// When both a global and a project config are loaded, the app highlights keys
// in the global config that are overridden by the project config ("superseded"
// fields are rendered in a dimmed style).
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
	"github.com/knightyoshi/git-config-manager/internal/schema"
	"github.com/knightyoshi/git-config-manager/internal/ui/blocks"
	uicanvas "github.com/knightyoshi/git-config-manager/internal/ui/canvas"
	"github.com/knightyoshi/git-config-manager/internal/ui/library"
)

// Run is the public entry point called by main.go.
// It creates the Fyne application, loads the global config, and shows the window.
// This function blocks until the user closes the window.
func Run(mainConfigPath string) {
	// app.NewWithID creates the Fyne application with a unique reverse-domain ID.
	// The ID is used by the OS for preferences, dock icons, etc.
	a := app.NewWithID("com.knightyoshi.gitconfigmanager")

	w := a.NewWindow("Git Config Manager — Visual Block Editor")
	w.Resize(fyne.NewSize(1200, 740))

	ui := newAppUI(a, w)

	// Load the global gitconfig on startup. Show a non-fatal error dialog if
	// the file cannot be read (e.g. it doesn't exist yet on a fresh machine).
	if err := ui.loadGlobal(mainConfigPath); err != nil {
		dialog.ShowError(
			fmt.Errorf("could not load %s:\n%w", mainConfigPath, err),
			w,
		)
	}

	w.SetContent(ui.buildLayout())
	w.ShowAndRun()
}

// ── Application state ─────────────────────────────────────────────────────────

// appUI owns all mutable UI state. It wires together the three panels and
// manages the global/project config lifecycle.
type appUI struct {
	fyneApp fyne.App
	window  fyne.Window

	// globalCanvas is the left config canvas showing ~/.gitconfig.
	globalCanvas *uicanvas.ConfigCanvas
	// globalPath is the absolute path to the global config file.
	globalPath string
	// globalCfg is the last parsed global config (used for superseded-key comparison).
	globalCfg *config.Config

	// projectCanvas is the right config canvas showing <repo>/.git/config.
	projectCanvas *uicanvas.ConfigCanvas
	// projectPath is the absolute path to the currently loaded project config.
	projectPath string

	// saveGlobalBtn and saveProjectBtn are kept as fields so we can enable/
	// disable them programmatically.
	saveGlobalBtn  *widget.Button
	saveProjectBtn *widget.Button

	// statusLabel is the bottom status bar.
	statusLabel *widget.Label
}

func newAppUI(a fyne.App, w fyne.Window) *appUI {
	u := &appUI{fyneApp: a, window: w}

	// Construct canvases with an onChange that marks unsaved state.
	u.globalCanvas = uicanvas.NewConfigCanvas(func() { u.markUnsaved("global") })
	u.projectCanvas = uicanvas.NewConfigCanvas(func() { u.markUnsaved("project") })

	return u
}

// ── Layout ─────────────────────────────────────────────────────────────────────

// buildLayout assembles the complete window content:
//
//	Border(top=nil, bottom=statusBar, left=libraryPanel, right=nil, center=split)
//	  split = HSplit(globalPanel, projectPanel)
func (u *appUI) buildLayout() fyne.CanvasObject {
	// ── Library panel (left, fixed ~200pt) ───────────────────────────────────
	libPanel := library.NewPanel(func(kind blocks.BlockKind, schemaName string) {
		// Clicking a library item adds a block to whichever canvas was last active.
		// For simplicity, we always add to the global canvas here.
		// Users can also drag blocks directly onto canvases.
		newBlock := blocks.CreateBlockFromPayload(blocks.DragPayload{
			Active:     true,
			Kind:       kind,
			SchemaName: schemaName,
		}, func() { u.markUnsaved("global") })
		if newBlock != nil {
			u.globalCanvas.AddBlock(newBlock)
		}
	})

	// ── Global config panel (centre-left) ────────────────────────────────────
	globalPanel := u.buildGlobalPanel()

	// ── Project config panel (centre-right) ──────────────────────────────────
	projectPanel := u.buildProjectPanel()

	// ── Split between the two canvases ───────────────────────────────────────
	contentSplit := container.NewHSplit(globalPanel, projectPanel)
	contentSplit.SetOffset(0.5) // 50/50 split

	// ── Library gets a fixed-ish width by putting it in HSplit at 15% ────────
	outerSplit := container.NewHSplit(libPanel, contentSplit)
	outerSplit.SetOffset(0.15) // ~15% for the library

	// ── Status bar ───────────────────────────────────────────────────────────
	u.statusLabel = widget.NewLabel("Ready.")
	statusBar := container.NewHBox(u.statusLabel)

	return container.NewBorder(nil, statusBar, nil, nil, outerSplit)
}

// buildGlobalPanel returns the widget for the global config column.
func (u *appUI) buildGlobalPanel() fyne.CanvasObject {
	titleLabel := widget.NewLabel("Global Config: ~/.gitconfig")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	u.saveGlobalBtn = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		u.saveGlobal()
	})

	header := container.NewBorder(nil, widget.NewSeparator(), nil, u.saveGlobalBtn, titleLabel)

	return container.NewBorder(header, nil, nil, nil, u.globalCanvas)
}

// buildProjectPanel returns the widget for the project config column.
func (u *appUI) buildProjectPanel() fyne.CanvasObject {
	titleLabel := widget.NewLabel("Project Config")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	openBtn := widget.NewButtonWithIcon("Open Project…", theme.FolderOpenIcon(), func() {
		u.showOpenProjectDialog()
	})

	u.saveProjectBtn = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		u.saveProject()
	})
	u.saveProjectBtn.Disable() // disabled until a project is loaded

	header := container.NewBorder(nil, widget.NewSeparator(), nil,
		container.NewHBox(openBtn, u.saveProjectBtn),
		titleLabel,
	)

	return container.NewBorder(header, nil, nil, nil, u.projectCanvas)
}

// ── Config loading ─────────────────────────────────────────────────────────────

// loadGlobal parses the file at path and populates the global canvas.
func (u *appUI) loadGlobal(path string) error {
	expanded, err := expandHome(path)
	if err != nil {
		return err
	}
	cfg, err := config.Parse(expanded)
	if err != nil {
		// If the file simply doesn't exist, start with an empty canvas.
		if os.IsNotExist(err) {
			u.globalPath = expanded
			u.globalCfg = &config.Config{FilePath: expanded}
			return nil
		}
		return err
	}
	u.globalPath = expanded
	u.globalCfg = cfg
	loadConfigToCanvas(cfg, u.globalCanvas, func() { u.markUnsaved("global") })
	return nil
}

// loadProject parses the gitconfig at path and populates the project canvas.
func (u *appUI) loadProject(path string) error {
	cfg, err := config.Parse(path)
	if err != nil {
		if os.IsNotExist(err) {
			u.projectPath = path
			u.projectCanvas.Clear()
			return nil
		}
		return err
	}
	u.projectPath = path
	u.projectCanvas.Clear()
	loadConfigToCanvas(cfg, u.projectCanvas, func() { u.markUnsaved("project") })
	u.saveProjectBtn.Enable()

	// After loading the project config, highlight any global keys it overrides.
	u.updateSupersededKeys(cfg)
	return nil
}

// loadConfigToCanvas walks the parsed config sections and creates the
// appropriate block for each one, then appends it to the canvas.
//
// Logic:
//   - [includeIf "…"] sections → ConditionalBlock (consecutive ones with the
//     same path are merged into one block with multiple condition rows)
//   - [include] sections → FileBlock (loads the included file recursively)
//   - Everything else → SectionBlock (if in schema) or CustomBlock (if not)
func loadConfigToCanvas(cfg *config.Config, c *uicanvas.ConfigCanvas, onChange func()) {
	i := 0
	for i < len(cfg.Sections) {
		sec := cfg.Sections[i]

		switch sec.Name {
		case "includeIf":
			// Collect all consecutive [includeIf] sections that point to the
			// same file and merge them into one ConditionalBlock.
			path := pathValue(sec)
			var conds []blocks.ConditionRowExport
			kind, val := blocks.ParseCondition(sec.Subsection)
			conds = append(conds, blocks.ConditionRowExport{Kind: kind, Value: val})

			j := i + 1
			for j < len(cfg.Sections) &&
				cfg.Sections[j].Name == "includeIf" &&
				pathValue(cfg.Sections[j]) == path {
				k2, v2 := blocks.ParseCondition(cfg.Sections[j].Subsection)
				conds = append(conds, blocks.ConditionRowExport{Kind: k2, Value: v2})
				j++
			}

			// Build the child FileBlock for the included file.
			fb := blocks.NewFileBlock(path, onChange)
			// Try to load the included file's contents into the FileBlock.
			loadIncludedFile(path, fb, onChange)

			// Build the ConditionalBlock from exported condition rows.
			condBlock := blocks.NewConditionalBlockFromExported(conds, fb, onChange)
			c.AddBlock(condBlock)
			i = j

		case "include":
			// Plain [include path=…] → FileBlock
			path := pathValue(sec)
			fb := blocks.NewFileBlock(path, onChange)
			loadIncludedFile(path, fb, onChange)
			c.AddBlock(fb)
			i++

		default:
			// Regular section: look up in schema registry.
			if s, ok := schema.Lookup(sec.Name); ok {
				sb := blocks.NewSectionBlock(s, sec.Subsection, sec.Entries, onChange)
				c.AddBlock(sb)
			} else {
				cb := blocks.NewCustomBlock(sec.Name, sec.Subsection, sec.Entries, onChange)
				c.AddBlock(cb)
			}
			i++
		}
	}
}

// loadIncludedFile attempts to parse the file at rawPath and add its sections
// as children of the given FileBlock. Errors are silently ignored (the file
// may not exist or be inaccessible — the user can still edit the path field).
func loadIncludedFile(rawPath string, fb *blocks.FileBlock, onChange func()) {
	if rawPath == "" {
		return
	}
	expanded, err := expandHome(rawPath)
	if err != nil {
		return
	}
	subCfg, err := config.Parse(expanded)
	if err != nil {
		return // file not found or parse error — silently skip
	}
	for _, sec := range subCfg.Sections {
		if sec.Name == "" {
			continue // skip preamble comment sections
		}
		if s, ok := schema.Lookup(sec.Name); ok {
			child := blocks.NewSectionBlock(s, sec.Subsection, sec.Entries, onChange)
			fb.AddChild(child)
		} else {
			child := blocks.NewCustomBlock(sec.Name, sec.Subsection, sec.Entries, onChange)
			fb.AddChild(child)
		}
	}
}

// pathValue extracts the "path" entry value from a config.Section.
func pathValue(sec config.Section) string {
	for _, e := range sec.Entries {
		if e.Key == "path" {
			return e.Value
		}
	}
	return ""
}

// ── Superseded key highlighting ───────────────────────────────────────────────

// updateSupersededKeys compares the project config against the global canvas
// and marks any global fields that are overridden by the project config.
//
// A field is "superseded" when the project config contains the same section
// name + subsection + key combination.
func (u *appUI) updateSupersededKeys(projectCfg *config.Config) {
	// Build a lookup set: "sectionName\x00subsection\x00key" → true
	superseded := make(map[string]bool)
	for _, sec := range projectCfg.Sections {
		for _, e := range sec.Entries {
			if e.Key != "" {
				key := sec.Name + "\x00" + sec.Subsection + "\x00" + e.Key
				superseded[key] = true
			}
		}
	}

	// Walk the global canvas blocks and mark matching entries.
	for _, obj := range u.globalCanvas.Blocks() {
		switch blk := obj.(type) {
		case *blocks.SectionBlock:
			for _, entry := range blk.Entries() {
				key := blk.SectionName() + "\x00" + blk.Subsection() + "\x00" + entry.Key
				blk.MarkSuperseded(entry.Key, superseded[key])
			}
		}
		// FileBlock and ConditionalBlock children are not checked — they represent
		// separate files that have their own config scope.
	}
}

// ── Save ──────────────────────────────────────────────────────────────────────

// saveGlobal serialises the global canvas back to disk.
func (u *appUI) saveGlobal() {
	if u.globalPath == "" {
		u.status("No global config path — nothing to save.")
		return
	}
	cfg := canvasToConfig(u.globalCanvas, u.globalPath)
	if err := config.Write(cfg); err != nil {
		dialog.ShowError(fmt.Errorf("save failed: %w", err), u.window)
		u.status("Save failed.")
		return
	}
	u.globalCfg = cfg
	u.status("Saved " + u.globalPath)
}

// saveProject serialises the project canvas back to disk.
func (u *appUI) saveProject() {
	if u.projectPath == "" {
		u.status("No project config loaded.")
		return
	}
	cfg := canvasToConfig(u.projectCanvas, u.projectPath)
	if err := config.Write(cfg); err != nil {
		dialog.ShowError(fmt.Errorf("save failed: %w", err), u.window)
		u.status("Save failed.")
		return
	}
	u.status("Saved " + u.projectPath)
}

// canvasToConfig walks the blocks on a canvas and reconstructs a config.Config.
func canvasToConfig(c *uicanvas.ConfigCanvas, filePath string) *config.Config {
	cfg := &config.Config{FilePath: filePath}
	for _, obj := range c.Blocks() {
		if ser, ok := obj.(blocks.BlockSerializer); ok {
			cfg.Sections = append(cfg.Sections, ser.ToSections()...)
		}
	}
	return cfg
}

// ── Open Project dialog ───────────────────────────────────────────────────────

// showOpenProjectDialog opens a folder picker. The selected directory's
// .git/config is loaded into the project canvas.
func (u *appUI) showOpenProjectDialog() {
	d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		repoPath := uri.Path()
		gitConfigPath := filepath.Join(repoPath, ".git", "config")

		if loadErr := u.loadProject(gitConfigPath); loadErr != nil {
			dialog.ShowError(
				fmt.Errorf("could not load %s:\n%w", gitConfigPath, loadErr),
				u.window,
			)
			return
		}

		u.status("Loaded project: " + repoPath)
		u.addRecentRepo(repoPath)
	}, u.window)
	d.Show()
}

// ── Recent repos ──────────────────────────────────────────────────────────────

// addRecentRepo prepends repoPath to the persisted recent-repos list, keeping
// at most 5 entries.
func (u *appUI) addRecentRepo(repoPath string) {
	prefs := u.fyneApp.Preferences()
	existing := prefs.StringList("recentRepos")

	// Remove any existing occurrence of this path (dedup).
	var filtered []string
	for _, p := range existing {
		if p != repoPath {
			filtered = append(filtered, p)
		}
	}

	// Prepend the new path.
	updated := append([]string{repoPath}, filtered...)

	// Keep at most 5 recent repos.
	if len(updated) > 5 {
		updated = updated[:5]
	}

	prefs.SetStringList("recentRepos", updated)
}

// recentRepos returns the stored list of recent repository paths.
func (u *appUI) recentRepos() []string {
	return u.fyneApp.Preferences().StringList("recentRepos")
}

// ── Status bar helpers ────────────────────────────────────────────────────────

// markUnsaved is called by canvas onChange callbacks to indicate unsaved changes.
func (u *appUI) markUnsaved(which string) {
	u.status("Unsaved changes in " + which + " config.")
}

// status updates the bottom status bar message.
func (u *appUI) status(msg string) {
	if u.statusLabel != nil {
		u.statusLabel.SetText(msg)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// expandHome replaces a leading ~ with the user's home directory.
// The Go standard library does not expand shell shortcuts automatically.
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
