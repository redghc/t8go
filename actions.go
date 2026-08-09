package t8go

import (
	"slices"

	"github.com/redghc/t8go/helpers"
)

// DrawPixel sets a pixel at the specified coordinates (x, y) in the display buffer.
// This is the most basic drawing primitive - a single point on the display.
func (t *T8Go) DrawPixel(x, y int16) {
	t.SetPixel(x, y, true)
}

// DrawLine draws a line between two points (startX, startY) and (endX, endY)
// using Bresenham's line algorithm for optimal pixel-perfect rendering.
// Both origin and destination pixels are included in the line.
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

	// The line's pixels stay within the bounding box of its endpoints.
	t.markDirty(startX, startY)
	t.markDirty(endX, endY)

	for currentXPos := startX; currentXPos <= endX; currentXPos++ {
		if isSteep {
			t.setPixelOn(currentYPos, currentXPos)
		} else {
			t.setPixelOn(currentXPos, currentYPos)
		}

		errorAccumulator -= deltaY
		if errorAccumulator < 0 {
			currentYPos += stepDirectionY
			errorAccumulator += deltaX
		}
	}
}

// DrawVLine draws a vertical line starting at (originX, originY) with the specified length.
// The length parameter specifies the number of pixels to draw, including the origin pixel.
// Supports negative length values (draws upward). No operation is performed if length is zero.
func (t *T8Go) DrawVLine(originX, originY, length int16) {
	direction := helpers.Direction(length)
	if direction == 0 {
		return
	}

	uLength := helpers.Abs(length)
	y0 := originY
	y1 := originY + (uLength-1)*direction
	if y0 > y1 {
		y0, y1 = y1, y0
	}

	// Clip to the visible area before writing.
	if originX < 0 || originX >= t.width {
		return
	}
	y0, y1, ok := helpers.ClipSpan(y0, y1, t.height)
	if !ok {
		return
	}

	t.markDirty(originX, y0)
	t.markDirty(originX, y1)
	t.display.FillVRun(originX, y0, y1, true)
}

// DrawHLine draws a horizontal line starting at (originX, originY) with the specified length.
// The length parameter specifies the number of pixels to draw, including the origin pixel.
// Supports negative length values (draws to the left). No operation is performed if length is zero.
func (t *T8Go) DrawHLine(originX, originY, length int16) {
	direction := helpers.Direction(length)
	if direction == 0 {
		return
	}

	uLength := helpers.Abs(length)
	x0 := originX
	x1 := originX + (uLength-1)*direction
	if x0 > x1 {
		x0, x1 = x1, x0
	}

	// Clip to the visible area before writing.
	if originY < 0 || originY >= t.height {
		return
	}
	x0, x1, ok := helpers.ClipSpan(x0, x1, t.width)
	if !ok {
		return
	}

	t.markDirty(x0, originY)
	t.markDirty(x1, originY)
	t.display.FillHRun(x0, x1, originY, true)
}

// DrawLineAngle draws a line from (originX, originY) with the specified length and angle.
// The angle is specified in units of 0-255, where 64=90°, 128=180°, 192=270°.
// The length includes the origin pixel. Quality matches Bresenham's algorithm by delegating to DrawLine.
func (t *T8Go) DrawLineAngle(originX, originY, length int16, angle uint8) {
	if length == 0 {
		return
	}
	endX, endY := helpers.AngleEndpoint(originX, originY, length, angle)
	t.DrawLine(originX, originY, endX, endY)
}

// DrawBox draws a rectangular outline starting from (originX, originY) with specified dimensions.
// The width and height parameters define the size of the rectangle.
// Supports negative width/height values to draw in the opposite direction.
// Must be at least 2x2 in absolute size to form a valid frame outline.
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

// DrawRoundBox draws a rectangular outline with rounded corners.
// The cornerRadius parameter controls the curvature of the corners.
// Corner radius is automatically clamped to fit within the rectangle dimensions.
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

// DrawBoxFill draws a filled rectangle starting from (originX, originY) with specified dimensions.
// The width and height parameters define the size of the filled rectangle.
// Supports negative width/height values to draw in the opposite direction.
// No operation is performed if width or height is zero.
func (t *T8Go) DrawBoxFill(originX, originY, width, height int16) {
	directionY := helpers.Direction(height)
	directionX := helpers.Direction(width)

	if directionX == 0 || directionY == 0 {
		return
	}

	uWidth := helpers.Abs(width)
	uHeight := helpers.Abs(height)

	x0 := originX
	x1 := originX + (uWidth-1)*directionX
	y0 := originY
	y1 := originY + (uHeight-1)*directionY

	t.fillRect(x0, y0, x1, y1)
}

