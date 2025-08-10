// Package helpers provides utility functions for mathematical operations
// and geometric calculations used throughout the t8go graphics library.
// It includes functions for absolute values, angle calculations,
// and coordinate transformations optimized for embedded systems.
package helpers

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

// NormalizeRect returns the top-left origin and the positive width/height
// for a rectangle defined by two corners (x0,y0)-(x1,y1), inclusive.
func NormalizeRect(x0, y0, x1, y1 int16) (originX, originY, width, height int16) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	return x0, y0, x1 - x0 + 1, y1 - y0 + 1
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
	ratio := uint32(offsetX) * 255 / uint32(offsetY)

	// Polynomial/rational fit (integer-only), adapted to map ratio to ~0..63.
	// Mirrors the original behavior while packaged as a helper.
	angle := ratio * (770195 - (ratio-255)*(ratio+941)) / 6137491

	if angle > 63 {
		return 63
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
	if d > 255-d {
		return 255 - d
	}
	return d
}

// AngleEndpoint returns the endpoint (endX, endY) of a line that starts at (originX, originY),
// has the given length (pixels, origin included), and angle in 0..255 units (64=90°, 128=180°, 192=270°).
// Negative length flips the direction by +180° and uses its absolute value.
func AngleEndpoint(originX, originY, length int16, angle uint8) (endX, endY int16) {
	if length == 0 {
		return originX, originY
	}

	// Normalize negative length by flipping the direction.
	if length < 0 {
		length = -length
		angle += 128
	}

	// Steps along the major axis (origin counts as the first pixel).
	chebyshev := length - 1
	if chebyshev <= 0 {
		return originX, originY
	}

	// Axis-aligned fast paths (exact endpoints, zero error).
	if angle%64 == 0 {
		switch angle {
		case 0:
			return originX + chebyshev, originY
		case 64:
			return originX, originY - chebyshev
		case 128:
			return originX - chebyshev, originY
		case 192:
			return originX, originY + chebyshev
		default:
			// e.g., 255≈360° → treat as 0°
			return originX + chebyshev, originY
		}
	}

	// Octant decomposition (32 units per octant).
	octant := int(angle / 32)  // 0..7
	intra := int16(angle % 32) // 0..31

	// Major axis per octant.
	majorIsX := (octant == 0 || octant == 3 || octant == 4 || octant == 7)

	// Integer minor component with rounding.
	var absDX, absDY int16
	if majorIsX {
		absDX = chebyshev
		absDY = int16((int32(chebyshev)*int32(intra) + 16) / 32) // ≈ round(chebyshev*intra/32)
	} else {
		absDY = chebyshev
		absDX = int16((int32(chebyshev)*int32(32-intra) + 16) / 32) // ≈ round(chebyshev*(32-intra)/32)
	}

	// Signs in screen coordinates (Y grows downward).
	switch octant {
	case 0, 1: // right/up-ish and up/right-ish
		absDX, absDY = +absDX, -absDY
	case 2, 3: // up/left-ish and left/up-ish
		absDX, absDY = -absDX, -absDY
	case 4, 5: // left/down-ish and down/left-ish
		absDX, absDY = -absDX, +absDY
	default: // 6, 7: down/right-ish and right/down-ish
		absDX, absDY = +absDX, +absDY
	}

	return originX + absDX, originY + absDY
}
