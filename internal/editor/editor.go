package editor

import (
	"fmt"
	"os"
	"time"

	"github.com/jonpalmisc/atto/internal/buffer"
	"github.com/jonpalmisc/atto/internal/config"
	"github.com/jonpalmisc/atto/internal/support"
	"github.com/nsf/termbox-go"
)

// Editor is the editor instance and manages the UI.
type Editor struct {

	// The editor's buffers and the index of the focused buffer.
	Buffers    []buffer.Buffer
	FocusIndex int

	// The editor's height and width measured in rows and columns, respectively.
	Width  int
	Height int

	// The current status message and the time it was set.
	StatusMessage     string
	StatusMessageTime time.Time

	// The prompt question, answer, and whether it is active or not.
	PromptQuestion string
	PromptAnswer   string
	PromptIsActive bool

	// The user's editor configuration.
	Config config.Config
}

// Create creates a new Editor instance.
func Create() (editor Editor) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// Attempt to load the user's editor configuration.
	cfg, err := config.Load()
	if err != nil {
		editor.SetStatusMessage("Failed to load config! (%v)", err)
	}

	editor.Config = cfg

	return editor
}

// Shutdown tears down the terminal screen and ends the process.
func (e *Editor) Shutdown() {
	termbox.Close()
	os.Exit(0)
}

// Run starts the editor.
func (e *Editor) Run(args []string) {

	// If we have arguments, create a new buffer for each argument.
	if len(args) != 0 {
		for _, path := range args {
			b, err := buffer.Create(&e.Config, path)
			if err != nil {
				e.SetStatusMessage("Error: %v", err)
				continue
			}

			e.Buffers = append(e.Buffers, b)
		}
	} else {
		b, err := buffer.Create(&e.Config, "Untitled")
		if err != nil {
			panic(err)
		}

		e.Buffers = []buffer.Buffer{b}
	}

	// Perform the initial draw of the UI.
	e.Draw()

	for {
		e.HandleEvent(termbox.PollEvent())

		// If there are no remaining buffers, terminate the program.
		if e.BufferCount() == 0 {
			e.Shutdown()
		}

		// If the last buffer was just closed, decrement the focus index.
		if e.FocusIndex >= e.BufferCount() {
			e.FocusIndex = e.BufferCount() - 1
		}

		e.Draw()
	}
}

// FB returns the focused buffer.
func (e *Editor) FB() *buffer.Buffer {
	return &e.Buffers[e.FocusIndex]
}

// BufferCount is a shorthand for getting the number of open buffers.
func (e *Editor) BufferCount() int {
	return len(e.Buffers)
}

// SetStatusMessage sets the status message and the time it was set at.
func (e *Editor) SetStatusMessage(format string, args ...interface{}) {
	e.StatusMessage = fmt.Sprintf(format, args...)
	e.StatusMessageTime = time.Now()
}

func (e *Editor) ShowHelp() {
	b := buffer.FromStrings(&e.Config, "Help.txt", support.HelpMessage)
	b.IsReadOnly = true

	e.Buffers = append(e.Buffers, b)
	e.FocusIndex = e.BufferCount() - 1
}
