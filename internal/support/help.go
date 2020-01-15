package support

const AttoVersion string = "0.5.6"

var HelpMessage = []string{
	"Atto - A lightweight, opinionated text editor written in Go.",
	"Copyright (c) 2019-2020 Jon Palmisciano",
	"",
	"1.  Usage",
	"",
	"    $ atto <files>",
	"",
	"2.  Shortcuts",
	"",
	"    ^R  Open a new buffer",
	"    ^O  Save the current buffer",
	"    ^X  Close the current buffer",
	"",
	"    ^P  Go to the next buffer",
	"    ^L  Go to the previous buffer",
	"",
	"    ^J  Jump to a specific line",
	"    ^A  Jump to the beginning of the line",
	"    ^E  Jump to the end of the line",

	"    ^C  Cancel the active operation",
	"    ^H  Show this help screen",
	"",
	"3.  Configuration",
	"",
	"    A configuration folder has been created for you at '~/.atto'. Inside you ",
	"    will find the file 'config.yml', which you can edit to change your editor",
	"    preferences.",
	"",
}