// fillRect sets the inclusive rectangle (x0,y0)-(x1,y1), clipped to the
// visible area, and marks it dirty. Coordinates are normalized internally.
func (t *T8Go) fillRect(x0, y0, x1, y1 int16) {
	x0, y0, x1, y1, ok := helpers.ClipRect(x0, y0, x1, y1, t.width, t.height)
	if !ok {
		return
	}

	t.markDirty(x0, y0)
	t.markDirty(x1, y1)
	t.display.FillRect(x0, y0, x1, y1, true)
}

// DrawBoxFillCoords draws a filled rectangle between two corners:
// top-left (startX, startY) and bottom-right (endX, endY), inclusive.
// The order of coordinates does not matter; they are normalized internally.
func (t *T8Go) DrawBoxFillCoords(startX, startY, endX, endY int16) {
	originX, originY, width, height := helpers.NormalizeRect(startX, startY, endX, endY)
	t.DrawBoxFill(originX, originY, width, height)
}

// DrawRoundBoxFill draws a filled rectangle with rounded corners.
// The cornerRadius parameter controls the curvature of the corners.
// Corner radius is automatically clamped to fit within the rectangle dimensions.
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

// DrawTriangle draws the outline of a triangle connecting three points.
// The triangle is drawn by connecting (x1,y1) to (x2,y2) to (x3,y3) and back to (x1,y1).
func (t *T8Go) DrawTriangle(x1, y1, x2, y2, x3, y3 int16) {
	t.DrawLine(x1, y1, x2, y2)
	t.DrawLine(x2, y2, x3, y3)
	t.DrawLine(x3, y3, x1, y1)
}

// DrawTriangleFill draws a filled triangle connecting three points.
// The triangle is filled using scanline rendering to ensure complete coverage
// with inclusive edges and no gaps.
func (t *T8Go) DrawTriangleFill(x1, y1, x2, y2, x3, y3 int16) {
	// Degenerate horizontal line (all y equal)
	if y1 == y2 && y2 == y3 {
		left := min(x1, min(x2, x3))
		right := max(x1, max(x2, x3))
		t.DrawHLine(left, y1, right-left+1)
		return
	}

	// Accumulate edge pixels into scanline spans (reused scratch buffer).
	// The edge spans already cover the outline pixels, so no separate
	// outline pass is needed.
	clear(t.scratch)
	scanAddLineToSpans(t.scratch, x1, y1, x2, y2)
	scanAddLineToSpans(t.scratch, x2, y2, x3, y3)
	scanAddLineToSpans(t.scratch, x3, y3, x1, y1)

	t.paintSpans(t.scratch)
}

// minPolygonVertices is the minimum number of vertices a polygon (outline or
// fill) needs to enclose an area.
const minPolygonVertices = 3

// DrawPolygon draws the boundary of the polygon connecting the given vertices
// in order. The outline is closed automatically: the last vertex connects back
// to the first one.
// At least three vertices are required to enclose an area; with fewer the call
// is a no-op. The outline is exactly the boundary of the region produced by
// DrawPolygonFill, so both functions always draw equivalent shapes.
func (t *T8Go) DrawPolygon(points ...Point) {
	if len(points) < minPolygonVertices {
		return
	}

	yMin, yMax := polygonBoundsY(points)
	rowCount := int(yMax - yMin + 1)

	// Compute the filled runs for every row so the boundary pass can check the
	// coverage of the rows above and below.
	offsets := make([]int, rowCount+1)
	buffer := t.polyBuf[:0]
	for i := range rowCount {
		offsets[i] = len(buffer)
		buffer = append(buffer, t.polyRuns(points, yMin+int16(i))...)
	}
	offsets[rowCount] = len(buffer)
	t.polyBuf = buffer

	for i := range rowCount {
		curr := buffer[offsets[i]:offsets[i+1]]
		var prev, next []int16
		if i > 0 {
			prev = buffer[offsets[i-1]:offsets[i]]
		}
		if i+1 < rowCount {
			next = buffer[offsets[i+1]:offsets[i+2]]
		}
		t.drawPolyBoundary(prev, curr, next, yMin+int16(i))
	}
}

