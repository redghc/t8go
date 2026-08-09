// Package helpers provides utility functions for mathematical operations
// and geometric calculations used throughout the t8go graphics library.
// It includes functions for absolute values, angle calculations,
// and coordinate transformations optimized for embedded systems.
package helpers

// Angle units used by the trigonometric helpers and arc drawing.
// Angles are expressed in 0..255 units where 64 = 90°, 128 = 180°, 192 = 270°.
const (
	// AngleQuarterTurn is a quarter turn (90°) in 0..255 angle units.
	AngleQuarterTurn = 64
	// AngleHalfTurn is a half turn (180°).
	AngleHalfTurn = 128
	// AngleThreeQuarterTurn is a three-quarter turn (270°).
	AngleThreeQuarterTurn = 192
	// AngleMax is the largest representable angle (≈360°, wraps to 0).
	AngleMax = 255
	// AngleFullCircle is the size of a full circle in angle units.
	AngleFullCircle = 256
	// OctantSize is the angular width of one octant (45°).
	OctantSize = 32
)

// Fixed-point parameters for the ApproxAtanUnit64 rational polynomial fit.
const (
	atanRatioScale  = 255     // scale of the offsetX/offsetY ratio
	atanOutputMax   = 63      // upper bound of the mapped output
	atanNumerator   = 770195  // numerator coefficient of the rational fit
	atanCrossTerm   = 941     // cross-term constant of the rational fit
	atanDenominator = 6137491 // denominator of the rational fit
)

// Octant indices in AngleEndpoint's octant decomposition, numbered
// counter-clockwise starting from the +X axis.
const (
	octant0 = iota // right/up-ish
	octant1        // up/right-ish
	octant2        // up/left-ish
	octant3        // left/up-ish
	octant4        // left/down-ish
	octant5        // down/left-ish
	octant6        // down/right-ish
	octant7        // right/down-ish
)

// Abs returns the absolute value of a signed integer (keeps original type).
func Abs[T ~int | ~int16 | ~int32 | ~int64](value T) T {
	if value < 0 {
		return -value
	}
	return value
}

// AbsDiff returns the absolute difference between two numbers of the same integer type.
// Works with any signed or unsigned integer type.
func AbsDiff[T ~int | ~int8 | ~int16 | ~int32 | ~int64](valueA, valueB T) T {
	if valueA > valueB {
		return valueA - valueB
	}
	return valueB - valueA
}

// Direction returns -1 if length is negative, 1 if positive, or 0 if zero.
// It is used to determine drawing direction for lines.
func Direction[T ~int | ~int16 | ~int32 | ~int64](length T) int16 {
	switch {
	case length < 0:
		return -1
	case length > 0:
		return 1
	default:
		return 0
	}
}

// NormalizeRect normalizes a rectangle defined by two corners (x0,y0)-(x1,y1),
// inclusive, returning the top-left `originX`/`originY` and the positive
// `width`/`height`.
func NormalizeRect(x0, y0, x1, y1 int16) (int16, int16, int16, int16) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	return x0, y0, x1 - x0 + 1, y1 - y0 + 1
}

// ClipRect clamps the inclusive rectangle (x0,y0)-(x1,y1) to the bounds
// [0,width) x [0,height), normalizing swapped corners. Returns the clipped
// `nx0`/`ny0`/`nx1`/`ny1` and `ok`=false when the rectangle lies entirely
// outside the bounds.
func ClipRect(x0, y0, x1, y1, width, height int16) (int16, int16, int16, int16, bool) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= width {
		x1 = width - 1
	}
	if y1 >= height {
		y1 = height - 1
	}
	if x1 < x0 || y1 < y0 {
		return 0, 0, 0, 0, false
	}
	return x0, y0, x1, y1, true
}

// ClipSpan clamps the inclusive 1-D span [a0, a1] to [0, n), normalizing a
// swapped range. Returns the clipped `na0`/`na1` and `ok`=false when the span
// lies entirely outside the bounds.
func ClipSpan(a0, a1, n int16) (int16, int16, bool) {
	if a0 > a1 {
		a0, a1 = a1, a0
	}
	if a0 < 0 {
		a0 = 0
	}
	if a1 >= n {
		a1 = n - 1
	}
	if a1 < a0 {
		return 0, 0, false
	}
	return a0, a1, true
}

