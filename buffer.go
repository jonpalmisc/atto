package main

import (
	"bufio"
	"fmt"
	"os"
)

type Buffer struct {
	Editor *Editor

	// The name and type of the file currently being edited.
	FileName string
	FileType FileType

	Lines   []BufferLine
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

// CreateBuffer creates a new buffer for a given path.
func CreateBuffer(editor *Editor, path string) (Buffer, error) {
	b := Buffer{
		Editor:   editor,
		FileName: path,
		FileType: GuessFileType(path),
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
	if len(b.Lines) == 0 {
		b.InsertLine(0, "")
	}

	f.Close()

	return b, nil
}

// Length returns the buffer's length (number of lines).
func (b *Buffer) Length() int {
	return len(b.Lines)
}

// FocusedRow returns the buffer's focused row.
func (b *Buffer) FocusedRow() *BufferLine {
	return &b.Lines[b.CursorY-1]
}

// InsertLine inserts a new line to the buffer at the given index.
func (b *Buffer) InsertLine(i int, text string) {

	// Ensure the index we are trying to insert at is valid.
	if i >= 0 && i <= len(b.Lines) {

		// https://github.com/golang/go/wiki/SliceTricks
		b.Lines = append(b.Lines, BufferLine{})
		copy(b.Lines[i+1:], b.Lines[i:])
		b.Lines[i] = MakeBufferLine(b, text)
	}
}

// RemoveLine removes the line at the given index from the buffer.
func (b *Buffer) RemoveLine(i int) {
	if i >= 0 && i < len(b.Lines) {
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
		text := b.FocusedRow().Text
		indent := b.FocusedRow().IndentLength()

		b.InsertLine(b.CursorY, text[:indent]+text[b.CursorX:])
		b.FocusedRow().Text = text[:b.CursorX]
		b.FocusedRow().Update()

		b.CursorX = indent
	}

	b.CursorY++
	b.IsDirty = true
}

// InsertChar inserts a character at the cursor's position.
func (b *Buffer) InsertChar(c rune) {
	if IsInsertable(c) {
		b.FocusedRow().InsertChar(b.CursorX, c)
		b.CursorX++
		b.IsDirty = true
	}
}

// DeleteChar deletes the character to the left of the cursor.
func (b *Buffer) DeleteChar() {
	if b.CursorX == 0 && b.CursorY-1 == 0 {
		return
	} else if b.CursorX > 0 {
		b.FocusedRow().DeleteChar(b.CursorX - 1)
		b.CursorX--
	} else {
		b.CursorX = len(b.Lines[b.CursorY-2].Text)
		b.Lines[b.CursorY-2].AppendString(b.FocusedRow().Text)
		b.RemoveLine(b.CursorY - 1)
		b.CursorY--
	}

	b.IsDirty = true
}
