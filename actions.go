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

// DrawArc draws an outlined arc centered at (centerX, centerY) with the given radius,
// starting from angleStart to angleEnd. Angles range from 0 to 255 where:
// 0   = 0° (right)
// 64  = 90° (up)
// 128 = 180° (left)
// 192 = 270° (down)
// 255 = 360° (full circle)
func (t *T8Go) DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8) {
	isFullArc := (angleStart == angleEnd)
	isInvertedRange := (angleStart > angleEnd)

	var rangeStart, rangeEnd uint8
	if isInvertedRange {
		rangeStart = angleEnd
		rangeEnd = angleStart
	} else {
		rangeStart = angleStart
		rangeEnd = angleEnd
	}

	x := int16(0)
	y := radius
	decision := radius - 1

	for y >= x {
		var angleRatio uint32
		if y != 0 {
			angleRatio = uint32(x) * 255 / uint32(y)
			angleRatio = angleRatio * (770195 - (angleRatio-255)*(angleRatio+941)) / 6137491
		}

		if isFullArc || ((angleRatio >= uint32(rangeStart) && angleRatio < uint32(rangeEnd)) != isInvertedRange) {
			t.DrawPixel(centerX+y, centerY-x)
		}
		if isFullArc || (((angleRatio+uint32(rangeEnd)) > 63 && (angleRatio+uint32(rangeStart)) <= 63) != isInvertedRange) {
			t.DrawPixel(centerX+x, centerY-y)
		}
		if isFullArc || (((angleRatio+64) >= uint32(rangeStart) && (angleRatio+64) < uint32(rangeEnd)) != isInvertedRange) {
			t.DrawPixel(centerX-x, centerY-y)
		}
		if isFullArc || (((angleRatio+uint32(rangeEnd)) > 127 && (angleRatio+uint32(rangeStart)) <= 127) != isInvertedRange) {
			t.DrawPixel(centerX-y, centerY-x)
		}
		if isFullArc || (((angleRatio+128) >= uint32(rangeStart) && (angleRatio+128) < uint32(rangeEnd)) != isInvertedRange) {
			t.DrawPixel(centerX-y, centerY+x)
		}
		if isFullArc || (((angleRatio+uint32(rangeEnd)) > 191 && (angleRatio+uint32(rangeStart)) <= 191) != isInvertedRange) {
			t.DrawPixel(centerX-x, centerY+y)
		}
		if isFullArc || (((angleRatio+192) >= uint32(rangeStart) && (angleRatio+192) < uint32(rangeEnd)) != isInvertedRange) {
			t.DrawPixel(centerX+x, centerY+y)
		}
		if isFullArc || (((angleRatio+uint32(rangeEnd)) > 255 && (angleRatio+uint32(rangeStart)) <= 255) != isInvertedRange) {
			t.DrawPixel(centerX+y, centerY+x)
		}

		if decision >= 2*x {
			decision -= 2*x + 1
			x++
		} else if decision < 2*(radius-y) {
			decision += 2*y - 1
			y--
		} else {
			decision += 2 * (y - x - 1)
			y--
			x++
		}
	}
}

// DrawDisc draws a filled circle (disc) centered at (centerX, centerY) with the given radius.
// The quadrants parameter determines which parts of the disc are drawn.
// If quadrants is empty or includes DRAW_FULL, the entire disc is rendered.
func (t *T8Go) DrawDisc(centerX, centerY, radius int16, quadrants []DrawQuadrants) {
	decision := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	t.drawDiscSection(offsetX, offsetY, centerX, centerY, quadrants)

	for offsetX < offsetY {
		if decision >= 0 {
			offsetY--
			deltaY += 2
			decision += deltaY
		}
		offsetX++
		deltaX += 2
		decision += deltaX

		t.drawDiscSection(offsetX, offsetY, centerX, centerY, quadrants)
	}
}

