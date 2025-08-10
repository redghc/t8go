package bitmap

import (
	"encoding/binary"
	"os"

	"github.com/redghc/t8go"
)

// display implements the t8go.Display interface for bitmap output
type display struct {
	width    uint16
	height   uint16
	filename string
	buffer   []byte
	bufSize  int
}

var _ t8go.Display = &display{}

// New creates a new bitmap display instance
func New(config Config) (t8go.Display, error) {
	if config.Width == 0 || config.Height == 0 {
		return nil, ErrInvalidDimensions
	}

	if config.Filename == "" {
		config.Filename = "display.bmp"
	}

	bufSize := int(config.Width) * int(config.Height) / 8
	if int(config.Height)%8 != 0 {
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

// Size returns the display dimensions
func (d *display) Size() (width, height uint16) {
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

// ClearBuffer clears the display buffer
func (d *display) ClearBuffer() {
	for i := range d.buffer {
		d.buffer[i] = 0
	}
}

// ClearDisplay clears the buffer and saves an empty bitmap
func (d *display) ClearDisplay() {
	d.ClearBuffer()
	_ = d.Display()
}

// Command is a no-op for bitmap display (maintains interface compatibility)
func (d *display) Command(cmd byte) error {
	// No commands needed for bitmap output
	return nil
}

// Display saves the current buffer as a BMP file
func (d *display) Display() error {
	return d.saveBMP()
}

// SetPixel sets a pixel at the given coordinates
func (d *display) SetPixel(x, y int16, color bool) {
	if x < 0 || y < 0 || x >= int16(d.width) || y >= int16(d.height) {
		return
	}

	// Use same buffer organization as SSD1306 for compatibility
	byteIndex := int(x) + (int(y)/8)*int(d.width)
	bitMask := uint8(1 << (y & 7))

	if byteIndex >= len(d.buffer) {
		return
	}

	if color {
		d.buffer[byteIndex] |= bitMask
	} else {
		d.buffer[byteIndex] &^= bitMask
	}
}

// GetPixel gets the state of a pixel at the given coordinates
func (d *display) GetPixel(x, y uint8) bool {
	if x >= uint8(d.width) || y >= uint8(d.height) {
		return false
	}

	byteIndex := int(x) + (int(y)/8)*int(d.width)
	bitMask := uint8(1 << (y & 7))

	if byteIndex >= len(d.buffer) {
		return false
	}

	return d.buffer[byteIndex]&bitMask != 0
}

// saveBMP saves the display buffer as a BMP file
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
	rowSize := (width + 31) / 32 * 4
	imageSize := rowSize * height
	fileSize := 54 + imageSize // 54 bytes for headers + image data

	// BMP File Header (14 bytes)
	bmpHeader := []byte{
		'B', 'M', // Signature
		0, 0, 0, 0, // File size (to be filled)
		0, 0, // Reserved
		0, 0, // Reserved
		54, 0, 0, 0, // Offset to pixel data
	}
	binary.LittleEndian.PutUint32(bmpHeader[2:6], uint32(fileSize))

	// BMP Info Header (40 bytes)
	infoHeader := make([]byte, 40)
	binary.LittleEndian.PutUint32(infoHeader[0:4], 40)                  // Header size
	binary.LittleEndian.PutUint32(infoHeader[4:8], uint32(width))       // Width
	binary.LittleEndian.PutUint32(infoHeader[8:12], uint32(height))     // Height
	binary.LittleEndian.PutUint16(infoHeader[12:14], 1)                 // Planes
	binary.LittleEndian.PutUint16(infoHeader[14:16], 1)                 // Bits per pixel (monochrome)
	binary.LittleEndian.PutUint32(infoHeader[16:20], 0)                 // Compression
	binary.LittleEndian.PutUint32(infoHeader[20:24], uint32(imageSize)) // Image size
	binary.LittleEndian.PutUint32(infoHeader[24:28], 2835)              // X pixels per meter
	binary.LittleEndian.PutUint32(infoHeader[28:32], 2835)              // Y pixels per meter
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

	for y := int16(0); y < int16(height); y++ {
		bmpRow := int(height) - 1 - int(y) // Flip vertically for BMP format
		rowStart := int64(bmpRow) * int64(rowSize)

		for x := int16(0); x < int16(width); x += 8 {
			// Get the source byte from our buffer
			srcByteIndex := int(x/8) + (int(y)/8)*int(d.width)
			var srcByte uint8
			if srcByteIndex < len(d.buffer) {
				// Extract the relevant bits for this row from the source byte
				bitOffset := y & 7
				srcByte = (d.buffer[srcByteIndex] >> bitOffset) & 1
				if srcByte != 0 {
					srcByte = 0xFF
				}
			}

			// Convert 8 pixels at a time
			dstByteIndex := rowStart + int64(x/8)
			if dstByteIndex < int64(len(pixelData)) {
				var dstByte uint8
				for bit := 0; bit < 8 && x+int16(bit) < int16(width); bit++ {
					if d.GetPixel(uint8(x+int16(bit)), uint8(y)) {
						dstByte |= 1 << (7 - bit) // MSB first for BMP
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
