package dev

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
)

var debugSet = os.Getenv("WANDER_DEBUG")

// dev
func Debug(msg string) {
	if debugSet != "" {
		f, err := tea.LogToFile("wander.log", "")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		log.Printf("%q", msg)
		defer f.Close()
	}
}
