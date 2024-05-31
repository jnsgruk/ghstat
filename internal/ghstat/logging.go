package ghstat

import (
	"log/slog"
	"os"
)

// SetupLogger creates a new default logger at the correct log level.
func SetupLogger(verbose bool) {
	// Create a default slog logger with the correct handlers.
	logLevel := new(slog.LevelVar)

	// Set the default log level to "INFO", and "DEBUG" if the verbose flag.
	// was specified
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	// Setup the TextHandler and ensure our configured logger is the default.
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	slog.SetDefault(logger)
	logLevel.Set(level)
}
