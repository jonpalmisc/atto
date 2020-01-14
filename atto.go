package main

import (
	"os"
)

func main() {
	editor := CreateEditor()
	editor.Run(os.Args[1:])
}
