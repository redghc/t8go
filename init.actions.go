package t8go

// Clear the display buffer
func (t *T8Go) ClearBuffer() {
	for i := range t.buffer {
		t.buffer[i] = 0
	}
}

func (t *T8Go) ClearDisplay() {
	t.ClearBuffer() // Clear the buffer
}
