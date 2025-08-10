package t8go

import "github.com/redghc/t8go/helpers"

// DrawPixel sets a pixel at the specified coordinates (x, y) in the display buffer.
func (t *T8Go) DrawPixel(x, y int16) {
	t.SetPixel(x, y, true)
}

// DrawLine draws a line between (startX, startY) and (endX, endY) using Bresenham's algorithm.
// Origin pixels are included
func (t *T8Go) DrawLine(startX, startY, endX, endY int16) {
	// Fast paths: vertical and horizontal lines
	if startX == endX {
		startYPos, endYPos := startY, endY
		if startYPos > endYPos {
			startYPos, endYPos = endYPos, startYPos
		}
		t.DrawVLine(startX, startYPos, endYPos-startYPos+1)
		return
	}

	if startY == endY {
		startXPos, endXPos := startX, endX
		if startXPos > endXPos {
			startXPos, endXPos = endXPos, startXPos
		}
		t.DrawHLine(startXPos, startY, endXPos-startXPos+1)
		return
	}

	// Determine if the line is steep (more vertical than horizontal).
	isSteep := helpers.AbsDiff(endY, startY) > helpers.AbsDiff(endX, startX)
	if isSteep {
		startX, startY = startY, startX
		endX, endY = endY, endX
	}

	// Ensure left-to-right progression.
	if startX > endX {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	deltaX := endX - startX
	deltaY := helpers.Abs(endY - startY)
	stepDirectionY := helpers.Direction(endY - startY)

	errorAccumulator := deltaX / 2
	currentYPos := startY

	for currentXPos := startX; currentXPos <= endX; currentXPos++ {
		if isSteep {
			t.SetPixel(currentYPos, currentXPos, true)
		} else {
			t.SetPixel(currentXPos, currentYPos, true)
		}

		errorAccumulator -= deltaY
		if errorAccumulator < 0 {
			currentYPos += stepDirectionY
			errorAccumulator += deltaX
		}
	}
}

// DrawVLine draws a vertical line starting at (originX, originY) with the given length.
// Length is the number of pixels; the origin pixel is included.
// Supports negative length (draws upward). No-op if length == 0.
func (t *T8Go) DrawVLine(originX, originY, length int16) {
	direction := helpers.Direction(length)
	if direction == 0 {
		return
	}

	uLength := helpers.Abs(length)
	for deltaY := range uLength {
		t.SetPixel(originX, originY+deltaY*direction, true)
	}
}

// DrawHLine draws a horizontal line starting at (originX, originY) with the given length.
// Length is the number of pixels; the origin pixel is included.
// Supports negative length (draws to the left). No-op if length == 0.
func (t *T8Go) DrawHLine(originX, originY, length int16) {
	direction := helpers.Direction(length)
	if direction == 0 {
		return
	}

	uLength := helpers.Abs(length)
	for deltaX := range uLength {
		t.SetPixel(originX+deltaX*direction, originY, true)
	}
}

// DrawLineAngle draws a line from (originX, originY) with the given length (origin included)
// and angle in 0..255 units. Quality matches Bresenham by delegating to DrawLine.
func (t *T8Go) DrawLineAngle(originX, originY, length int16, angle uint8) {
	if length == 0 {
		return
	}
	endX, endY := helpers.AngleEndpoint(originX, originY, length, angle)
	t.DrawLine(originX, originY, endX, endY)
}

// DrawBox draws a rectangular outline starting from the top-left corner (originX, originY) with the specified width and height.
// Supports negative width/height to draw in the opposite direction.
// Must be at least 2x2 in absolute size to form a valid frame.
// Origin pixel is included.
func (t *T8Go) DrawBox(originX, originY, width, height int16) {
	directionX := helpers.Direction(width)
	directionY := helpers.Direction(height)

	uWidth := helpers.Abs(width)
	uHeight := helpers.Abs(height)

	// Need at least 2 pixels in each dimension to form a proper outline
	if uWidth <= 1 || uHeight <= 1 {
		return
	}

	// Calculate far corner coordinates based on direction
	maxX := originX + (uWidth-1)*directionX
	maxY := originY + (uHeight-1)*directionY

	// Top and bottom horizontal edges
	t.DrawHLine(originX, originY, width)
	t.DrawHLine(originX, maxY, width)

	// Left and right vertical edges (excluding the corners already drawn)
	t.DrawVLine(originX, originY+directionY, height-2*directionY)
	t.DrawVLine(maxX, originY+directionY, height-2*directionY)
}

// DrawBoxCoords draws a rectangular outline between two corners:
// top-left (startX, startY) and bottom-right (endX, endY), inclusive.
// The order of coordinates does not matter; they are normalized internally.
func (t *T8Go) DrawBoxCoords(startX, startY, endX, endY int16) {
	originX, originY, width, height := helpers.NormalizeRect(startX, startY, endX, endY)
	t.DrawBox(originX, originY, width, height)
}

// DrawRoundBox draws a rectangle outline with rounded corners.
// Corner radius is clamped to fit within width/height.
func (t *T8Go) DrawRoundBox(originX, originY, width, height, cornerRadius int16) {
	uWidth := helpers.Abs(width)
	uHeight := helpers.Abs(height)
	if uWidth <= 1 || uHeight <= 1 {
		return
	}
	if cornerRadius < 0 {
		cornerRadius = 0
	}

	// Fast path: if cornerRadius == 0, just draw a box.
	if cornerRadius == 0 {
		t.DrawBox(originX, originY, width, height)
		return
	}

	// Clamp so there is at least 1px of straight edge per side.
	limit := min(uWidth, uHeight)
	maxCorner := (limit - 1) / 2
	cornerRadius = min(cornerRadius, maxCorner)

	// Normalize bounds.
	rawMaxX := originX + width - 1
	rawMaxY := originY + height - 1
	minX, maxX := min(originX, rawMaxX), max(originX, rawMaxX)
	minY, maxY := min(originY, rawMaxY), max(originY, rawMaxY)

	hLen := (maxX - minX + 1) - 2*cornerRadius
	vLen := (maxY - minY + 1) - 2*cornerRadius

	// Straight edges.
	t.DrawHLine(minX+cornerRadius, minY, hLen)
	t.DrawHLine(minX+cornerRadius, maxY, hLen)
	t.DrawVLine(minX, minY+cornerRadius, vLen)
	t.DrawVLine(maxX, minY+cornerRadius, vLen)

	// Rounded corners.
	t.DrawCircle(minX+cornerRadius, minY+cornerRadius, cornerRadius, DrawTopLeft)
	t.DrawCircle(maxX-cornerRadius, minY+cornerRadius, cornerRadius, DrawTopRight)
	t.DrawCircle(maxX-cornerRadius, maxY-cornerRadius, cornerRadius, DrawBottomRight)
	t.DrawCircle(minX+cornerRadius, maxY-cornerRadius, cornerRadius, DrawBottomLeft)
}

// DrawBoxFill draws a filled rectangle starting from the top-left corner (originX, originY)
// with the specified dimensions: width and height.
// Supports negative width/height to draw in the opposite direction.
// Origin pixel is included. No-op if width == 0 or height == 0.
func (t *T8Go) DrawBoxFill(originX, originY, width, height int16) {
	directionY := helpers.Direction(height)
	directionX := helpers.Direction(width)

	if directionX == 0 || directionY == 0 {
		return
	}

	uHeight := helpers.Abs(height)

	for offsetY := range uHeight {
		t.DrawHLine(
			originX,
			originY+offsetY*directionY,
			width,
		)
	}
}

// DrawBoxFillCoords draws a filled rectangle between two corners:
// top-left (startX, startY) and bottom-right (endX, endY), inclusive.
// The order of coordinates does not matter; they are normalized internally.
func (t *T8Go) DrawBoxFillCoords(startX, startY, endX, endY int16) {
	originX, originY, width, height := helpers.NormalizeRect(startX, startY, endX, endY)
	t.DrawBoxFill(originX, originY, width, height)
}

// DrawRoundBoxFill draws a filled rectangle with rounded corners.
// Corner radius is clamped to fit within width/height.
func (t *T8Go) DrawRoundBoxFill(originX, originY, width, height, cornerRadius int16) {
	uWidth := helpers.Abs(width)
	uHeight := helpers.Abs(height)
	if uWidth <= 0 || uHeight <= 0 {
		return
	}
	if cornerRadius < 0 {
		cornerRadius = 0
	}

	// Fast path: if cornerRadius == 0, just draw a box filled.
	if cornerRadius == 0 {
		t.DrawBoxFill(originX, originY, width, height)
		return
	}

	// Clamp so there is at least 1px of straight edge per side.
	limit := min(uWidth, uHeight)
	maxCorner := (limit - 1) / 2
	cornerRadius = min(cornerRadius, maxCorner)

	// Normalize bounds.
	rawMaxX := originX + width - 1
	rawMaxY := originY + height - 1
	minX, maxX := min(originX, rawMaxX), max(originX, rawMaxX)
	minY, maxY := min(originY, rawMaxY), max(originY, rawMaxY)

	// Middle slabs: draw both orientations to handle capsules in either axis.
	centerWidth := (maxX - minX + 1) - 2*cornerRadius  // vertical middle slab width
	centerHeight := (maxY - minY + 1) - 2*cornerRadius // horizontal middle slab height

	// Vertical middle slab (covers tall shapes and general case).
	if centerWidth > 0 {
		t.DrawBoxFill(minX+cornerRadius, minY, centerWidth, uHeight)
	}

	// Horizontal middle slab (covers wide shapes and general case).
	if centerHeight > 0 {
		t.DrawBoxFill(minX, minY+cornerRadius, uWidth, centerHeight)
	}

	// Rounded corners (quarter-disc fills).
	t.DrawCircleFill(minX+cornerRadius, minY+cornerRadius, cornerRadius, DrawTopLeft)
	t.DrawCircleFill(maxX-cornerRadius, minY+cornerRadius, cornerRadius, DrawTopRight)
	t.DrawCircleFill(maxX-cornerRadius, maxY-cornerRadius, cornerRadius, DrawBottomRight)
	t.DrawCircleFill(minX+cornerRadius, maxY-cornerRadius, cornerRadius, DrawBottomLeft)
}

// DrawTriangle draws the outline of a triangle connecting points (x1,y1), (x2,y2), (x3,y3).
func (t *T8Go) DrawTriangle(x1, y1, x2, y2, x3, y3 int16) {
	t.DrawLine(x1, y1, x2, y2)
	t.DrawLine(x2, y2, x3, y3)
	t.DrawLine(x3, y3, x1, y1)
}

// DrawTriangleFill draws a filled triangle connecting points (x1,y1), (x2,y2), (x3,y3).
// The edges are inclusive, ensuring no gaps.
func (t *T8Go) DrawTriangleFill(x1, y1, x2, y2, x3, y3 int16) {
	t.DrawTriangle(x1, y1, x2, y2, x3, y3)

	// Degenerate horizontal line (all y equal)
	if y1 == y2 && y2 == y3 {
		left := min(x1, min(x2, x3))
		right := max(x1, max(x2, x3))
		t.DrawHLine(left, y1, right-left+1)
		return
	}

	minY := min(y1, min(y2, y3))
	maxY := max(y1, max(y2, y3))

	// Accumulate edge pixels into spans keyed by Y (map value's zero state has initialized=false).
	spans := make(map[int16]scanSpan, int(maxY-minY)+1)
	scanAddLineToSpans(spans, x1, y1, x2, y2)
	scanAddLineToSpans(spans, x2, y2, x3, y3)
	scanAddLineToSpans(spans, x3, y3, x1, y1)

	for y := minY; y <= maxY; y++ {
		row := spans[y]
		if !row.initialized {
			continue
		}
		startXPos, endXPos := row.minX, row.maxX
		if startXPos > endXPos {
			startXPos, endXPos = endXPos, startXPos
		}
		t.DrawHLine(startXPos, y, endXPos-startXPos+1)
	}
}

// DrawCircle draws an outlined circle centered at (centerX, centerY) with the given radius.
// The diameter of the circle is 2*radius + 1.
// The mask parameter determines which quadrants of the circle will be drawn.
// If mask is DrawNone, the entire circle is rendered.
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants) {
	if radius <= 0 {
		return
	}

	// Midpoint circle algorithm with integer arithmetic.
	errorAccumulator := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	t.drawCircleSection(offsetX, offsetY, centerX, centerY, mask)

	for offsetX < offsetY {
		if errorAccumulator >= 0 {
			offsetY--
			deltaY += 2
			errorAccumulator += deltaY
		}
		offsetX++
		deltaX += 2
		errorAccumulator += deltaX

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

// DrawCircleFill draws a filled circle centered at (centerX, centerY) with the given radius.
// The mask parameter selects which quadrants are filled. If mask is DrawNone, the entire disc is filled.
func (t *T8Go) DrawCircleFill(centerX, centerY, radius int16, mask DrawQuadrants) {
	if radius <= 0 {
		return
	}

	// Midpoint circle algorithm with integer arithmetic.
	errorAccumulator := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	t.drawCircleFillSection(offsetX, offsetY, centerX, centerY, mask)

	for offsetX < offsetY {
		if errorAccumulator >= 0 {
			offsetY--
			deltaY += 2
			errorAccumulator += deltaY
		}
		offsetX++
		deltaX += 2
		errorAccumulator += deltaX

		t.drawCircleFillSection(offsetX, offsetY, centerX, centerY, mask)
	}
}

// drawCircleFillSection draws the vertical spans for the 8-way symmetric points of a circle,
// filtered by the mask to fill only the selected quadrants.
func (t *T8Go) drawCircleFillSection(offsetX, offsetY, centerX, centerY int16, mask DrawQuadrants) {
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

// DrawEllipse draws an outlined ellipse centered at (centerX, centerY) with radiusX and radiusY.
// The mask parameter determines which quadrants will be drawn.
// If mask is DrawNone, the entire ellipse outline is rendered.
// No-op if radiusX <= 0 or radiusY <= 0.
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants) {
	if radiusX <= 0 || radiusY <= 0 {
		return
	}

	// Use int32 internally to avoid overflow on products.
	rx := int32(radiusX)
	ry := int32(radiusY)
	rx2 := rx * rx
	ry2 := ry * ry
	rx2x2 := rx2 * 2
	ry2x2 := ry2 * 2

	// Region 1 (|dy/dx| < 1)
	offsetX := rx
	offsetY := int32(0)
	errorAccumulator := (1 - 2*rx) * ry2
	deltaX := rx2
	deltaY := rx2
	stopX := ry2x2 * rx
	stopY := int32(0)

	for stopX >= stopY {
		t.drawEllipseSection(int16(offsetX), int16(offsetY), centerX, centerY, mask)

		offsetY++
		stopY += rx2x2
		errorAccumulator += deltaY
		deltaY += rx2x2

		if 2*errorAccumulator+deltaX > 0 {
			offsetX--
			stopX -= ry2x2
			errorAccumulator += (1 - 2*offsetX) * ry2
			deltaX += ry2x2
		}
	}

	// Region 2 (|dy/dx| >= 1)
	offsetX = 0
	offsetY = ry
	errorAccumulator = (1 - 2*ry) * rx2
	deltaX = ry2
	deltaY = rx2
	stopX = 0
	stopY = rx2x2 * ry

	for stopX <= stopY {
		t.drawEllipseSection(int16(offsetX), int16(offsetY), centerX, centerY, mask)

		offsetX++
		stopX += ry2x2
		errorAccumulator += deltaX
		deltaX += ry2x2

		if 2*errorAccumulator+deltaY > 0 {
			offsetY--
			stopY -= rx2x2
			errorAccumulator += (1 - 2*offsetY) * rx2
			deltaY += rx2x2
		}
	}
}

// drawEllipseSection plots the symmetric points of an ellipse for the given offsets,
// filtered by the mask to draw only the selected quadrants.
func (t *T8Go) drawEllipseSection(offsetX, offsetY, centerX, centerY int16, mask DrawQuadrants) {
	if offsetY < 0 {
		return
	}
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

// DrawEllipseFill draws a filled ellipse centered at (centerX, centerY) with radiusX and radiusY.
// The mask parameter selects which quadrants are filled. If mask is DrawNone, the entire area is filled.
// No-op if radiusX <= 0 or radiusY <= 0.
func (t *T8Go) DrawEllipseFill(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants) {
	if radiusX <= 0 || radiusY <= 0 {
		return
	}

	rx := int32(radiusX)
	ry := int32(radiusY)
	rx2 := rx * rx   // rx^2
	ry2 := ry * ry   // ry^2
	rx2x2 := rx2 * 2 // 2*rx^2
	ry2x2 := ry2 * 2 // 2*ry^2

	// Region 1 (horizontal-dominant): stopX >= stopY
	offsetX := int32(radiusX)
	offsetY := int32(0)

	deltaXChange := (1 - 2*int32(radiusX)) * ry2
	deltaYChange := rx2
	errorAccumulator := int32(0)

	stopX := ry2x2 * int32(radiusX)
	stopY := int32(0)

	for stopX >= stopY {
		t.drawEllipseFillSection(int16(offsetX), int16(offsetY), centerX, centerY, mask)

		offsetY++
		stopY += rx2x2
		errorAccumulator += deltaYChange
		deltaYChange += rx2x2

		if 2*errorAccumulator+deltaXChange > 0 {
			offsetX--
			stopX -= ry2x2
			errorAccumulator += deltaXChange
			deltaXChange += ry2x2
		}
	}

	// Region 2 (vertical-dominant): stopX <= stopY
	offsetX = 0
	offsetY = int32(radiusY)

	deltaXChange = ry2
	deltaYChange = (1 - 2*int32(radiusY)) * rx2
	errorAccumulator = 0

	stopX = 0
	stopY = rx2x2 * int32(radiusY)

	for stopX <= stopY {
		t.drawEllipseFillSection(int16(offsetX), int16(offsetY), centerX, centerY, mask)

		offsetX++
		stopX += ry2x2
		errorAccumulator += deltaXChange
		deltaXChange += ry2x2

		if 2*errorAccumulator+deltaYChange > 0 {
			offsetY--
			stopY -= rx2x2
			errorAccumulator += deltaYChange
			deltaYChange += rx2x2
		}
	}
}

// drawEllipseFillSection draws vertical spans for the 4-way symmetric points of an ellipse,
// filtered by the mask to fill only the selected quadrants.
// For each (offsetX, offsetY), spans have length offsetY+1.
func (t *T8Go) drawEllipseFillSection(offsetX, offsetY, centerX, centerY int16, mask DrawQuadrants) {
	if offsetY < 0 {
		return
	}

	// Upper quadrants: start at y0 - offsetY, length offsetY+1
	if mask.has(DrawTopRight) {
		t.DrawVLine(centerX+offsetX, centerY-offsetY, offsetY+1)
	}
	if mask.has(DrawTopLeft) {
		t.DrawVLine(centerX-offsetX, centerY-offsetY, offsetY+1)
	}

	// Lower quadrants: start at y0, length offsetY+1
	if mask.has(DrawBottomRight) {
		t.DrawVLine(centerX+offsetX, centerY, offsetY+1)
	}
	if mask.has(DrawBottomLeft) {
		t.DrawVLine(centerX-offsetX, centerY, offsetY+1)
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
// The arc is rendered from angleStart (inclusive) to angleEnd (exclusive).
// If angleStart == angleEnd, a full circle is drawn.
func (t *T8Go) DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8) {
	if radius <= 0 {
		return
	}

	// Fast path: full arc
	isFullArc := angleStart == angleEnd
	if isFullArc {
		t.DrawCircle(centerX, centerY, radius, DrawAll)
		return
	}

	// Midpoint circle algorithm (integer arithmetic).
	errorAccumulator := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	t.drawArcSection(offsetX, offsetY, centerX, centerY, angleStart, angleEnd)

	for offsetX < offsetY {
		if errorAccumulator >= 0 {
			offsetY--
			deltaY += 2
			errorAccumulator += deltaY
		}
		offsetX++
		deltaX += 2
		errorAccumulator += deltaX

		t.drawArcSection(offsetX, offsetY, centerX, centerY, angleStart, angleEnd)
	}
}

// drawArcSection plots the 8 symmetric points for the given offsets if their angles
// fall within the requested range. Uses 0..255 unit angles (64 = 90°).
func (t *T8Go) drawArcSection(offsetX, offsetY, centerX, centerY int16, angleStart, angleEnd uint8) {
	// Base angle in the first octant [0..64], approximated using integer math.
	// The helper is assumed to be monotonic and tuned for small integer inputs.
	baseOctantAngle := helpers.ApproxAtanUnit64(offsetX, offsetY) // uint8 in [0..64]

	// Compose angles (0..255) for the 8-way symmetry. All done in uint8 space.
	a0 := baseOctantAngle                      // (+y, -x) →   0.. 64
	a1 := 64 - baseOctantAngle                 // (+x, -y) →   0.. 64
	a2 := 64 + baseOctantAngle                 // (-x, -y) →  64..128
	a3 := 128 - baseOctantAngle                // (-y, -x) →  64..128
	a4 := 128 + baseOctantAngle                // (-y, +x) → 128..192
	a5 := 192 - baseOctantAngle                // (-x, +y) → 128..192
	a6 := 192 + baseOctantAngle                // (+x, +y) → 192..256
	a7 := uint8(256 - uint16(baseOctantAngle)) // (+y, +x) → wrap to 0..255

	// Only plot points whose angle falls inside [angleStart, angleEnd).
	// If the caller asked for a full arc, this function is not invoked (fast-path above).
	if helpers.InAngleRange(a0, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetY, centerY-offsetX)
	}
	if helpers.InAngleRange(a1, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetX, centerY-offsetY)
	}
	if helpers.InAngleRange(a2, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetX, centerY-offsetY)
	}
	if helpers.InAngleRange(a3, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetY, centerY-offsetX)
	}
	if helpers.InAngleRange(a4, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetY, centerY+offsetX)
	}
	if helpers.InAngleRange(a5, angleStart, angleEnd) {
		t.DrawPixel(centerX-offsetX, centerY+offsetY)
	}
	if helpers.InAngleRange(a6, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetX, centerY+offsetY)
	}
	if helpers.InAngleRange(a7, angleStart, angleEnd) {
		t.DrawPixel(centerX+offsetY, centerY+offsetX)
	}
}

