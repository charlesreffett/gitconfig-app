# gitconfig-app

A desktop application for visually editing `.gitconfig` files and conditional includes. Built with [Wails v2](https://wails.io) — a Go backend with a Vue 3 + TypeScript frontend running in the native OS webview.

## Features

- Edit your config as a list of typed blocks — known sections (`[user]`, `[core]`, `[remote]`, …) get schema-driven fields with toggles, dropdowns, and text inputs; unknown sections become free key/value blocks
- Manage `[include]` and `[includeIf "gitdir:..."]` entries, including multiple conditions pointing at the same file
- See nested config files (e.g. `.gitconfig_work`) inline under the include that loads them
- Global (`~/.gitconfig`) and project (`.git/config`) configs side by side, with project values that override global ones highlighted
- Drag and drop to reorder blocks; add new ones from the library panel
- Lossless round-trips: comments and blank lines are preserved
- Safe writes: changes are written atomically via temp-file rename

## Requirements

- Go 1.25+
- Node.js 20+ and [pnpm](https://pnpm.io)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- macOS, Linux, or Windows (Linux needs the GTK3 and WebKit2GTK dev packages — run `wails doctor` to check)

## Development

```bash
wails dev
```

Runs the app with hot reload for both the Go backend and the Vue frontend.

## Building

```bash
wails build
```

Produces a single binary with the frontend embedded, in `build/bin/`.

## Testing

```bash
go test ./...
```

```bash
cd frontend && pnpm test
```
