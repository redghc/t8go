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

type DrawQuadrants int

const (
	DRAW_FULL DrawQuadrants = iota
	DRAW_TOP_LEFT
	DRAW_TOP_RIGHT
	DRAW_BOTTOM_RIGHT
	DRAW_BOTTOM_LEFT
)

// DrawCircle draws a filled circle with center at (x0, y0) and specified radius.
// The diameter of the circle is 2*radius + 1.
// The options parameter allows drawing specific sections of the circle.
// If options is empty or contains DRAW_FULL, the entire circle is drawn.
func (t *T8Go) DrawCircle(x0, y0, rad uint8, options []DrawQuadrants) {
	cx0 := int16(x0)
	cy0 := int16(y0)
	r := int16(rad)

	f := int16(1 - r)
	ddF_x := int16(1)
	ddF_y := -2 * r
	x := int16(0)
	y := r

	t.drawCircleSection(x, y, cx0, cy0, options)

	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x

		t.drawCircleSection(x, y, cx0, cy0, options)
	}
}

func (t *T8Go) drawCircleSection(x, y, x0, y0 int16, options []DrawQuadrants) {
	draw := func(x, y int16) {
		if x >= 0 && y >= 0 && x < 128 && y < 64 {
			t.DrawPixel(uint8(x), uint8(y))
		}
	}

	if shouldDraw(options, DRAW_TOP_RIGHT) {
		draw(x0+x, y0-y)
		draw(x0+y, y0-x)
	}
	if shouldDraw(options, DRAW_TOP_LEFT) {
		draw(x0-x, y0-y)
		draw(x0-y, y0-x)
	}
	if shouldDraw(options, DRAW_BOTTOM_RIGHT) {
		draw(x0+x, y0+y)
		draw(x0+y, y0+x)
	}
	if shouldDraw(options, DRAW_BOTTOM_LEFT) {
		draw(x0-x, y0+y)
		draw(x0-y, y0+x)
	}
}

func shouldDraw(options []DrawQuadrants, section DrawQuadrants) bool {
	if len(options) == 0 {
		return section == DRAW_FULL
	}
	for _, opt := range options {
		if opt == DRAW_FULL || opt == section {
			return true
		}
	}
	return false
}
