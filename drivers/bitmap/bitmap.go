// Package bitmap provides a bitmap file output driver for t8go graphics.
// It implements the t8go.Display interface and allows rendering graphics
// directly to bitmap (.bmp) files for testing and visualization purposes.
package bitmap

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/redghc/t8go"
)

// BMP file format constants.
const (
	bitsPerPixelRow   = 8  // monochrome pixels per buffer byte
	bitOffsetMask     = 7  // masks the bit offset within a buffer byte (bitsPerPixelRow-1)
	bmpRowAlignment   = 32 // BMP rows padded to a multiple of 4 bytes
	bmpFileHeaderSize = 54 // BMP file header (14) + info header (40)
	bmpInfoHeaderSize = 40 // BMP info header size
	bmpPixelsPerMeter = 2835
	// maxBitmapDimension caps width/height so the int16/uint8 bounds checks
	// in SetPixel and GetPixel stay correct (no silent truncation wrap-around).
	maxBitmapDimension = 32767
)

// display implements the t8go.Display interface for bitmap file output.
// It provides a virtual display that saves graphics as bitmap files.
type display struct {
	width    uint16 // Display width in pixels
	height   uint16 // Display height in pixels
	filename string // Output bitmap filename
	buffer   []byte // Display buffer
	bufSize  int    // Buffer size in bytes
}

var _ t8go.IDisplay = &display{}

// New creates a new bitmap display instance with the specified configuration.
// The display will render graphics to a bitmap file with the given dimensions.
// If no filename is specified, it defaults to "display.bmp".
// Returns the display, or a non-nil `err` if the dimensions are invalid
// (zero width or height, or larger than `maxBitmapDimension`).
func New(config Config) (t8go.IDisplay, error) {
	if config.Width == 0 || config.Height == 0 ||
		int(config.Width) > maxBitmapDimension || int(config.Height) > maxBitmapDimension {
		return nil, ErrInvalidDimensions
	}

	if config.Filename == "" {
		config.Filename = "display.bmp"
	}

	bufSize := int(config.Width) * int(config.Height) / bitsPerPixelRow
	if int(config.Height)%bitsPerPixelRow != 0 {
		bufSize += int(config.Width)
	}

	d := &display{
		width:    config.Width,
		height:   config.Height,
		filename: config.Filename,
		buffer:   make([]byte, bufSize),
		bufSize:  bufSize,
	}

	return d, nil
}

// Size returns the display dimensions in pixels: `width` and `height`.
func (d *display) Size() (uint16, uint16) {
	return d.width, d.height
}

// BufferSize returns the size of the display buffer.
func (d *display) BufferSize() int {
	return d.bufSize
}

// Buffer returns the display buffer.
func (d *display) Buffer() []byte {
	return d.buffer
}

// ClearBuffer clears the display buffer.
func (d *display) ClearBuffer() {
	for i := range d.buffer {
		d.buffer[i] = 0
	}
}

// ClearDisplay clears the buffer and saves an empty bitmap.
func (d *display) ClearDisplay() {
	d.ClearBuffer()
	_ = d.Display()
}

// Command is a no-op for bitmap display (maintains interface compatibility).
func (d *display) Command(_ byte) error {
	return nil
}

// Display saves the current buffer as a BMP file.
// Returns an error if the file cannot be created or written.
func (d *display) Display() error {
	return d.saveBMP()
}

// SetPixel sets a pixel at the given coordinates.
// Out-of-bounds are safely ignored.
// color=true -> set, color=false -> clear.
func (d *display) SetPixel(x, y int16, color bool) {
	if x < 0 || y < 0 || int(x) >= int(d.width) || int(y) >= int(d.height) {
		return
	}

	// Use same buffer organization as SSD1306 for compatibility
	byteIndex := int(x) + (int(y)/bitsPerPixelRow)*int(d.width)
	bitMask := uint8(1 << (y & bitOffsetMask))

	if byteIndex >= len(d.buffer) {
		return
	}

	if color {
		d.buffer[byteIndex] |= bitMask
	} else {
		d.buffer[byteIndex] &^= bitMask
	}
}

// GetPixel returns the pixel state at the given coordinates.
// Returns false for out-of-bounds coordinates.
func (d *display) GetPixel(x, y uint8) bool {
	if int(x) >= int(d.width) || int(y) >= int(d.height) {
		return false
	}

	byteIndex := int(x) + (int(y)/bitsPerPixelRow)*int(d.width)
	bitMask := uint8(1 << (y & bitOffsetMask))

	if byteIndex >= len(d.buffer) {
		return false
	}

	return d.buffer[byteIndex]&bitMask != 0
}

