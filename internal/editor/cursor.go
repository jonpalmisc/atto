package editor

import "strconv"

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

// MoveCursor moves the cursor according to the operation provided.
func (e *Editor) MoveCursor(move CursorMove) {
	rowLength := len(e.FB().FocusedLine().Text)

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
			e.FB().CursorX = len(e.FB().PreviousLine().Text)
			e.FB().CursorY--
		}
	case CursorMoveRight:
		if e.FB().CursorX < rowLength {
			e.FB().CursorX++
		} else if e.FB().CursorX == rowLength && e.FB().CursorY != e.FB().Length() {
			e.FB().CursorX = 0
			e.FB().CursorY++
		}
	case CursorMoveLineStart:

		// Move the cursor to the end of the indent if the cursor is not there
		// already, otherwise, move it to the start of the line.
		if e.FB().CursorX != e.FB().FocusedLine().IndentLength() {
			e.FB().CursorX = e.FB().FocusedLine().IndentLength()
		} else {
			e.FB().CursorX = 0
		}
	case CursorMoveLineEnd:
		e.FB().CursorX = rowLength
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
	rowLength = len(e.FB().FocusedLine().Text)
	if e.FB().CursorX > rowLength {
		e.FB().CursorX = rowLength
	}
}

// JumpToLine asks the user for a line number and attempts to jump to that line.
func (e *Editor) JumpToLine() {

	// Prompt the user to enter a line number.
	answer, err := e.Ask("Line: ", "")
	if err != nil {
		e.SetStatusMessage("Jump cancelled.")
		return
	}

	// Attempt to convert the provided input into an integer.
	i, err := strconv.Atoi(answer)
	if err != nil {
		e.SetStatusMessage("Error: Invalid input.")
		return
	}

	lineCount := e.FB().Length()

	// Check if the target line is out of bounds, then jump to the correct line.
	if i <= 1 {
		e.FB().CursorY = 1
	} else if i > lineCount {
		e.FB().CursorY = lineCount
	} else {
		e.FB().CursorY = i
	}

	// Automatically move the cursor to the start of the new line.
	e.MoveCursor(CursorMoveLineStart)
}
