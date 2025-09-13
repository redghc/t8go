package t8go

import "github.com/redghc/t8go/helpers"

// Display represents a generic display interface that all display drivers must implement.
// It provides low-level operations for drawing pixels and managing the display buffer.
type Display interface {
	Size() (width, height uint16) // Size returns the display dimensions in pixels
	BufferSize() int              // BufferSize returns the size in bytes of the display buffer
	Buffer() []byte               // Buffer returns the underlying display buffer

	ClearBuffer()                 // ClearBuffer clears the display buffer without updating the physical display
	ClearDisplay()                // ClearDisplay clears both buffer and physical display
	Command(cmd byte) error       // Command sends a command byte to the display
	Display() error               // Display sends the current buffer to the physical display
	SetPixel(x, y int16, on bool) // SetPixel sets a pixel at (x, y) to on/off
	GetPixel(x, y uint8) bool     // GetPixel returns the state of a pixel at (x, y)
}

// ----------

// DisplayDrawer provides a comprehensive interface for drawing operations on display devices.
// It combines basic display management with advanced drawing capabilities including geometric shapes,
// lines, and pixel manipulation. The interface supports both outlined and filled shapes, with
// specialized methods for circles, ellipses, arcs, and various box types.
//
// Basic Operations:
//   - Display management: get display info, buffer operations, clearing
//   - Pixel operations: set, get, and draw individual pixels
//   - Command execution and display updates
//
// Drawing Capabilities:
//   - Lines: straight lines, vertical/horizontal lines, angled lines
//   - Rectangles: basic boxes, rounded boxes, coordinate-based boxes (both outlined and filled)
//   - Triangles: outlined and filled triangles
//   - Circles: full or partial circles with quadrant masking (outlined and filled)
//   - Ellipses: full or partial ellipses with quadrant masking (outlined and filled)
//   - Arcs: circular arcs with start/end angles (outlined and filled)
//
// Coordinate System:
//   - Uses int16 for most coordinates to support negative values and larger displays
//   - Pixel queries use uint8 for optimization on smaller displays
//   - Display dimensions returned as uint16
//
// The interface is designed to work with various display technologies and provides
// a unified API for both simple and complex drawing operations.
type DisplayDrawer interface {
	GetDisplay() Display
	Size() (width, height uint16)
	BufferSize() int
	Buffer() []byte
	ClearBuffer()
	ClearDisplay()
	Command(cmd byte) error
	Display() error
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
	display Display // The underlying display interface
	buffer  []byte  // Internal buffer for graphics operations
}

var _ DisplayDrawer = (*T8Go)(nil) // Ensure T8Go implements DisplayDrawer

// ----------

// DrawQuadrants represents which quadrants of a circle or ellipse should be drawn.
// It uses bitwise flags to specify combinations of quadrants.
type DrawQuadrants uint8

const (
	DrawNone        DrawQuadrants = 0
	DrawTopLeft     DrawQuadrants = 1 << 0
	DrawTopRight    DrawQuadrants = 1 << 1
	DrawBottomRight DrawQuadrants = 1 << 2
	DrawBottomLeft  DrawQuadrants = 1 << 3
	DrawAll                       = DrawTopLeft | DrawTopRight | DrawBottomRight | DrawBottomLeft
)

// has checks if the given quadrant should be drawn.
// If mask is DrawNone, it is interpreted as "draw all".
func (mask DrawQuadrants) has(quadrant DrawQuadrants) bool {
	return mask == DrawNone || (mask&quadrant) != 0
}

// ----------

// scanSpan stores the min/max X coordinates to fill for a given scanline Y.
// This is used internally for filled shape rendering.
type scanSpan struct {
	minX        int16 // Minimum X coordinate for this scanline
	maxX        int16 // Maximum X coordinate for this scanline
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
	spans map[int16]scanSpan,
	centerX, centerY, offsetX, offsetY int16,
	angleStart, angleEnd uint8,
) {
	baseOctantAngle := helpers.ApproxAtanUnit64(offsetX, offsetY) // 0..64

	candidates := [8]struct {
		ang uint8
		x   int16
		y   int16
	}{
		{baseOctantAngle, centerX + offsetY, centerY - offsetX},                      // a0
		{64 - baseOctantAngle, centerX + offsetX, centerY - offsetY},                 // a1
		{64 + baseOctantAngle, centerX - offsetX, centerY - offsetY},                 // a2
		{128 - baseOctantAngle, centerX - offsetY, centerY - offsetX},                // a3
		{128 + baseOctantAngle, centerX - offsetY, centerY + offsetX},                // a4
		{192 - baseOctantAngle, centerX - offsetX, centerY + offsetY},                // a5
		{192 + baseOctantAngle, centerX + offsetX, centerY + offsetY},                // a6
		{uint8(256 - uint16(baseOctantAngle)), centerX + offsetY, centerY + offsetX}, // a7
	}

	for _, c := range candidates {
		if helpers.InAngleRange(c.ang, angleStart, angleEnd) {
			updateSpan(spans, c.x, c.y)

			if d := helpers.ArcAngleDistance(c.ang, angleStart); d < accum.bestStartAngleDiff {
				accum.bestStartAngleDiff, accum.startEndX, accum.startEndY = d, c.x, c.y
			}
			if d := helpers.ArcAngleDistance(c.ang, angleEnd); d < accum.bestEndAngleDiff {
				accum.bestEndAngleDiff, accum.endEndX, accum.endEndY = d, c.x, c.y
			}
		}
	}
}
