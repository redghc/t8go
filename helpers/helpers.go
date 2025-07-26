package helpers

func Abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func Direction(distance int) int {
	if distance < 0 {
		return -1
	} else if distance > 0 {
		return 1
	}
	return 0
}
