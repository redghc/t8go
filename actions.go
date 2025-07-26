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
