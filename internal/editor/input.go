package editor

import (
	"errors"

	"github.com/jonpalmisc/atto/internal/buffer"
	"github.com/nsf/termbox-go"
)

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
			e.FB().DeleteRune()
		case termbox.KeyEnter:
			e.FB().BreakLine()
		case termbox.KeyTab:
			e.FB().InsertRune('\t')
		case termbox.KeySpace:
			e.FB().InsertRune(' ')
		default:
			e.FB().InsertRune(event.Ch)
		}
	}
}

// InsertPromptChar inserts a character into the current prompt answer.
func (e *Editor) InsertPromptChar(c rune) {
	if buffer.IsInsertable(c) {
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

// AskChar prompts the user to respond to a question with a single character.
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
				if buffer.IsInsertable(event.Ch) {
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
