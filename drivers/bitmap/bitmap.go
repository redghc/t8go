// Package bitmap provides a bitmap file output driver for t8go graphics.
// It implements the t8go.Display interface and allows rendering graphics
// directly to bitmap (.bmp) files for testing and visualization purposes.
package bitmap

import (
	"encoding/binary"
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
// (zero width or height).
func New(config Config) (t8go.IDisplay, error) {
	if config.Width == 0 || config.Height == 0 {
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
func (d *display) Command(cmd byte) error {
	// No commands needed for bitmap output
	return nil
}

// Display saves the current buffer as a BMP file.
// Returns an error if the file cannot be created or written.
func (d *display) Display() error {
	return d.saveBMP()
}

// SetPixel sets a pixel at the given coordinates.
func (d *display) SetPixel(x, y int16, color bool) {
	if x < 0 || y < 0 || x >= int16(d.width) || y >= int16(d.height) {
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

// GetPixel gets the state of a pixel at the given coordinates.
func (d *display) GetPixel(x, y uint8) bool {
	if x >= uint8(d.width) || y >= uint8(d.height) {
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
func (d *display) DisplayRegion(x0, y0, x1, y1 int) error {
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

	// BMP file structure
	width := int32(d.width)
	height := int32(d.height)

	// Calculate padding for BMP format (rows must be multiple of 4 bytes)
	rowSize := (width + bmpRowAlignment - 1) / bmpRowAlignment * (bmpRowAlignment / bitsPerPixelRow)
	imageSize := rowSize * height
	fileSize := bmpFileHeaderSize + imageSize

	// BMP File Header (14 bytes)
	bmpHeader := []byte{
		'B', 'M', // Signature
		0, 0, 0, 0, // File size (to be filled)
		0, 0, // Reserved
		0, 0, // Reserved
		bmpFileHeaderSize, 0, 0, 0, // Offset to pixel data
	}
	binary.LittleEndian.PutUint32(bmpHeader[2:6], uint32(fileSize))

	// BMP Info Header (40 bytes)
	infoHeader := make([]byte, bmpInfoHeaderSize)
	binary.LittleEndian.PutUint32(infoHeader[0:4], bmpInfoHeaderSize)   // Header size
	binary.LittleEndian.PutUint32(infoHeader[4:8], uint32(width))       // Width
	binary.LittleEndian.PutUint32(infoHeader[8:12], uint32(height))     // Height
	binary.LittleEndian.PutUint16(infoHeader[12:14], 1)                 // Planes
	binary.LittleEndian.PutUint16(infoHeader[14:16], 1)                 // Bits per pixel (monochrome)
	binary.LittleEndian.PutUint32(infoHeader[16:20], 0)                 // Compression
	binary.LittleEndian.PutUint32(infoHeader[20:24], uint32(imageSize)) // Image size
	binary.LittleEndian.PutUint32(infoHeader[24:28], bmpPixelsPerMeter) // X pixels per meter
	binary.LittleEndian.PutUint32(infoHeader[28:32], bmpPixelsPerMeter) // Y pixels per meter
	binary.LittleEndian.PutUint32(infoHeader[32:36], 2)                 // Colors used (black and white)
	binary.LittleEndian.PutUint32(infoHeader[36:40], 0)                 // Important colors

	// Write headers
	if _, err := file.Write(bmpHeader); err != nil {
		return ErrFileWrite
	}
	if _, err := file.Write(infoHeader); err != nil {
		return ErrFileWrite
	}

	// Color palette for monochrome BMP (8 bytes)
	colorPalette := []byte{
		0x00, 0x00, 0x00, 0x00, // Black (BGRA)
		0xFF, 0xFF, 0xFF, 0x00, // White (BGRA)
	}
	if _, err := file.Write(colorPalette); err != nil {
		return ErrFileWrite
	}

	// Convert buffer to BMP pixel data
	// BMP rows are stored bottom-to-top, so we need to flip the image
	pixelData := make([]byte, imageSize)

	for y := range int(height) {
		bmpRow := int(height) - 1 - y // Flip vertically for BMP format
		rowStart := int64(bmpRow) * int64(rowSize)

		for x := int16(0); x < int16(width); x += bitsPerPixelRow {
			// Convert 8 pixels at a time
			dstByteIndex := rowStart + int64(x/bitsPerPixelRow)
			if dstByteIndex < int64(len(pixelData)) {
				var dstByte uint8
				for bit := 0; bit < bitsPerPixelRow && x+int16(bit) < int16(width); bit++ {
					if d.GetPixel(uint8(x+int16(bit)), uint8(y)) {
						dstByte |= 1 << (bitOffsetMask - bit) // MSB first for BMP
					}
				}
				pixelData[dstByteIndex] = dstByte
			}
		}
	}

	// Write pixel data
	if _, err := file.Write(pixelData); err != nil {
		return ErrFileWrite
	}

	return nil
}
