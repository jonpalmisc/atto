package buffer

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"unicode"

	"github.com/jonpalmisc/atto/internal/config"
	"github.com/jonpalmisc/atto/internal/support"
)

// IsInsertable tells whether a character is insertable into the buffer or not.
func IsInsertable(c rune) bool {
	switch unicode.ToLower(c) {
	case '!', '@', '#', '$', '%', '^', '&', '*', '(', ')',
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p',
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l',
		'z', 'x', 'c', 'v', 'b', 'n', 'm',
		'`', '~', '-', '=', '+', '\t', '[', '{', ']', '}', '\\', '|',
		';', ':', '\'', '"', ',', '<', '.', '>', '/', '?', ' ':
		return true
	default:
		return false
	}
}

// Buffer represents a text buffer corresponding to a file.
type Buffer struct {
	Config *config.Config

	// The name and type of the file currently being edited.
	FileName string
	FileType support.FileType

	// The buffer's lines and condition.
	Lines   []Line
	IsDirty bool

	// The cursor's position. The Y value must always be decremented by one when
	// accessing buffer elements since the editor's title bar occupies the first
	// row of the screen. CursorDX is the cursor's X position, with compensation
	// for extra space introduced by rendering tabs.
	CursorX  int
	CursorDX int
	CursorY  int

	// The viewport's column and row offsets.
	OffsetX int
	OffsetY int
}

// Create creates a new buffer for a given path.
func Create(config *config.Config, path string) (Buffer, error) {
	b := Buffer{
		Config:   config,
		FileName: path,
		FileType: support.GuessFileType(path),
		CursorY:  1,
	}

	// Attempt to open the file at the given path.
	f, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return Buffer{}, fmt.Errorf("%v (%v)", path, err)
	}

	// Read the file line by line and append each line to end of the buffer.
	s := bufio.NewScanner(f)
	for s.Scan() {
		b.InsertLine(b.Length(), s.Text())
	}

	// If the file is completely empty, add an empty line to the buffer.
	if b.Length() == 0 {
		b.InsertLine(0, "")
	}

	f.Close()

	return b, nil
}

// Length returns the buffer's length (number of lines).
func (b *Buffer) Length() int {
	return len(b.Lines)
}

// FocusedLine returns the buffer's focused line.
func (b *Buffer) FocusedLine() *Line {
	return &b.Lines[b.CursorY-1]
}

func (b *Buffer) Write(path string) error {
	var text string

	// Append each line of the buffer (plus a newline) to the string.
	for i := 0; i < b.Length(); i++ {
		text += b.Lines[i].Text + "\n"
	}

	err := ioutil.WriteFile(path, []byte(text), os.ModePerm)
	if err != nil {
		return err
	} else {
		b.FileName = path
		b.FileType = support.GuessFileType(path)
		b.IsDirty = false

		return nil
	}
}
