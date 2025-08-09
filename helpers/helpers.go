package helpers

func Abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func Abs16(value int16) int16 {
	if value < 0 {
		return -value
	}
	return value
}

func AbsDiff(valueA, valueB int16) int16 {
	if valueA > valueB {
		return valueA - valueB
	}
	return valueB - valueA
}

// SwapInt16 swaps the values pointed by a and b.
func SwapInt16(a, b *int16) {
	*a, *b = *b, *a
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
