// Package t8go provides a lightweight graphics library for monochrome displays.
// It is designed for embedded systems and micro-controllers, providing comprehensive
// 2D drawing capabilities optimized for resource-constrained environments.
// The library supports various drawing operations including pixels, lines, rectangles,
// circles, ellipses, arcs, and triangles.
package t8go

// New creates a new T8Go graphics context with the specified display.
// The display parameter must implement the Display interface.
// Returns a pointer to a T8Go instance that can be used for drawing operations.
func New(display IDisplay) IDisplayDrawer {
	w, h := display.Size()

	// colScratch holds the vertical extent of each column for filled
	// circle/ellipse rendering; index by the horizontal offset from center.
	maxDim := max(w, h)

	return &T8Go{
		display:    display,
		width:      int16(w),
		height:     int16(h),
		scratch:    make([]scanSpan, int(h)),
		colScratch: make([]scanSpan, int(maxDim)/2+1),
	}
}

// GetDisplay returns the underlying display interface.
func (t *T8Go) GetDisplay() IDisplay {
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
// The whole screen is marked dirty so a subsequent Display flushes it.
func (t *T8Go) ClearBuffer() {
	t.display.ClearBuffer()
	t.markDirtyAll()
}

// ClearDisplay clears both the buffer and the physical display.
func (t *T8Go) ClearDisplay() {
	t.display.ClearDisplay()
	t.dirty.valid = false
}

// Command sends a command byte to the display.
// Returns an error if the command fails to send.
func (t *T8Go) Command(cmd byte) error {
	return t.display.Command(cmd)
}

// Display flushes only the region modified since the last flush to the
// physical display. It is a no-op if nothing has changed.
// Returns an error if the display update fails.
//
// Note: if you write pixels directly to Buffer() (e.g. blitting a bitmap),
// those changes are not tracked, so use DisplayAll instead.
func (t *T8Go) Display() error {
	if !t.dirty.valid {
		return nil
	}

	err := t.display.DisplayRegion(
		int(t.dirty.x0), int(t.dirty.y0),
		int(t.dirty.x1), int(t.dirty.y1),
	)
	t.dirty.valid = false
	return err
}

// DisplayAll flushes the entire display buffer to the physical display,
// regardless of the tracked dirty region. Use this after modifying Buffer()
// directly.
func (t *T8Go) DisplayAll() error {
	err := t.display.Display()
	t.dirty.valid = false
	return err
}

// SetPixel sets a pixel at the specified coordinates (x, y).
// If on is true, the pixel is turned on; if false, it's turned off.
// Out-of-bounds pixels are ignored.
func (t *T8Go) SetPixel(x, y int16, on bool) {
	if x < 0 || y < 0 || x >= t.width || y >= t.height {
		return
	}
	t.markDirty(x, y)
	t.display.SetPixel(x, y, on)
}

// setPixel writes an on-pixel without dirty tracking or bounds handling.
// Drawing primitives use it for their per-pixel loops and mark the shape's
// bounding box once, avoiding per-pixel overhead.
func (t *T8Go) setPixel(x, y int16) {
	t.display.SetPixel(x, y, true)
}

// GetPixel returns the state of a pixel at the specified coordinates (x, y).
// Returns true if the pixel is on, false if it's off.
func (t *T8Go) GetPixel(x, y uint8) bool {
	return t.display.GetPixel(x, y)
}

// markDirty expands the dirty region to include the pixel (x, y), clamped to
// the visible area.
func (t *T8Go) markDirty(x, y int16) {
	if x < 0 {
		x = 0
	} else if x >= t.width {
		x = t.width - 1
	}
	if y < 0 {
		y = 0
	} else if y >= t.height {
		y = t.height - 1
	}

	if !t.dirty.valid {
		t.dirty.valid = true
		t.dirty.x0, t.dirty.y0, t.dirty.x1, t.dirty.y1 = x, y, x, y
		return
	}
	if x < t.dirty.x0 {
		t.dirty.x0 = x
	}
	if x > t.dirty.x1 {
		t.dirty.x1 = x
	}
	if y < t.dirty.y0 {
		t.dirty.y0 = y
	}
	if y > t.dirty.y1 {
		t.dirty.y1 = y
	}
}

// markDirtyAll marks the entire screen as dirty.
func (t *T8Go) markDirtyAll() {
	t.dirty = dirtyRect{valid: true, x0: 0, y0: 0, x1: t.width - 1, y1: t.height - 1}
}
