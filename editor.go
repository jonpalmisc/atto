package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// Editor is the editor instance and manages the UI.
type Editor struct {

	// The editor's height and width measured in rows and columns, respectively.
	Width  int
	Height int

	// The cursor's position. The Y value must always be decremented by one when
	// accessing buffer elements since the editor's title bar occupies the first
	// row of the screen. CursorDX is the cursor's X position, with compensation
	// for extra space introduced by rendering tabs.
	CursorX  int
	CursorDX int
	CursorY  int

	// The viewport's
	OffsetX int // The viewport's column offset.
	OffsetY int // The viewport's row offset.

	// The name and type of the file currently being edited.
	FileName string
	FileType FileType

	// The buffer for the current file and whether it has been modifed or not.
	Buffer []BufferLine
	Dirty  bool

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

	editor.CursorY = 1
	editor.Config = config

	if err = termbox.Init(); err != nil {
		panic(err)
	}

	return editor
}

// Quit closes the editor and terminates the program.
func (e *Editor) Quit() {
	if e.Dirty {
		a, _ := e.Ask("File has unsaved changes - quit anyways? [Y/N]: ", "")
		switch a {
		case "Y", "y":
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
		case termbox.KeyCtrlS:
			e.Save()
		case termbox.KeyDelete:
		case termbox.KeyBackspace:
		case termbox.KeyBackspace2:
			e.DeleteChar()
		case termbox.KeyEnter:
			e.BreakLine()
		case termbox.KeyTab:
			e.InsertChar('\t')
		case termbox.KeySpace:
			e.InsertChar(' ')
		default:
			e.InsertChar(event.Ch)
		}
	}
}

// Run starts the editor.
func (e *Editor) Run() {
	e.Draw()

	for {
		e.HandleEvent(termbox.PollEvent())
		e.Draw()
	}
}

/* ---------------------------------- I/O ----------------------------------- */

// Open reads a file into a the buffer.
func (e *Editor) Open(path string) {
	e.FileName = path
	e.FileType = GuessFileType(path)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		_, err := os.Create(path)
		if err != nil {
			panic(err)
		}

		e.InsertLine(0, "")
	}

	f, err := os.Open(path)
	if err != nil {
		e.InsertLine(0, "")
		e.SetStatusMessage("Error: Couldn't open file: %v (%v)", path, err)
		return
	}

	// Read the file line by line and append each line to end of the buffer.
	s := bufio.NewScanner(f)
	for s.Scan() {
		e.InsertLine(len(e.Buffer), s.Text())
	}

	// If the file is completely empty, add an empty line to the buffer.
	if len(e.Buffer) == 0 {
		e.InsertLine(0, "")
	}

	// The file can now be closed since it is loaded into memory.
	f.Close()
}

// Save writes the current buffer back to the file it was read from.
func (e *Editor) Save() {
	filename, err := e.Ask("Save as: ", e.FileName)
	if err != nil {
		e.SetStatusMessage("Save cancelled.")
		return
	}

	var text string

	// Append each line of the buffer (plus a newline) to the string.
	bufferLen := len(e.Buffer)
	for i := 0; i < bufferLen; i++ {
		text += e.Buffer[i].Text + "\n"
	}

	if err := ioutil.WriteFile(filename, []byte(text), 0644); err != nil {
		e.SetStatusMessage("Error: %v.", err)
	} else {
		e.SetStatusMessage("File saved successfully. (%v)", filename)
		e.FileName = filename
		e.Dirty = false
	}
}

/* --------------------------------- Buffer --------------------------------- */

// InsertLine inserts a new line to the buffer at the given index.
func (e *Editor) InsertLine(i int, text string) {

	// Ensure the index we are trying to insert at is valid.
	if i >= 0 && i <= len(e.Buffer) {

		// https://github.com/golang/go/wiki/SliceTricks
		e.Buffer = append(e.Buffer, BufferLine{})
		copy(e.Buffer[i+1:], e.Buffer[i:])
		e.Buffer[i] = MakeBufferLine(e, text)
	}
}

