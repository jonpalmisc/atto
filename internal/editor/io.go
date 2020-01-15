package editor

import (
	"github.com/jonpalmisc/atto/internal/buffer"
)

// Open prompts the user for a path and creates a new buffer for it.
func (e *Editor) Open() {
	path, err := e.Ask("Open file: ", "")
	if err != nil {
		e.SetStatusMessage("User cancelled operation.")
		return
	}

	b, err := buffer.Create(&e.Config, path)
	if err != nil {
		e.SetStatusMessage("Error: %v", err)
	}

	e.Buffers = append(e.Buffers, b)
	e.FocusIndex = e.BufferCount() - 1
}

// Save writes the current buffer back to the file it was read from.
func (e *Editor) Save() {
	path, err := e.Ask("Save as: ", e.FB().FileName)
	if err != nil {
		e.SetStatusMessage("Save cancelled.")
		return
	}

	err = e.FB().Write(path)
	if err != nil {
		e.SetStatusMessage("Error: %v.", err)
	} else {
		e.SetStatusMessage("File saved successfully. (%v)", path)
	}
}

// Close closes the focused buffer.
func (e *Editor) Close(i int) {
	b := &e.Buffers[i]

	if b.IsDirty {
		switch e.AskBool("Save changes? [Y/N]: ") {
		case BoolAnswerYes:
			defer e.Save()
			return
		case BoolAnswerNo:
			break
		case BoolAnswerCancel:
			return
		}
	}

	e.Buffers = append(e.Buffers[:i], e.Buffers[i+1:]...)
}
