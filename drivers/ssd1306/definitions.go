package ssd1306

type AddressMode = byte

const (
	Address_128_32 AddressMode = 0x3D // I2C address for 128x32 SSD1306
	Address_128_64 AddressMode = 0x3C // I2C address for 128x64 SSD1306
)

// -----

type VCCMode byte

const (
	// VCC Mode
	VCC_MODE_EXTERNAL   = 0x01
	VCC_MODE_SWITCH_CAP = 0x02
)

// -----

type CommandMode = byte

const (
	CONTROL_CMD_SINGLE CommandMode = 0x00 // Control byte for single command (Co byte = 0, D/C# byte = 0)
	CONTROL_CMD_STREAM CommandMode = 0x40 // Control byte for command stream (Co byte = 0, D/C# byte = 1)
)
