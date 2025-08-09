package t8go

import "github.com/redghc/t8go/helpers"

// DrawPixel sets a pixel at the specified coordinates (x, y) in the display buffer.
func (t *T8Go) DrawPixel(x, y int16) {
	t.SetPixel(x, y, true)
}

// DrawLine draws a line between (startX, startY) and (endX, endY) using Bresenham's algorithm.
// The result is rendered into the display buffer.
func (t *T8Go) DrawLine(startX, startY, endX, endY int16) {
	// Fast paths: vertical and horizontal
	if startX == endX {
		y0, y1 := startY, endY
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		t.DrawVLine(startX, y0, y1-y0+1)
		return
	}

	if startY == endY {
		x0, x1 := startX, endX
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		t.DrawHLine(x0, startY, x1-x0+1)
		return
	}

	// Determine if the line is steep (more vertical than horizontal).
	swapXY := helpers.AbsDiff(endY, startY) > helpers.AbsDiff(endX, startX)
	if swapXY {
		helpers.SwapInt16(&startX, &startY)
		helpers.SwapInt16(&endX, &endY)
	}

	// Ensure left-to-right progression.
	if startX > endX {
		helpers.SwapInt16(&startX, &endX)
		helpers.SwapInt16(&startY, &endY)
	}

	deltaX := endX - startX
	deltaY := endY - startY
	stepY := int16(1)
	if deltaY < 0 {
		stepY = -1
		deltaY = -deltaY
	}

	errorAccumulator := deltaX / 2
	currentY := startY

	for currentX := startX; currentX <= endX; currentX++ {
		if swapXY {
			t.SetPixel(currentY, currentX, true)
		} else {
			t.SetPixel(currentX, currentY, true)
		}

		errorAccumulator -= deltaY
		if errorAccumulator < 0 {
			currentY += stepY
			errorAccumulator += deltaX
		}
	}
}

// DrawVLine draws a vertical line starting at (originX, originY) with the given length.
// Length is the number of pixels; origin is included. No-op if length <= 0.
func (t *T8Go) DrawVLine(originX, originY, length int16) {
	if length <= 0 {
		return
	}

	for deltaY := range length {
		t.SetPixel(originX, originY+deltaY, true)
	}
}

// DrawHLine draws a horizontal line starting at (startX, startY) with the given length.
// Length is the number of pixels; origin is included. No-op if length <= 0.
func (t *T8Go) DrawHLine(startX, startY, length int16) {
	if length <= 0 {
		return
	}

	for deltaX := range length {
		t.SetPixel(startX+deltaX, startY, true)
	}
}

// DrawBox draws a filled rectangle starting from the top-left corner (originX, originY)
// with the specified dimensions: width and height. No-op if width <= 0 or height <= 0.
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
	originX, originY, width, height := helpers.NormalizeRect(startX, startY, endX, endY)
	t.DrawBox(originX, originY, width, height)
}

// DrawFrame draws a rectangular outline starting from the top-left corner (originX, originY)
// with the specified width and height. Must be at least 2x2 to form a valid frame.
func (t *T8Go) DrawFrame(originX, originY, width, height int16) {
	if width <= 1 || height <= 1 {
		return
	}

	maxX := originX + width - 1
	maxY := originY + height - 1

	// Top and bottom horizontal edges (full width).
	t.DrawHLine(originX, originY, width)
	t.DrawHLine(originX, maxY, width)

	// Left and right vertical edges (excluding corners already drawn).
	t.DrawVLine(originX, originY+1, height-2)
	t.DrawVLine(maxX, originY+1, height-2)
}

// DrawFrameCoords draws a rectangular outline between two corners: top-left (startX, startY)
// and bottom-right (endX, endY), inclusive.
func (t *T8Go) DrawFrameCoords(startX, startY, endX, endY int16) {
	originX, originY, width, height := helpers.NormalizeRect(startX, startY, endX, endY)
	t.DrawFrame(originX, originY, width, height)
}

// DrawCircle draws an outlined circle centered at (centerX, centerY) with the given radius.
// The diameter of the circle is 2*radius + 1.
// The mask parameter determines which quadrants of the circle will be drawn.
// If mask is DrawNone, the entire circle is rendered.
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants) {
	if radius <= 0 {
		return
	}

	// Midpoint circle algorithm using integer arithmetic for performance.
	decision := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	t.drawCircleSection(offsetX, offsetY, centerX, centerY, mask)

	for offsetX < offsetY {
		if decision >= 0 {
			offsetY--
			deltaY += 2
			decision += deltaY
		}
		offsetX++
		deltaX += 2
		decision += deltaX

		t.drawCircleSection(offsetX, offsetY, centerX, centerY, mask)
	}
}

