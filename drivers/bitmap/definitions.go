package bitmap

import "errors"

// Config holds the configuration for the bitmap display
type Config struct {
	Width    uint16
	Height   uint16
	Filename string // Path to save the BMP file
}

// Common errors
var (
	ErrInvalidDimensions = errors.New("invalid display dimensions")
	ErrFileWrite         = errors.New("failed to write bitmap file")
)
