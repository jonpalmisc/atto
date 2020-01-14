package buffer

import (
	"unicode"

	"github.com/jonpalmisc/atto/internal/syntax"
	"github.com/nsf/termbox-go"
)

func isSeparator(c rune) bool {
	switch c {
	case ' ', ',', '.', ';', '(', ')', '[', ']', '+', '-', '/', '*', '=', '%':
		return true
	default:
		return false
	}
}

func fill(slice *[]TokenType, start, length int, fill TokenType) {
	for i := 0; i < length; i++ {
		(*slice)[start+i] = fill
	}
}

// TokenType represents the type of token a rune belongs to.
type TokenType int

const (

	// TokenTypeText is the default token type.
	TokenTypeText TokenType = iota

	// TokenTypeKeyword represents a language keyword.
	TokenTypeKeyword

	// TokenTypeNumber represents a number.
	TokenTypeNumber

	// TokenTypeString represents a string.
	TokenTypeString

	// TokenTypeComment represents a comment.
	TokenTypeComment
)

// Color returns the appropriate highlighting color for a highlight type.
func (t TokenType) Color() termbox.Attribute {
	switch t {
	case TokenTypeKeyword:
		return termbox.ColorMagenta
	case TokenTypeNumber:
		return termbox.ColorBlue
	case TokenTypeString:
		return termbox.ColorGreen
	case TokenTypeComment:
		return termbox.ColorCyan
	default:
		return termbox.ColorDefault
	}
}

// Highlight updates the character to highlighting mapping for a line.
func (l *Line) Highlight(s *syntax.Syntax) {
	T := &l.TokenTypes

	// Keep track of whether we are inside a string and whether the last rune
	// was a separator.
	insideString := false
	afterSeparator := true

	// Get the display text for the line as an array of runes & the its length.
	text := []rune(l.DisplayText)
	length := len(text)

	for i := 0; i < length; i++ {
		r := text[i]

		// Get the type of the last token if accessible. or default to text.
		lastTokenType := TokenTypeText
		if i > 0 {
			lastTokenType = (*T)[i-1]
		}

		// If we are already within a string, keep highlighting until we hit
		// another quote character.
		if insideString {
			(*T)[i] = TokenTypeString

			insideString = r != '"'
			continue
		}

		// If we hit the beginning of a single line comment, highlight the rest
		// of the line and break out of the loop.
		scsPattern := &s.Patterns.SingleLineCommentStart
		scsLength := len(*scsPattern)
		if i+scsLength <= length && string(text[i:i+scsLength]) == *scsPattern {
			fill(T, i, length-i, TokenTypeComment)
			break
		}

		// If we hit a quotation mark, set insideString to true and highlight it.
		if r == '"' || r == '\'' {
			(*T)[i] = TokenTypeString
			insideString = true
			continue
		}

		// If our character is a digit, is after a separator or trailing another
		// digit, or is a decimal trailing a digit, highlight it as a number.
		isDigit := unicode.IsDigit(r)
		isAfterDigit := lastTokenType == TokenTypeNumber
		isAfterDecimal := r == '.' && lastTokenType == TokenTypeNumber
		if isDigit && (afterSeparator || isAfterDigit) || (isAfterDecimal) {
			(*T)[i] = TokenTypeNumber
			continue
		}

		// If the current rune is after a separator, check if it is the start of
		// a keyword.
		if afterSeparator {
			for _, keyword := range s.Keywords {
				keywordLength := len(keyword)

				// If testing the keyword will cause an index out of bounds
				// error, just skip to the next keyword.
				if i+keywordLength >= length {
					continue
				}

				// If the character after the keyword is not a separator, it may
				// be part of a compound word, and should not be highlighted.
				if !isSeparator(text[i+keywordLength]) {
					continue
				}

				// If we have made it this far and the slice we are inspecting
				// matches the keyword, highlight it.
				if string(text[i:i+keywordLength]) == keyword {
					fill(T, i, keywordLength, TokenTypeKeyword)
				}
			}
		}

		afterSeparator = isSeparator(r)
	}
}
