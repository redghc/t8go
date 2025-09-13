// Package t8go provides a lightweight graphics library for monochrome displays.
// It is designed for embedded systems and microcontrollers, providing comprehensive
// 2D drawing capabilities optimized for resource-constrained environments.
// The library supports various drawing operations including pixels, lines, rectangles,
// circles, ellipses, arcs, and triangles.
package t8go

// New creates a new T8Go graphics context with the specified display.
// The display parameter must implement the Display interface.
// Returns a pointer to a T8Go instance that can be used for drawing operations.
func New(display Display) DisplayDrawer {
	bufferSize := display.BufferSize()

	return &T8Go{
		display: display,
		buffer:  make([]byte, bufferSize),
	}
}

// GetDisplay returns the underlying display interface
func (t *T8Go) GetDisplay() Display {
	return t.display
}

// Size returns the display dimensions as width and height in pixels.
func (t *T8Go) Size() (width, height uint16) {
	return t.display.Size()
}

// BufferSize returns the size in bytes of the display buffer.
func (t *T8Go) BufferSize() int {
	return t.display.BufferSize()
}

// Buffer returns the underlying display buffer as a byte slice.
func (t *T8Go) Buffer() []byte {
	return t.display.Buffer()
}

// ClearBuffer clears the display buffer without updating the physical display.
func (t *T8Go) ClearBuffer() {
	t.display.ClearBuffer()
}

// ClearDisplay clears both the buffer and the physical display.
func (t *T8Go) ClearDisplay() {
	t.display.ClearDisplay()
}

// Command sends a command byte to the display.
// Returns an error if the command fails to send.
func (t *T8Go) Command(cmd byte) error {
	return t.display.Command(cmd)
}

// Display sends the current buffer contents to the physical display.
// Returns an error if the display update fails.
func (t *T8Go) Display() error {
	return t.display.Display()
}

// SetPixel sets a pixel at the specified coordinates (x, y).
// If on is true, the pixel is turned on; if false, it's turned off.
func (t *T8Go) SetPixel(x, y int16, on bool) {
	t.display.SetPixel(x, y, on)
}

// GetPixel returns the state of a pixel at the specified coordinates (x, y).
// Returns true if the pixel is on, false if it's off.
func (t *T8Go) GetPixel(x, y uint8) bool {
	return t.display.GetPixel(x, y)
}
