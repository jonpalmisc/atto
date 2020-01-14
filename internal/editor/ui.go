package editor

import (
	"fmt"
	"time"

	"github.com/jonpalmisc/atto/internal/support"
	"github.com/nsf/termbox-go"
)

// DrawTitleBar draws the editor's title bar at the top of the screen.
func (e *Editor) DrawTitleBar() {
	info := "Atto " + support.Version
	localTime := time.Now().Local().Format("2006-01-02 15:04")

	name := fmt.Sprintf("%v (%v/%v)", e.FB().FileName, e.FocusIndex+1, e.BufferCount())
	if e.FB().IsDirty {
		name = "*" + name
	}

	nameLen := len(name)
	timeLen := len(localTime)

	// Draw the title bar canvas.
	for x := 0; x < e.Width; x++ {
		termbox.SetCell(x, 0, ' ', termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the info banner on the left.
	for x := 0; x < len(info); x++ {
		termbox.SetCell(x, 0, rune(info[x]),
			termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the current file's name in the center.
	namePadding := (e.Width - nameLen) / 2
	for x := 0; x < nameLen; x++ {
		termbox.SetCell(namePadding+x, 0, rune(name[x]),
			termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the system time on the right.
	for x := 0; x < timeLen; x++ {
		termbox.SetCell(e.Width-timeLen+x, 0, rune(localTime[x]),
			termbox.ColorBlack, termbox.ColorWhite)
	}
}

// DrawStatusBar draws the editor's status bar on the bottom of the screen.
func (e *Editor) DrawStatusBar() {
	left := ""
	if e.PromptIsActive {
		left = e.PromptQuestion + e.PromptAnswer
	} else if time.Now().Before(e.StatusMessageTime.Add(3 * time.Second)) {
		left = e.StatusMessage
	}

	right := fmt.Sprintf(" | %v | Line %v, Column %v", e.FB().FileType, e.FB().CursorY, e.FB().CursorDX+1)

	leftLen := len(left)
	rightLen := len(right)

	// Draw the status bar canvas.
	for x := 0; x < e.Width; x++ {
		termbox.SetCell(x, e.Height-1, ' ', termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the current prompt or status message on the left if it hasn't expired.
	for x := 0; x < leftLen; x++ {
		termbox.SetCell(x, e.Height-1, rune(left[x]), termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the file type and position on the right.
	for x := 0; x < rightLen; x++ {
		termbox.SetCell(e.Width-rightLen+x, e.Height-1, rune(right[x]), termbox.ColorBlack, termbox.ColorWhite)
	}
}

// DrawBuffer draws the editor's focused buffer.
func (e *Editor) DrawBuffer() {
	for y := 0; y < e.Height-2; y++ {
		i := y + e.FB().OffsetY

		if i < e.FB().Length() {
			line := e.FB().Lines[i]
			length := len(line.DisplayText) - e.FB().OffsetX

			text := line.DisplayText
			tokens := line.TokenTypes

			if length > 0 {
				for x, c := range text[e.FB().OffsetX : e.FB().OffsetX+length] {
					termbox.SetCell(x, y+1, c, tokens[x].Color(), 0)
				}
			}
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
