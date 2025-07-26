package t8go

// Clear the display buffer
func (t *T8Go) ClearBuffer() {
	for i := range t.buffer {
		t.buffer[i] = 0
	}
}

// ClearDisplay clears the image buffer and clear the display
func (t *T8Go) ClearDisplay() {
	t.ClearBuffer()
	t.display.Display(t.buffer)
}

// SendBuffer sends the buffer to the display
func (t *T8Go) SendBuffer() error {
	return t.display.Display(t.buffer)
}
