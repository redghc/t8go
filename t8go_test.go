package t8go_test

import (
	"os"
	"strings"
	"testing"

	"github.com/redghc/t8go"
)

// fakeDisplay mirrors the SSD1306 buffer layout (byte = column, bit = row
// within an 8-row page). It implements the t8go.IDisplay contract including
// the batch fill methods. FillHRun/FillRect are deliberately naive (loops of
// SetPixel) so they are layout-independent for golden comparison.
type fakeDisplay struct {
	width, height uint16
	stride        int
	buffer        []byte
}

func newFake(w, h uint16) *fakeDisplay {
	return &fakeDisplay{
		width:  w,
		height: h,
		stride: int(w),
		buffer: make([]byte, int(w)*int(h)/8),
	}
}

func (f *fakeDisplay) Size() (uint16, uint16)                 { return f.width, f.height }
func (f *fakeDisplay) BufferSize() int                        { return len(f.buffer) }
func (f *fakeDisplay) Buffer() []byte                         { return f.buffer }
func (f *fakeDisplay) ClearBuffer()                           { clear(f.buffer) }
func (f *fakeDisplay) ClearDisplay()                          { f.ClearBuffer() }
func (f *fakeDisplay) Command(byte) error                     { return nil }
func (f *fakeDisplay) Display() error                         { return nil }
func (f *fakeDisplay) DisplayRegion(int, int, int, int) error { return nil }

func (f *fakeDisplay) SetPixel(x, y int16, on bool) {
	if x < 0 || y < 0 || x >= int16(f.width) || y >= int16(f.height) {
		return
	}
	bi := int(x) + (int(y)>>3)*f.stride
	bit := uint8(1 << (y & 7))
	if on {
		f.buffer[bi] |= bit
	} else {
		f.buffer[bi] &^= bit
	}
}

func (f *fakeDisplay) GetPixel(x, y uint8) bool {
	if x >= uint8(f.width) || y >= uint8(f.height) {
		return false
	}
	bi := int(x) + (int(y)>>3)*f.stride
	return f.buffer[bi]&(1<<(y&7)) != 0
}

func (f *fakeDisplay) FillHRun(x0, x1, y int16, on bool) {
	for x := x0; x <= x1; x++ {
		f.SetPixel(x, y, on)
	}
}

func (f *fakeDisplay) FillRect(x0, y0, x1, y1 int16, on bool) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			f.SetPixel(x, y, on)
		}
	}
}

type shapeCase struct {
	name string
	draw func(g t8go.IDisplayDrawer)
}

