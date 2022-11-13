package dev

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"log"
	"os"
)

// dev
func Debug(msg string) {
	if constants.DebugSet {
		f, err := tea.LogToFile("wander.log", "")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		log.Printf("%q", msg)
		defer f.Close()
	}
}
