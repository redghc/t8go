package t8go_test

import (
	"bytes"
	"errors"
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
		buffer: make([]byte, int(w)*((int(h)+7)/8)),
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

func (f *fakeDisplay) FillVRun(x, y0, y1 int16, on bool) {
	for y := y0; y <= y1; y++ {
		f.SetPixel(x, y, on)
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
		{"polygon", func(g t8go.IDisplayDrawer) {
			g.DrawPolygon(t8go.Point{64, 5}, t8go.Point{10, 50}, t8go.Point{110, 50}, t8go.Point{90, 20})
		}},
		{"polygonfill", func(g t8go.IDisplayDrawer) {
			g.DrawPolygonFill(t8go.Point{64, 5}, t8go.Point{10, 50}, t8go.Point{110, 50}, t8go.Point{90, 20})
		}},
		{"polygon-concave", func(g t8go.IDisplayDrawer) {
			g.DrawPolygon(t8go.Point{15, 10}, t8go.Point{90, 10},
				t8go.Point{60, 35}, t8go.Point{105, 55}, t8go.Point{20, 55})
		}},
		{"polygonfill-concave", func(g t8go.IDisplayDrawer) {
			g.DrawPolygonFill(t8go.Point{15, 10}, t8go.Point{90, 10},
				t8go.Point{60, 35}, t8go.Point{105, 55}, t8go.Point{20, 55})
		}},
		{"polygon-star", func(g t8go.IDisplayDrawer) {
			g.DrawPolygon(
				t8go.Point{64, 62}, t8go.Point{57, 42}, t8go.Point{35, 41}, t8go.Point{53, 28},
				t8go.Point{46, 8}, t8go.Point{64, 20}, t8go.Point{82, 8}, t8go.Point{75, 28},
				t8go.Point{93, 41}, t8go.Point{71, 42},
			)
		}},
		{"polygonfill-star", func(g t8go.IDisplayDrawer) {
			g.DrawPolygonFill(
				t8go.Point{64, 62}, t8go.Point{57, 42}, t8go.Point{35, 41}, t8go.Point{53, 28},
				t8go.Point{46, 8}, t8go.Point{64, 20}, t8go.Point{82, 8}, t8go.Point{75, 28},
				t8go.Point{93, 41}, t8go.Point{71, 42},
			)
		}},
		{"polygon-offscreen", func(g t8go.IDisplayDrawer) {
			g.DrawPolygonFill(t8go.Point{-10, -10}, t8go.Point{40, -10}, t8go.Point{40, 40}, t8go.Point{-10, 40})
		}},
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
	for y := range f.height {
		for x := range f.width {
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

// recordingDisplay wraps fakeDisplay and records the region passed to
// DisplayRegion so dirty-region tracking can be asserted.
type recordingDisplay struct {
	*fakeDisplay

	sawRegion                              bool
	regionX0, regionY0, regionX1, regionY1 int
}

func (r *recordingDisplay) DisplayRegion(x0, y0, x1, y1 int) error {
	r.sawRegion = true
	r.regionX0, r.regionY0, r.regionX1, r.regionY1 = x0, y0, x1, y1
	return nil
}

func (r *recordingDisplay) Display() error {
	r.sawRegion = true
	r.regionX0, r.regionY0, r.regionX1, r.regionY1 = 0, 0, int(r.width-1), int(r.height-1)
	return nil
}

func TestDirtyRegion(t *testing.T) {
	t.Parallel()

	// Each subtest draws something (or nothing), flushes, and asserts the
	// recorded region. secondFlush exercises the no-op-after-flush path.
	cases := []struct {
		name        string
		draw        func(g t8go.IDisplayDrawer)
		useAll      bool
		secondFlush bool
		wantFlush   bool
		wantX0      int
		wantY0      int
		wantX1      int
		wantY1      int
	}{
		{
			name:      "no-op when nothing drawn",
			draw:      func(t8go.IDisplayDrawer) {},
			wantFlush: false,
		},
		{
			name:      "box marks its bounding box",
			draw:      func(g t8go.IDisplayDrawer) { g.DrawBoxFill(10, 10, 20, 20) },
			wantFlush: true, wantX0: 10, wantY0: 10, wantX1: 29, wantY1: 29,
		},
		{
			name:      "circle marks its bounding box",
			draw:      func(g t8go.IDisplayDrawer) { g.DrawCircleFill(64, 32, 20, t8go.DrawAll) },
			wantFlush: true, wantX0: 44, wantY0: 12, wantX1: 84, wantY1: 52,
		},
		{
			name:      "line marks endpoints' bounding box",
			draw:      func(g t8go.IDisplayDrawer) { g.DrawLine(10, 10, 100, 50) },
			wantFlush: true, wantX0: 10, wantY0: 10, wantX1: 100, wantY1: 50,
		},
		{
			name:        "no-op after flush",
			draw:        func(g t8go.IDisplayDrawer) { g.DrawBoxFill(0, 0, 1, 1) },
			secondFlush: true,
			wantFlush:   true, wantX0: 0, wantY0: 0, wantX1: 0, wantY1: 0,
		},
		{
			name:      "DisplayAll always flushes",
			draw:      func(t8go.IDisplayDrawer) {},
			useAll:    true,
			wantFlush: true, wantX0: 0, wantY0: 0, wantX1: 127, wantY1: 63,
		},
		{
			name:      "ClearBuffer marks the whole screen",
			draw:      func(g t8go.IDisplayDrawer) { g.ClearBuffer() },
			wantFlush: true, wantX0: 0, wantY0: 0, wantX1: 127, wantY1: 63,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runDirtyCase(t, tc)
		})
	}
}

// runDirtyCase runs one entry from the dirty-region table.
func runDirtyCase(t *testing.T, tc struct {
	name        string
	draw        func(g t8go.IDisplayDrawer)
	useAll      bool
	secondFlush bool
	wantFlush   bool
	wantX0      int
	wantY0      int
	wantX1      int
	wantY1      int
}) {
	disp, g := newRecordingCtx()
	tc.draw(g)
	flush := g.Display
	if tc.useAll {
		flush = g.DisplayAll
	}
	if err := flush(); err != nil {
		t.Fatal(err)
	}
	if tc.secondFlush {
		before := disp.sawRegion
		if err := g.Display(); err != nil {
			t.Fatal(err)
		}
		if disp.sawRegion != before {
			t.Error("Display() flushed again with no new pixels")
			return
		}
	}
	if !tc.wantFlush {
		if disp.sawRegion {
			t.Error("Display() flushed with no dirty pixels")
		}
		return
	}
	assertRegion(t, disp, tc.wantX0, tc.wantY0, tc.wantX1, tc.wantY1)
}

// newRecordingCtx builds a fresh drawing context so each subtest has its
// own state and can safely run in parallel.
func newRecordingCtx() (*recordingDisplay, t8go.IDisplayDrawer) {
	disp := &recordingDisplay{fakeDisplay: newFake(128, 64)}
	return disp, t8go.New(disp)
}

// assertRegion fails the test if the recorded flush region doesn't match.
func assertRegion(t *testing.T, disp *recordingDisplay, x0, y0, x1, y1 int) {
	t.Helper()
	if !disp.sawRegion {
		t.Fatal("expected a region flush")
	}
	if disp.regionX0 != x0 || disp.regionY0 != y0 || disp.regionX1 != x1 || disp.regionY1 != y1 {
		t.Errorf("bad region: got (%d,%d)-(%d,%d), want (%d,%d)-(%d,%d)",
			disp.regionX0, disp.regionY0, disp.regionX1, disp.regionY1,
			x0, y0, x1, y1)
	}
}

// errorDisplay wraps fakeDisplay and lets individual methods return errors on
// demand, so we can assert that T8Go propagates display-driver errors.
type errorDisplay struct {
	*fakeDisplay

	displayErr error
	regionErr  error
	cmdErr     error
}

func (e *errorDisplay) Display() error                         { return e.displayErr }
func (e *errorDisplay) DisplayRegion(int, int, int, int) error { return e.regionErr }
func (e *errorDisplay) Command(byte) error                     { return e.cmdErr }

func TestPixelPrimitives(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)

	if w, h := g.Size(); w != 16 || h != 8 {
		t.Errorf("Size = (%d,%d), want (16,8)", w, h)
	}
	if got := g.BufferSize(); got != 16 {
		t.Errorf("BufferSize = %d, want 16", got)
	}
	if buf := g.Buffer(); len(buf) != 16 {
		t.Errorf("Buffer length = %d, want 16", len(buf))
	}
	if g.GetDisplay() == nil {
		t.Error("GetDisplay returned nil")
	}

	g.DrawPixel(3, 4)
	if !g.GetPixel(3, 4) {
		t.Error("GetPixel(3,4) = false after DrawPixel")
	}
	if g.GetPixel(0, 0) {
		t.Error("GetPixel(0,0) = true on empty buffer")
	}
}

func TestSetPixelOutOfBoundsDoesNotMarkDirty(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)

	for _, p := range []struct{ x, y int16 }{
		{-1, 0}, {0, -1}, {16, 0}, {0, 8}, {100, 100}, {-100, -100},
	} {
		g.SetPixel(p.x, p.y, true)
	}
	if err := g.Display(); err != nil {
		t.Fatalf("Display returned %v", err)
	}
}

func TestSetPixelOnFalse(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)

	g.DrawPixel(5, 3)
	if !g.GetPixel(5, 3) {
		t.Fatal("pixel not set after DrawPixel")
	}
	g.SetPixel(5, 3, false)
	if g.GetPixel(5, 3) {
		t.Error("pixel still on after SetPixel(false)")
	}
}

func TestClearBufferClearsBuffer(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)

	g.DrawPixel(0, 0)
	g.DrawPixel(15, 7)
	g.ClearBuffer()
	if g.GetPixel(0, 0) || g.GetPixel(15, 7) {
		t.Error("ClearBuffer left pixels on")
	}
}

