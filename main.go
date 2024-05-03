package main

import (
	"github.com/robinovitch61/wander/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
