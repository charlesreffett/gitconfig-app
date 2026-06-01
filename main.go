package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("com.knightyoshi.gitconfigmanager")
	w := a.NewWindow("Git Config Manager")
	w.SetContent(widget.NewLabel("Loading…"))
	w.ShowAndRun()
}
