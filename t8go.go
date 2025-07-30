package t8go

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
func (t *T8Go) Size() (width, height uint16) {
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

func (t *T8Go) SetPixel(x, y int16, on bool) {
	t.display.SetPixel(x, y, on)
}

func (t *T8Go) GetPixel(x, y uint8) bool {
	return t.display.GetPixel(x, y)
}