func shapeCases() []shapeCase {
	return []shapeCase{
		{"hline", func(g t8go.IDisplayDrawer) { g.DrawHLine(10, 20, 40) }},
		{"vline", func(g t8go.IDisplayDrawer) { g.DrawVLine(60, 5, 30) }},
		{"line-diag", func(g t8go.IDisplayDrawer) { g.DrawLine(0, 0, 127, 63) }},
		{"line-slope", func(g t8go.IDisplayDrawer) { g.DrawLine(10, 10, 120, 20) }},
		{"box", func(g t8go.IDisplayDrawer) { g.DrawBox(10, 10, 50, 30) }},
		{"boxfill", func(g t8go.IDisplayDrawer) { g.DrawBoxFill(5, 5, 60, 40) }},
		{"boxfill-neg", func(g t8go.IDisplayDrawer) { g.DrawBoxFill(100, 50, -60, -40) }},
		{"boxcoords", func(g t8go.IDisplayDrawer) { g.DrawBoxCoords(30, 15, 90, 55) }},
		{"boxfillcoords", func(g t8go.IDisplayDrawer) { g.DrawBoxFillCoords(20, 10, 80, 50) }},
		{"roundbox", func(g t8go.IDisplayDrawer) { g.DrawRoundBox(5, 5, 60, 40, 8) }},
		{"roundboxfill", func(g t8go.IDisplayDrawer) { g.DrawRoundBoxFill(5, 5, 60, 40, 8) }},
		{"triangle", func(g t8go.IDisplayDrawer) { g.DrawTriangle(64, 5, 10, 50, 110, 50) }},
		{"trianglefill", func(g t8go.IDisplayDrawer) { g.DrawTriangleFill(64, 5, 10, 50, 110, 50) }},
		{"trianglefill-off", func(g t8go.IDisplayDrawer) { g.DrawTriangleFill(0, 0, 20, 40, 40, 0) }},
		{"circle", func(g t8go.IDisplayDrawer) { g.DrawCircle(64, 32, 20, t8go.DrawAll) }},
		{"circle-masked", func(g t8go.IDisplayDrawer) { g.DrawCircle(64, 32, 20, t8go.DrawTopRight) }},
		{"circlefill", func(g t8go.IDisplayDrawer) { g.DrawCircleFill(64, 32, 20, t8go.DrawAll) }},
		{"circlefill-masked", func(g t8go.IDisplayDrawer) { g.DrawCircleFill(64, 32, 20, t8go.DrawBottomLeft) }},
		{"ellipse", func(g t8go.IDisplayDrawer) { g.DrawEllipse(64, 32, 30, 20, t8go.DrawAll) }},
		{"ellipsefill", func(g t8go.IDisplayDrawer) { g.DrawEllipseFill(64, 32, 30, 20, t8go.DrawAll) }},
		{"arc", func(g t8go.IDisplayDrawer) { g.DrawArc(64, 32, 25, 32, 160) }},
		{"arcfill", func(g t8go.IDisplayDrawer) { g.DrawArcFill(64, 32, 25, 32, 160) }},
		{"lineangle", func(g t8go.IDisplayDrawer) { g.DrawLineAngle(64, 32, 40, 96) }},
		{"hline-offscreen", func(g t8go.IDisplayDrawer) { g.DrawHLine(100, 0, 50) }},
		{"vline-offscreen", func(g t8go.IDisplayDrawer) { g.DrawVLine(0, 60, 30) }},
		{"circle-offscreen", func(g t8go.IDisplayDrawer) { g.DrawCircle(0, 0, 30, t8go.DrawAll) }},
		{"circlefill-offscreen", func(g t8go.IDisplayDrawer) { g.DrawCircleFill(0, 0, 25, t8go.DrawAll) }},
	}
}

// renderASCII converts the display buffer into a human-readable ASCII art
// rendering: one line per pixel row, '#' for a set pixel and '.' for a clear
// pixel. The result is viewable in any text editor.
func renderASCII(f *fakeDisplay) string {
	var b strings.Builder
	b.Grow(int(f.height) * (int(f.width) + 1))
	for y := uint16(0); y < f.height; y++ {
		for x := uint16(0); x < f.width; x++ {
			bi := int(x) + (int(y)>>3)*f.stride
			bit := uint8(1 << (y & 7))
			if f.buffer[bi]&bit != 0 {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// TestShapesGolden verifies that shape rendering is identical to a golden
// reference captured from a previous build. Goldens are stored as ASCII art
// (one line per pixel row, '#' = pixel on, '.' = pixel off) so they can be
// inspected directly in any text editor.
// Set T8GO_WRITE_GOLDEN=1 to (re)generate the golden files from current code.
func TestShapesGolden(t *testing.T) {
	t.Parallel()
	writeMode := os.Getenv("T8GO_WRITE_GOLDEN") == "1"
	dir := "testdata"
	if writeMode {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range shapeCases() {
		disp := newFake(128, 64)
		g := t8go.New(disp)
		tc.draw(g)
		got := renderASCII(disp)

		path := dir + "/" + tc.name + ".txt"
		if writeMode {
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Logf("wrote %s", path)
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if string(want) != got {
			t.Errorf("%s: rendering mismatch", tc.name)
		}
	}
}
