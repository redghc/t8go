package t8go

import "github.com/redghc/t8go/helpers"

// DrawPixel sets a pixel at the specified coordinates (x, y) in the display buffer.
func (t *T8Go) DrawPixel(x, y int16) {
	t.SetPixel(x, y, true)
}

// DrawLine draws a line between two points (x1, y1) and (x2, y2) using Bresenham's algorithm to the display buffer.
func (t *T8Go) DrawLine(x1, y1, x2, y2 int16) {
	swapXY := false
	if helpers.AbsDiff(y2, y1) > helpers.AbsDiff(x2, x1) {
		x1, y1 = y1, x1
		x2, y2 = y2, x2
		swapXY = true
	}

	if x1 > x2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	deltaX := x2 - x1
	deltaY := y2 - y1
	accumulatedError := deltaX / 2
	var yStep int16 = 1
	if y2 < y1 {
		yStep = -1
	}
	y := y1

	for x := x1; x <= x2; x++ {
		if swapXY {
			t.SetPixel(y, x, true)
		} else {
			t.SetPixel(x, y, true)
		}
		accumulatedError -= helpers.Abs16(deltaY)
		if accumulatedError < 0 {
			y += yStep
			accumulatedError += deltaX
		}
	}
}

// DrawBox draws a filled rectangle with the top-left corner at (x, y) and the specified width and height.
func (t *T8Go) DrawBox(x, y, width, height int16) {
	if width <= 0 || height <= 0 {
		return
	}

	for j := int16(0); j < height; j++ {
		offsetY := y + j
		for i := int16(0); i < width; i++ {
			t.SetPixel(x+i, offsetY, true)
		}
	}
}

// DrawBoxCoords draws a filled rectangle with the top-left corner at (x1, y1) and the bottom-right corner at (x2, y2).
func (t *T8Go) DrawBoxCoords(x1, y1, x2, y2 int16) {
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}

	width := x2 - x1 + 1
	height := y2 - y1 + 1

	t.DrawBox(x1, y1, width, height)
}

// DrawFrame draws a rectangle outline with the top-left corner at (x, y) and the specified width and height.
func (t *T8Go) DrawFrame(x, y, width, height int16) {
	if width <= 1 || height <= 1 {
		return
	}

	right := x + width - 1
	bottom := y + height - 1

	// Top and bottom edges
	for i := x; i <= right; i++ {
		t.SetPixel(i, y, true)
		t.SetPixel(i, bottom, true)
	}

	// Draw left and right edges, excluding corners already drawn
	for j := y + 1; j < bottom; j++ {
		t.SetPixel(x, j, true)
		t.SetPixel(right, j, true)
	}
}

// DrawFrameCoords draws a rectangle outline with the top-left corner at (x1, y1) and the bottom-right corner at (x2, y2).
func (t *T8Go) DrawFrameCoords(x1, y1, x2, y2 int16) {
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}

	width := x2 - x1 + 1
	height := y2 - y1 + 1

	t.DrawFrame(x1, y1, width, height)
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
			t.DrawPixel(x, y)
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