// DrawPolygonFill draws a filled polygon connecting the given vertices in order.
// The polygon is closed automatically: the last vertex connects back to the
// first one.
// At least three vertices are required to enclose an area; with fewer the call
// is a no-op. Filling uses the even-odd scanline rule with pixel-center
// sampling: for each row the edge crossings are collected, sorted and filled
// pairwise, so concave and self-intersecting polygons are rendered correctly.
func (t *T8Go) DrawPolygonFill(points ...Point) {
	if len(points) < minPolygonVertices {
		return
	}

	yMin, yMax := polygonBoundsY(points)
	for y := yMin; y <= yMax; y++ {
		runs := t.polyRuns(points, y)
		for i := 0; i+1 < len(runs); i += 2 {
			t.DrawHLine(runs[i], y, runs[i+1]-runs[i]+1)
		}
	}
}

// polyRuns returns the filled runs of the polygon on scanline y as sorted
// [start, end] pairs, using even-odd pixel-center crossings. The returned
// slice aliases t.polyXS and is overwritten on the next call.
func (t *T8Go) polyRuns(points []Point, y int16) []int16 {
	xs := t.polyXS
	if cap(xs) < len(points) {
		xs = make([]int16, 0, len(points))
	}
	xs = xs[:0]

	count := len(points)
	for i := range count {
		p1, p2 := points[i], points[(i+1)%count]
		if x, ok := polyCrossingStart(p1, p2, y); ok {
			xs = append(xs, x)
		}
	}
	slices.Sort(xs)

	runs := xs[:0]
	for i := 0; i+1 < len(xs); i += 2 {
		start, end := xs[i], xs[i+1]-1
		if end >= start {
			runs = append(runs, start, end)
		}
	}
	t.polyXS = runs
	return runs
}

// drawPolyBoundary draws the boundary pixels of the filled runs in curr, using
// prev and next (the filled runs of the adjacent scanlines) to determine which
// interior pixels are exposed on their top or bottom edge.
func (t *T8Go) drawPolyBoundary(prev, curr, next []int16, y int16) {
	for i := 0; i+1 < len(curr); i += 2 {
		start, end := curr[i], curr[i+1]
		t.DrawPixel(start, y)
		t.DrawPixel(end, y)
		if end > start {
			t.drawPolyExposed(start+1, end-1, prev, y)
			t.drawPolyExposed(start+1, end-1, next, y)
		}
	}
}

// drawPolyExposed draws the pixels of [lo, hi] that are not covered by any of
// the given runs, i.e. the pixels exposed on the top or bottom edge of the
// filled region.
func (t *T8Go) drawPolyExposed(lo, hi int16, runs []int16, y int16) {
	x := lo
	for i := 0; i+1 < len(runs) && x <= hi; i += 2 {
		start, end := runs[i], runs[i+1]
		if end < x {
			continue
		}
		if start > hi {
			break
		}
		if start > x {
			t.DrawHLine(x, y, start-x)
		}
		if end+1 > x {
			x = end + 1
		}
	}
	if x <= hi {
		t.DrawHLine(x, y, hi-x+1)
	}
}

// polygonBoundsY returns the minimum and maximum Y coordinates of the given
// vertices, as `yMin` and `yMax`.
func polygonBoundsY(points []Point) (int16, int16) {
	yMin, yMax := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		if p.Y < yMin {
			yMin = p.Y
		}
		if p.Y > yMax {
			yMax = p.Y
		}
	}
	return yMin, yMax
}

// scanPixelScale is the scale factor used for pixel-center scanline sampling.
const scanPixelScale = 2

// polyCrossingStart returns `x`, the coordinate of the first pixel whose center
// lies to the right of the intersection of the edge p1-p2 with the center of
// scanline y, and `ok`, whether the edge crosses that scanline. Horizontal
// edges never cross. The half-open interval [lo, hi) guarantees that a vertex
// shared by two edges is counted exactly once.
func polyCrossingStart(p1, p2 Point, y int16) (int16, bool) {
	x1, y1 := p1.X, p1.Y
	x2, y2 := p2.X, p2.Y
	if y1 == y2 {
		return 0, false
	}

	lo, hi := y1, y2
	xLo, xHi := x1, x2
	if lo > hi {
		lo, hi = hi, lo
		xLo, xHi = xHi, xLo
	}
	if y < lo || y >= hi {
		return 0, false
	}

	// Crossing at the pixel center y+0.5, expressed in integer math:
	//   x = xLo + (2*(y-lo)+1) * (xHi-xLo) / (2*(hi-lo))
	// The first filled pixel is the one whose center is above the crossing:
	//   floor(x - 0.5) + 1 = xLo + floor((2*num-den)/(2*den)) + 1
	dx := int32(xHi - xLo)
	dy := int32(hi - lo)
	num := (scanPixelScale*int64(y-lo) + 1) * int64(dx)
	den := scanPixelScale * int64(dy)
	pixel := int64(xLo) + floorDiv(scanPixelScale*num-den, scanPixelScale*den) + 1
	// The result is bounded by the vertex coordinates (display-size), so the
	// narrowing cast cannot overflow in practice.
	return int16(pixel), true
}

