package t8go

import "github.com/redghc/t8go/helpers"

// IDisplay represents a generic display interface that all display drivers must implement.
// It provides low-level operations for drawing pixels and managing the display buffer.
type IDisplay interface {
	Size() (width, height uint16) // Size returns the display dimensions in pixels
	BufferSize() int              // BufferSize returns the size in bytes of the display buffer
	Buffer() []byte               // Buffer returns the underlying display buffer

	ClearBuffer()                           // ClearBuffer clears the display buffer without updating the physical display
	ClearDisplay()                          // ClearDisplay clears both buffer and physical display
	Command(cmd byte) error                 // Command sends a command byte to the display
	Display() error                         // Display sends the current buffer to the physical display
	DisplayRegion(x0, y0, x1, y1 int) error // DisplayRegion updates a rectangular region of the physical display
	SetPixel(x, y int16, on bool)           // SetPixel sets a pixel at (x, y) to on/off
	GetPixel(x, y uint8) bool               // GetPixel returns the state of a pixel at (x, y)

	FillHRun(x0, x1, y int16, on bool)      // FillHRun sets or clears a horizontal run of pixels from x0 to x1 inclusive at row y
	FillRect(x0, y0, x1, y1 int16, on bool) // FillRect sets or clears a rectangular region with inclusive bounds
	FillVRun(x, y0, y1 int16, on bool)      // FillVRun sets or clears a vertical run of pixels from y0 to y1 inclusive at column x
}

// ----------

// IDisplayDrawer is the high-level graphics surface: display management plus
// shape primitives (lines, boxes, triangles, polygons, circles, ellipses, arcs).
// Coordinates use int16 so negative (offscreen) values are expressible;
// pixel queries use uint8 since embedded displays rarely exceed 255 px on a side.
type IDisplayDrawer interface {
	GetDisplay() IDisplay
	Size() (width, height uint16) // Size returns the display dimensions in pixels
	BufferSize() int
	Buffer() []byte
	ClearBuffer()
	ClearDisplay()
	Command(cmd byte) error
	Display() error
	DisplayAll() error
	SetPixel(x, y int16, on bool)
	GetPixel(x, y uint8) bool

	DrawPixel(x, y int16)

	DrawLine(startX, startY, endX, endY int16)
	DrawVLine(originX, originY, length int16)
	DrawHLine(originX, originY, length int16)
	DrawLineAngle(originX, originY, length int16, angle uint8)

	DrawBox(originX, originY, width, height int16)
	DrawBoxCoords(startX, startY, endX, endY int16)
	DrawRoundBox(originX, originY, width, height, cornerRadius int16)
	DrawBoxFill(originX, originY, width, height int16)
	DrawBoxFillCoords(startX, startY, endX, endY int16)
	DrawRoundBoxFill(originX, originY, width, height, cornerRadius int16)

	DrawTriangle(x1, y1, x2, y2, x3, y3 int16)
	DrawTriangleFill(x1, y1, x2, y2, x3, y3 int16)

	DrawPolygon(points ...Point)
	DrawPolygonFill(points ...Point)

	DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants)
	DrawCircleFill(centerX, centerY, radius int16, mask DrawQuadrants)

	DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)
	DrawEllipseFill(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)

	DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8)
	DrawArcFill(centerX, centerY, radius int16, angleStart, angleEnd uint8)
}

// T8Go is the main graphics context that provides high-level drawing operations.
// It wraps a Display interface and provides methods for drawing various shapes
// such as lines, rectangles, circles, and other geometric primitives.
type T8Go struct {
	display    IDisplay   // The underlying display interface
	width      int16      // Display width in pixels (cached)
	height     int16      // Display height in pixels (cached)
	spans      []scanSpan // Reusable per-row scanline accumulator for arc/triangle fills
	colExtents []scanSpan // Reusable per-column vertical extent for filled circles/ellipses
	crossings  []int16    // Reusable x-intersection buffer for polygon fill
	runRows    []int16    // Reusable per-row run storage for polygon boundary
	dirty      dirtyRect  // Dirty region tracked for partial display updates
}

var _ IDisplayDrawer = (*T8Go)(nil) // Ensure T8Go implements DisplayDrawer

// dirtyRect tracks the bounding box of pixels modified since the last flush.
// When valid is false, nothing has been drawn and Display can be skipped.
type dirtyRect struct {
	valid          bool
	x0, y0, x1, y1 int16
}

// reset marks the dirty region as empty (nothing pending to flush).
func (d *dirtyRect) reset() {
	d.valid = false
}

// ----------

