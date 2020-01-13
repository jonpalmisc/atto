package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// Editor is the editor instance and manages the UI.
type Editor struct {
	Buffers    []Buffer
	FocusIndex int

	// The editor's height and width measured in rows and columns, respectively.
	Width  int
	Height int

	// The current status message and the time it was set.
	StatusMessage     string
	StatusMessageTime time.Time

	PromptActive bool
	Question     string
	Answer       string

	Config Config
}

// MakeEditor creates a new Editor instance.
func MakeEditor() Editor {
	editor := Editor{}

	config, err := LoadConfig()
	if err != nil {
		editor.SetStatusMessage("Failed to load config! (%v)", err)
	}

	editor.Config = config

	if err = termbox.Init(); err != nil {
		panic(err)
	}

	return editor
}

// Quit closes the editor and terminates the program.
func (e *Editor) Quit() {

	// TODO: Make this check all buffers not just the focused one!
	if e.FB().Dirty {
		choices := []rune{'y', 'n'}
		a, _ := e.AskChar("Save changes? [Y=Save / N=Quit / Ctrl+C=Cancel]: ", choices)
		switch a {
		case 'Y', 'y':
			defer e.Save()
			return
		case 'N', 'n':
			break
		default:
			return
		}
	}

	termbox.Close()
	os.Exit(0)
}

func (e *Editor) HandleEvent(event termbox.Event) {
	switch event.Type {
	case termbox.EventKey:
		switch event.Key {
		case termbox.KeyArrowUp:
			e.MoveCursor(CursorMoveUp)
		case termbox.KeyArrowDown:
			e.MoveCursor(CursorMoveDown)
		case termbox.KeyArrowLeft:
			e.MoveCursor(CursorMoveLeft)
		case termbox.KeyArrowRight:
			e.MoveCursor(CursorMoveRight)
		case termbox.KeyPgup:
			e.MoveCursor(CursorMovePageUp)
		case termbox.KeyPgdn:
			e.MoveCursor(CursorMovePageDown)
		case termbox.KeyCtrlA:
			e.MoveCursor(CursorMoveLineStart)
		case termbox.KeyCtrlE:
			e.MoveCursor(CursorMoveLineEnd)
		case termbox.KeyCtrlX:
			e.Quit()
		case termbox.KeyCtrlO:
			e.Save()
		case termbox.KeyCtrlR:
			e.Open()
		case termbox.KeyCtrlP:
			if e.FocusIndex+1 < e.BufferCount() {
				e.FocusIndex++
			}
		case termbox.KeyCtrlL:
			if e.FocusIndex > 0 {
				e.FocusIndex--
			}
		case termbox.KeyDelete:
		case termbox.KeyBackspace:
		case termbox.KeyBackspace2:
			e.FB().DeleteChar()
		case termbox.KeyEnter:
			e.FB().BreakLine()
		case termbox.KeyTab:
			e.FB().InsertChar('\t')
		case termbox.KeySpace:
			e.FB().InsertChar(' ')
		default:
			e.FB().InsertChar(event.Ch)
		}
	}
}

// Run starts the editor.
func (e *Editor) Run(args []string) {

	// If we have arguments, create a new buffer for each argument.
	if len(args) != 0 {
		for _, file := range args {
			e.Read(file)
		}

		e.FB().CursorY = 1
	} else {
		b, err := CreateBuffer(e, "Untitled")
		if err != nil {
			panic(err)
		}

		e.Buffers = []Buffer{b}
	}

	e.Draw()

	for {
		e.HandleEvent(termbox.PollEvent())
		e.Draw()
	}
}

/* ---------------------------------- I/O ----------------------------------- */

// Read reads a file into a new buffer.
func (e *Editor) Read(path string) {
	b, err := CreateBuffer(e, path)
	if err != nil {
		e.SetStatusMessage("Error: %v", err)
	} else {
		e.Buffers = append(e.Buffers, b)
	}
}

func (e *Editor) Open() {
	filename, err := e.Ask("Read file: ", "")
	if err != nil {
		e.SetStatusMessage("User cancelled operation.")
		return
	}

	e.Read(filename)
	e.FocusIndex = e.BufferCount() - 1
}

// Save writes the current buffer back to the file it was read from.
func (e *Editor) Save() {
	filename, err := e.Ask("Save as: ", e.FB().FileName)
	if err != nil {
		e.SetStatusMessage("Save cancelled.")
		return
	}

	var text string

	// Append each line of the buffer (plus a newline) to the string.
	bufferLen := e.FB().Length()
	for i := 0; i < bufferLen; i++ {
		text += e.FB().Lines[i].Text + "\n"
	}

	if err := ioutil.WriteFile(filename, []byte(text), 0644); err != nil {
		e.SetStatusMessage("Error: %v.", err)
	} else {
		e.SetStatusMessage("File saved successfully. (%v)", filename)

		e.FB().FileName = filename
		e.FB().FileType = GuessFileType(filename)
		e.FB().Dirty = false
	}
}

