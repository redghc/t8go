package ssd1306

import (
	"machine"

	"github.com/redghc/t8go"
)

// * ----- Definitions -----s

type displayConfig struct {
	width   int16
	height  int16
	vccMode VCCMode
}

type display struct {
	bus     *machine.I2C
	address AddressMode

	vccMode VCCMode // Default: VCC_MODE_EXTERNAL

	width  int16 // Default: 128
	height int16 // Default: 64

	buffer []byte // Buffer to hold the display data
}

var _ t8go.Display = &display{}

// * ----- Constructors -----

// NewI2C creates a new SSD1306 display instance using I2C communication
func NewI2C(bus *machine.I2C, address AddressMode) t8go.Display {
	return &display{
		bus:     bus,
		address: address,
		vccMode: VCC_MODE_EXTERNAL,
	}
}

// Init initializes the display with the given configuration
func (d *display) Init(config displayConfig) error {
	if config.width != 0 {
		d.width = config.width
	} else {
		d.width = 128 // Default width
	}

	if config.height != 0 {
		d.height = config.height
	} else {
		d.height = 64 // Default height
	}

	if config.vccMode != 0 {
		d.vccMode = config.vccMode
	} else {
		d.vccMode = VCC_MODE_EXTERNAL // Default VCC mode
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

// * ----- Display methods -----

// ClearBuffer clears the display buffer
func (d *display) ClearBuffer() {
	for i := 0; i < len(d.buffer); i++ {
		d.buffer[i] = 0
	}
}

// ClearDisplay clears the image buffer and display
func (d *display) ClearDisplay() {
	d.ClearBuffer()
	d.Display()
}

// Command sends a command byte to the display
func (d *display) Command(cmd byte) error {
	return d.bus.WriteRegister(d.address, CONTROL_CMD_SINGLE, []byte{cmd})
}

// Display sends the current buffer to the display
func (d *display) Display() error {
	seq := []byte{
		SET_COLUMN_ADDRESS, 0x00, byte(d.width - 1),
		SET_PAGE_ADDRESS, 0x00, byte((d.height / 8) - 1),
	}

	return d.bus.WriteRegister(d.address, CONTROL_CMD_STREAM, seq)
}

// * ----- Getter methods -----

// Size returns the display dimensions
func (d *display) Size() (width, height int16) {
	return d.width, d.height
}

// BufferSize returns the size of the display buffer
func (d *display) BufferSize() int {
	return len(d.buffer)
}

// Buffer returns the buffer
func (d *display) Buffer() []byte {
	return d.buffer
}