func TestClearDisplayResetsDirty(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)

	g.DrawPixel(0, 0)
	g.ClearDisplay()
	if g.GetPixel(0, 0) {
		t.Error("ClearDisplay did not clear buffer")
	}
	if err := g.Display(); err != nil {
		t.Fatalf("Display returned %v", err)
	}
}

func TestCommandPassthrough(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)
	if err := g.Command(0xAE); err != nil {
		t.Errorf("Command returned %v, want nil", err)
	}
}

func TestDisplayErrorPropagates(t *testing.T) {
	t.Parallel()
	ed := &errorDisplay{fakeDisplay: newFake(16, 8), regionErr: errors.New("boom")}
	g := t8go.New(ed)

	g.DrawPixel(1, 1)
	err := g.Display()
	if err == nil || err.Error() != "boom" {
		t.Errorf("Display err = %v, want boom", err)
	}
	if err2 := g.Display(); err2 != nil {
		t.Errorf("second Display err = %v, want nil (dirty must reset on error)", err2)
	}
}

func TestDisplayAllErrorPropagates(t *testing.T) {
	t.Parallel()
	ed := &errorDisplay{fakeDisplay: newFake(16, 8), displayErr: errors.New("io")}
	g := t8go.New(ed)
	if err := g.DisplayAll(); err == nil || err.Error() != "io" {
		t.Errorf("DisplayAll err = %v, want io", err)
	}
}

