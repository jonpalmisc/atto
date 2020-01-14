package buffer

import (
	"github.com/jonpalmisc/atto/internal/support"
	"github.com/jonpalmisc/atto/internal/syntax"
	"strings"
)

// Line represents a single line in a buffer.
type Line struct {
	Buffer       *Buffer
	Text         string
	DisplayText  string
	Highlighting []HighlightType
}

// MakeBufferLine creates a new Line with the given text.
func MakeBufferLine(buffer *Buffer, text string) (bl Line) {
	bl = Line{
		Buffer: buffer,
		Text:   text,
	}

	bl.Update()
	return bl
}

// InsertChar inserts a character into the line at the given index.
func (l *Line) InsertChar(i int, c rune) {

	// If a tab is being inserted and the editor is using soft tabs insert a
	// tab's width worth of spaces instead.
	if c == '\t' && l.Buffer.Config.SoftTabs {
		l.Text = l.Text[:i] + strings.Repeat(" ", l.Buffer.Config.TabSize) + l.Text[i:]
		l.Buffer.CursorX += l.Buffer.Config.TabSize - 1
	} else {
		l.Text = l.Text[:i] + string(c) + l.Text[i:]
	}

	l.Update()
}

// DeleteChar deletes a character from the line at the given index.
func (l *Line) DeleteChar(i int) {
	if i >= 0 && i < len(l.Text) {
		l.Text = l.Text[:i] + l.Text[i+1:]
		l.Update()
	}
}

// AppendString appends a string to the line.
func (l *Line) AppendString(s string) {
	l.Text += s
	l.Update()
}

// Update refreshes the DisplayText field.
func (l *Line) Update() {
	// Expand tabs to spaces.
	l.DisplayText = strings.ReplaceAll(l.Text, "\t", strings.Repeat(" ", l.Buffer.Config.TabSize))

	l.Highlighting = make([]HighlightType, len(l.DisplayText))

	if l.Buffer.Config.UseHighlighting {
		switch l.Buffer.FileType {
		case support.FileTypeC, support.FileTypeCPP:
			HighlightLine(l, &syntax.SyntaxC)
		case support.FileTypeGo:
			HighlightLine(l, &syntax.SyntaxGo)
		}
	}
}

// AdjustX corrects the cursor's X position to compensate for rendering effects.
func (l *Line) AdjustX(x int) int {
	tabSize := l.Buffer.Config.TabSize
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
func (l *Line) IndentLength() (indent int) {
	for j := 0; j < len(l.Text) && (l.Text[j] == ' ' || l.Text[j] == '\t'); j++ {
		indent++
	}

	return indent
}
