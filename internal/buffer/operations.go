package buffer

// InsertLine inserts a new line to the buffer at the given index.
func (b *Buffer) InsertLine(i int, text string) {
	if b.IsReadOnly {
		return
	}

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
	if b.IsReadOnly {
		return
	}

	if i >= 0 && i < b.Length() {
		b.Lines = append(b.Lines[:i], b.Lines[i+1:]...)
		b.IsDirty = true
	}
}

// BreakLine inserts a newline character and breaks the line at the cursor.
func (b *Buffer) BreakLine() {
	if b.IsReadOnly {
		return
	}

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

// InsertRune inserts a rune at the cursor's position.
func (b *Buffer) InsertRune(c rune) {
	if b.IsReadOnly {
		return
	}

	if IsInsertable(c) {
		b.FocusedLine().InsertRune(b.CursorX, c)
		b.CursorX++
		b.IsDirty = true
	}
}

// DeleteRune deletes the rune to the left of the cursor.
func (b *Buffer) DeleteRune() {
	if b.IsReadOnly {
		return
	}

	if b.CursorX == 0 && b.CursorY-1 == 0 {
		return
	} else if b.CursorX > 0 {
		b.FocusedLine().DeleteRune(b.CursorX - 1)
		b.CursorX--
	} else {
		b.CursorX = len(b.Lines[b.CursorY-2].Text)
		b.Lines[b.CursorY-2].AppendString(b.FocusedLine().Text)
		b.RemoveLine(b.CursorY - 1)
		b.CursorY--
	}

	b.IsDirty = true
}
