// Package ssd1306 provides a driver for SSD1306 OLED displays.
// It implements the t8go.Display interface and supports I2C communication
// with configurable display dimensions and VCC modes.
package ssd1306

import (
	"machine"

	"github.com/redghc/t8go"
	"github.com/redghc/t8go/helpers"
)

// Buffer geometry: pixels are packed 8 per byte (one "page").
const (
	bitsPerPixelRow = 8    // pixels packed per buffer byte (page height)
	bitOffsetMask   = 7    // bitsPerPixelRow - 1; masks the bit position within a byte
	fullPageMask    = 0xFF // all bits set in a page byte
)

// SSD1306 panel variants.
const (
	panelHeight32 = 32 // 128x32 panel uses an alternate COM-pin configuration
)

// SSD1306 initialization register values (see datasheet).
const (
	displayClockSuggested = 0x80 // display clock divide ratio (datasheet suggested)
	displayOffsetNone     = 0x00 // no vertical display offset
	startLineOffsetNone   = 0x00 // start-line register set to 0
	segmentRemapEnable    = 0x01 // SetSegmentRemap flag bit
	vcomDeselectLevel     = 0x20 // VCOMH deselect level (~0.77 * Vcc)
)

// * ----- Definitions -----

// Config holds the configuration parameters for an SSD1306 display.
type Config struct {
	Width   uint8   // Display width in pixels (default: 128)
	Height  uint8   // Display height in pixels (default: 64)
	VCCMode VCCMode // VCC generation mode (default: VCCSwitchCap)
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

var _ t8go.IDisplay = &display{}

// * ----- Constructors -----

// NewI2C creates a new SSD1306 display instance using I2C communication.
// Returns the display, or a non-nil `err` if the I2C bus is nil or the
// initialization sequence fails.
func NewI2C(bus *machine.I2C, address AddressMode, config Config) (t8go.IDisplay, error) {
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
		config.VCCMode = VCCSwitchCap
	}

	bufferSize := int(config.Width) * int(config.Height) / bitsPerPixelRow

	d := &display{
		bus:       bus,
		address:   address,
		width:     config.Width,
		height:    config.Height,
		pageCount: config.Height / bitsPerPixelRow,
		stride:    int(config.Width),
		vccMode:   config.VCCMode,
		buffer:    make([]byte, bufferSize),
		bufSize:   bufferSize,
	}

	// Initialize the display
	if err := d.init(d.height); err != nil {
		return nil, err
	}

	return d, nil
}

// init initializes the display.
// Returns an error if any command in the initialization sequence fails to send.
func (d *display) init(height uint8) error {
	// Determine VCC-dependent settings
	var chargePump, contrast, preCharge uint8
	if d.vccMode == VCCExternal {
		chargePump = ChargePumpSettingOff
		contrast = 0x9F
		preCharge = 0x22
	} else {
		chargePump = ChargePumpSettingOn
		contrast = 0xCF
		preCharge = 0xF1
	}

	var comPins uint8
	if height == panelHeight32 {
		comPins = 0x02
	} else {
		comPins = 0x12
	}

	// Build initialization sequence in pre-allocated buffer
	cmdSeq := d.cmdBuf[:0]
	cmdSeq = append(cmdSeq,
		SetDisplayOff,
		SetDisplayClockDivideRatio, displayClockSuggested,
		SetMultiplexRatio, height-1,
		SetDisplayOffset, displayOffsetNone,
		SetStartLine|startLineOffsetNone,
		ChargePumpSetting, chargePump,
		SetMemoryAddressingMode, horizontalAddressingMode,

		SetSegmentRemap|segmentRemapEnable,
		SetCOMOutputScanDirectionDec,

		SetCOMPins, comPins,
		SetContrast, contrast,

		SetPreChargePeriod, preCharge,
		SetVCOMDeselectLevel, vcomDeselectLevel,
		DisplayAllOnResume,
		SetNormalDisplay,
		DeactivateScroll,
		SetDisplayOn,
	)
	return d.CommandStream(cmdSeq...)
}

// * ----- Getter methods -----

// Size returns the display dimensions in pixels: `width` and `height`.
func (d *display) Size() (uint16, uint16) {
	return uint16(d.width), uint16(d.height)
}

// BufferSize returns the size of the display buffer.
func (d *display) BufferSize() int {
	return d.bufSize
}

// Buffer returns the display buffer.
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

// Command sends a single command byte to the display.
// Returns an error if the I2C write fails.
func (d *display) Command(cmd byte) error {
	d.cmdBuf[0] = cmd
	return d.bus.WriteRegister(d.address, ControlCMDSingle, d.cmdBuf[:1])
}

// CommandStream writes multiple command bytes with a single control prefix.
// Returns an error if the I2C write fails.
func (d *display) CommandStream(cmds ...byte) error {
	return d.bus.WriteRegister(d.address, ControlCMDStream, cmds)
}

// Display flushes the full backbuffer to the panel using horizontal addressing.
// Returns an error if any I2C write during the flush fails.
func (d *display) Display() error {
	// Set addressing window to full screen.
	addrSeq := d.addrBuf[:6]
	addrSeq[0] = SetColumnAddress
	addrSeq[1] = 0x00
	addrSeq[2] = d.width - 1
	addrSeq[3] = SetPageAddress
	addrSeq[4] = 0x00
	addrSeq[5] = d.pageCount - 1

	// Send addressing commands
	if err := d.CommandStream(addrSeq...); err != nil {
		return err
	}

	return d.bus.WriteRegister(d.address, ControlDataStream, d.buffer)
}

