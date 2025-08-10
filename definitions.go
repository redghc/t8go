package t8go

import "github.com/redghc/t8go/helpers"

type Display interface {
	Size() (width, height uint16) // Size returns the display dimensions
	BufferSize() int              // BufferSize returns the size of the display buffer
	Buffer() []byte               // Buffer returns the buffer

	ClearBuffer()                 // ClearBuffer clears the display buffer
	ClearDisplay()                // ClearDisplay clears the image buffer and display
	Command(cmd byte) error       // Send a command to the display
	Display() error               // Send the current buffer to the display
	SetPixel(x, y int16, on bool) // SetPixel sets a pixel at (x, y) to on/off
	GetPixel(x, y uint8) bool     // GetPixel returns the state of a pixel at (x, y)
}

// ----------

type T8Go struct {
	display Display
	buffer  []byte
}

// ----------

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

// scanSpan stores the min/max X to fill for a given scanline Y.
type scanSpan struct {
	minX        int16
	maxX        int16
	initialized bool
}

// AddPoint expands the span to include x.
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

// IsEmpty reports whether the span has no points.
func (s *scanSpan) IsEmpty() bool {
	return !s.initialized
}

// arcAccum tracks perimeter points closest to start/end angles.
type arcAccum struct {
	bestStartAngleDiff uint8
	bestEndAngleDiff   uint8
	startEndX          int16
	startEndY          int16
	endEndX            int16
	endEndY            int16
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
