package t8go

// SendBuffer sends the buffer to the display
func (t *T8Go) SendBuffer() error {
	return t.display.Display()
}