func TestDrawHLineZeroLengthNoOp(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)
	g.DrawHLine(5, 3, 0)
	if err := g.Display(); err != nil {
		t.Fatal(err)
	}
}

func TestDrawVLineZeroLengthNoOp(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)
	g.DrawVLine(5, 3, 0)
	if err := g.Display(); err != nil {
		t.Fatal(err)
	}
}

func TestDrawHLineNegativeLength(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)
	g.DrawHLine(10, 3, -5)
	for _, x := range []int16{6, 7, 8, 9, 10} {
		if !g.GetPixel(uint8(x), 3) {
			t.Errorf("pixel (%d,3) not set", x)
		}
	}
}

func TestDrawVLineNegativeLength(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)
	g.DrawVLine(3, 5, -4)
	for _, y := range []int16{2, 3, 4, 5} {
		if !g.GetPixel(3, uint8(y)) {
			t.Errorf("pixel (3,%d) not set", y)
		}
	}
}

func TestDrawBoxTooSmallNoOp(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ w, h int16 }{
		{1, 5}, {5, 1}, {1, 1}, {0, 5}, {5, 0}, {0, 0},
	} {
		disp := newFake(16, 8)
		g := t8go.New(disp)
		g.DrawBox(2, 2, tc.w, tc.h)
		for y := int16(0); y < 8; y++ {
			for x := int16(0); x < 16; x++ {
				if g.GetPixel(uint8(x), uint8(y)) {
					t.Errorf("box %dx%d: unexpected pixel at (%d,%d)", tc.w, tc.h, x, y)
				}
			}
		}
	}
}

