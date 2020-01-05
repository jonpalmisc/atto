package main

import (
	"strings"
)

// BufferLine represents a single line in a buffer.
type BufferLine struct {
	Editor       *Editor
	Text         string
	DisplayText  string
	Highlighting []HighlightType
}

// MakeBufferLine creates a new BufferLine with the given text.
func MakeBufferLine(editor *Editor, text string) (bl BufferLine) {
	bl = BufferLine{
		Editor: editor,
		Text:   text,
	}

	bl.Update()
	return bl
}

// InsertChar inserts a character into the line at the given index.
func (l *BufferLine) InsertChar(i int, c rune) {
	l.Text = l.Text[:i] + string(c) + l.Text[i:]
	l.Update()
}

// DeleteChar deletes a character from the line at the given index.
func (l *BufferLine) DeleteChar(i int) {

	// TODO: Condense this when my brain works.
	if i < 0 || i >= len(l.Text) {
		return
	}

	l.Text = l.Text[:i] + l.Text[i+1:]
	l.Update()
}

// AppendString appends a string to the line.
func (l *BufferLine) AppendString(s string) {
	l.Text += s
	l.Update()
}

// Update refreshes the DisplayText field.
func (l *BufferLine) Update() {
	// Expand tabs to spaces.
	l.DisplayText = strings.ReplaceAll(l.Text, "\t", "    ")

	l.Highlighting = make([]HighlightType, len(l.DisplayText))

	if l.Editor.Config.UseHighlighting {
		switch l.Editor.FileType {
		case FileTypeC, FileTypeCPP:
			HighlightLineC(l)
		}
	}
}

// AdjustX corrects the cursor's X position to compensate for rendering effects.
func (l *BufferLine) AdjustX(x int) int {
	delta := 0

	for _, c := range l.Text[:x] {
		if c == '\t' {
			delta += 3 - (delta % 4)
		}

		delta++
	}

	return delta
}
