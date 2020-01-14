package buffer

import (
	"bufio"
	"fmt"
	"os"
	"unicode"

	"github.com/jonpalmisc/atto/internal/config"
	"github.com/jonpalmisc/atto/internal/support"
)

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

// InsertLine inserts a new line to the buffer at the given index.
func (b *Buffer) InsertLine(i int, text string) {

	// Ensure the index we are trying to insert at is valid.
	if i >= 0 && i <= b.Length() {

		// https://github.com/golang/go/wiki/SliceTricks
		b.Lines = append(b.Lines, Line{})
		copy(b.Lines[i+1:], b.Lines[i:])
		b.Lines[i] = MakeBufferLine(b, text)
	}
}

// RemoveLine removes the line at the given index from the buffer.
func (b *Buffer) RemoveLine(i int) {
	if i >= 0 && i < b.Length() {
		b.Lines = append(b.Lines[:i], b.Lines[i+1:]...)
		b.IsDirty = true
	}
}

// BreakLine inserts a newline character and breaks the line at the cursor.
func (b *Buffer) BreakLine() {
	if b.CursorX == 0 {
		b.InsertLine(b.CursorY-1, "")
		b.CursorX = 0
	} else {
		text := b.FocusedLine().Text
		indent := b.FocusedLine().IndentLength()

		b.InsertLine(b.CursorY, text[:indent]+text[b.CursorX:])
		b.FocusedLine().Text = text[:b.CursorX]
		b.FocusedLine().Update()

		b.CursorX = indent
	}

	b.CursorY++
	b.IsDirty = true
}

// InsertChar inserts a character at the cursor's position.
func (b *Buffer) InsertChar(c rune) {
	if IsInsertable(c) {
		b.FocusedLine().InsertChar(b.CursorX, c)
		b.CursorX++
		b.IsDirty = true
	}
}

// DeleteChar deletes the character to the left of the cursor.
func (b *Buffer) DeleteChar() {
	if b.CursorX == 0 && b.CursorY-1 == 0 {
		return
	} else if b.CursorX > 0 {
		b.FocusedLine().DeleteChar(b.CursorX - 1)
		b.CursorX--
	} else {
		b.CursorX = len(b.Lines[b.CursorY-2].Text)
		b.Lines[b.CursorY-2].AppendString(b.FocusedLine().Text)
		b.RemoveLine(b.CursorY - 1)
		b.CursorY--
	}

	b.IsDirty = true
}
