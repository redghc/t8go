package bitmap

import "errors"

// Config holds the configuration parameters for a bitmap display instance.
type Config struct {
	Width    uint16 // Display width in pixels (must be > 0)
	Height   uint16 // Display height in pixels (must be > 0)
	Filename string // Output bitmap filename (defaults to "display.bmp" if empty)
}

// Common errors returned by the bitmap driver.
var (
	ErrInvalidDimensions = errors.New("invalid display dimensions")  // Width or height is zero
	ErrFileWrite         = errors.New("failed to write bitmap file") // Bitmap file write failed
)
