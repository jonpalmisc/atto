package editor

import (
	"errors"
	"fmt"
	"github.com/jonpalmisc/atto/internal/buffer"
	"github.com/jonpalmisc/atto/internal/config"
	"github.com/jonpalmisc/atto/internal/support"
	"io/ioutil"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// Editor is the editor instance and manages the UI.
type Editor struct {

	// The editor's buffers and the index of the focused buffer.
	Buffers    []buffer.Buffer
	FocusIndex int

	// The editor's height and width measured in rows and columns, respectively.
	Width  int
	Height int

	// The current status message and the time it was set.
	StatusMessage     string
	StatusMessageTime time.Time

	// The prompt question, answer, and whether it is active or not.
	PromptQuestion string
	PromptAnswer   string
	PromptIsActive bool

	// The user's editor configuration.
	Config config.Config
}

// CreateEditor creates a new Editor instance.
func CreateEditor() (editor Editor) {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	// Attempt to load the user's editor configuration.
	config, err := config.LoadConfig()
	if err != nil {
		editor.SetStatusMessage("Failed to load config! (%v)", err)
	}

	editor.Config = config

	return editor
}

// Shutdown tears down the terminal screen and ends the process.
func (e *Editor) Shutdown() {
	termbox.Close()
	os.Exit(0)
}

// HandleEvent executes the appropriate code in response to an event.
func (e *Editor) HandleEvent(event termbox.Event) {
	switch event.Type {
	case termbox.EventKey:
		switch event.Key {

		// Handle cursor movement keys.
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

		// Handle buffer operation keys.
		case termbox.KeyCtrlR:
			e.Open()
		case termbox.KeyCtrlO:
			e.Save()
		case termbox.KeyCtrlW:
			e.Close(e.FocusIndex)
		case termbox.KeyCtrlP:
			if e.FocusIndex+1 < e.BufferCount() {
				e.FocusIndex++
			}
		case termbox.KeyCtrlL:
			if e.FocusIndex > 0 {
				e.FocusIndex--
			}

		// Handle regular input keys.
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
	} else {
		b, err := buffer.CreateBuffer(&e.Config, "Untitled")
		if err != nil {
			panic(err)
		}

		e.Buffers = []buffer.Buffer{b}
	}

	e.Draw()

	for {
		e.HandleEvent(termbox.PollEvent())

		// If there are no remaining buffers, terminate the program.
		if e.BufferCount() == 0 {
			e.Shutdown()
		}

		// If the last buffer was just closed, decrement the focus index.
		if e.FocusIndex >= e.BufferCount() {
			e.FocusIndex = e.BufferCount() - 1
		}

		e.Draw()
	}
}

/* ---------------------------------- I/O ----------------------------------- */

// Read reads a file into a new buffer.
func (e *Editor) Read(path string) {
	b, err := buffer.CreateBuffer(&e.Config, path)
	if err != nil {
		e.SetStatusMessage("Error: %v", err)
	} else {
		e.Buffers = append(e.Buffers, b)
	}
}

// Open prompts the user for a path and creates a new buffer for it.
func (e *Editor) Open() {
	filename, err := e.Ask("Open file: ", "")
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
		e.FB().FileType = support.GuessFileType(filename)
		e.FB().IsDirty = false
	}
}

// Close closes the focused buffer.
func (e *Editor) Close(i int) {
	b := &e.Buffers[i]

	if b.IsDirty {
		a, _ := e.AskChar("Save changes? [Y/N]: ", []rune{'y', 'n'})

		switch a {
		case 'y':
			defer e.Save()
			return
		case 'n':
			break
		default:
			return
		}
	}

	e.Buffers = append(e.Buffers[:i], e.Buffers[i+1:]...)
}

/* --------------------------------- Buffer --------------------------------- */

// InsertPromptChar inserts a character into the current prompt answer.
func (e *Editor) InsertPromptChar(c rune) {
	if support.IsInsertable(c) {
		i := e.FB().CursorX - len(e.PromptQuestion)

		e.PromptAnswer = e.PromptAnswer[:i] + string(c) + e.PromptAnswer[i:]
		e.FB().CursorX++
	}
}

// DeletePromptChar removes a character from the current prompt answer.
func (e *Editor) DeletePromptChar() {
	x := e.FB().CursorX - len(e.PromptQuestion) - 1

	if x >= 0 && x < len(e.PromptAnswer) {
		e.PromptAnswer = e.PromptAnswer[:x] + e.PromptAnswer[x+1:]
		e.FB().CursorX--
	}
}

/* ----------------------------- User Interface ----------------------------- */

// DrawTitleBar draws the editor's title bar at the top of the screen.
func (e *Editor) DrawTitleBar() {
	info := support.ProgramName + " " + support.ProgramVersion
	systemTime := time.Now().Local().Format("2006-01-02 15:04")

	indicator := ""
	if e.FB().IsDirty {
		indicator = "*"
	}

	name := fmt.Sprintf("%v%v (%v/%v)", indicator, e.FB().FileName, e.FocusIndex+1, e.BufferCount())

	nameLen := len(name)
	timeLen := len(systemTime)

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

	if e.PromptIsActive {
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
	if e.PromptIsActive {
		return
	}

	e.FB().CursorDX = e.FB().FocusedRow().AdjustX(e.FB().CursorX)

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
			e.FB().CursorX = len(e.FB().FocusedRow().Text)
		}
	case CursorMoveRight:
		if e.FB().CursorX < len(e.FB().FocusedRow().Text) {
			e.FB().CursorX++
		} else if e.FB().CursorX == len(e.FB().FocusedRow().Text) && e.FB().CursorY != e.FB().Length() {
			e.FB().CursorX = 0
			e.FB().CursorY++
		}
	case CursorMoveLineStart:

		// Move the cursor to the end of the indent if the cursor is not there
		// already, otherwise, move it to the start of the line.
		if e.FB().CursorX != e.FB().FocusedRow().IndentLength() {
			e.FB().CursorX = e.FB().FocusedRow().IndentLength()
		} else {
			e.FB().CursorX = 0
		}
	case CursorMoveLineEnd:
		e.FB().CursorX = len(e.FB().FocusedRow().Text)
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
	rowLength := len(e.FB().FocusedRow().Text)
	if e.FB().CursorX > rowLength {
		e.FB().CursorX = rowLength
	}
}

// MovePromptCursor moves the cursor inside of the current prompt.
func (e *Editor) MovePromptCursor(move CursorMove) {
	x := e.FB().CursorX - len(e.PromptQuestion)

	switch move {
	case CursorMoveLeft:
		if x != 0 {
			e.FB().CursorX--
		}
	case CursorMoveRight:
		if x < len(e.PromptAnswer) {
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
func (e *Editor) FB() *buffer.Buffer {
	return &e.Buffers[e.FocusIndex]
}

// DirtyBufferCount returns the number of dirty buffers.
func (e *Editor) DirtyBufferCount() (count int) {
	for _, b := range e.Buffers {
		if b.IsDirty {
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
		e.PromptIsActive = false
	}()

	e.PromptIsActive = true
	e.PromptQuestion, e.PromptAnswer = q, a

	e.FB().CursorY = e.Height
	e.FB().CursorX = len(e.PromptQuestion) + len(e.PromptAnswer)

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
				return e.PromptAnswer, nil
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
		e.PromptIsActive = false
	}()

	e.PromptIsActive = true
	e.PromptQuestion, e.PromptAnswer = q, ""

	e.FB().CursorY = e.Height
	e.FB().CursorX = len(e.PromptQuestion) + len(e.PromptAnswer)

	for {
		e.Draw()

		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return '\x00', errors.New("user cancelled")
			default:
				if support.IsInsertable(event.Ch) {
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