// ApproxAtanUnit64 approximates atan(offsetX/offsetY) mapped to 0..64 units (i.e., 0..90°)
// using integer arithmetic. For offsetY == 0, it returns 0.
//
// The approximation keeps good monotonicity for arc inclusion tests and is derived
// from a tuned rational polynomial to fit arctan on [0, 1].
func ApproxAtanUnit64(offsetX, offsetY int16) uint8 {
	if offsetY == 0 {
		return 0
	}

	// ratio in 0..255 (represents offsetX/offsetY scaled to 0..255)
	ratio := uint32(offsetX) * atanRatioScale / uint32(offsetY)

	// Polynomial/rational fit (integer-only), adapted to map ratio to ~0..63.
	// Mirrors the original behavior while packaged as a helper.
	angle := ratio * (atanNumerator - (ratio-atanRatioScale)*(ratio+atanCrossTerm)) / atanDenominator

	if angle > atanOutputMax {
		return atanOutputMax
	}
	return uint8(angle)
}

// InAngleRange reports whether angle (0..255) is within [start, end) on a circular scale.
// If start == end, the caller should treat it as a full circle (handled by isFullArc).
func InAngleRange(angle, start, end uint8) bool {
	if start <= end {
		return angle >= start && angle < end
	}

	// Wrapped interval (e.g., 224..16)
	return angle >= start || angle < end
}

// ArcAngleDistance returns the minimal circular distance in 0..255 units.
func ArcAngleDistance(a, b uint8) uint8 {
	var d uint8
	if a >= b {
		d = a - b
	} else {
		d = b - a
	}
	if d > AngleMax-d {
		return AngleMax - d
	}
	return d
}

// AngleEndpoint returns the `endX`/`endY` endpoint of a line that starts at
// (originX, originY), has the given length (pixels, origin included), and angle
// in 0..255 units (64=90°, 128=180°, 192=270°).
// Negative length flips the direction by +180° and uses its absolute value.
func AngleEndpoint(originX, originY, length int16, angle uint8) (int16, int16) {
	if length == 0 {
		return originX, originY
	}

	// Normalize negative length by flipping the direction.
	if length < 0 {
		length = -length
		angle += AngleHalfTurn
	}

	// Steps along the major axis (origin counts as the first pixel).
	chebyshev := length - 1
	if chebyshev <= 0 {
		return originX, originY
	}

	// Axis-aligned fast paths (exact endpoints, zero error).
	if angle%AngleQuarterTurn == 0 {
		switch angle {
		case 0:
			return originX + chebyshev, originY
		case AngleQuarterTurn:
			return originX, originY - chebyshev
		case AngleHalfTurn:
			return originX - chebyshev, originY
		case AngleThreeQuarterTurn:
			return originX, originY + chebyshev
		default:
			// e.g., 255≈360° → treat as 0°
			return originX + chebyshev, originY
		}
	}

	// Octant decomposition (32 units per octant).
	octant := int(angle / OctantSize)  // 0..7
	intra := int16(angle % OctantSize) // 0..31

	// Major axis per octant.
	majorIsX := (octant == octant0 || octant == octant3 || octant == octant4 || octant == octant7)

	// Integer minor component with rounding.
	var absDX, absDY int16
	if majorIsX {
		absDX = chebyshev
		absDY = int16((int32(chebyshev)*int32(intra) + OctantSize/2) / OctantSize) // ≈ round(chebyshev*intra/32)
	} else {
		absDY = chebyshev
		absDX = int16((int32(chebyshev)*int32(OctantSize-intra) + OctantSize/2) / OctantSize) // ≈ round(chebyshev*(32-intra)/32)
	}

	// Signs in screen coordinates (Y grows downward).
	switch octant {
	case octant0, octant1: // right/up-ish and up/right-ish
		absDX, absDY = +absDX, -absDY
	case octant2, octant3: // up/left-ish and left/up-ish
		absDX, absDY = -absDX, -absDY
	case octant4, octant5: // left/down-ish and down/left-ish
		absDX, absDY = -absDX, +absDY
	default: // octant6, octant7: down/right-ish and right/down-ish
		absDX, absDY = +absDX, +absDY
	}

	return originX + absDX, originY + absDY
}