// RemoveLine removes the line at the given index from the buffer.
func (e *Editor) RemoveLine(index int) {
	if index >= 0 && index < len(e.Buffer) {
		e.Buffer = append(e.Buffer[:index], e.Buffer[index+1:]...)
		e.Dirty = true
	}
}

// BreakLine inserts a newline character and breaks the line at the cursor.
func (e *Editor) BreakLine() {
	if e.CursorX == 0 {
		e.InsertLine(e.CursorY-1, "")
		e.CursorX = 0
	} else {
		text := e.CurrentRow().Text
		indent := e.CurrentRow().IndentLength()

		e.InsertLine(e.CursorY, text[:indent]+text[e.CursorX:])
		e.CurrentRow().Text = text[:e.CursorX]
		e.CurrentRow().Update()

		e.CursorX = indent
	}

	e.CursorY++
	e.Dirty = true
}

// InsertChar inserts a character at the cursor's position.
func (e *Editor) InsertChar(c rune) {
	if IsInsertable(c) {
		if e.CursorY == len(e.Buffer) {
			e.InsertLine(len(e.Buffer), "")
		}

		e.CurrentRow().InsertChar(e.CursorX, c)
		e.CursorX++
		e.Dirty = true
	}
}

// InsertPromptChar is a variant of InsertChar for modifying the prompt answer.
func (e *Editor) InsertPromptChar(c rune) {
	if IsInsertable(c) {
		i := e.CursorX - len(e.Question)

		e.Answer = e.Answer[:i] + string(c) + e.Answer[i:]
		e.CursorX++
	}
}

// DeleteChar deletes the character to the left of the cursor.
func (e *Editor) DeleteChar() {
	if e.CursorX == 0 && e.CursorY-1 == 0 {
		return
	} else if e.CursorX > 0 {
		e.CurrentRow().DeleteChar(e.CursorX - 1)
		e.CursorX--
	} else {
		e.CursorX = len(e.Buffer[e.CursorY-2].Text)
		e.Buffer[e.CursorY-2].AppendString(e.CurrentRow().Text)
		e.RemoveLine(e.CursorY - 1)
		e.CursorY--
	}

	e.Dirty = true
}

// DeletePromptChar is a variant of DeleteChar for modifying the prompt answer.
func (e *Editor) DeletePromptChar() {
	x := e.CursorX - len(e.Question) - 1

	if x >= 0 && x < len(e.Answer) {
		e.Answer = e.Answer[:x] + e.Answer[x+1:]
		e.CursorX--
	}
}

/* ----------------------------- User Interface ----------------------------- */

// DrawTitleBar draws the editor's title bar at the top of the screen.
func (e *Editor) DrawTitleBar() {
	banner := ProgramName + " " + ProgramVersion
	time := time.Now().Local().Format("2006-01-02 15:04")

	name := e.FileName
	if e.Dirty {
		name += " (*)"
	}

	nameLen := len(name)
	timeLen := len(time)

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

	// Draw the time on the right.
	for x := 0; x < timeLen; x++ {
		termbox.SetCell(e.Width-timeLen+x, 0, rune(time[x]),
			termbox.ColorBlack, termbox.ColorWhite)
	}
}

