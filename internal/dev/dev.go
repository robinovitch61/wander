package dev

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
)

var debugSet = os.Getenv("WANDER_DEBUG")

// Returns a function that prints a message to the log file if the WANDER_DEBUG
// environment variable is set.
func createDebug(path string) func(string) {
	return func (msg string) {
		if debugSet != "" {
			f, err := tea.LogToFile(path, "")
			if err != nil {
				fmt.Println("fatal:", err)
				os.Exit(1)
			}
			log.Printf("%q", msg)
			defer f.Close()
		}
	}
}

var Debug = createDebug("wander.log")
var MyDebug = createDebug("/tmp/mydebug.log")
