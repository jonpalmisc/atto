package main

// Config holds the editor's configuration and settings.
type Config struct {
	TabSize         int
	SoftTabs        bool
	UseHighlighting bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		TabSize:         4,
		SoftTabs:        false,
		UseHighlighting: true,
	}
}
