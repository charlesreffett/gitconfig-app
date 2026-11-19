# Changelog

## 0.1.0 (2026-11-19)


### ⚠ BREAKING CHANGES

* the Go module path is now github.com/knightyoshi/gitconfig-app and user prefs are read from <config dir>/gitconfig-app/prefs.json.

### Features

* add internal/schema package with full gitconfig section schemas ([2ed647b](https://github.com/charlesreffett/gitconfig-app/commit/2ed647b82fc0e6b968964620537690c357aaed2c))
* add Reset button to discard unsaved config changes ([3d41802](https://github.com/charlesreffett/gitconfig-app/commit/3d41802ad0805fb0e69692153f0f0af8baebb2bb))
* implement gitconfig parser, writer, and Fyne desktop UI ([4aeca0e](https://github.com/charlesreffett/gitconfig-app/commit/4aeca0ee4b819bbac6677c620cc621ccc23685d3))
* implement visual drag-and-drop block editor (Phases 2-6) ([2278437](https://github.com/charlesreffett/gitconfig-app/commit/2278437cad7effb86b65297eb6f20ac785be2905))
* migrate desktop UI from Fyne to Wails v2 + Vue 3 ([82920b3](https://github.com/charlesreffett/gitconfig-app/commit/82920b38fa71878f077993b0e8c72eec176b1f06))


### Bug Fixes

* resolve relative include paths like git ([9e88afe](https://github.com/charlesreffett/gitconfig-app/commit/9e88afe0065ecd930406d0e42207751d2f5f6afe))
* save edits to included file contents ([22f966f](https://github.com/charlesreffett/gitconfig-app/commit/22f966f0cba282d2d3ef0eac6355ed02dc7a8632))
* use dialog.NewCustomConfirm + Resize for properly sized input fields ([09f9d96](https://github.com/charlesreffett/gitconfig-app/commit/09f9d9638049f0821345629a2c06b78f60b6538a))
