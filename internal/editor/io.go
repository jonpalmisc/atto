package editor

import (
	"io/ioutil"

	"github.com/jonpalmisc/atto/internal/buffer"
	"github.com/jonpalmisc/atto/internal/support"
)

// Read reads a file into a new buffer.
func (e *Editor) Read(path string) {
	b, err := buffer.Create(&e.Config, path)
	if err != nil {
		e.SetStatusMessage("Error: %v", err)
	} else {
		e.Buffers = append(e.Buffers, b)
	}
}

// Open prompts the user for a path and creates a new buffer for it.
func (e *Editor) Open() {
	filename, err := e.Ask("Open file: ", "")
	if err != nil {
		e.SetStatusMessage("User cancelled operation.")
		return
	}

	e.Read(filename)
	e.FocusIndex = e.BufferCount() - 1
}

// Save writes the current buffer back to the file it was read from.
func (e *Editor) Save() {
	filename, err := e.Ask("Save as: ", e.FB().FileName)
	if err != nil {
		e.SetStatusMessage("Save cancelled.")
		return
	}

	var text string

	// Append each line of the buffer (plus a newline) to the string.
	bufferLen := e.FB().Length()
	for i := 0; i < bufferLen; i++ {
		text += e.FB().Lines[i].Text + "\n"
	}

	if err := ioutil.WriteFile(filename, []byte(text), 0644); err != nil {
		e.SetStatusMessage("Error: %v.", err)
	} else {
		e.SetStatusMessage("File saved successfully. (%v)", filename)

		e.FB().FileName = filename
		e.FB().FileType = support.GuessFileType(filename)
		e.FB().IsDirty = false
	}
}

// Close closes the focused buffer.
func (e *Editor) Close(i int) {
	b := &e.Buffers[i]

	if b.IsDirty {
		a, _ := e.AskRune("Save changes? [Y/N]: ", []rune{'y', 'n'})

		switch a {
		case 'y':
			defer e.Save()
			return
		case 'n':
			break
		default:
			return
		}
	}

	e.Buffers = append(e.Buffers[:i], e.Buffers[i+1:]...)
}