// drawCircleSection plots the symmetric points of the circle for the given offsets,
// filtered by the mask to draw only the selected quadrants.
func (t *T8Go) drawCircleSection(offsetX, offsetY, centerX, centerY int16, mask DrawQuadrants) {
	if mask.has(DrawTopRight) {
		t.DrawPixel(centerX+offsetX, centerY-offsetY)
		t.DrawPixel(centerX+offsetY, centerY-offsetX)
	}
	if mask.has(DrawTopLeft) {
		t.DrawPixel(centerX-offsetX, centerY-offsetY)
		t.DrawPixel(centerX-offsetY, centerY-offsetX)
	}
	if mask.has(DrawBottomRight) {
		t.DrawPixel(centerX+offsetX, centerY+offsetY)
		t.DrawPixel(centerX+offsetY, centerY+offsetX)
	}
	if mask.has(DrawBottomLeft) {
		t.DrawPixel(centerX-offsetX, centerY+offsetY)
		t.DrawPixel(centerX-offsetY, centerY+offsetX)
	}
}

// DrawArc draws an outlined arc centered at (centerX, centerY) with the given radius.
// Angles are expressed in 0..255 units where:
//
//	0   =   0° (right)
//	64  =  90° (up)
//	128 = 180° (left)
//	192 = 270° (down)
//	255 = 360° (wrap to 0)
//
// The arc is rendered from angleStart (inclusive) to angleEnd (exclusive). If angleStart == angleEnd,
// a full circle is drawn.
func (t *T8Go) DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8) {
	if radius <= 0 {
		return
	}

	isFullArc := angleStart == angleEnd

	// Midpoint circle algorithm (integer arithmetic).
	decision := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	// Plot first set of symmetric points.
	t.drawArcSection(offsetX, offsetY, centerX, centerY, angleStart, angleEnd, isFullArc)

	for offsetX < offsetY {
		if decision >= 0 {
			offsetY--
			deltaY += 2
			decision += deltaY
		}
		offsetX++
		deltaX += 2
		decision += deltaX

		t.drawArcSection(offsetX, offsetY, centerX, centerY, angleStart, angleEnd, isFullArc)
	}
}

// drawArcSection plots the 8 symmetric points for the given offsets if their angles
// fall within the requested range.
func (t *T8Go) drawArcSection(offsetX, offsetY, centerX, centerY int16, angleStart, angleEnd uint8, isFullArc bool) {
	// Base angle in the first octant [0..64], derived from atan(offsetX/offsetY) in integer arithmetic.
	// This avoids floating-point and uses a tuned polynomial/rational approximation.
	baseOctantAngle := helpers.ApproxAtanUnit64(offsetX, offsetY) // 0..64

	// Compose the 8 angles (0..255) corresponding to the 8 symmetric points.
	// Note: we keep comparisons in uint8 space and rely on inAngleRange for wrap handling.
	a0 := baseOctantAngle                      // ( +y, -x )   → 0..64
	a1 := 64 - baseOctantAngle                 // ( +x, -y )   → 0..64
	a2 := 64 + baseOctantAngle                 // ( -x, -y )   → 64..128
	a3 := 128 - baseOctantAngle                // ( -y, -x )   → 64..128
	a4 := 128 + baseOctantAngle                // ( -y, +x )   → 128..192
	a5 := 192 - baseOctantAngle                // ( -x, +y )   → 128..192
	a6 := 192 + baseOctantAngle                // ( +x, +y )   → 192..256
	a7 := uint8(256 - uint16(baseOctantAngle)) // ( +y, +x )   → wraps to 0..256

	if isFullArc || helpers.InAngleRange(a0, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetY, centerY-offsetX)
	}
	if isFullArc || helpers.InAngleRange(a1, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetX, centerY-offsetY)
	}
	if isFullArc || helpers.InAngleRange(a2, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetX, centerY-offsetY)
	}
	if isFullArc || helpers.InAngleRange(a3, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetY, centerY-offsetX)
	}
	if isFullArc || helpers.InAngleRange(a4, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetY, centerY+offsetX)
	}
	if isFullArc || helpers.InAngleRange(a5, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetX, centerY+offsetY)
	}
	if isFullArc || helpers.InAngleRange(a6, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetX, centerY+offsetY)
	}
	if isFullArc || helpers.InAngleRange(a7, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetY, centerY+offsetX)
	}
}

// DrawDisc draws a filled circle (disc) centered at (centerX, centerY) with the given radius.
// The mask parameter selects which quadrants are filled. If mask is DrawNone, the entire disc is filled.
func (t *T8Go) DrawDisc(centerX, centerY, radius int16, mask DrawQuadrants) {
	if radius <= 0 {
		return
	}

	// Midpoint circle algorithm (integer arithmetic).
	decision := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	t.drawDiscSection(offsetX, offsetY, centerX, centerY, mask)

	for offsetX < offsetY {
		if decision >= 0 {
			offsetY--
			deltaY += 2
			decision += deltaY
		}
		offsetX++
		deltaX += 2
		decision += deltaX

		t.drawDiscSection(offsetX, offsetY, centerX, centerY, mask)
	}
}

