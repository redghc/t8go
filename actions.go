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

// DrawVLine draws a vertical line starting from (originX, originY) with the given length.
func (t *T8Go) DrawVLine(originX, originY, length int16) {
	if length <= 0 {
		return
	}

	for offsetY := range length {
		t.SetPixel(originX, originY+offsetY, true)
	}
}

// DrawHLine draws a horizontal line starting from (startX, startY) with the given length.
func (t *T8Go) DrawHLine(startX, startY, length int16) {
	if length <= 0 {
		return
	}

	for offsetX := range length {
		t.SetPixel(startX+offsetX, startY, true)
	}
}

// DrawBox draws a filled rectangle starting from the top-left corner (originX, originY)
// with the specified dimensions: width and height.
func (t *T8Go) DrawBox(originX, originY, width, height int16) {
	if width <= 0 || height <= 0 {
		return
	}

	for offsetY := range height {
		t.DrawHLine(originX, originY+offsetY, width)
	}
}

// DrawBoxCoords draws a filled rectangle between two corners: top-left (startX, startY)
// and bottom-right (endX, endY), inclusive.
func (t *T8Go) DrawBoxCoords(startX, startY, endX, endY int16) {
	if endX < startX {
		startX, endX = endX, startX
	}
	if endY < startY {
		startY, endY = endY, startY
	}

	width := endX - startX + 1
	height := endY - startY + 1

	t.DrawBox(startX, startY, width, height)
}

// DrawFrame draws a rectangular outline starting from the top-left corner (originX, originY)
// with the specified width and height. Must be at least 2x2 to form a valid frame.
func (t *T8Go) DrawFrame(originX, originY, width, height int16) {
	if width <= 1 || height <= 1 {
		return
	}

	maxX := originX + width - 1
	maxY := originY + height - 1

	// Top and bottom horizontal edges
	t.DrawHLine(originX, originY, width)
	t.DrawHLine(originX, maxY, width)

	// Left and right vertical edges (excluding corners already drawn)
	t.DrawVLine(originX, originY+1, height-2)
	t.DrawVLine(maxX, originY+1, height-2)
}

// DrawFrameCoords draws a rectangular outline starting from the top-left corner (startX, startY)
// and bottom-right (endX, endY), inclusive.
func (t *T8Go) DrawFrameCoords(startX, startY, endX, endY int16) {
	if endX < startX {
		startX, endX = endX, startX
	}
	if endY < startY {
		startY, endY = endY, startY
	}

	width := endX - startX + 1
	height := endY - startY + 1

	t.DrawFrame(startX, startY, width, height)
}

// DrawCircle draws an outlined circle centered at (centerX, centerY) with the given radius.
// The diameter of the circle is 2*radius + 1.
// The quadrants parameter determines which sections of the circle are drawn.
// If quadrants is empty or includes DRAW_FULL, the entire circle is rendered.
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, quadrants []DrawQuadrants) {
	decision := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	circleX := int16(0)
	circleY := radius

	t.drawCircleSection(circleX, circleY, centerX, centerY, quadrants)

	for circleX < circleY {
		if decision >= 0 {
			circleY--
			deltaY += 2
			decision += deltaY
		}
		circleX++
		deltaX += 2
		decision += deltaX

		t.drawCircleSection(circleX, circleY, centerX, centerY, quadrants)
	}
}

// drawCircleSection draws the selected circle segments based on quadrant flags.
func (t *T8Go) drawCircleSection(offsetX, offsetY, centerX, centerY int16, quadrants []DrawQuadrants) {
	if t.shouldDraw(quadrants, DRAW_TOP_RIGHT) {
		t.DrawPixel(centerX+offsetX, centerY-offsetY)
		t.DrawPixel(centerX+offsetY, centerY-offsetX)
	}
	if t.shouldDraw(quadrants, DRAW_TOP_LEFT) {
		t.DrawPixel(centerX-offsetX, centerY-offsetY)
		t.DrawPixel(centerX-offsetY, centerY-offsetX)
	}
	if t.shouldDraw(quadrants, DRAW_BOTTOM_RIGHT) {
		t.DrawPixel(centerX+offsetX, centerY+offsetY)
		t.DrawPixel(centerX+offsetY, centerY+offsetX)
	}
	if t.shouldDraw(quadrants, DRAW_BOTTOM_LEFT) {
		t.DrawPixel(centerX-offsetX, centerY+offsetY)
		t.DrawPixel(centerX-offsetY, centerY+offsetX)
	}
}

// shouldDraw returns true if a specific quadrant should be drawn.
func (t *T8Go) shouldDraw(quadrants []DrawQuadrants, section DrawQuadrants) bool {
	if len(quadrants) == 0 {
		return section == DRAW_FULL
	}
	for _, option := range quadrants {
		if option == DRAW_FULL || option == section {
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
			t.DrawPixel(x0+y, y0-x)
		}
		if full || (((ratio+uint32(aEnd)) > 63 && (ratio+uint32(aStart)) <= 63) != inverted) {
			t.DrawPixel(x0+x, y0-y)
		}
		if full || (((ratio+64) >= uint32(aStart) && (ratio+64) < uint32(aEnd)) != inverted) {
			t.DrawPixel(x0-x, y0-y)
		}
		if full || (((ratio+uint32(aEnd)) > 127 && (ratio+uint32(aStart)) <= 127) != inverted) {
			t.DrawPixel(x0-y, y0-x)
		}
		if full || (((ratio+128) >= uint32(aStart) && (ratio+128) < uint32(aEnd)) != inverted) {
			t.DrawPixel(x0-y, y0+x)
		}
		if full || (((ratio+uint32(aEnd)) > 191 && (ratio+uint32(aStart)) <= 191) != inverted) {
			t.DrawPixel(x0-x, y0+y)
		}
		if full || (((ratio+192) >= uint32(aStart) && (ratio+192) < uint32(aEnd)) != inverted) {
			t.DrawPixel(x0+x, y0+y)
		}
		if full || (((ratio+uint32(aEnd)) > 255 && (ratio+uint32(aStart)) <= 255) != inverted) {
			t.DrawPixel(x0+y, y0+x)
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
	if t.shouldDraw(options, DRAW_TOP_RIGHT) {
		t.DrawVLine(x0+x, y0-y, y+1)
		t.DrawVLine(x0+y, y0-x, x+1)
	}

	// Upper left
	if t.shouldDraw(options, DRAW_TOP_LEFT) {
		t.DrawVLine(x0-x, y0-y, y+1)
		t.DrawVLine(x0-y, y0-x, x+1)
	}

	// Lower right
	if t.shouldDraw(options, DRAW_BOTTOM_RIGHT) {
		t.DrawVLine(x0+x, y0, y+1)
		t.DrawVLine(x0+y, y0, x+1)
	}

	// Lower left
	if t.shouldDraw(options, DRAW_BOTTOM_LEFT) {
		t.DrawVLine(x0-x, y0, y+1)
		t.DrawVLine(x0-y, y0, x+1)
	}
}
