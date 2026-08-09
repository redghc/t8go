package helpers_test

import (
	"testing"

	"github.com/redghc/t8go/helpers"
)

func TestAbs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want int16
	}{
		{0, 0},
		{5, 5},
		{-5, 5},
		{-32768, -32768}, // overflow-safe: -(-32768) is still negative on int16
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := helpers.Abs(tc.in); got != tc.want {
				t.Errorf("Abs(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestAbsDiff(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b, want int16
	}{
		{10, 3, 7},
		{3, 10, 7},
		{-5, 5, 10},
		{0, 0, 0},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := helpers.AbsDiff(tc.a, tc.b); got != tc.want {
				t.Errorf("AbsDiff(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestDirection(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   int16
		want int16
	}{
		{0, 0},
		{5, 1},
		{-5, -1},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := helpers.Direction(tc.in); got != tc.want {
				t.Errorf("Direction(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeRect(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		x0, y0, x1, y1 int16
		wantOX, wantOY int16
		wantW, wantH   int16
	}{
		{"already normalized", 2, 3, 7, 9, 2, 3, 6, 7},
		{"swap x", 7, 3, 2, 9, 2, 3, 6, 7},
		{"swap y", 2, 9, 7, 3, 2, 3, 6, 7},
		{"both swapped", 7, 9, 2, 3, 2, 3, 6, 7},
		{"single row", 0, 5, 4, 5, 0, 5, 5, 1},
		{"single col", 5, 0, 5, 4, 5, 0, 1, 5},
		{"single pixel", 3, 4, 3, 4, 3, 4, 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ox, oy, w, h := helpers.NormalizeRect(tc.x0, tc.y0, tc.x1, tc.y1)
			if ox != tc.wantOX || oy != tc.wantOY || w != tc.wantW || h != tc.wantH {
				t.Errorf("got (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					ox, oy, w, h, tc.wantOX, tc.wantOY, tc.wantW, tc.wantH)
			}
		})
	}
}

func TestInAngleRange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		angle, start, end uint8
		want              bool
	}{
		{10, 10, 20, true},    // in [10,20)
		{20, 10, 20, false},   // end is exclusive
		{10, 10, 20, true},    // start is inclusive
		{5, 10, 20, false},    // before
		{15, 10, 20, true},    // middle
		{250, 240, 10, true},  // wrapped: 250 in [240,10)
		{5, 240, 10, true},    // wrapped: 5 in [240,10)
		{100, 240, 10, false}, // outside wrapped interval
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := helpers.InAngleRange(tc.angle, tc.start, tc.end); got != tc.want {
				t.Errorf("InAngleRange(%d, %d, %d) = %v, want %v",
					tc.angle, tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestArcAngleDistance(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b, want uint8
	}{
		{0, 0, 0},
		{0, 64, 64},
		{0, 192, 64}, // wrap: 192 → 64 via 0
		{100, 200, 100},
		{200, 100, 100},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := helpers.ArcAngleDistance(tc.a, tc.b); got != tc.want {
				t.Errorf("ArcAngleDistance(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestApproxAtanUnit64(t *testing.T) {
	t.Parallel()
	// Known reference values (within ~1 unit tolerance).
	cases := []struct {
		x, y int16
		want uint8
	}{
		{0, 0, 0},      // degenerate
		{0, 100, 0},    // 0/anything → 0
		{100, 100, 32}, // 1:1 → ≈45° (octant boundary)
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			got := helpers.ApproxAtanUnit64(tc.x, tc.y)
			// 1-unit tolerance: integer rounding may give 31 or 33.
			diff := int(got) - int(tc.want)
			if diff < 0 {
				diff = -diff
			}
			if diff > 1 {
				t.Errorf("ApproxAtanUnit64(%d, %d) = %d, want ~%d", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

func TestAngleEndpoint(t *testing.T) {
	t.Parallel()
	originX, originY := int16(10), int16(10)

	t.Run("zero length returns origin", func(t *testing.T) {
		t.Parallel()
		x, y := helpers.AngleEndpoint(originX, originY, 0, 64)
		if x != originX || y != originY {
			t.Errorf("got (%d,%d), want origin", x, y)
		}
	})

	t.Run("axis aligned 0deg (right)", func(t *testing.T) {
		t.Parallel()
		x, y := helpers.AngleEndpoint(originX, originY, 5, 0)
		if x != originX+4 || y != originY {
			t.Errorf("got (%d,%d), want (%d,%d)", x, y, originX+4, originY)
		}
	})

	t.Run("axis aligned 90deg (up)", func(t *testing.T) {
		t.Parallel()
		x, y := helpers.AngleEndpoint(originX, originY, 5, helpers.AngleQuarterTurn)
		if x != originX || y != originY-4 {
			t.Errorf("got (%d,%d), want (%d,%d)", x, y, originX, originY-4)
		}
	})

	t.Run("axis aligned 180deg (left)", func(t *testing.T) {
		t.Parallel()
		x, y := helpers.AngleEndpoint(originX, originY, 5, helpers.AngleHalfTurn)
		if x != originX-4 || y != originY {
			t.Errorf("got (%d,%d), want (%d,%d)", x, y, originX-4, originY)
		}
	})

	t.Run("axis aligned 270deg (down)", func(t *testing.T) {
		t.Parallel()
		x, y := helpers.AngleEndpoint(originX, originY, 5, helpers.AngleThreeQuarterTurn)
		if x != originX || y != originY+4 {
			t.Errorf("got (%d,%d), want (%d,%d)", x, y, originX, originY+4)
		}
	})

	t.Run("negative length flips direction 180", func(t *testing.T) {
		t.Parallel()
		// -5 at angle 0 should land 4 left of origin (180° flipped from +5).
		xNeg, yNeg := helpers.AngleEndpoint(originX, originY, -5, 0)
		if xNeg != originX-4 || yNeg != originY {
			t.Errorf("got (%d,%d), want (%d,%d)", xNeg, yNeg, originX-4, originY)
		}
	})
}
