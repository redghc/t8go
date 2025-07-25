package ssd1306

import (
	"machine"

	"github.com/redghc/t8go"
)

type AddressMode uint16

func (a AddressMode) uint8() uint8 {
	return uint8(a)
}

func (a AddressMode) uint16() uint16 {
	return uint16(a)
}

type display struct {
	bus     *machine.I2C
	address AddressMode
	width   int16
	height  int16
}

var _ t8go.Display = &display{}

// NewI2C creates a new SSD1306 display instance using I2C communication
func NewI2C(bus *machine.I2C, address AddressMode, width, height int16) t8go.Display {
	return &display{
		bus:     bus,
		address: address,
		width:   width,
		height:  height,
	}
}

// Command sends a command byte to the display
func (d *display) Command(cmd byte) error {
	return d.bus.WriteRegister(d.address.uint8(), CONTROL_CMD_SINGLE, []byte{cmd})
}

// Size returns the display dimensions
func (d *display) Size() (width, height int16) {
	return d.width, d.height
}