// DrawArcFill draws a filled arc (sector) centered at (centerX, centerY) with the given radius.
// Angles are expressed in 0..255 units where 64 = 90°, 128 = 180°, 192 = 270°.
// The arc is rendered from angleStart (inclusive) to angleEnd (exclusive).
// If angleStart == angleEnd, a full disc is filled.
func (t *T8Go) DrawArcFill(centerX, centerY, radius int16, angleStart, angleEnd uint8) {
	if radius <= 0 {
		return
	}

	// Fast path: full sector -> full circle fill
	if angleStart == angleEnd {
		t.DrawCircleFill(centerX, centerY, radius, DrawAll)
		return
	}

	// Perimeter sampling (midpoint circle) → accumulate spans and arc endpoints.
	spans := make(map[int16]scanSpan, int(radius)*2+3)
	accum := arcAccum{bestStartAngleDiff: 255, bestEndAngleDiff: 255}

	errorAccumulator := int16(1 - radius)
	deltaX := int16(1)
	deltaY := int16(-2 * radius)
	offsetX := int16(0)
	offsetY := radius

	accum.arcProcessPerimeter(spans, centerX, centerY, offsetX, offsetY, angleStart, angleEnd)

	for offsetX < offsetY {
		if errorAccumulator >= 0 {
			offsetY--
			deltaY += 2
			errorAccumulator += deltaY
		}
		offsetX++
		deltaX += 2
		errorAccumulator += deltaX

		accum.arcProcessPerimeter(spans, centerX, centerY, offsetX, offsetY, angleStart, angleEnd)
	}

	// Add radial boundaries (center → endpoints) into spans.
	updateSpan(spans, centerX, centerY)
	scanAddLineToSpans(spans, centerX, centerY, accum.startEndX, accum.startEndY)
	scanAddLineToSpans(spans, centerX, centerY, accum.endEndX, accum.endEndY)

	// Paint consolidated spans.
	for yPos, row := range spans {
		if row.IsEmpty() {
			continue
		}
		startXPos, endXPos := row.minX, row.maxX
		if startXPos > endXPos {
			startXPos, endXPos = endXPos, startXPos
		}
		length := endXPos - startXPos + 1
		t.DrawHLine(startXPos, yPos, length)
	}
}

