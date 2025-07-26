package ssd1306

import (
	"machine"

	"github.com/redghc/t8go"
)

type displayConfig struct {
	vccMode VCCMode
}

type display struct {
	bus     *machine.I2C
	address AddressMode
	width   int16
	height  int16

	vccMode VCCMode // Default: VCC_MODE_EXTERNAL
}

var _ t8go.Display = &display{}

// NewI2C creates a new SSD1306 display instance using I2C communication
func NewI2C(bus *machine.I2C, address AddressMode, width, height int16) t8go.Display {
	return &display{
		bus:     bus,
		address: address,
		width:   width,
		height:  height,
		vccMode: VCC_MODE_EXTERNAL, // Default VCC mode
	}
}

// Command sends a command byte to the display
func (d *display) Command(cmd byte) error {
	return d.bus.WriteRegister(d.address, CONTROL_CMD_SINGLE, []byte{cmd})
}

// Init initializes the display with the given configuration
func (d *display) Init(config displayConfig) error {
	if config.vccMode != 0 {
		d.vccMode = config.vccMode
	}

	// --

	var pumpMode byte
	var contrast byte
	var chargePeriod byte
	if d.vccMode == VCC_MODE_EXTERNAL {
		pumpMode = CHARGE_PUMP_SETTING_OFF
		contrast = 0x9F
		chargePeriod = 0x22
	} else {
		pumpMode = CHARGE_PUMP_SETTING_ON
		contrast = 0xCF
		chargePeriod = 0xF1
	}

	seq := []byte{
		SET_DISPLAY_OFF,
		SET_DISPLAY_CLOCK_DIVIDE_RATIO, 0x80,
		SET_MULTIPLEX_RATIO, 0x3F,
		SET_DISPLAY_OFFSET, 0x00,
		SET_START_LINE | 0x00,
		CHARGE_PUMP_SETTING, pumpMode,
		SET_SEGMENT_REMAP | 0x01,
		SET_COM_OUTPUT_SCAN_DIRECTION_DEC,
		SET_COM_PINS, 0x12,
		SET_CONTRAST, contrast,
		SET_PRE_CHARGE_PERIOD, chargePeriod,
		SET_VCOM_DESELECT_LEVEL, 0x20,
		DISPLAY_ALL_ON_RESUME,
		SET_NORMAL_DISPLAY,
		SET_DISPLAY_ON,
	}

	return d.bus.WriteRegister(d.address, CONTROL_CMD_STREAM, seq)
}

// Size returns the display dimensions
func (d *display) Size() (width, height int16) {
	return d.width, d.height
}

// Display sends the current buffer to the display
func (d *display) Display(buffer []byte) error {
	seq := []byte{
		SET_COLUMN_ADDRESS, 0x00, byte(d.width - 1),
		SET_PAGE_ADDRESS, 0x00, byte((d.height / 8) - 1),
	}

	return d.bus.WriteRegister(d.address, CONTROL_CMD_STREAM, seq)
}
