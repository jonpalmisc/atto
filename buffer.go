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

	// If a tab is being inserted and the editor is using soft tabs insert a
	// tab's width worth of spaces instead.
	if c == '\t' && l.Editor.Config.SoftTabs {
		l.Text = l.Text[:i] + strings.Repeat(" ", l.Editor.Config.TabSize) + l.Text[i:]
		l.Editor.CursorX += l.Editor.Config.TabSize - 1
	} else {
		l.Text = l.Text[:i] + string(c) + l.Text[i:]
	}

	l.Update()
}

// DeleteChar deletes a character from the line at the given index.
func (l *BufferLine) DeleteChar(i int) {
	if i >= 0 && i < len(l.Text) {
		l.Text = l.Text[:i] + l.Text[i+1:]
		l.Update()
	}
}

// AppendString appends a string to the line.
func (l *BufferLine) AppendString(s string) {
	l.Text += s
	l.Update()
}

// Update refreshes the DisplayText field.
func (l *BufferLine) Update() {
	// Expand tabs to spaces.
	l.DisplayText = strings.ReplaceAll(l.Text, "\t", strings.Repeat(" ", l.Editor.Config.TabSize))

	l.Highlighting = make([]HighlightType, len(l.DisplayText))

	if l.Editor.Config.UseHighlighting {
		switch l.Editor.FileType {
		case FileTypeC, FileTypeCPP:
			HighlightLine(l, &SyntaxC)
		case FileTypeGo:
			HighlightLine(l, &SyntaxGo)
		}
	}
}

// AdjustX corrects the cursor's X position to compensate for rendering effects.
func (l *BufferLine) AdjustX(x int) int {
	tabSize := l.Editor.Config.TabSize
	delta := 0

	for _, c := range l.Text[:x] {
		if c == '\t' {
			delta += (tabSize - 1) - (delta % tabSize)
		}

		delta++
	}

	return delta
}

// IndentLength gets the line's level of indentation in columns.
func (l *BufferLine) IndentLength() (indent int) {
	for j := 0; j < len(l.Text) && (l.Text[j] == ' ' || l.Text[j] == '\t'); j++ {
		indent++
	}

	return indent
}
