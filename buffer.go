package main

type Buffer struct {
	Editor *Editor

	Lines []BufferLine
	Dirty  bool

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

	// The name and type of the file currently being edited.
	FileName string
	FileType FileType
}

func MakeBuffer(editor *Editor) Buffer {
	return Buffer{
		Editor: editor,
		CursorY: 1,
	}
}

func (b *Buffer) Length() int {
	return len(b.Lines)
}

func (b *Buffer) CurrentRow() *BufferLine {
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
		b.Dirty = true
	}
}

// BreakLine inserts a newline character and breaks the line at the cursor.
func (b *Buffer) BreakLine() {
	if b.CursorX == 0 {
		b.InsertLine(b.CursorY-1, "")
		b.CursorX = 0
	} else {
		text := b.CurrentRow().Text
		indent := b.CurrentRow().IndentLength()

		b.InsertLine(b.CursorY, text[:indent]+text[b.CursorX:])
		b.CurrentRow().Text = text[:b.CursorX]
		b.CurrentRow().Update()

		b.CursorX = indent
	}

	b.CursorY++
	b.Dirty = true
}

// InsertChar inserts a character at the cursor's position.
func (b *Buffer) InsertChar(c rune) {
	if IsInsertable(c) {
		b.CurrentRow().InsertChar(b.CursorX, c)
		b.CursorX++
		b.Dirty = true
	}
}

// DeleteChar deletes the character to the left of the cursor.
func (b *Buffer) DeleteChar() {
	if b.CursorX == 0 && b.CursorY-1 == 0 {
		return
	} else if b.CursorX > 0 {
		b.CurrentRow().DeleteChar(b.CursorX - 1)
		b.CursorX--
	} else {
		b.CursorX = len(b.Lines[b.CursorY-2].Text)
		b.Lines[b.CursorY-2].AppendString(b.CurrentRow().Text)
		b.RemoveLine(b.CursorY - 1)
		b.CursorY--
	}

	b.Dirty = true
}