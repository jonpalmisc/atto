package buffer

import (
	"strings"

	"github.com/jonpalmisc/atto/internal/support"
	"github.com/jonpalmisc/atto/internal/syntax"
)

// Line represents a single line in a buffer.
type Line struct {
	Buffer      *Buffer
	Text        string
	DisplayText string
	TokenTypes  []TokenType
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

// InsertRune inserts a rune into the line at the given index.
func (l *Line) InsertRune(i int, c rune) {
	tabSize := l.Buffer.Config.TabSize

	// If a tab is being inserted and the editor is using soft tabs insert a
	// tab's width worth of spaces instead.
	if c == '\t' && l.Buffer.Config.UseSoftTabs {
		l.Text = l.Text[:i] + strings.Repeat(" ", tabSize) + l.Text[i:]
		l.Buffer.CursorX += tabSize - 1
	} else {
		l.Text = l.Text[:i] + string(c) + l.Text[i:]
	}

	l.Update()
}

// DeleteRune deletes a rune from the line at the given index.
func (l *Line) DeleteRune(i int) {
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
	tabFill := strings.Repeat(" ", l.Buffer.Config.TabSize)
	l.DisplayText = strings.ReplaceAll(l.Text, "\t", tabFill)

	l.TokenTypes = make([]TokenType, len(l.DisplayText))
	if l.Buffer.Config.UseHighlighting {
		switch l.Buffer.FileType {
		case support.FileTypeC, support.FileTypeCPP:
			l.Highlight(&syntax.LanguageC)
		case support.FileTypeGo:
			l.Highlight(&syntax.LanguageGo)
		}
	}
}

// AdjustedX returns the cursor's X position compensated for tab expansion.
func (l *Line) AdjustedX(x int) int {
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