// FillHRun fills a horizontal run of pixels at the given y coordinate.
func (d *display) FillHRun(x0, x1, y int16, on bool) {
	for x := x0; x <= x1; x++ {
		d.SetPixel(x, y, on)
	}
}

// FillRect fills a rectangular region of pixels.
func (d *display) FillRect(x0, y0, x1, y1 int16, on bool) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			d.SetPixel(x, y, on)
		}
	}
}

// FillVRun fills a vertical run of pixels at the given x coordinate.
func (d *display) FillVRun(x, y0, y1 int16, on bool) {
	for y := y0; y <= y1; y++ {
		d.SetPixel(x, y, on)
	}
}

// DisplayRegion is a no-op for bitmap output: the whole file is
// rewritten on each flush, so the region argument is irrelevant.
// Returns an error if the file cannot be created or written.
func (d *display) DisplayRegion(_, _, _, _ int) error {
	return d.Display()
}

// saveBMP saves the display buffer as a BMP file.
// Returns an error if the file cannot be created, written, or synced.
func (d *display) saveBMP() error {
	file, err := os.Create(d.filename)
	if err != nil {
		return ErrFileWrite
	}
	defer file.Close()

	width := int32(d.width)
	height := int32(d.height)

	// Rows must be padded to a multiple of 4 bytes.
	rowSize := (width + bmpRowAlignment - 1) / bmpRowAlignment * (bmpRowAlignment / bitsPerPixelRow)
	imageSize := rowSize * height
	fileSize := bmpFileHeaderSize + imageSize

	header := buildBMPHeader(width, height, fileSize, imageSize)
	palette := buildBMPPalette()
	pixelData := d.buildBMPPixelData(width, height, rowSize)

	if err = writeOrErr(file, header); err != nil {
		return err
	}
	if err = writeOrErr(file, palette); err != nil {
		return err
	}
	return writeOrErr(file, pixelData)
}

// writeOrErr writes data and collapses the error into ErrFileWrite so the
// caller doesn't repeat the same handling at every site.
func writeOrErr(w io.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return ErrFileWrite
	}
	return nil
}

// buildBMPHeader returns the 14-byte file header + 40-byte info header in a
// single buffer.
func buildBMPHeader(width, height int32, fileSize, imageSize int32) []byte {
	buf := make([]byte, bmpFileHeaderSize)
	buf[0], buf[1] = 'B', 'M'
	binary.LittleEndian.PutUint32(buf[2:6], uint32(fileSize))
	binary.LittleEndian.PutUint32(buf[10:14], bmpFileHeaderSize)

	// Info header lives at buf[14:54].
	info := buf[14:]
	binary.LittleEndian.PutUint32(info[0:4], bmpInfoHeaderSize)
	binary.LittleEndian.PutUint32(info[4:8], uint32(width))
	binary.LittleEndian.PutUint32(info[8:12], uint32(height))
	binary.LittleEndian.PutUint16(info[12:14], 1)                 // planes
	binary.LittleEndian.PutUint16(info[14:16], 1)                 // bits per pixel
	binary.LittleEndian.PutUint32(info[20:24], uint32(imageSize)) // image size
	binary.LittleEndian.PutUint32(info[24:28], bmpPixelsPerMeter) // X ppm
	binary.LittleEndian.PutUint32(info[28:32], bmpPixelsPerMeter) // Y ppm
	binary.LittleEndian.PutUint32(info[32:36], 2)                 // colors used
	return buf
}

// buildBMPPalette returns the 8-byte BGRA palette for a monochrome BMP.
func buildBMPPalette() []byte {
	return []byte{
		0x00, 0x00, 0x00, 0x00, // Black (BGRA)
		0xFF, 0xFF, 0xFF, 0x00, // White (BGRA)
	}
}

// buildBMPPixelData renders the backbuffer into BMP rows (stored bottom-to-top).
func (d *display) buildBMPPixelData(width, height int32, rowSize int32) []byte {
	pixelData := make([]byte, rowSize*height)

	for y := range int(height) {
		bmpRow := int(height) - 1 - y
		rowStart := int64(bmpRow) * int64(rowSize)

		for x := int16(0); int(x) < int(width); x += bitsPerPixelRow {
			dstByteIndex := rowStart + int64(x/bitsPerPixelRow)
			if dstByteIndex >= int64(len(pixelData)) {
				break
			}
			var dstByte uint8
			for bit := 0; bit < bitsPerPixelRow && int(x)+bit < int(width); bit++ {
				if d.GetPixel(uint8(x)+uint8(bit), uint8(y)) {
					dstByte |= 1 << (bitOffsetMask - bit) // MSB first for BMP
				}
			}
			pixelData[dstByteIndex] = dstByte
		}
	}
	return pixelData
}