func (t *T8Go) drawDiscSection(offsetX, offsetY, centerX, centerY int16, quadrants []DrawQuadrants) {
	if t.shouldDraw(quadrants, DRAW_TOP_RIGHT) {
		t.DrawVLine(centerX+offsetX, centerY-offsetY, offsetY+1)
		t.DrawVLine(centerX+offsetY, centerY-offsetX, offsetX+1)
	}
	if t.shouldDraw(quadrants, DRAW_TOP_LEFT) {
		t.DrawVLine(centerX-offsetX, centerY-offsetY, offsetY+1)
		t.DrawVLine(centerX-offsetY, centerY-offsetX, offsetX+1)
	}
	if t.shouldDraw(quadrants, DRAW_BOTTOM_RIGHT) {
		t.DrawVLine(centerX+offsetX, centerY, offsetY+1)
		t.DrawVLine(centerX+offsetY, centerY, offsetX+1)
	}
	if t.shouldDraw(quadrants, DRAW_BOTTOM_LEFT) {
		t.DrawVLine(centerX-offsetX, centerY, offsetY+1)
		t.DrawVLine(centerX-offsetY, centerY, offsetX+1)
	}
}

// DrawEllipse draws an outlined ellipse centered at (centerX, centerY) with the given radii.
// The quadrants parameter determines which sections of the ellipse are drawn.
// If quadrants is empty or includes DRAW_FULL, the entire ellipse is rendered.
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, quadrants []DrawQuadrants) {
	var x, y int16
	var xchange, ychange int32
	var err int32
	var rxrx2, ryry2 int32
	var stopx, stopy int32

	// Calculate squared and doubled radii
	rxrx2 = int32(radiusX)
	rxrx2 *= int32(radiusX)
	rxrx2 *= 2

	ryry2 = int32(radiusY)
	ryry2 *= int32(radiusY)
	ryry2 *= 2

	// First region: |dx/dy| < 1
	x = radiusX
	y = 0

	xchange = 1
	xchange -= int32(radiusX)
	xchange -= int32(radiusX)
	xchange *= int32(radiusY)
	xchange *= int32(radiusY)

	ychange = int32(radiusX)
	ychange *= int32(radiusX)

	err = 0

	stopx = ryry2
	stopx *= int32(radiusX)
	stopy = 0

	for stopx >= stopy {
		t.drawEllipseSection(x, y, centerX, centerY, quadrants)
		y++
		stopy += rxrx2
		err += ychange
		ychange += rxrx2
		if 2*err+xchange > 0 {
			x--
			stopx -= ryry2
			err += xchange
			xchange += ryry2
		}
	}

	// Second region: |dx/dy| >= 1
	x = 0
	y = radiusY

	xchange = int32(radiusY)
	xchange *= int32(radiusY)

	ychange = 1
	ychange -= int32(radiusY)
	ychange -= int32(radiusY)
	ychange *= int32(radiusX)
	ychange *= int32(radiusX)

	err = 0

	stopx = 0

	stopy = rxrx2
	stopy *= int32(radiusY)

	for stopx <= stopy {
		t.drawEllipseSection(x, y, centerX, centerY, quadrants)
		x++
		stopx += ryry2
		err += xchange
		xchange += ryry2
		if 2*err+ychange > 0 {
			y--
			stopy -= rxrx2
			err += ychange
			ychange += rxrx2
		}
	}
}

// drawEllipseSection draws the selected ellipse segments based on quadrant flags.
func (t *T8Go) drawEllipseSection(offsetX, offsetY, centerX, centerY int16, quadrants []DrawQuadrants) {
	if t.shouldDraw(quadrants, DRAW_TOP_RIGHT) {
		t.DrawPixel(centerX+offsetX, centerY-offsetY)
	}
	if t.shouldDraw(quadrants, DRAW_TOP_LEFT) {
		t.DrawPixel(centerX-offsetX, centerY-offsetY)
	}
	if t.shouldDraw(quadrants, DRAW_BOTTOM_RIGHT) {
		t.DrawPixel(centerX+offsetX, centerY+offsetY)
	}
	if t.shouldDraw(quadrants, DRAW_BOTTOM_LEFT) {
		t.DrawPixel(centerX-offsetX, centerY+offsetY)
	}
}
