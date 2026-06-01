// Git Config Manager — a desktop application for editing .gitconfig files.
//
// This is the entry point. Its only job is to find the user's ~/.gitconfig
// and hand control to the UI package, which owns everything from here on.
//
// # Why is this file so small?
//
// Go encourages keeping main() thin. The real logic lives in packages under
// internal/ so it can be tested without running the GUI. main() is just the
// "wiring" that connects the OS environment (home directory path) to the
// application logic.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/knightyoshi/git-config-manager/internal/ui"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		// log.Fatalf prints the message and calls os.Exit(1).
		// We use it here because without a home directory we cannot find
		// .gitconfig — there is nothing sensible to do but exit.
		log.Fatalf("cannot determine home directory: %v", err)
	}

	mainConfigPath := filepath.Join(home, ".gitconfig")
	ui.Run(mainConfigPath)
}
