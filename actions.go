package t8go

import "github.com/redghc/t8go/helpers"

// DrawPixel sets a pixel at the specified coordinates (x, y) in the display buffer.
func (t *T8Go) DrawPixel(x, y uint8) {
	t.SetPixel(x, y, true)
}

// DrawLine draws a line between two points (x1, y1) and (x2, y2) using Bresenham's algorithm to the display buffer.
func (t *T8Go) DrawLine(x1, y1, x2, y2 uint8) {
	width, height := t.Size()
	if x1 >= width || x2 >= width || y1 >= height || y2 >= height {
		return // Out of bounds
	}

	x, y := int(x1), int(y1)
	distanceX, distanceY := int(x2)-x, int(y2)-y
	directionX, directionY := helpers.Direction(distanceX), helpers.Direction(distanceY)
	absDx, absDy := helpers.Abs(distanceX), helpers.Abs(distanceY)

	var accumulatedError int
	if absDx > absDy {
		accumulatedError = absDx / 2
		for i := 0; i <= absDx; i++ {
			t.SetPixel(uint8(x), uint8(y), true)
			x += directionX
			accumulatedError -= absDy
			if accumulatedError < 0 {
				y += directionY
				accumulatedError += absDx
			}
		}
	} else {
		accumulatedError := absDy / 2
		for i := 0; i <= absDy; i++ {
			t.SetPixel(uint8(x), uint8(y), true)
			y += directionY
			accumulatedError -= absDx
			if accumulatedError < 0 {
				x += directionX
				accumulatedError += absDy
			}
		}
	}
}

// DrawBox draws a filled rectangle with the top-left corner at (x, y) and the specified width and height.
func (t *T8Go) DrawBox(x, y, width, height uint8) {
	displayWidth, displayHeight := t.Size()
	if x >= displayWidth || y >= displayHeight || x+width > displayWidth || y+height > displayHeight {
		return // Out of bounds
	}

	for i := uint8(0); i < width; i++ {
		for j := uint8(0); j < height; j++ {
			t.SetPixel(x+i, y+j, true)
		}
	}
}

// DrawBoxCoords draws a filled rectangle with the top-left corner at (x1, y1) and the bottom-right corner at (x2, y2).
func (t *T8Go) DrawBoxCoords(x1, y1, x2, y2 uint8) {
	width, height := t.Size()
	if x1 >= width || x2 >= width || y1 >= height || y2 >= height {
		return // Out of bounds
	}

	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			t.SetPixel(x, y, true)
		}
	}
}

// DrawFrame draws a rectangle outline with the top-left corner at (x, y) and the specified width and height.
func (t *T8Go) DrawFrame(x, y, width, height uint8) {
	displayWidth, displayHeight := t.Size()
	if x >= displayWidth || y >= displayHeight || x+width > displayWidth || y+height > displayHeight {
		return // Out of bounds
	}

	// Draw top and bottom edges
	for i := range width {
		t.SetPixel(x+i, y, true)          // Top edge
		t.SetPixel(x+i, y+height-1, true) // Bottom edge
	}

	// Draw left and right edges
	for j := range height {
		t.SetPixel(x, y+j, true)         // Left edge
		t.SetPixel(x+width-1, y+j, true) // Right edge
	}
}
