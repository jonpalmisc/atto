package editor

import (
	"fmt"
	"time"

	"github.com/jonpalmisc/atto/internal/support"
	"github.com/nsf/termbox-go"
)

const (

	// BarForeground is the foreground color of title/status bars.
	BarForeground = termbox.ColorBlack

	// BarBackground is the background color of title/status bars.
	BarBackground = termbox.ColorWhite
)

// drawText is a helper function for drawing an array of runes left to right.
func drawText(text []rune, ox, y int, fg, bg termbox.Attribute) {
	for i := 0; i < len(text); i++ {
		termbox.SetCell(ox+i, y, text[i], fg, bg)
	}
}

// DrawTitleBar draws the editor's title bar at the top of the screen.
func (e *Editor) DrawTitleBar() {
	info := "Atto " + support.AttoVersion
	localTime := time.Now().Local().Format("2006-01-02 15:04")
	name := fmt.Sprintf("%v (%v/%v)", e.FB().FileName(), e.FocusIndex+1, e.BufferCount())

	// Prepend an asterisk in front of the filename if it is unsaved.
	if e.FB().IsDirty {
		name = "*" + name
	}

	// Calculate the offsets for the filename and time. The name must be
	// centered and the time must be right-aligned.
	nameOffset := (e.Width - len(name)) / 2
	timeOffset := e.Width - len(localTime)

	// Draw the bar canvas.
	for x := 0; x < e.Width; x++ {
		termbox.SetCell(x, 0, ' ', BarForeground, BarBackground)
	}

	// Draw the bar elements.
	drawText([]rune(info), 0, 0, BarForeground, BarBackground)
	drawText([]rune(name), nameOffset, 0, BarForeground, BarBackground)
	drawText([]rune(localTime), timeOffset, 0, BarForeground, BarBackground)
}

// statusBarMessage is a shorthand for getting the message for the status bar.
func (e *Editor) statusBarMessage() string {
	if e.PromptIsActive {
		return e.PromptQuestion + e.PromptAnswer
	} else if time.Now().Before(e.StatusMessageTime.Add(3 * time.Second)) {
		return e.StatusMessage
	}

	return ""
}

// DrawStatusBar draws the editor's status bar on the bottom of the screen.
func (e *Editor) DrawStatusBar() {
	message := e.statusBarMessage()

	// Format the file info string.
	info := fmt.Sprintf(" | %v | %v:%v", e.FB().FileType, e.FB().CursorY, e.FB().CursorDX+1)
	infoOffset := e.Width - len(info)

	// Draw the bar canvas.
	for x := 0; x < e.Width; x++ {
		termbox.SetCell(x, e.Height-1, ' ', BarForeground, BarBackground)
	}

	// Draw the bar elements.
	drawText([]rune(message), 0, e.Height-1, BarForeground, BarBackground)
	drawText([]rune(info), infoOffset, e.Height-1, BarForeground, BarBackground)
}

// DrawBuffer draws the editor's focused buffer.
func (e *Editor) DrawBuffer() {
	for y := 0; y < e.Height-2; y++ {
		i := y + e.FB().OffsetY

		// Return early if we reach the end of the buffer.
		if i >= e.FB().Length() {
			return
		}

		line := e.FB().Lines[i]
		length := len(line.DisplayText) - e.FB().OffsetX

		// Skip to the next line if we have nothing to draw.
		if length <= 0 {
			continue
		}

		text, tokens := line.DisplayText, line.TokenTypes
		startIndex, endIndex := e.FB().OffsetX, e.FB().OffsetX+length

		for x, c := range text[startIndex:endIndex] {
			termbox.SetCell(x, y+1, c, tokens[x].Color(), 0)
		}
	}
}

// ScrollView recalculates the offsets for the view window.
func (e *Editor) ScrollView() {

	// If the prompt is currently active, everything below can be skipped.
	if e.PromptIsActive {
		return
	}

	e.FB().CursorDX = e.FB().FocusedLine().AdjustedX(e.FB().CursorX)

	if e.FB().CursorY-1 < e.FB().OffsetY {
		e.FB().OffsetY = e.FB().CursorY - 1
	}

	if e.FB().CursorY+2 >= e.FB().OffsetY+e.Height {
		e.FB().OffsetY = e.FB().CursorY - e.Height + 2
	}

	if e.FB().CursorDX < e.FB().OffsetX {
		e.FB().OffsetX = e.FB().CursorDX
	}

	if e.FB().CursorDX >= e.FB().OffsetX+e.Width {
		e.FB().OffsetX = e.FB().CursorDX - e.Width + 1
	}
}

// Draw draws the entire editor - UI, buffer, etc. - to the screen & updates the
// cursor's position.
func (e *Editor) Draw() {

	// The screen's height and width should be updated on each render to account
	// for the user resizing the window.
	e.Width, e.Height = termbox.Size()
	e.ScrollView()

	err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if err != nil {
		panic(err)
	}

	e.DrawTitleBar()
	e.DrawBuffer()
	e.DrawStatusBar()

	if e.PromptIsActive {
		termbox.SetCursor(e.FB().CursorX, e.Height)
	} else {
		termbox.SetCursor(e.FB().CursorDX-e.FB().OffsetX, e.FB().CursorY-e.FB().OffsetY)
	}

	err = termbox.Flush()
	if err != nil {
		panic(err)
	}
}