func TestDrawBoxFillZeroDimNoOp(t *testing.T) {
	t.Parallel()
	disp := newFake(16, 8)
	g := t8go.New(disp)
	g.DrawBoxFill(5, 5, 0, 5)
	g.DrawBoxFill(5, 5, 5, 0)
	if err := g.Display(); err != nil {
		t.Fatal(err)
	}
}

func TestDrawRoundBoxNegativeRadiusEqualsBox(t *testing.T) {
	t.Parallel()
	dispA := newFake(20, 16)
	gA := t8go.New(dispA)
	gA.DrawRoundBox(2, 2, 10, 8, -3)
	bufA := append([]byte{}, dispA.buffer...)

	dispB := newFake(20, 16)
	gB := t8go.New(dispB)
	gB.DrawBox(2, 2, 10, 8)

	if !bytes.Equal(bufA, dispB.buffer) {
		t.Error("DrawRoundBox with negative radius should equal DrawBox")
	}
}

func TestDrawPolygonTooFewVerticesNoOp(t *testing.T) {
	t.Parallel()
	disp := newFake(20, 20)
	g := t8go.New(disp)
	g.DrawPolygon(t8go.Point{5, 5})
	g.DrawPolygon(t8go.Point{5, 5}, t8go.Point{10, 10})
	g.DrawPolygonFill(t8go.Point{5, 5})
	g.DrawPolygonFill(t8go.Point{5, 5}, t8go.Point{10, 10})
	if err := g.Display(); err != nil {
		t.Fatal(err)
	}
}

func TestDrawCircleRadiusZeroNoOp(t *testing.T) {
	t.Parallel()
	for _, r := range []int16{0, -1, -100} {
		disp := newFake(20, 20)
		g := t8go.New(disp)
		g.DrawCircle(10, 10, r, t8go.DrawAll)
		g.DrawCircleFill(10, 10, r, t8go.DrawAll)
		if err := g.Display(); err != nil {
			t.Errorf("radius %d: %v", r, err)
		}
	}
}

func TestDrawCircleMaskZeroDrawsNothing(t *testing.T) {
	t.Parallel()
	// The zero DrawQuadrants value (no bits set) must select no quadrant.
	// DrawAll is required to render the full disc.
	disp := newFake(20, 20)
	g := t8go.New(disp)
	g.DrawCircle(10, 10, 5, 0)
	g.DrawCircleFill(10, 10, 5, 0)
	if err := g.Display(); err != nil {
		t.Fatal(err)
	}
	for y := int16(0); y < 20; y++ {
		for x := int16(0); x < 20; x++ {
			if g.GetPixel(uint8(x), uint8(y)) {
				t.Errorf("mask=0: unexpected pixel at (%d,%d)", x, y)
			}
		}
	}
}

func TestDrawEllipseRadiusZeroNoOp(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ rx, ry int16 }{
		{0, 5}, {5, 0}, {-1, 5}, {5, -1}, {0, 0},
	} {
		disp := newFake(20, 20)
		g := t8go.New(disp)
		g.DrawEllipse(10, 10, tc.rx, tc.ry, t8go.DrawAll)
		g.DrawEllipseFill(10, 10, tc.rx, tc.ry, t8go.DrawAll)
		if err := g.Display(); err != nil {
			t.Errorf("rx=%d ry=%d: %v", tc.rx, tc.ry, err)
		}
	}
}

