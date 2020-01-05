package main

import (
	"unicode"

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

// These are just temporary until language definitions are added.
var PrimaryKeywordsC = [...]string{
	"if", "else", "for", "while", "switch", "case", "break", "continue",
	"return", "struct", "union", "typedef", "static", "enum",
}

var SecondaryKeywordsC = [...]string{
	"int", "long", "double", "float", "char", "unsigned", "signed", "void", "NULL",
}

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

func HighlightLineC(l *BufferLine) {
	afterSeparator := true
	inString := false
	inComment := false

	for i, c := range l.DisplayText {
		lastHighlight := HighlightTypeNormal
		if i > 0 {
			lastHighlight = l.Highlighting[i-1]
		}

		// Highlight comment markers and all succeeding characters.
		if inComment {
			l.Highlighting[i] = HighlightTypeComment
			continue
		} else if !inString && i+2 <= len(l.DisplayText) {
			if l.DisplayText[i:i+2] == "//" {
				l.Highlighting[i] = HighlightTypeComment
				inComment = true
			}
		}

		// Highlight characters in strings and quotes.
		// TODO: Highlight escaped characters differently.
		if c == '"' || c == '\'' {
			inString = !inString
			l.Highlighting[i] = HighlightTypeString
			continue
		} else if inString {
			l.Highlighting[i] = HighlightTypeString
			continue
		}

		// Highlight numbers and decimal points.
		if unicode.IsDigit(c) &&
			(afterSeparator || lastHighlight == HighlightTypeNumber) ||
			(c == '.' && lastHighlight == HighlightTypeNumber) {
			l.Highlighting[i] = HighlightTypeNumber
			continue
		}

		// Highlight keywords and primitive types.
		if afterSeparator {
			for _, keyword := range PrimaryKeywordsC {
				keywordLen := len(keyword)

				if i+keywordLen < len(l.DisplayText) && l.DisplayText[i:i+keywordLen] == keyword {
					for j := 0; j < keywordLen; j++ {
						l.Highlighting[i+j] = HighlightTypePrimaryKeyword
					}

					i += keywordLen - 1
					break
				}
			}

			for _, keyword := range SecondaryKeywordsC {
				keywordLen := len(keyword)

				if i+keywordLen < len(l.DisplayText) && l.DisplayText[i:i+keywordLen] == keyword {
					for j := 0; j < keywordLen; j++ {
						l.Highlighting[i+j] = HighlightTypeSecondaryKeyword
					}

					i += keywordLen - 1
					break
				}
			}

			afterSeparator = false
			continue
		}

		afterSeparator = IsSeparator(c)
	}
}
