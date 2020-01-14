package editor

import (
	"errors"
	"unicode"

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

// restoreCursor just sets the cursor position but this function is meant to be
// called using defer inside prompt-related methods.
func (e *Editor) restoreCursor(x, y int) {
	e.FB().CursorX, e.FB().CursorY = x, y
}

// activatePrompt opens the prompt and moves the cursor to the prompt.
func (e *Editor) activatePrompt(question, answer string) {
	e.PromptQuestion, e.PromptAnswer, e.PromptIsActive = question, answer, true
	e.FB().CursorY, e.FB().CursorX = e.Height, len(question) + len(answer)
}

// closePrompt is just syntactic sugar for setting PromptIsActive to false, but
// it is meant to be called using defer inside prompt-related methods.
func (e *Editor) closePrompt() {
	e.PromptIsActive = false
}

// Ask prompts the user to answer a question and assumes control over all input
// until the question is answered or the request is cancelled.
func (e *Editor) Ask(question, answer string) (string, error) {

	// Restore the cursor position and close the prompt when the function exits.
	defer e.restoreCursor(e.FB().CursorX, e.FB().CursorY)
	defer e.closePrompt()

	// Activate the prompt and poll events until the user responds or cancels.
	e.activatePrompt(question, answer)
	for {
		e.Draw()

		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyCtrlC:
				return "", errors.New("user cancelled")
			case termbox.KeyEnter:
				return e.PromptAnswer, nil
			case termbox.KeyArrowLeft:
				e.MovePromptCursor(CursorMoveLeft)
			case termbox.KeyArrowRight:
				e.MovePromptCursor(CursorMoveRight)
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

// BoolAnswer represents a boolean answer choice.
type BoolAnswer int

const (

	// BoolAnswerCancel represents a cancelled prompt.
	BoolAnswerCancel BoolAnswer = -1

	// BoolAnswerNo represents a "No" answer.
	BoolAnswerNo BoolAnswer = 0

	// BoolAnswerYes represents a "Yes" answer.
	BoolAnswerYes BoolAnswer = 1
)

// AskBool asks a yes or no question with the choice of cancelling.
func (e *Editor) AskBool(question string) BoolAnswer {

	// Restore the cursor position and close the prompt when the function exits.
	defer e.restoreCursor(e.FB().CursorX, e.FB().CursorY)
	defer e.closePrompt()

	// Activate the prompt and poll events until the user responds or cancels.
	e.activatePrompt(question, "")
	for {
		e.Draw()

		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyCtrlC:
				return BoolAnswerCancel
			default:
				switch unicode.ToUpper(event.Ch) {
				case 'Y':
					return BoolAnswerYes
				case 'N':
					return BoolAnswerNo
				}
			}
		}
	}
}