// updateSpan widens the span at (yPos) to include xPos.
func updateSpan(spans map[int16]scanSpan, xPos, yPos int16) {
	row := spans[yPos]
	row.AddPoint(xPos)
	spans[yPos] = row
}

// scanAddLineToSpans rasterizes a line into spans using Bresenham rules (no clipping).
func scanAddLineToSpans(spans map[int16]scanSpan, x0, y0, x1, y1 int16) {
	// Vertical
	if x0 == x1 {
		startYPos, endYPos := y0, y1
		if startYPos > endYPos {
			startYPos, endYPos = endYPos, startYPos
		}
		for currentYPos := startYPos; currentYPos <= endYPos; currentYPos++ {
			updateSpan(spans, x0, currentYPos)
		}
		return
	}
	// Horizontal
	if y0 == y1 {
		startXPos, endXPos := x0, x1
		if startXPos > endXPos {
			startXPos, endXPos = endXPos, startXPos
		}
		for currentXPos := startXPos; currentXPos <= endXPos; currentXPos++ {
			updateSpan(spans, currentXPos, y0)
		}
		return
	}

	// General Bresenham (mirrors your DrawLine semantics).
	steep := helpers.AbsDiff(y1, y0) > helpers.AbsDiff(x1, x0)
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	deltaX := x1 - x0
	deltaY := helpers.Abs(y1 - y0)
	stepDirectionY := helpers.Direction(y1 - y0)

	errorAccumulator := deltaX / 2
	currentYPos := y0

	for currentXPos := x0; currentXPos <= x1; currentXPos++ {
		if steep {
			updateSpan(spans, currentYPos, currentXPos)
		} else {
			updateSpan(spans, currentXPos, currentYPos)
		}
		errorAccumulator -= deltaY
		if errorAccumulator < 0 {
			currentYPos += stepDirectionY
			errorAccumulator += deltaX
		}
	}
}
