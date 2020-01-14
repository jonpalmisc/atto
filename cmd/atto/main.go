package main

import (
	"os"

	"github.com/jonpalmisc/atto/internal/editor"
)

func main() {
	editor := editor.Create()
	editor.Run(os.Args[1:])
}