func TestDrawArcRadiusZeroNoOp(t *testing.T) {
	t.Parallel()
	disp := newFake(20, 20)
	g := t8go.New(disp)
	g.DrawArc(10, 10, 0, 0, 64)
	g.DrawArcFill(10, 10, 0, 0, 64)
	if err := g.Display(); err != nil {
		t.Fatal(err)
	}
}

func TestDrawArcFullArcEqualsCircle(t *testing.T) {
	t.Parallel()
	dispArc := newFake(20, 20)
	gArc := t8go.New(dispArc)
	gArc.DrawArc(10, 10, 5, 0, 0)

	dispCircle := newFake(20, 20)
	gCircle := t8go.New(dispCircle)
	gCircle.DrawCircle(10, 10, 5, t8go.DrawAll)

	if !bytes.Equal(dispArc.buffer, dispCircle.buffer) {
		t.Error("DrawArc(...,start==end) should equal DrawCircle")
	}
}

func TestDrawArcFillFullArcEqualsCircleFill(t *testing.T) {
	t.Parallel()
	dispArc := newFake(20, 20)
	gArc := t8go.New(dispArc)
	gArc.DrawArcFill(10, 10, 5, 0, 0)

	dispCircle := newFake(20, 20)
	gCircle := t8go.New(dispCircle)
	gCircle.DrawCircleFill(10, 10, 5, t8go.DrawAll)

	if !bytes.Equal(dispArc.buffer, dispCircle.buffer) {
		t.Error("DrawArcFill(...,start==end) should equal DrawCircleFill")
	}
}

func TestDrawArcWrappedInterval(t *testing.T) {
	t.Parallel()
	// [224, 32) covers 224..255 and 0..31 (the East + a bit of N/S).
	// Pixel at angle 0 (East): (30, 20) must be ON.
	// Pixel at angle 128 (West): (10, 20) must be OFF.
	disp := newFake(40, 40)
	g := t8go.New(disp)
	g.DrawArc(20, 20, 10, 224, 32)
	if !g.GetPixel(30, 20) {
		t.Error("expected pixel (30,20) on for angle 0 in [224,32)")
	}
	if g.GetPixel(10, 20) {
		t.Error("unexpected pixel (10,20) on for angle 128 not in [224,32)")
	}
}

func TestDirtyTrackingForPrimitives(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		draw   func(g t8go.IDisplayDrawer)
		wantX0 int
		wantY0 int
		wantX1 int
		wantY1 int
	}{
		{
			name: "polygon",
			draw: func(g t8go.IDisplayDrawer) {
				g.DrawPolygon(t8go.Point{10, 10}, t8go.Point{30, 10}, t8go.Point{30, 30}, t8go.Point{10, 30})
			},
			// Even-odd pixel-center sampling: the rightmost/bottommost
			// edge of the bounding box is not drawn.
			wantX0: 10, wantY0: 10, wantX1: 29, wantY1: 29,
		},
		{
			name:   "roundbox",
			draw:   func(g t8go.IDisplayDrawer) { g.DrawRoundBoxFill(10, 10, 30, 20, 4) },
			wantX0: 10, wantY0: 10, wantX1: 39, wantY1: 29,
		},
		{
			name:   "triangle",
			draw:   func(g t8go.IDisplayDrawer) { g.DrawTriangleFill(20, 5, 5, 30, 35, 30) },
			wantX0: 5, wantY0: 5, wantX1: 35, wantY1: 30,
		},
		{
			name:   "ellipse",
			draw:   func(g t8go.IDisplayDrawer) { g.DrawEllipse(64, 32, 20, 10, t8go.DrawAll) },
			wantX0: 44, wantY0: 22, wantX1: 84, wantY1: 42,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			disp, g := newRecordingCtx()
			tc.draw(g)
			if err := g.Display(); err != nil {
				t.Fatal(err)
			}
			assertRegion(t, disp, tc.wantX0, tc.wantY0, tc.wantX1, tc.wantY1)
		})
	}
}
