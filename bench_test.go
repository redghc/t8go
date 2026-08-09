package t8go_test

import (
	"testing"

	"github.com/redghc/t8go"
)

// optFake mirrors the SSD1306 byte-column layout and reproduces the OPTIMIZED
// FillHRun/FillRect of the real ssd1306 driver (page-aligned row masks, 0xFF
// direct stores). Benchmarks therefore measure the real buffer-write cost as
// it would run on hardware. On the pre-refactor code the drawing layer never
// calls these methods (it writes via SetPixel), which is exactly the point.
type optFake struct {
	width, height uint16
	stride        int
	buffer        []byte
}

func newOptFake(w, h uint16) *optFake {
	return &optFake{
		width:  w,
		height: h,
		stride: int(w),
		buffer: make([]byte, int(w)*int(h)/8),
	}
}

func (f *optFake) Size() (uint16, uint16)                 { return f.width, f.height }
func (f *optFake) BufferSize() int                        { return len(f.buffer) }
func (f *optFake) Buffer() []byte                         { return f.buffer }
func (f *optFake) ClearBuffer()                           { clear(f.buffer) }
func (f *optFake) ClearDisplay()                          { f.ClearBuffer() }
func (f *optFake) Command(byte) error                     { return nil }
func (f *optFake) Display() error                         { return nil }
func (f *optFake) DisplayRegion(int, int, int, int) error { return nil }

func (f *optFake) SetPixel(x, y int16, on bool) {
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

func (f *optFake) GetPixel(x, y uint8) bool {
	if x >= uint8(f.width) || y >= uint8(f.height) {
		return false
	}
	bi := int(x) + (int(y)>>3)*f.stride
	return f.buffer[bi]&(1<<(y&7)) != 0
}

func (f *optFake) FillHRun(x0, x1, y int16, on bool) {
	if y < 0 || y >= int16(f.height) {
		return
	}
	if x0 < 0 {
		x0 = 0
	}
	if x1 >= int16(f.width) {
		x1 = int16(f.width) - 1
	}
	if x1 < x0 {
		return
	}
	mask := uint8(1 << (y & 7))
	base := int(y>>3) * f.stride
	for c := int(x0); c <= int(x1); c++ {
		if on {
			f.buffer[base+c] |= mask
		} else {
			f.buffer[base+c] &^= mask
		}
	}
}

func (f *optFake) FillRect(x0, y0, x1, y1 int16, on bool) {
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= int16(f.width) {
		x1 = int16(f.width) - 1
	}
	if y1 >= int16(f.height) {
		y1 = int16(f.height) - 1
	}
	if x1 < x0 || y1 < y0 {
		return
	}
	startPage := int(y0 >> 3)
	endPage := int(y1 >> 3)
	for page := startPage; page <= endPage; page++ {
		var lo, hi uint8
		if page == startPage {
			lo = uint8(y0 & 7)
		}
		if page == endPage {
			hi = uint8(y1 & 7)
		} else {
			hi = 7
		}
		var mask uint8
		for b := lo; b <= hi; b++ {
			mask |= 1 << b
		}
		rowOffset := page * f.stride
		for c := int(x0); c <= int(x1); c++ {
			bi := rowOffset + c
			if on {
				if mask == 0xFF {
					f.buffer[bi] = 0xFF
				} else {
					f.buffer[bi] |= mask
				}
			} else {
				f.buffer[bi] &^= mask
			}
		}
	}
}

func (f *optFake) FillVRun(x, y0, y1 int16, on bool) {
	if x < 0 || x >= int16(f.width) {
		return
	}
	if y0 < 0 {
		y0 = 0
	}
	if y1 >= int16(f.height) {
		y1 = int16(f.height) - 1
	}
	if y1 < y0 {
		return
	}
	col := int(x)
	startPage := int(y0 >> 3)
	endPage := int(y1 >> 3)
	for page := startPage; page <= endPage; page++ {
		var lo, hi uint8
		if page == startPage {
			lo = uint8(y0 & 7)
		}
		if page == endPage {
			hi = uint8(y1 & 7)
		} else {
			hi = 7
		}
		var mask uint8
		for b := lo; b <= hi; b++ {
			mask |= 1 << b
		}
		bi := page*f.stride + col
		if on {
			if mask == 0xFF {
				f.buffer[bi] = 0xFF
			} else {
				f.buffer[bi] |= mask
			}
		} else {
			f.buffer[bi] &^= mask
		}
	}
}

// benchDraw renders one shape per iteration into a fresh buffer, reusing a
// single graphics context so setup cost (and any scratch allocation) happens
// outside the timed loop.
func benchDraw(b *testing.B, draw func(g t8go.IDisplayDrawer)) {
	b.ReportAllocs()
	disp := newOptFake(128, 64)
	g := t8go.New(disp)
	for range b.N {
		disp.ClearBuffer()
		draw(g)
	}
}

func BenchmarkDrawHLine(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawHLine(0, 32, 100) })
}
func BenchmarkDrawVLine(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawVLine(60, 5, 50) })
}
func BenchmarkDrawLineDiag(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawLine(0, 0, 127, 63) })
}
func BenchmarkDrawBox(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawBox(10, 10, 60, 40) })
}
func BenchmarkDrawBoxFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawBoxFill(10, 10, 60, 40) })
}
func BenchmarkDrawRoundBoxFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawRoundBoxFill(10, 10, 60, 40, 8) })
}
func BenchmarkDrawCircle(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawCircle(64, 32, 30, t8go.DrawAll) })
}
func BenchmarkDrawCircleFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawCircleFill(64, 32, 30, t8go.DrawAll) })
}
func BenchmarkDrawEllipseFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawEllipseFill(64, 32, 30, 20, t8go.DrawAll) })
}
func BenchmarkDrawTriangleFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawTriangleFill(64, 5, 10, 50, 110, 50) })
}
func BenchmarkDrawPolygon(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) {
		g.DrawPolygon(t8go.Point{64, 5}, t8go.Point{10, 50}, t8go.Point{110, 50}, t8go.Point{90, 20})
	})
}
func BenchmarkDrawPolygonFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) {
		g.DrawPolygonFill(t8go.Point{64, 5}, t8go.Point{10, 50}, t8go.Point{110, 50}, t8go.Point{90, 20})
	})
}
func BenchmarkDrawArcFill(b *testing.B) {
	benchDraw(b, func(g t8go.IDisplayDrawer) { g.DrawArcFill(64, 32, 30, 32, 160) })
}
