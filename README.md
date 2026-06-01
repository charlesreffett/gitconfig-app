# Git Config Manager

A desktop application for managing `.gitconfig` files and conditional includes, built in Go with the Fyne UI framework.

## Features

- View and edit all sections of your `.gitconfig`
- Manage conditional `[includeIf "gitdir:..."]` entries
- Open and edit nested config files (e.g., `.gitconfig_work`) in the same interface
- Safe writes: changes are written atomically via temp-file rename

## Requirements

- Go 1.22+
- macOS, Linux, or Windows

## Running

```bash
go run .
```
