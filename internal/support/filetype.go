package support

import (
	"path/filepath"
)

// FileType represents a type of file.
type FileType string

const (
	FileTypeMakefile FileType = "Makefile"
	FileTypeCMake    FileType = "CMake"

	// -- Go --
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
func GuessFileType(path string) FileType {

	// In theory the file name could be passed to this function but this is more
	// convenient since this function is used when constructing buffers.
	_, name := filepath.Split(path)

	// Handle file types which have specific names.
	switch name {
	case "Makefile":
		return FileTypeMakefile
	case "CMakeLists.txt":
		return FileTypeCMake
	}

	// Attempt to determine the file's type by the extension.
	switch filepath.Ext(name) {
	case ".go":
		return FileTypeGo
	case ".mod":
		return FileTypeGoModule
	case ".h", ".c":
		return FileTypeC
	case ".hpp", ".cpp", ".cc":
		return FileTypeCPP
	case ".md":
		return FileTypeMarkdown
	case ".txt":
		return FileTypePlaintext
	}

	return FileTypeUnknown
}
