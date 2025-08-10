package ssd1306

// AddressMode represents the I2C address configuration for SSD1306 displays.
type AddressMode = byte

const (
	ADDRESS_GND AddressMode = 0x3C // SA0 connected to GND (default address)
	ADDRESS_VCC AddressMode = 0x3D // SA0 connected to VCC (alternate address)
)

// -----

// VCCMode represents the voltage supply configuration for SSD1306 displays.
type VCCMode byte

const (
	VCC_EXTERNAL   VCCMode = 0x01 // External VCC supply
	VCC_SWITCH_CAP VCCMode = 0x02 // Internal charge pump (default, most common)
)

// -----

// CommandMode represents the control byte modes for I2C communication.
// These determine how command and data bytes are interpreted by the display.
type CommandMode = byte

const (
	CONTROL_CMD_SINGLE  CommandMode = 0x80 // Single command byte (Co=1, D/C#=0)
	CONTROL_CMD_STREAM  CommandMode = 0x00 // Command stream mode (Co=0, D/C#=0)
	CONTROL_DATA_STREAM CommandMode = 0x40 // Data stream mode (D/C#=1)
	CONTROL_DATA_SINGLE CommandMode = 0xC0 // Single data byte (Co=1, D/C#=1)
)

// -----

// Memory addressing modes for the SSD1306 display controller.
// These control how the display buffer is accessed and updated.
const (
	horizontalAddressingMode = 0x00 // Horizontal addressing (recommended)
	verticalAddressingMode   = 0x01 // Vertical addressing
	pageAddressingMode       = 0x02 // Page addressing (legacy)
)