// drawDiscSection draws the vertical spans for the 8-way symmetric points of a circle,
// filtered by the mask to fill only the selected quadrants.
func (t *T8Go) drawDiscSection(offsetX, offsetY, centerX, centerY int16, mask DrawQuadrants) {
	if mask.has(DrawTopRight) {
		t.DrawVLine(centerX+offsetX, centerY-offsetY, offsetY+1)
		t.DrawVLine(centerX+offsetY, centerY-offsetX, offsetX+1)
	}
	if mask.has(DrawTopLeft) {
		t.DrawVLine(centerX-offsetX, centerY-offsetY, offsetY+1)
		t.DrawVLine(centerX-offsetY, centerY-offsetX, offsetX+1)
	}
	if mask.has(DrawBottomRight) {
		t.DrawVLine(centerX+offsetX, centerY, offsetY+1)
		t.DrawVLine(centerX+offsetY, centerY, offsetX+1)
	}
	if mask.has(DrawBottomLeft) {
		t.DrawVLine(centerX-offsetX, centerY, offsetY+1)
		t.DrawVLine(centerX-offsetY, centerY, offsetX+1)
	}
}

// DrawEllipse draws an outlined ellipse centered at (centerX, centerY) with the given radii.
// The mask parameter determines which quadrants of the ellipse are drawn.
// If mask is DrawNone, the entire ellipse is rendered.
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants) {
	if radiusX == 0 && radiusY == 0 {
		return
	}

	if radiusX < 0 {
		radiusX = -radiusX
	}
	if radiusY < 0 {
		radiusY = -radiusY
	}

	// Pre-calculate squared and doubled radius values for optimization.
	radiusXSq := int32(radiusX) * int32(radiusX)
	radiusYSq := int32(radiusY) * int32(radiusY)
	radiusXSquaredDouble := radiusXSq * 2
	radiusYSquaredDouble := radiusYSq * 2

	// Region 1: horizontal-dominant (|dx/dy| < 1)
	ellipseX := radiusX
	ellipseY := int16(0)

	// Error terms
	deltaXChange := (1 - 2*int32(radiusX)) * radiusYSq
	deltaYChange := radiusXSq
	errorAccumulator := int32(0)

	// Stop conditions for region 1
	stopX := radiusYSquaredDouble * int32(radiusX)
	stopY := int32(0)

	for stopX >= stopY {
		t.drawEllipseSection(ellipseX, ellipseY, centerX, centerY, mask)

		ellipseY++
		stopY += radiusXSquaredDouble
		errorAccumulator += deltaYChange
		deltaYChange += radiusXSquaredDouble

		if 2*errorAccumulator+deltaXChange > 0 {
			ellipseX--
			stopX -= radiusYSquaredDouble
			errorAccumulator += deltaXChange
			deltaXChange += radiusYSquaredDouble
		}
	}

	// Region 2: vertical-dominant (|dx/dy| >= 1)
	ellipseX = 0
	ellipseY = radiusY

	// Recompute error terms for region 2
	deltaXChange = radiusYSq
	deltaYChange = (1 - 2*int32(radiusY)) * radiusXSq
	errorAccumulator = 0

	stopX = 0
	stopY = radiusXSquaredDouble * int32(radiusY)

	for stopX <= stopY {
		t.drawEllipseSection(ellipseX, ellipseY, centerX, centerY, mask)

		ellipseX++
		stopX += radiusYSquaredDouble
		errorAccumulator += deltaXChange
		deltaXChange += radiusYSquaredDouble

		if 2*errorAccumulator+deltaYChange > 0 {
			ellipseY--
			stopY -= radiusXSquaredDouble
			errorAccumulator += deltaYChange
			deltaYChange += radiusXSquaredDouble
		}
	}
}

// drawEllipseSection plots the symmetric points of an ellipse for the given offsets,
// filtered by the mask to draw only the selected quadrants.
func (t *T8Go) drawEllipseSection(offsetX, offsetY, centerX, centerY int16, mask DrawQuadrants) {
	if mask.has(DrawTopRight) {
		t.DrawPixel(centerX+offsetX, centerY-offsetY)
	}
	if mask.has(DrawTopLeft) {
		t.DrawPixel(centerX-offsetX, centerY-offsetY)
	}
	if mask.has(DrawBottomRight) {
		t.DrawPixel(centerX+offsetX, centerY+offsetY)
	}
	if mask.has(DrawBottomLeft) {
		t.DrawPixel(centerX-offsetX, centerY+offsetY)
	}
}
