package t8go

type Display interface {
	Size() (width, height uint8) // Size returns the display dimensions
	BufferSize() int             // BufferSize returns the size of the display buffer
	Buffer() []byte              // Buffer returns the buffer

	ClearBuffer()                 // ClearBuffer clears the display buffer
	ClearDisplay()                // ClearDisplay clears the image buffer and display
	Command(cmd byte) error       // Send a command to the display
	Display() error               // Send the current buffer to the display
	SetPixel(x, y uint8, on bool) // SetPixel sets a pixel at (x, y) to on/off
	GetPixel(x, y uint8) bool     // GetPixel returns the state of a pixel at (x, y)
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
func (t *T8Go) Size() (width, height uint8) {
	return t.display.Size()
}

func (t *T8Go) BufferSize() int {
	return t.display.BufferSize()
}

func (t *T8Go) Buffer() []byte {
	return t.display.Buffer()
}

func (t *T8Go) ClearBuffer() {
	t.display.ClearBuffer()
}

func (t *T8Go) ClearDisplay() {
	t.display.ClearDisplay()
}

func (t *T8Go) Command(cmd byte) error {
	return t.display.Command(cmd)
}

func (t *T8Go) Display() error {
	return t.display.Display()
}

func (t *T8Go) SetPixel(x, y uint8, on bool) {
	t.display.SetPixel(x, y, on)
}

func (t *T8Go) GetPixel(x, y uint8) bool {
	return t.display.GetPixel(x, y)
}
