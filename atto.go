package main

import (
	"os"
)

func main() {
	args := os.Args[1:]

	editor := MakeEditor()
	editor.Run(args)
}