// DisplayRegion updates a rectangular region aligned to page rows.
// It reduces I²C traffic when drawing incrementally.
// Returns an error if any I2C write during the region flush fails.
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

	startPage := uint8(y0 / bitsPerPixelRow)
	endPage := uint8(y1 / bitsPerPixelRow)

	// Setup column/page window
	addr := d.addrBuf[:6]
	addr[0] = SetColumnAddress
	addr[1] = byte(x0)
	addr[2] = byte(x1)
	addr[3] = SetPageAddress
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
		if err := d.bus.WriteRegister(d.address, ControlDataStream, d.buffer[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// FillHRun sets or clears a horizontal run of pixels from x0 to x1 inclusive at row y.
// Out-of-bounds coordinates are clipped, and runs outside the display are ignored.
func (d *display) FillHRun(x0, x1, y int16, on bool) {
	if y < 0 || y >= int16(d.height) {
		return
	}
	x0, x1, ok := helpers.ClipSpan(x0, x1, int16(d.width))
	if !ok {
		return
	}

	mask := uint8(1 << (y & bitOffsetMask))
	base := int(y/bitsPerPixelRow) * d.stride
	for c := int(x0); c <= int(x1); c++ {
		if on {
			d.buffer[base+c] |= mask
		} else {
			d.buffer[base+c] &^= mask
		}
	}
}

// FillRect sets or clears a rectangular region with inclusive bounds.
// Out-of-bounds coordinates are clipped, and regions outside the display are ignored.
func (d *display) FillRect(x0, y0, x1, y1 int16, on bool) {
	x0, y0, x1, y1, ok := helpers.ClipRect(x0, y0, x1, y1, int16(d.width), int16(d.height))
	if !ok {
		return
	}

	startPage := int(y0 / bitsPerPixelRow)
	endPage := int(y1 / bitsPerPixelRow)

	for page := startPage; page <= endPage; page++ {
		mask := pageMask(y0, y1, page, startPage, endPage)

		rowOffset := page * d.stride
		for c := int(x0); c <= int(x1); c++ {
			bi := rowOffset + c
			if on {
				if mask == fullPageMask {
					d.buffer[bi] = fullPageMask
				} else {
					d.buffer[bi] |= mask
				}
			} else {
				d.buffer[bi] &^= mask
			}
		}
	}
}

// FillVRun sets or clears a vertical run of pixels from y0 to y1 inclusive at column x.
// Out-of-bounds coordinates are clipped, and runs outside the display are ignored.
// Pages fully covered by the run are written with a single byte store.
func (d *display) FillVRun(x, y0, y1 int16, on bool) {
	if x < 0 || x >= int16(d.width) {
		return
	}
	y0, y1, ok := helpers.ClipSpan(y0, y1, int16(d.height))
	if !ok {
		return
	}

	col := int(x)
	startPage := int(y0 / bitsPerPixelRow)
	endPage := int(y1 / bitsPerPixelRow)

	for page := startPage; page <= endPage; page++ {
		mask := pageMask(y0, y1, page, startPage, endPage)

		bi := page*d.stride + col
		if on {
			if mask == fullPageMask {
				d.buffer[bi] = fullPageMask
			} else {
				d.buffer[bi] |= mask
			}
		} else {
			d.buffer[bi] &^= mask
		}
	}
}

// pageMask returns the bit mask for the pixels of [y0, y1] that fall in `page`,
// where `page` lies within [startPage, endPage]. Bits [lo, hi] within the
// 8-pixel page are set.
func pageMask(y0, y1 int16, page, startPage, endPage int) uint8 {
	var lo, hi uint8
	if page == startPage {
		lo = uint8(y0 & bitOffsetMask)
	}
	if page == endPage {
		hi = uint8(y1 & bitOffsetMask)
	} else {
		hi = bitOffsetMask
	}

	var mask uint8
	for b := lo; b <= hi; b++ {
		mask |= 1 << b
	}
	return mask
}

// SetPixel sets a pixel at the given coordinates
// Out-of-bounds are safely ignored.
// color=true -> set, color=false -> clear.
func (d *display) SetPixel(x, y int16, color bool) {
	if x < 0 || y < 0 || x >= int16(d.width) || y >= int16(d.height) {
		return
	}

	byteIndex := int(x) + (int(y)/bitsPerPixelRow)*d.stride
	bitMask := uint8(1 << (y & bitOffsetMask))

	if color {
		d.buffer[byteIndex] |= bitMask
	} else {
		d.buffer[byteIndex] &^= bitMask
	}
}

// GetPixel returns the current pixel state from the backbuffer.
func (d *display) GetPixel(x, y uint8) bool {
	if x >= d.width || y >= d.height {
		return false
	}

	byteIndex := int(x) + (int(y)/bitsPerPixelRow)*d.stride
	bitMask := uint8(1 << (y & bitOffsetMask))

	return (d.buffer[byteIndex] & bitMask) != 0
}