// Point represents a pixel coordinate in the display space.
// X and Y are 16-bit signed to allow negative (offscreen) coordinates.
type Point struct {
	X, Y int16
}

// ----------

// DrawQuadrants represents which quadrants of a circle or ellipse should be drawn.
// It uses bitwise flags to specify combinations of quadrants.
type DrawQuadrants uint8

const (
	DrawTopLeft     DrawQuadrants = 1 << 0
	DrawTopRight    DrawQuadrants = 1 << 1
	DrawBottomRight DrawQuadrants = 1 << 2
	DrawBottomLeft  DrawQuadrants = 1 << 3
	DrawAll         DrawQuadrants = DrawTopLeft | DrawTopRight | DrawBottomRight | DrawBottomLeft
)

// has checks if the given quadrant should be drawn.
// The zero value (no bits set) selects no quadrant, drawing nothing.
// Use DrawAll (or the explicit bitwise OR of every flag) to render the
// complete circle or ellipse.
func (mask DrawQuadrants) has(quadrant DrawQuadrants) bool {
	return mask&quadrant != 0
}

// ----------

// scanSpan stores the min/max X coordinates to fill for a given scan-line Y.
// This is used internally for filled shape rendering.
type scanSpan struct {
	minX        int16 // Minimum X coordinate for this scan-line
	maxX        int16 // Maximum X coordinate for this scan-line
	initialized bool  // Whether this span has been initialized with values
}

// AddPoint expands the span to include the given X coordinate.
// If this is the first point, it initializes the span with that coordinate.
func (s *scanSpan) AddPoint(x int16) {
	if !s.initialized {
		s.minX, s.maxX, s.initialized = x, x, true
		return
	}
	if x < s.minX {
		s.minX = x
	}
	if x > s.maxX {
		s.maxX = x
	}
}

// IsEmpty reports whether the span has no points (is uninitialized).
func (s *scanSpan) IsEmpty() bool {
	return !s.initialized
}

// arcAccum tracks perimeter points closest to start/end angles for arc rendering.
// This is used internally to find the optimal endpoints when drawing arcs.
type arcAccum struct {
	bestStartAngleDiff uint8 // Smallest angle difference found for start angle
	bestEndAngleDiff   uint8 // Smallest angle difference found for end angle
	startEndX          int16 // X coordinate of best start angle match
	startEndY          int16 // Y coordinate of best start angle match
	endEndX            int16 // X coordinate of best end angle match
	endEndY            int16 // Y coordinate of best end angle match
}

// arcProcessPerimeter samples 8-way symmetric perimeter points, filters by angle range,
// widens spans, and updates endpoints closest to angleStart/angleEnd.
func (accum *arcAccum) arcProcessPerimeter(
	spans []scanSpan,
	centerX, centerY, offsetX, offsetY int16,
	angleStart, angleEnd uint8,
) {
	baseOctantAngle := helpers.ApproxAtanUnit64(offsetX, offsetY) // 0..64

	candidates := [8]struct {
		ang uint8
		x   int16
		y   int16
	}{
		{baseOctantAngle, centerX + offsetY, centerY - offsetX},                                          // a0
		{helpers.AngleQuarterTurn - baseOctantAngle, centerX + offsetX, centerY - offsetY},               // a1
		{helpers.AngleQuarterTurn + baseOctantAngle, centerX - offsetX, centerY - offsetY},               // a2
		{helpers.AngleHalfTurn - baseOctantAngle, centerX - offsetY, centerY - offsetX},                  // a3
		{helpers.AngleHalfTurn + baseOctantAngle, centerX - offsetY, centerY + offsetX},                  // a4
		{helpers.AngleThreeQuarterTurn - baseOctantAngle, centerX - offsetX, centerY + offsetY},          // a5
		{helpers.AngleThreeQuarterTurn + baseOctantAngle, centerX + offsetX, centerY + offsetY},          // a6
		{uint8(helpers.AngleFullCircle - uint16(baseOctantAngle)), centerX + offsetY, centerY + offsetX}, // a7
	}

	for _, c := range candidates {
		if helpers.InAngleRange(c.ang, angleStart, angleEnd) {
			addSpan(spans, c.x, c.y)

			if d := helpers.ArcAngleDistance(c.ang, angleStart); d < accum.bestStartAngleDiff {
				accum.bestStartAngleDiff, accum.startEndX, accum.startEndY = d, c.x, c.y
			}
			if d := helpers.ArcAngleDistance(c.ang, angleEnd); d < accum.bestEndAngleDiff {
				accum.bestEndAngleDiff, accum.endEndX, accum.endEndY = d, c.x, c.y
			}
		}
	}
}
