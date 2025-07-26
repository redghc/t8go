package ssd1306

// I2C addresses for SSD1306 displays
type AddressMode = byte

const (
	ADDRESS_GND AddressMode = 0x3C // SA0 to GND mode
	ADDRESS_VCC AddressMode = 0x3D // SA0 to VCC mode
)

// -----

// VCC modes for SSD1306 displays
type VCCMode byte

const (
	VCC_EXTERNAL   VCCMode = 0x01 // External VCC
	VCC_SWITCH_CAP VCCMode = 0x02 // Internal charge pump
)

// -----

// Command modes for SSD1306
type CommandMode = byte

const (
	CONTROL_CMD_SINGLE  CommandMode = 0x00 // Single command byte
	CONTROL_CMD_STREAM  CommandMode = 0x00 // Command stream
	CONTROL_DATA_STREAM CommandMode = 0x40 // Data stream
)

// -----

// Memory addressing modes
const (
	horizontalAddressingMode = 0x00
	verticalAddressingMode   = 0x01
	pageAddressingMode       = 0x02
)
