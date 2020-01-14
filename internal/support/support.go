package support

import (
	"strings"
	"unicode"
)

const (
	ProgramName    string = "Atto"
	ProgramVersion string = "0.4.0"
	ProgramAuthor  string = "Jon Palmisciano <jonpalmisc@gmail.com>"
)

func IsInsertable(c rune) bool {
	switch unicode.ToLower(c) {
	case '!', '@', '#', '$', '%', '^', '&', '*', '(', ')',
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p',
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l',
		'z', 'x', 'c', 'v', 'b', 'n', 'm',
		'`', '~', '-', '=', '+', '\t', '[', '{', ']', '}', '\\', '|',
		';', ':', '\'', '"', ',', '<', '.', '>', '/', '?', ' ':
		return true
	default:
		return false
	}
}

// FileType represents a type of file.
type FileType string

const (
	FileTypeMakefile FileType = "Makefile"
	FileTypeCMake    FileType = "CMake"

	FileTypeGo       FileType = "Go"
	FileTypeGoModule FileType = "Go Module"

	// -- C/C++ --
	FileTypeC   FileType = "C"
	FileTypeCPP FileType = "C++"

	// -- Text Files --
	FileTypeMarkdown  FileType = "Markdown"
	FileTypePlaintext FileType = "Plaintext"

	FileTypeUnknown FileType = "Unknown"
)

// GuessFileType attempts to deduce a file's type from its name and extension.
func GuessFileType(name string) FileType {

	// Handle filetypes which have specific names.
	switch name {
	case "Makefile":
		return FileTypeMakefile
	case "CMakeLists.txt":
		return FileTypeCMake
	}

	parts := strings.Split(name, ".")

	// Return unknown if the file has no extension and wasn't matched earlier.
	if len(parts) < 2 {
		return FileTypeUnknown
	}

	// Attempt to determine the file's type by the extension.
	switch parts[1] {
	case "go":
		return FileTypeGo
	case "mod":
		return FileTypeGoModule
	case "h", "c":
		return FileTypeC
	case "hpp", "cpp", "cc":
		return FileTypeCPP
	case "md":
		return FileTypeMarkdown
	case "txt":
		return FileTypePlaintext
	}

	return FileTypeUnknown
}