// floorDiv divides a by b rounding toward negative infinity (b must be > 0).
func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

// DrawCircle draws an outlined circle centered at (centerX, centerY) with the specified radius.
// The circle diameter will be 2*radius + 1 pixels.
// The mask parameter controls which quadrants are drawn using DrawQuadrants flags.
// Use DrawNone or DrawAll to draw the complete circle.
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants) {
	if radius <= 0 {
		return
	}

	t.markDirty(centerX-radius, centerY-radius)
	t.markDirty(centerX+radius, centerY+radius)

	// Midpoint circle algorithm with integer arithmetic.
	errorAccumulator := 1 - radius
	deltaX := int16(1)
	deltaY := -2 * radius
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
		t.setPixelOn(centerX+offsetX, centerY-offsetY)
		t.setPixelOn(centerX+offsetY, centerY-offsetX)
	}
	if mask.has(DrawTopLeft) {
		t.setPixelOn(centerX-offsetX, centerY-offsetY)
		t.setPixelOn(centerX-offsetY, centerY-offsetX)
	}
	if mask.has(DrawBottomRight) {
		t.setPixelOn(centerX+offsetX, centerY+offsetY)
		t.setPixelOn(centerX+offsetY, centerY+offsetX)
	}
	if mask.has(DrawBottomLeft) {
		t.setPixelOn(centerX-offsetX, centerY+offsetY)
		t.setPixelOn(centerX-offsetY, centerY+offsetX)
	}
}

