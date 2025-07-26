package ssd1306

import (
	"machine"

	"github.com/redghc/t8go"
)

// * ----- Definitions -----

type Config struct {
	Width   uint8
	Height  uint8
	VCCMode VCCMode
}

type display struct {
	bus     *machine.I2C
	address AddressMode

	width   uint8   // Default: 128
	height  uint8   // Default: 64
	vccMode VCCMode // Default: VCC_SWITCH_CAP

	buffer  []byte
	bufSize int

	// Pre-allocated command buffers to avoid allocations
	cmdBuf  [32]byte
	addrBuf [6]byte
}

var _ t8go.Display = &display{}

// * ----- Constructors -----

// NewI2C creates a new SSD1306 display instance using I2C communication
func NewI2C(bus *machine.I2C, address AddressMode, config Config) (t8go.Display, error) {
	if bus == nil {
		return nil, ErrI2CBusNil
	}

	if config.Width == 0 {
		config.Width = 128 // Default width
	}
	if config.Height == 0 {
		config.Height = 64 // Default height
	}
	if config.VCCMode == 0 {
		config.VCCMode = VCC_SWITCH_CAP
	}

	bufferSize := int(config.Width) * int(config.Height) / 8

	d := &display{
		bus:     bus,
		address: address,
		width:   config.Width,
		height:  config.Height,
		vccMode: config.VCCMode,
		buffer:  make([]byte, bufferSize),
		bufSize: bufferSize,
	}

	// Initialize the display
	if err := d.init(d.width, d.height); err != nil {
		return nil, err
	}

	return d, nil
}

// init initializes the display
func (d *display) init(width, height uint8) error {
	// Determine VCC-dependent settings
	var chargePump, contrast, preCharge uint8
	if d.vccMode == VCC_EXTERNAL {
		chargePump = CHARGE_PUMP_SETTING_OFF
		contrast = 0x9F
		preCharge = 0x22
	} else {
		chargePump = CHARGE_PUMP_SETTING_ON
		contrast = 0xCF
		preCharge = 0xF1
	}

	var comPins uint8
	if height == 32 {
		comPins = 0x02
	} else {
		comPins = 0x12
	}

	// Build initialization sequence in pre-allocated buffer
	cmdSeq := d.cmdBuf[:0]
	cmdSeq = append(cmdSeq,
		SET_DISPLAY_OFF,
		SET_DISPLAY_CLOCK_DIVIDE_RATIO, 0x80,
		SET_MULTIPLEX_RATIO, uint8(height-1),
		SET_DISPLAY_OFFSET, 0x00,
		SET_START_LINE|0x00,
		CHARGE_PUMP_SETTING, chargePump,
		SET_MEMORY_ADDRESSING_MODE, horizontalAddressingMode,

		SET_SEGMENT_REMAP|0x01,
		SET_COM_OUTPUT_SCAN_DIRECTION_DEC,

		SET_COM_PINS, comPins,
		SET_CONTRAST, contrast,

		SET_PRE_CHARGE_PERIOD, preCharge,
		SET_VCOM_DESELECT_LEVEL, 0x20,
		DISPLAY_ALL_ON_RESUME,
		SET_NORMAL_DISPLAY,
		DEACTIVATE_SCROLL,
		SET_DISPLAY_ON,
	)

	return d.bus.WriteRegister(d.address, CONTROL_CMD_STREAM, cmdSeq)
}

// * ----- Getter methods -----

// Size returns the display dimensions as uint8 for interface compatibility
func (d *display) Size() (width, height uint8) {
	return d.width, d.height
}

// BufferSize returns the size of the display buffer
func (d *display) BufferSize() int {
	return d.bufSize
}

// Buffer returns the display buffer
func (d *display) Buffer() []byte {
	return d.buffer
}

// * ----- Display methods -----

// ClearBuffer clears the display buffer
func (d *display) ClearBuffer() {
	for i := range d.buffer {
		d.buffer[i] = 0
	}
}

// ClearDisplay clears the buffer and immediately updates the display
func (d *display) ClearDisplay() {
	d.ClearBuffer()
	_ = d.Display()
}

// Command sends a single command byte to the display
func (d *display) Command(cmd byte) error {
	return d.bus.WriteRegister(d.address, CONTROL_CMD_SINGLE, []byte{cmd})
}

// Display sends the current buffer to the display
func (d *display) Display() error {
	// Set addressing window using pre-allocated buffer
	addrSeq := d.addrBuf[:6]
	addrSeq[0] = SET_COLUMN_ADDRESS
	addrSeq[1] = 0x00
	addrSeq[2] = d.width - 1
	addrSeq[3] = SET_PAGE_ADDRESS
	addrSeq[4] = 0x00
	addrSeq[5] = (d.height / 8) - 1

	// Send addressing commands
	if err := d.bus.WriteRegister(d.address, CONTROL_CMD_STREAM, addrSeq); err != nil {
		return err
	}

	return d.bus.WriteRegister(d.address, CONTROL_DATA_STREAM, d.buffer)
}

// SetPixel sets a pixel at the given coordinates
func (d *display) SetPixel(x, y uint8, color bool) {
	if x >= d.width || y >= d.height {
		return
	}

	byteIndex := int(x) + (int(y)/8)*int(d.width)
	bitMask := uint8(1 << (y & 7))

	if color {
		d.buffer[byteIndex] |= bitMask
	} else {
		d.buffer[byteIndex] &^= bitMask
	}
}

// GetPixel gets the state of a pixel at the given coordinates
func (d *display) GetPixel(x, y uint8) bool {
	if x >= d.width || y >= d.height {
		return false
	}

	byteIndex := int(x) + (int(y)/8)*int(d.width)
	bitMask := uint8(1 << (y & 7))

	return d.buffer[byteIndex]&bitMask != 0
}
