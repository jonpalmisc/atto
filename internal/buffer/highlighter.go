package buffer

import (
	"unicode"

	"github.com/jonpalmisc/atto/internal/syntax"
	"github.com/nsf/termbox-go"
)

type HighlightType int

const (
	HighlightTypeNormal HighlightType = iota
	HighlightTypePrimaryKeyword
	HighlightTypeSecondaryKeyword
	HighlightTypeNumber
	HighlightTypeString
	HighlightTypeComment
)

func (t HighlightType) Color() termbox.Attribute {
	switch t {
	case HighlightTypePrimaryKeyword:
		return termbox.ColorRed
	case HighlightTypeSecondaryKeyword:
		return termbox.ColorMagenta
	case HighlightTypeNumber:
		return termbox.ColorBlue
	case HighlightTypeString:
		return termbox.ColorGreen
	case HighlightTypeComment:
		return termbox.ColorCyan
	default:
		return termbox.ColorDefault
	}
}

func IsSeparator(c rune) bool {
	switch c {
	case ' ', ',', '.', ';', '(', ')', '[', ']', '+', '-', '/', '*', '=', '%':
		return true
	default:
		return false
	}
}

func HighlightLine(l *Line, s *syntax.Syntax) {
	H := &l.Highlighting

	inString := false
	afterSeparator := true

	text := []rune(l.DisplayText)
	for i := 0; i < len(text); i++ {
		c := text[i]

		lastHT := HighlightTypeNormal
		if i > 0 {
			lastHT = (*H)[i-1]
		}

		// If we are already within a string, keep highlighting until we hit
		// another quote character.
		if inString {
			(*H)[i] = HighlightTypeString

			if c == '"' {
				inString = false
			}

			continue
		}

		// If we hit the beginning of a single line comment, highlight the rest
		// of the line and break out of the loop.
		scsPattern := &s.Patterns.SingleLineCommentStart
		scsLength := len(*scsPattern)
		if i+scsLength <= len(text) && string(text[i:i+scsLength]) == *scsPattern {
			for j := i; j < len(text); j++ {
				(*H)[j] = HighlightTypeComment
			}

			break
		}

		// If we hit a quotation mark, set inString to true and highlight it.
		if c == '"' || c == '\'' {
			(*H)[i] = HighlightTypeString
			inString = true
			continue
		}

		// If our character is a digit, is after a separator or trailing another
		// digit, or is a decimal trailing a digit, highlight it as a number.
		if unicode.IsDigit(c) &&
			(afterSeparator || lastHT == HighlightTypeNumber) ||
			(c == '.' && lastHT == HighlightTypeNumber) {
			(*H)[i] = HighlightTypeNumber
			continue
		}

		if afterSeparator {
			for _, k := range s.Keywords {
				kl := len(k)

				tail := ' '
				if i+kl < len(text) {
					tail = text[i+kl]
				}

				if i+kl <= len(text) && string(text[i:i+kl]) == k && IsSeparator(tail) {
					for j := 0; j < kl; j++ {
						(*H)[i+j] = HighlightTypeSecondaryKeyword
					}
				}
			}
		}

		afterSeparator = IsSeparator(c)
	}
}
