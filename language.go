package main

type SyntaxPatterns struct {
	SingleLineCommentStart string
	MultiLineCommendStart  string
	MultiLineCommentEnd    string
}

type Langauge struct {
	Keywords []string
	Patterns SyntaxPatterns
}

var LanguageC Langauge = Langauge{
	Keywords: []string{
		"#define", "#include", "NULL", "auto", "break", "case", "char", "const",
		"continue", "default", "do", "double", "else", "enum", "extern", "float",
		"for", "goto", "if", "int", "long", "register", "return", "short",
		"signed", "sizeof", "static", "struct", "switch", "typedef", "union",
		"unsigned", "void", "volatile", "while",
	},
	Patterns: SyntaxPatterns{
		SingleLineCommentStart: "//",
		MultiLineCommendStart:  "/*",
		MultiLineCommentEnd:    "*/",
	},
}

var LanguageGo Langauge = Langauge{
	Keywords: []string{
		"append", "bool", "break", "byte", "cap", "case", "chan", "close",
		"complex", "complex128", "complex64", "const", "continue", "copy",
		"default", "defer", "delete", "else", "error", "fallthrough", "false",
		"float32", "float64", "for", "func", "go", "goto", "if", "imag",
		"import", "int", "int16", "int32", "int64", "int8", "interface", "len",
		"make", "map", "new", "nil", "package", "panic", "range", "real",
		"recover", "return", "rune", "select", "string", "struct", "switch",
		"true", "type", "uint", "uint16", "uint32", "uint64", "uint8", "uintptr",
		"var",
	},
	Patterns: SyntaxPatterns{
		SingleLineCommentStart: "//",
		MultiLineCommendStart:  "/*",
		MultiLineCommentEnd:    "*/",
	},
}
