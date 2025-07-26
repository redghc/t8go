package t8go

type Display interface {
	Size() (width, height int16) // Size returns the display dimensions
	BufferSize() int             // BufferSize returns the size of the display buffer
	Buffer() []byte              // Buffer returns the buffer

	ClearBuffer()           // ClearBuffer clears the display buffer
	ClearDisplay()          // ClearDisplay clears the image buffer and display
	Command(cmd byte) error // Send a command to the display
	Display() error         // Send the current buffer to the display
}

// ----------

type T8Go struct {
	display Display
	buffer  []byte
}

// ----------

func New(display Display) *T8Go {
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

// Size returns the display dimensions
func (t *T8Go) Size() (width, height int16) {
	return t.display.Size()
}

// Command sends a command to the display
func (t *T8Go) Command(cmd byte) error {
	return t.display.Command(cmd)
}