// DrawCircleFill draws a filled circle centered at (centerX, centerY) with the specified radius.
// The mask parameter controls which quadrants are filled using DrawQuadrants flags.
// Use DrawNone or DrawAll to fill the complete circle disc.
func (t *T8Go) DrawCircleFill(centerX, centerY, radius int16, mask DrawQuadrants) {
	if radius <= 0 {
		return
	}

	// Full disc fast path: compute each column's vertical extent once and
	// paint it with a page-aligned FillVRun, instead of 8 overlapping
	// vertical runs per perimeter step.
	if (mask == DrawNone || mask == DrawAll) && radius < int16(len(t.colScratch)) {
		t.fillCircle(centerX, centerY, radius)
		return
	}

	// Midpoint circle algorithm with integer arithmetic.
	errorAccumulator := 1 - radius
	deltaX := int16(1)
	deltaY := -2 * radius
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

// fillCircle fills a complete disc. The midpoint perimeter is walked once to
// record the vertical half-extent of each column, then every column is painted
// exactly once.
func (t *T8Go) fillCircle(centerX, centerY, radius int16) {
	clear(t.colScratch)

	errorAccumulator := 1 - radius
	deltaX := int16(1)
	deltaY := -2 * radius
	offsetX := int16(0)
	offsetY := radius

	t.recordCircleExtent(offsetX, offsetY)
	for offsetX < offsetY {
		if errorAccumulator >= 0 {
			offsetY--
			deltaY += 2
			errorAccumulator += deltaY
		}
		offsetX++
		deltaX += 2
		errorAccumulator += deltaX

		t.recordCircleExtent(offsetX, offsetY)
	}

	t.paintColumnExtents(centerX, centerY, radius)
}

// recordCircleExtent stores the disc's vertical half-extent for the columns at
// offsets offsetX and offsetY.
func (t *T8Go) recordCircleExtent(offsetX, offsetY int16) {
	t.setColExtent(offsetX, offsetY)
	t.setColExtent(offsetY, offsetX)
}

// setColExtent records (or grows) the vertical half-extent for column offset.
func (t *T8Go) setColExtent(offset, extent int16) {
	if offset < 0 || int(offset) >= len(t.colScratch) {
		return
	}
	row := t.colScratch[offset]
	row.AddPoint(extent)
	t.colScratch[offset] = row
}

// paintColumnExtents fills one vertical run per recorded column offset.
func (t *T8Go) paintColumnExtents(centerX, centerY, radius int16) {
	t.markDirty(centerX-radius, centerY-radius)
	t.markDirty(centerX+radius, centerY+radius)

	for offsetX := int16(0); offsetX <= radius; offsetX++ {
		row := t.colScratch[offsetX]
		if row.IsEmpty() {
			continue
		}
		half := row.maxX
		y0 := centerY - half
		height := 2*half + 1
		if offsetX == 0 {
			t.DrawVLine(centerX, y0, height)
		} else {
			t.DrawVLine(centerX+offsetX, y0, height)
			t.DrawVLine(centerX-offsetX, y0, height)
		}
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

// DrawEllipse draws an outlined ellipse centered at (centerX, centerY) with specified radii.
// The radiusX and radiusY parameters define the horizontal and vertical extents.
// The mask parameter controls which quadrants are drawn using DrawQuadrants flags.
// Use DrawNone or DrawAll to draw the complete ellipse outline.
// No operation is performed if either radius is less than or equal to zero.
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants) {
	if radiusX <= 0 || radiusY <= 0 {
		return
	}

	t.markDirty(centerX-radiusX, centerY-radiusY)
	t.markDirty(centerX+radiusX, centerY+radiusY)

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
		t.setPixelOn(centerX+offsetX, centerY-offsetY)
	}
	if mask.has(DrawTopLeft) {
		t.setPixelOn(centerX-offsetX, centerY-offsetY)
	}
	if mask.has(DrawBottomRight) {
		t.setPixelOn(centerX+offsetX, centerY+offsetY)
	}
	if mask.has(DrawBottomLeft) {
		t.setPixelOn(centerX-offsetX, centerY+offsetY)
	}
}

// DrawEllipseFill draws a filled ellipse centered at (centerX, centerY) with specified radii.
// The radiusX and radiusY parameters define the horizontal and vertical extents.
// The mask parameter controls which quadrants are filled using DrawQuadrants flags.
// Use DrawNone or DrawAll to fill the complete ellipse area.
// No operation is performed if either radius is less than or equal to zero.
func (t *T8Go) DrawEllipseFill(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants) {
	if radiusX <= 0 || radiusY <= 0 {
		return
	}

	// Full ellipse fast path: compute each column's vertical extent once and
	// paint it with a page-aligned FillVRun.
	if (mask == DrawNone || mask == DrawAll) && radiusX < int16(len(t.colScratch)) {
		t.fillEllipse(centerX, centerY, radiusX, radiusY)
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

// fillEllipse fills a complete ellipse. The midpoint perimeter (both regions)
// is walked once to record the vertical half-extent of each column, then every
// column is painted exactly once.
func (t *T8Go) fillEllipse(centerX, centerY, radiusX, radiusY int16) {
	clear(t.colScratch)

	rx := int32(radiusX)
	ry := int32(radiusY)
	rx2 := rx * rx
	ry2 := ry * ry
	rx2x2 := rx2 * 2
	ry2x2 := ry2 * 2

	// Region 1 (horizontal-dominant).
	offsetX := int32(radiusX)
	offsetY := int32(0)
	deltaXChange := (1 - 2*int32(radiusX)) * ry2
	deltaYChange := rx2
	errorAccumulator := int32(0)
	stopX := ry2x2 * int32(radiusX)
	stopY := int32(0)

	for stopX >= stopY {
		t.recordEllipseExtent(int16(offsetX), int16(offsetY))

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

	// Region 2 (vertical-dominant).
	offsetX = 0
	offsetY = int32(radiusY)
	deltaXChange = ry2
	deltaYChange = (1 - 2*int32(radiusY)) * rx2
	errorAccumulator = 0
	stopX = 0
	stopY = rx2x2 * int32(radiusY)

	for stopX <= stopY {
		t.recordEllipseExtent(int16(offsetX), int16(offsetY))

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

	t.paintColumnExtents(centerX, centerY, radiusX)
}

// recordEllipseExtent stores the ellipse's vertical half-extent for the column
// at offset offsetX. Unlike the circle, the fill only covers the cx±offsetX
// column family, so no symmetric offsetY record is needed.
func (t *T8Go) recordEllipseExtent(offsetX, offsetY int16) {
	t.setColExtent(offsetX, offsetY)
}

// DrawArc draws an outlined arc (partial circle) centered at (centerX, centerY) with the specified radius.
// Angles are expressed in 0-255 units where 0=0°, 64=90°, 128=180°, 192=270°.
// The arc is rendered from angleStart (inclusive) to angleEnd (exclusive).
// If angleStart equals angleEnd, a complete circle is drawn.
//
// Angle reference:
//   - 0   =   0° (right)
//   - 64  =  90° (up)
//   - 128 = 180° (left)
//   - 192 = 270° (down)
//   - 255 = 360° (wraps to 0)
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
	errorAccumulator := 1 - radius
	deltaX := int16(1)
	deltaY := -2 * radius
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
	a0 := baseOctantAngle                                          // (+y, -x) →   0.. 64
	a1 := helpers.AngleQuarterTurn - baseOctantAngle               // (+x, -y) →   0.. 64
	a2 := helpers.AngleQuarterTurn + baseOctantAngle               // (-x, -y) →  64..128
	a3 := helpers.AngleHalfTurn - baseOctantAngle                  // (-y, -x) →  64..128
	a4 := helpers.AngleHalfTurn + baseOctantAngle                  // (-y, +x) → 128..192
	a5 := helpers.AngleThreeQuarterTurn - baseOctantAngle          // (-x, +y) → 128..192
	a6 := helpers.AngleThreeQuarterTurn + baseOctantAngle          // (+x, +y) → 192..256
	a7 := uint8(helpers.AngleFullCircle - uint16(baseOctantAngle)) // (+y, +x) → wrap to 0..255

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

// DrawArcFill draws a filled arc (sector/pie slice) centered at (centerX, centerY) with the specified radius.
// Angles are expressed in 0-255 units where 64=90°, 128=180°, 192=270°.
// The arc is rendered from angleStart (inclusive) to angleEnd (exclusive).
// If angleStart equals angleEnd, a complete filled circle is drawn.
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
	clear(t.scratch)
	accum := arcAccum{bestStartAngleDiff: helpers.AngleFullCircle / 2, bestEndAngleDiff: helpers.AngleFullCircle / 2}

	errorAccumulator := 1 - radius
	deltaX := int16(1)
	deltaY := -2 * radius
	offsetX := int16(0)
	offsetY := radius

	accum.arcProcessPerimeter(t.scratch, centerX, centerY, offsetX, offsetY, angleStart, angleEnd)

	for offsetX < offsetY {
		if errorAccumulator >= 0 {
			offsetY--
			deltaY += 2
			errorAccumulator += deltaY
		}
		offsetX++
		deltaX += 2
		errorAccumulator += deltaX

		accum.arcProcessPerimeter(t.scratch, centerX, centerY, offsetX, offsetY, angleStart, angleEnd)
	}

	// Add radial boundaries (center → endpoints) into spans.
	addSpan(t.scratch, centerX, centerY)
	scanAddLineToSpans(t.scratch, centerX, centerY, accum.startEndX, accum.startEndY)
	scanAddLineToSpans(t.scratch, centerX, centerY, accum.endEndX, accum.endEndY)

	t.paintSpans(t.scratch)
}

// addSpan widens the span at row yPos to include xPos.
// Out-of-bounds rows are ignored.
func addSpan(spans []scanSpan, xPos, yPos int16) {
	if yPos < 0 || int(yPos) >= len(spans) {
		return
	}
	row := spans[yPos]
	row.AddPoint(xPos)
	spans[yPos] = row
}

// paintSpans fills every scanline span with a horizontal run, skipping empty rows.
func (t *T8Go) paintSpans(spans []scanSpan) {
	for yPos := range spans {
		row := spans[yPos]
		if row.IsEmpty() {
			continue
		}
		startXPos, endXPos := row.minX, row.maxX
		if startXPos > endXPos {
			startXPos, endXPos = endXPos, startXPos
		}
		t.DrawHLine(startXPos, int16(yPos), endXPos-startXPos+1)
	}
}

// scanAddLineToSpans rasterizes a line into spans using Bresenham rules.
func scanAddLineToSpans(spans []scanSpan, x0, y0, x1, y1 int16) {
	// Vertical
	if x0 == x1 {
		startYPos, endYPos := y0, y1
		if startYPos > endYPos {
			startYPos, endYPos = endYPos, startYPos
		}
		for currentYPos := startYPos; currentYPos <= endYPos; currentYPos++ {
			addSpan(spans, x0, currentYPos)
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
			addSpan(spans, currentXPos, y0)
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
			addSpan(spans, currentYPos, currentXPos)
		} else {
			addSpan(spans, currentXPos, currentYPos)
		}
		errorAccumulator -= deltaY
		if errorAccumulator < 0 {
			currentYPos += stepDirectionY
			errorAccumulator += deltaX
		}
	}
}
