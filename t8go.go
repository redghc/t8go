package t8go

type Display interface {
	Size() (width, height int16) // Size returns the display dimensions

	Command(cmd byte) error // Send a command to the display
}

// ----------

type T8Go struct {
	display Display
	buffer  []byte
}

// ----------

func New(display Display) *T8Go {
	width, height := display.Size()
	bufferSize := int(width) * int(height) // Calculate buffer size based on display dimensions

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