/* --------------------------------- Buffer --------------------------------- */

// InsertPromptChar is a variant of InsertChar for modifying the prompt answer.
func (e *Editor) InsertPromptChar(c rune) {
	if IsInsertable(c) {
		i := e.FB().CursorX - len(e.Question)

		e.Answer = e.Answer[:i] + string(c) + e.Answer[i:]
		e.FB().CursorX++
	}
}

// DeletePromptChar is a variant of DeleteChar for modifying the prompt answer.
func (e *Editor) DeletePromptChar() {
	x := e.FB().CursorX - len(e.Question) - 1

	if x >= 0 && x < len(e.Answer) {
		e.Answer = e.Answer[:x] + e.Answer[x+1:]
		e.FB().CursorX--
	}
}

/* ----------------------------- User Interface ----------------------------- */

// DrawTitleBar draws the editor's title bar at the top of the screen.
func (e *Editor) DrawTitleBar() {
	banner := ProgramName + " " + ProgramVersion
	systemTime := time.Now().Local().Format("2006-01-02 15:04")

	indicator := ""
	if e.FB().Dirty {
		indicator = "*"
	}

	name := fmt.Sprintf("%v%v (%v/%v)", indicator, e.FB().FileName, e.FocusIndex+1, e.BufferCount())

	nameLen := len(name)
	timeLen := len(systemTime)

	// Draw the title bar canvas.
	for x := 0; x < e.Width; x++ {
		termbox.SetCell(x, 0, ' ', termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the banner on the left.
	for x := 0; x < len(banner); x++ {
		termbox.SetCell(x, 0, rune(banner[x]),
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
		termbox.SetCell(e.Width-timeLen+x, 0, rune(systemTime[x]),
			termbox.ColorBlack, termbox.ColorWhite)
	}
}

// DrawBuffer draws the editor's focused buffer.
func (e *Editor) DrawBuffer() {
	for y := 0; y < e.Height-2; y++ {
		bufferRow := y + e.FB().OffsetY

		if bufferRow < e.FB().Length() {
			line := e.FB().Lines[bufferRow]
			length := len(line.DisplayText) - e.FB().OffsetX

			if length > 0 {
				for x, c := range line.DisplayText[e.FB().OffsetX : e.FB().OffsetX+length] {
					termbox.SetCell(x, y+1, c, line.Highlighting[x].Color(), 0)
				}
			}
		}
	}
}

// DrawStatusBar draws the editor's status bar on the bottom of the screen.
func (e *Editor) DrawStatusBar() {
	left := ""
	if e.PromptActive {
		left = e.Question + e.Answer
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

// Draw draws the entire editor - UI, buffer, etc. - to the screen & updates the
// cursor's position.
func (e *Editor) Draw() {
	defer termbox.Flush()

	// The screen's height and width should be updated on each render to account
	// for the user resizing the window.
	e.Width, e.Height = termbox.Size()
	e.ScrollView()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	e.DrawTitleBar()
	e.DrawBuffer()
	e.DrawStatusBar()

	if e.PromptActive {
		termbox.SetCursor(e.FB().CursorX, e.Height)
	} else {
		termbox.SetCursor(e.FB().CursorDX-e.FB().OffsetX, e.FB().CursorY-e.FB().OffsetY)
	}
}

/* ----------------------------- Input Handling ----------------------------- */

// CursorMove is a type of cursor movement.
type CursorMove int

const (
	// CursorMoveUp moves the cursor up one row.
	CursorMoveUp CursorMove = 0

	// CursorMoveDown moves the cursor down one row.
	CursorMoveDown CursorMove = 1

	// CursorMoveLeft moves the cursor left one column.
	CursorMoveLeft CursorMove = 2

	// CursorMoveRight moves the cursor right one column.
	CursorMoveRight CursorMove = 3

	// CursorMoveLineStart moves the cursor to the first non-whitespace
	// character of the line, or the first character of the line if the cursor
	// is already on the first non-whitespace character.
	CursorMoveLineStart CursorMove = 4

	// CursorMoveLineEnd moves the cursor to the end of the line.
	CursorMoveLineEnd CursorMove = 5

	// CursorMovePageUp moves the cursor up by the  height of the screen.
	CursorMovePageUp CursorMove = 6

	// CursorMovePageDown moves the cursor down by the  height of the screen.
	CursorMovePageDown CursorMove = 7
)

// ScrollView recalculates the offsets for the view window.
func (e *Editor) ScrollView() {

	// If the prompt is currently active, everything below can be skipped.
	if e.PromptActive {
		return
	}

	e.FB().CursorDX = e.FB().CurrentRow().AdjustX(e.FB().CursorX)

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

// MoveCursor moves the cursor according to the operation provided.
func (e *Editor) MoveCursor(move CursorMove) {
	switch move {
	case CursorMoveUp:
		if e.FB().CursorY > 1 {
			e.FB().CursorY--
		}
	case CursorMoveDown:
		if e.FB().CursorY < e.FB().Length() {
			e.FB().CursorY++
		}
	case CursorMoveLeft:
		if e.FB().CursorX != 0 {
			e.FB().CursorX--
		} else if e.FB().CursorY > 1 {
			e.FB().CursorY--
			e.FB().CursorX = len(e.FB().CurrentRow().Text)
		}
	case CursorMoveRight:
		if e.FB().CursorX < len(e.FB().CurrentRow().Text) {
			e.FB().CursorX++
		} else if e.FB().CursorX == len(e.FB().CurrentRow().Text) && e.FB().CursorY != e.FB().Length() {
			e.FB().CursorX = 0
			e.FB().CursorY++
		}
	case CursorMoveLineStart:

		// Move the cursor to the end of the indent if the cursor is not there
		// already, otherwise, move it to the start of the line.
		if e.FB().CursorX != e.FB().CurrentRow().IndentLength() {
			e.FB().CursorX = e.FB().CurrentRow().IndentLength()
		} else {
			e.FB().CursorX = 0
		}
	case CursorMoveLineEnd:
		e.FB().CursorX = len(e.FB().CurrentRow().Text)
	case CursorMovePageUp:
		if e.Height > e.FB().CursorY {
			e.FB().CursorY = 1
		} else {
			e.FB().CursorY -= e.Height - 2
		}
	case CursorMovePageDown:
		e.FB().CursorY += e.Height - 2
		e.FB().OffsetY += e.Height

		if e.FB().CursorY > e.FB().Length() {
			e.FB().CursorY = e.FB().Length() - 1
		}
	}

	// Prevent the user from moving past the end of the line.
	rowLength := len(e.FB().CurrentRow().Text)
	if e.FB().CursorX > rowLength {
		e.FB().CursorX = rowLength
	}
}

// MovePromptCursor moves the cursor inside of the current prompt.
func (e *Editor) MovePromptCursor(move CursorMove) {
	x := e.FB().CursorX - len(e.Question)

	switch move {
	case CursorMoveLeft:
		if x != 0 {
			e.FB().CursorX--
		}
	case CursorMoveRight:
		if x < len(e.Answer) {
			e.FB().CursorX++
		}
	}
}

/* -------------------------------- Internal -------------------------------- */

// BufferCount is a shorthand for getting the number of open buffers.
func (e *Editor) BufferCount() int {
	return len(e.Buffers)
}

// FB returns the focused buffer.
func (e *Editor) FB() *Buffer {
	return &e.Buffers[e.FocusIndex]
}

// DirtyBufferCount returns the number of dirty buffers.
func (e *Editor) DirtyBufferCount() (count int) {
	for _, b := range e.Buffers {
		if b.Dirty {
			count++
		}
	}

	return count
}

// SetStatusMessage sets the status message and the time it was set at.
func (e *Editor) SetStatusMessage(format string, args ...interface{}) {
	e.StatusMessage = fmt.Sprintf(format, args...)
	e.StatusMessageTime = time.Now()
}

// Ask prompts the user to answer a question and assumes control over all input
// until the question is answered or the request is cancelled.
func (e *Editor) Ask(q, a string) (string, error) {
	savedX, savedY := e.FB().CursorX, e.FB().CursorY

	defer func() {
		e.FB().CursorX, e.FB().CursorY = savedX, savedY
		e.PromptActive = false
	}()

	e.PromptActive = true
	e.Question, e.Answer = q, a

	e.FB().CursorY = e.Height
	e.FB().CursorX = len(e.Question) + len(e.Answer)

	for {
		e.Draw()

		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return "", errors.New("user cancelled")
			case termbox.KeyArrowLeft:
				e.MovePromptCursor(CursorMoveLeft)
			case termbox.KeyArrowRight:
				e.MovePromptCursor(CursorMoveRight)
			case termbox.KeyEnter:
				return e.Answer, nil
			case termbox.KeyBackspace2:
				e.DeletePromptChar()
			case termbox.KeySpace:
				e.InsertPromptChar(' ')
			default:
				e.InsertPromptChar(event.Ch)
			}
		}
	}
}

func (e *Editor) AskChar(q string, choices []rune) (rune, error) {
	savedX, savedY := e.FB().CursorX, e.FB().CursorY

	defer func() {
		e.FB().CursorX, e.FB().CursorY = savedX, savedY
		e.PromptActive = false
	}()

	e.PromptActive = true
	e.Question, e.Answer = q, ""

	e.FB().CursorY = e.Height
	e.FB().CursorX = len(e.Question) + len(e.Answer)

	for {
		e.Draw()

		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return '\x00', errors.New("user cancelled")
			default:
				if IsInsertable(event.Ch) {
					for _, r := range choices {
						if r == event.Ch {
							return r, nil
						}
					}
				}
			}
		}
	}
}
