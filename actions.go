package t8go

import "github.com/redghc/t8go/helpers"

// DrawPixel sets a pixel at the specified coordinates (x, y) in the display buffer.
func (t *T8Go) DrawPixel(x, y int16) {
	t.SetPixel(x, y, true)
}

// DrawLine draws a line between two points (startX, startY) and (endX, endY)
// using Bresenham's algorithm. The result is rendered into the display buffer.
func (t *T8Go) DrawLine(startX, startY, endX, endY int16) {
	swapXY := false

	// Determine if the line is steep (more vertical than horizontal).
	if helpers.AbsDiff(endY, startY) > helpers.AbsDiff(endX, startX) {
		startX, startY = startY, startX
		endX, endY = endY, endX
		swapXY = true
	}

	// ? Always draw from left to right
	if startX > endX {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	deltaX := endX - startX
	deltaY := endY - startY
	errorAccumulator := deltaX / 2

	// ? Determine Y direction
	yStep := int16(1)
	if endY < startY {
		yStep = -1
	}

	currentY := startY

	for currentX := startX; currentX <= endX; currentX++ {
		if swapXY {
			t.SetPixel(currentY, currentX, true)
		} else {
			t.SetPixel(currentX, currentY, true)
		}

		errorAccumulator -= helpers.Abs16(deltaY)
		if errorAccumulator < 0 {
			currentY += yStep
			errorAccumulator += deltaX
		}
	}
}

// DrawBox draws a filled rectangle with the top-left corner at (x, y) and the specified width and height.
func (t *T8Go) DrawBox(x, y, width, height int16) {
	if width <= 0 || height <= 0 {
		return
	}

	for deltaY := range height {
		for deltaX := range width {
			t.SetPixel(x+deltaX, y+deltaY, true)
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

// DrawCircle draws a outlined circle with center at (x0, y0) and specified radius.
// The diameter of the circle is 2*radius + 1.
// The options parameter allows drawing specific sections of the circle.
// If options is empty or contains DRAW_FULL, the entire circle is drawn.
func (t *T8Go) DrawCircle(x0, y0, rad int16, options []DrawQuadrants) {
	f := int16(1 - rad)
	ddF_x := int16(1)
	ddF_y := -2 * rad
	x := int16(0)
	y := rad

	t.drawCircleSection(x, y, x0, y0, options)

	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x

		t.drawCircleSection(x, y, x0, y0, options)
	}
}

func (t *T8Go) drawCircleSection(x, y, x0, y0 int16, options []DrawQuadrants) {
	if shouldDraw(options, DRAW_TOP_RIGHT) {
		t.drawArcPixel(x0+x, y0-y)
		t.drawArcPixel(x0+y, y0-x)
	}
	if shouldDraw(options, DRAW_TOP_LEFT) {
		t.drawArcPixel(x0-x, y0-y)
		t.drawArcPixel(x0-y, y0-x)
	}
	if shouldDraw(options, DRAW_BOTTOM_RIGHT) {
		t.drawArcPixel(x0+x, y0+y)
		t.drawArcPixel(x0+y, y0+x)
	}
	if shouldDraw(options, DRAW_BOTTOM_LEFT) {
		t.drawArcPixel(x0-x, y0+y)
		t.drawArcPixel(x0-y, y0+x)
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

// DrawArc draws an outlined arc with center at (x0, y0), specified radius, and angular range from start to end.
// The start and end angles are specified as values from 0 to 255, where:
// - 0 represents 0 degrees (right)
// - 64 represents 90 degrees (up)
// - 128 represents 180 degrees (left)
// - 192 represents 270 degrees (down)
// - 255 represents 360 degrees (full circle)
func (t *T8Go) DrawArc(x0, y0, rad int16, start, end uint8) {
	// Manage angle inputs
	full := (start == end)
	inverted := (start > end)
	var aStart, aEnd uint8
	if inverted {
		aStart = end
		aEnd = start
	} else {
		aStart = start
		aEnd = end
	}

	// Initialize variables
	x := int16(0)
	y := int16(rad)
	d := int16(rad) - 1

	// Trace arc radius with the Andres circle algorithm (process each pixel of a 1/8th circle of radius rad)
	for y >= x {
		// Get the percentage of 1/8th circle drawn with a fast approximation of arctan(x/y)
		var ratio uint32
		if y != 0 {
			ratio = uint32(x) * 255 / uint32(y)                          // x/y [0..255]
			ratio = ratio * (770195 - (ratio-255)*(ratio+941)) / 6137491 // arctan(x/y) [0..32]
		}

		// Fill the pixels of the 8 sections of the circle, but only on the arc defined by the angles (start and end)
		if full || ((ratio >= uint32(aStart) && ratio < uint32(aEnd)) != inverted) {
			t.drawArcPixel(x0+y, y0-x)
		}
		if full || (((ratio+uint32(aEnd)) > 63 && (ratio+uint32(aStart)) <= 63) != inverted) {
			t.drawArcPixel(x0+x, y0-y)
		}
		if full || (((ratio+64) >= uint32(aStart) && (ratio+64) < uint32(aEnd)) != inverted) {
			t.drawArcPixel(x0-x, y0-y)
		}
		if full || (((ratio+uint32(aEnd)) > 127 && (ratio+uint32(aStart)) <= 127) != inverted) {
			t.drawArcPixel(x0-y, y0-x)
		}
		if full || (((ratio+128) >= uint32(aStart) && (ratio+128) < uint32(aEnd)) != inverted) {
			t.drawArcPixel(x0-y, y0+x)
		}
		if full || (((ratio+uint32(aEnd)) > 191 && (ratio+uint32(aStart)) <= 191) != inverted) {
			t.drawArcPixel(x0-x, y0+y)
		}
		if full || (((ratio+192) >= uint32(aStart) && (ratio+192) < uint32(aEnd)) != inverted) {
			t.drawArcPixel(x0+x, y0+y)
		}
		if full || (((ratio+uint32(aEnd)) > 255 && (ratio+uint32(aStart)) <= 255) != inverted) {
			t.drawArcPixel(x0+y, y0+x)
		}

		// Run Andres circle algorithm to get to the next pixel
		if d >= 2*x {
			d = d - 2*x - 1
			x = x + 1
		} else if d < 2*(int16(rad)-y) {
			d = d + 2*y - 1
			y = y - 1
		} else {
			d = d + 2*(y-x-1)
			y = y - 1
			x = x + 1
		}
	}
}

// drawArcPixel draws a pixel if it's within the display bounds
func (t *T8Go) drawArcPixel(x, y int16) {
	if x >= 0 && y >= 0 && x < 128 && y < 64 {
		t.DrawPixel(x, y)
	}
}

// DrawVLine draws a vertical line starting at (x, y) with the specified length.
func (t *T8Go) DrawVLine(x, y, length int16) {
	if length <= 0 {
		return
	}
	for i := int16(0); i < length; i++ {
		t.SetPixel(x, y+i, true)
	}
}

// DrawDisc draws a filled circle (disc) with center at (x0, y0) and specified radius.
// The options parameter allows drawing specific sections of the disc.
// If options is empty or contains DRAW_FULL, the entire disc is drawn.
func (t *T8Go) DrawDisc(x0, y0, rad int16, options []DrawQuadrants) {
	f := int16(1 - rad)
	ddF_x := int16(1)
	ddF_y := -2 * rad
	x := int16(0)
	y := rad

	t.drawDiscSection(x, y, x0, y0, options)

	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x

		t.drawDiscSection(x, y, x0, y0, options)
	}
}

func (t *T8Go) drawDiscSection(x, y, x0, y0 int16, options []DrawQuadrants) {
	// Upper right
	if shouldDraw(options, DRAW_TOP_RIGHT) {
		t.DrawVLine(x0+x, y0-y, y+1)
		t.DrawVLine(x0+y, y0-x, x+1)
	}

	// Upper left
	if shouldDraw(options, DRAW_TOP_LEFT) {
		t.DrawVLine(x0-x, y0-y, y+1)
		t.DrawVLine(x0-y, y0-x, x+1)
	}

	// Lower right
	if shouldDraw(options, DRAW_BOTTOM_RIGHT) {
		t.DrawVLine(x0+x, y0, y+1)
		t.DrawVLine(x0+y, y0, x+1)
	}

	// Lower left
	if shouldDraw(options, DRAW_BOTTOM_LEFT) {
		t.DrawVLine(x0-x, y0, y+1)
		t.DrawVLine(x0-y, y0, x+1)
	}
}
