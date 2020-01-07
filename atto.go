package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: atto <file>")
		os.Exit(1)
	}

	editor := MakeEditor()
	editor.Open(args[0])
	editor.Run()
}