// DrawBuffer draws the editor's buffer.
func (e *Editor) DrawBuffer() {
	for y := 0; y < e.Height-2; y++ {
		bufferRow := y + e.OffsetY

		if bufferRow < len(e.Buffer) {
			line := e.Buffer[bufferRow]
			length := len(line.DisplayText) - e.OffsetX

			if length > 0 {
				for x, c := range line.DisplayText[e.OffsetX : e.OffsetX+length] {
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

	right := fmt.Sprintf(" | %v | Line %v, Column %v", e.FileType, e.CursorY, e.CursorDX+1)

	leftLen := len(left)
	rightLen := len(right)

	// Draw the status bar canvas.
	for x := 0; x < e.Width; x++ {
		termbox.SetCell(x, e.Height-1, ' ',
			termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the current prompt or status message on the left if it hasn't expired.
	for x := 0; x < leftLen; x++ {
		termbox.SetCell(x, e.Height-1, rune(left[x]),
			termbox.ColorBlack, termbox.ColorWhite)
	}

	// Draw the file type and position on the right.
	for x := 0; x < rightLen; x++ {
		termbox.SetCell(e.Width-rightLen+x, e.Height-1, rune(right[x]),
			termbox.ColorBlack, termbox.ColorWhite)
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
		termbox.SetCursor(e.CursorX, e.Height)
	} else {
		termbox.SetCursor(e.CursorDX-e.OffsetX, e.CursorY-e.OffsetY)
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

	e.CursorDX = e.CurrentRow().AdjustX(e.CursorX)

	if e.CursorY-1 < e.OffsetY {
		e.OffsetY = e.CursorY - 1
	}

	if e.CursorY+2 >= e.OffsetY+e.Height {
		e.OffsetY = e.CursorY - e.Height + 2
	}

	if e.CursorDX < e.OffsetX {
		e.OffsetX = e.CursorDX
	}

	if e.CursorDX >= e.OffsetX+e.Width {
		e.OffsetX = e.CursorDX - e.Width + 1
	}
}

// MoveCursor moves the cursor according to the operation provided.
func (e *Editor) MoveCursor(move CursorMove) {
	switch move {
	case CursorMoveUp:
		if e.CursorY > 1 {
			e.CursorY--
		}
	case CursorMoveDown:
		if e.CursorY < len(e.Buffer) {
			e.CursorY++
		}
	case CursorMoveLeft:
		if e.CursorX != 0 {
			e.CursorX--
		} else if e.CursorY > 1 {
			e.CursorY--
			e.CursorX = len(e.CurrentRow().Text)
		}
	case CursorMoveRight:
		if e.CursorX < len(e.CurrentRow().Text) {
			e.CursorX++
		} else if e.CursorX == len(e.CurrentRow().Text) && e.CursorY != len(e.Buffer) {
			e.CursorX = 0
			e.CursorY++
		}
	case CursorMoveLineStart:

		// Move the cursor to the end of the indent if the cursor is not there
		// already, otherwise, move it to the start of the line.
		if e.CursorX != e.CurrentRow().IndentLength() {
			e.CursorX = e.CurrentRow().IndentLength()
		} else {
			e.CursorX = 0
		}
	case CursorMoveLineEnd:
		e.CursorX = len(e.CurrentRow().Text)
	case CursorMovePageUp:
		if e.Height > e.CursorY {
			e.CursorY = 1
		} else {
			e.CursorY -= e.Height - 2
		}
	case CursorMovePageDown:
		e.CursorY += e.Height - 2
		e.OffsetY += e.Height

		if e.CursorY > len(e.Buffer) {
			e.CursorY = len(e.Buffer) - 1
		}
	}

	// Prevent the user from moving past the end of the line.
	rowLength := len(e.CurrentRow().Text)
	if e.CursorX > rowLength {
		e.CursorX = rowLength
	}
}

// MovePromptCursor moves the cursor inside of the current prompt.
func (e *Editor) MovePromptCursor(move CursorMove) {
	x := e.CursorX - len(e.Question)

	switch move {
	case CursorMoveLeft:
		if x != 0 {
			e.CursorX--
		}
	case CursorMoveRight:
		if x < len(e.Answer) {
			e.CursorX++
		}
	}
}

/* -------------------------------- Internal -------------------------------- */

func (e *Editor) CurrentRow() *BufferLine {
	return &e.Buffer[e.CursorY-1]
}

// SetStatusMessage sets the status message and the time it was set at.
func (e *Editor) SetStatusMessage(format string, args ...interface{}) {
	e.StatusMessage = fmt.Sprintf(format, args...)
	e.StatusMessageTime = time.Now()
}

// Ask prompts the user to answer a question and assumes control over all input
// until the question is answered or the request is cancelled.
func (e *Editor) Ask(q, a string) (string, error) {
	savedX, savedY := e.CursorX, e.CursorY

	defer func() {
		e.CursorX, e.CursorY = savedX, savedY
		e.PromptActive = false
	}()

	e.PromptActive = true
	e.Question, e.Answer = q, a

	e.CursorY = e.Height
	e.CursorX = len(e.Question)

	for {
		e.Draw()

		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyEsc, termbox.KeyCtrlX:
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
