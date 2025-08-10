package ssd1306

import "errors"

var (
	ErrI2CBusNil = errors.New("I2C bus cannot be nil")
)
