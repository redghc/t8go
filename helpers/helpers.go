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

func Direction(distance int) int {
	if distance < 0 {
		return -1
	} else if distance > 0 {
		return 1
	}
	return 0
}
