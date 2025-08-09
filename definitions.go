package t8go

type Display interface {
	Size() (width, height uint16) // Size returns the display dimensions
	BufferSize() int              // BufferSize returns the size of the display buffer
	Buffer() []byte               // Buffer returns the buffer

	ClearBuffer()                 // ClearBuffer clears the display buffer
	ClearDisplay()                // ClearDisplay clears the image buffer and display
	Command(cmd byte) error       // Send a command to the display
	Display() error               // Send the current buffer to the display
	SetPixel(x, y int16, on bool) // SetPixel sets a pixel at (x, y) to on/off
	GetPixel(x, y uint8) bool     // GetPixel returns the state of a pixel at (x, y)
}

// ----------

type T8Go struct {
	display Display
	buffer  []byte
}

// ----------

type DrawQuadrants uint8

const (
	DrawNone        DrawQuadrants = 0
	DrawTopLeft     DrawQuadrants = 1 << 0
	DrawTopRight    DrawQuadrants = 1 << 1
	DrawBottomRight DrawQuadrants = 1 << 2
	DrawBottomLeft  DrawQuadrants = 1 << 3
	DrawAll                       = DrawTopLeft | DrawTopRight | DrawBottomRight | DrawBottomLeft
)

// has checks if the given quadrant should be drawn.
// If mask is DrawNone, it is interpreted as "draw all".
func (mask DrawQuadrants) has(quadrant DrawQuadrants) bool {
	return mask == DrawNone || (mask&quadrant) != 0
}
