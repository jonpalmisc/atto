package main

import (
	"fmt"
	"os"

	"github.com/nsf/termbox-go"
)

func main() {
	// TODO: Move to editor construction.
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: atto <file>")
		os.Exit(1)
	}

	editor := MakeEditor()
	editor.Open(args[0])
	editor.Run()
}
