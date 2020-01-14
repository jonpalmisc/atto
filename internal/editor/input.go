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

// InsertPromptRune inserts a rune into the current prompt answer.
func (e *Editor) InsertPromptRune(c rune) {
	if buffer.IsInsertable(c) {
		i := e.FB().CursorX - len(e.PromptQuestion)

		e.PromptAnswer = e.PromptAnswer[:i] + string(c) + e.PromptAnswer[i:]
		e.FB().CursorX++
	}
}

// DeletePromptRune deletes a rune from the current prompt answer.
func (e *Editor) DeletePromptRune() {
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

	// Save the current cursor position so it can be restored later.
	savedX, savedY := e.FB().CursorX, e.FB().CursorY

	// Restore the cursor position and close the prompt when the function exits.
	defer func() {
		e.FB().CursorX, e.FB().CursorY = savedX, savedY
		e.PromptIsActive = false
	}()

	// Activate and update the prompt.
	e.PromptIsActive = true
	e.PromptQuestion, e.PromptAnswer = q, a

	// Move the cursor to the prompt.
	e.FB().CursorY = e.Height
	e.FB().CursorX = len(e.PromptQuestion) + len(e.PromptAnswer)

	// Keep polling events until the user responds or cancels.
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
				e.DeletePromptRune()
			case termbox.KeySpace:
				e.InsertPromptRune(' ')
			default:
				e.InsertPromptRune(event.Ch)
			}
		}
	}
}

// AskRune prompts the user to respond to a question with a single rune.
func (e *Editor) AskRune(q string, choices []rune) (rune, error) {

	// Save the current cursor position so it can be restored later.
	savedX, savedY := e.FB().CursorX, e.FB().CursorY

	// Restore the cursor position and close the prompt when the function exits.
	defer func() {
		e.FB().CursorX, e.FB().CursorY = savedX, savedY
		e.PromptIsActive = false
	}()

	// Activate and update the prompt.
	e.PromptIsActive = true
	e.PromptQuestion, e.PromptAnswer = q, ""

	// Move the cursor to the prompt.
	e.FB().CursorY = e.Height
	e.FB().CursorX = len(e.PromptQuestion) + len(e.PromptAnswer)

	// Keep polling events until the user responds or cancels.
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
