// Package ssd1306 provides a driver for SSD1306 OLED displays.
// It implements the t8go.Display interface and supports I2C communication
// with configurable display dimensions and VCC modes.
package ssd1306

import (
	"machine"

	"github.com/redghc/t8go"
)

// * ----- Definitions -----

// Config holds the configuration parameters for an SSD1306 display.
type Config struct {
	Width   uint8   // Display width in pixels (default: 128)
	Height  uint8   // Display height in pixels (default: 64)
	VCCMode VCCMode // VCC generation mode (default: VCC_SWITCH_CAP)
}

// display represents an SSD1306 OLED display instance.
type display struct {
	bus     *machine.I2C // I2C bus interface
	address AddressMode  // I2C device address

	width     uint8   // Display width in pixels
	height    uint8   // Display height in pixels
	pageCount uint8   // Number of 8-pixel high pages (height / 8)
	stride    int     // Bytes per page (equals width)
	vccMode   VCCMode // VCC generation mode

	buffer  []byte // Display buffer
	bufSize int    // Buffer size in bytes

	// Pre-allocated command buffers to avoid allocations
	cmdBuf  [32]byte // Command buffer for sending display commands
	addrBuf [6]byte  // Address buffer for I2C operations
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
		bus:       bus,
		address:   address,
		width:     config.Width,
		height:    config.Height,
		pageCount: config.Height / 8,
		stride:    int(config.Width),
		vccMode:   config.VCCMode,
		buffer:    make([]byte, bufferSize),
		bufSize:   bufferSize,
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
		SET_MULTIPLEX_RATIO, height-1,
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
	return d.CommandStream(cmdSeq...)
}

// * ----- Getter methods -----

// Size returns the display dimensions as uint16 for interface compatibility
func (d *display) Size() (width, height uint16) {
	return uint16(d.width), uint16(d.height)
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

// ClearBuffer zeros the internal backbuffer.
func (d *display) ClearBuffer() {
	clear(d.buffer)
}

// ClearDisplay clears the buffer and flushes to the panel.
func (d *display) ClearDisplay() {
	d.ClearBuffer()
	_ = d.Display()
}

// Command sends a single command byte to the display
func (d *display) Command(cmd byte) error {
	return d.bus.WriteRegister(d.address, CONTROL_CMD_SINGLE, []byte{cmd})
}

// CommandStream writes multiple command bytes with a single control prefix.
func (d *display) CommandStream(cmds ...byte) error {
	return d.bus.WriteRegister(d.address, CONTROL_CMD_STREAM, cmds)
}

// Display flushes the full backbuffer to the panel using horizontal addressing.
func (d *display) Display() error {
	// Set addressing window to full screen.
	addrSeq := d.addrBuf[:6]
	addrSeq[0] = SET_COLUMN_ADDRESS
	addrSeq[1] = 0x00
	addrSeq[2] = d.width - 1
	addrSeq[3] = SET_PAGE_ADDRESS
	addrSeq[4] = 0x00
	addrSeq[5] = d.pageCount - 1

	// Send addressing commands
	if err := d.CommandStream(addrSeq...); err != nil {
		return err
	}

	return d.bus.WriteRegister(d.address, CONTROL_DATA_STREAM, d.buffer)
}

// DisplayRegion updates a rectangular region aligned to page rows.
// It reduces I²C traffic when drawing incrementally.
func (d *display) DisplayRegion(x0, y0, x1, y1 int) error {
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
	if x1 >= int(d.width) {
		x1 = int(d.width) - 1
	}
	if y1 >= int(d.height) {
		y1 = int(d.height) - 1
	}

	startPage := uint8(y0 >> 3)
	endPage := uint8(y1 >> 3)

	// Setup column/page window
	addr := d.addrBuf[:6]
	addr[0] = SET_COLUMN_ADDRESS
	addr[1] = byte(x0)
	addr[2] = byte(x1)
	addr[3] = SET_PAGE_ADDRESS
	addr[4] = startPage
	addr[5] = endPage

	if err := d.CommandStream(addr...); err != nil {
		return err
	}

	// Stream contiguous chunks page by page
	for page := int(startPage); page <= int(endPage); page++ {
		rowOffset := page * d.stride
		start := rowOffset + x0
		end := rowOffset + x1 + 1
		if err := d.bus.WriteRegister(d.address, CONTROL_DATA_STREAM, d.buffer[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// SetPixel sets a pixel at the given coordinates
// Out-of-bounds are safely ignored.
// color=true -> set, color=false -> clear.
func (d *display) SetPixel(x, y int16, color bool) {
	if x < 0 || y < 0 || x >= int16(d.width) || y >= int16(d.height) {
		return
	}

	byteIndex := int(x) + (int(y)>>3)*d.stride
	bitMask := uint8(1 << (y & 7))

	if color {
		d.buffer[byteIndex] |= bitMask
	} else {
		d.buffer[byteIndex] &^= bitMask
	}
}

// GetPixel returns the current pixel state from the backbuffer.
func (d *display) GetPixel(x, y uint8) bool {
	if x < 0 || y < 0 || x >= d.width || y >= d.height {
		return false
	}

	byteIndex := int(x) + (int(y)>>3)*d.stride
	bitMask := uint8(1 << (y & 7))

	return (d.buffer[byteIndex] & bitMask) != 0
}
