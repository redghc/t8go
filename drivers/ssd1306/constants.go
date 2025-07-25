package ssd1306

// SSD1306 constants
const (
	Address_128_32 AddressMode = 0x3D // I2C address for 128x32 SSD1306
	Address_128_64 AddressMode = 0x3C // I2C address for 128x64 SSD1306

	// 1. Fundamental Command
	SET_CONTRAST          = 0x81
	DISPLAY_ALL_ON_RESUME = 0xA4
	DISPLAY_ALL_ON        = 0xA5
	SET_NORMAL_DISPLAY    = 0xA6
	SET_INVERT_DISPLAY    = 0xA7
	SET_DISPLAY_OFF       = 0xAE
	SET_DISPLAY_ON        = 0xAF

	// 2. Scrolling Command
	RIGHT_HORIZONTAL_SCROLL              = 0x26
	LEFT_HORIZONTAL_SCROLL               = 0x27
	VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL = 0x29
	VERTICAL_AND_LEFT_HORIZONTAL_SCROLL  = 0x2A
	DEACTIVATE_SCROLL                    = 0x2E
	ACTIVATE_SCROLL                      = 0x2F
	SET_VERTICAL_SCROLL_AREA             = 0xA3

	// 3. Addressing Setting Command
	SET_LOWER_COLUMN           = 0x00
	SET_HIGHER_COLUMN          = 0x10
	SET_MEMORY_ADDRESSING_MODE = 0x20
	SET_COLUMN_ADDRESS         = 0x21
	SET_PAGE_ADDRESS           = 0x22
	// Pages?

	// 4. Hardware Configuration (Panel resolution & layout related) Command
	SET_START_LINE                    = 0x40
	SET_SEGMENT_REMAP_RESET           = 0xA0
	SET_SEGMENT_REMAP                 = 0xA1
	SET_MULTIPLEX_RATIO               = 0xA8
	SET_COM_OUTPUT_SCAN_DIRECTION_INC = 0xC0
	SET_COM_OUTPUT_SCAN_DIRECTION_DEC = 0xC8
	SET_DISPLAY_OFFSET                = 0xD3
	SET_COM_PINS                      = 0xDA

	// 5. Timing & Driving Scheme Setting Command
	SET_DISPLAY_CLOCK_DIVIDE_RATIO = 0xD5
	SET_PRE_CHARGE_PERIOD          = 0xD9
	SET_VCOM_DESELECT_LEVEL        = 0xDB
	NOP                            = 0xE3

	// 1. Charge Pump Command
	CHARGE_PUMP_SETTING = 0x8D

	EXTERNAL_VCC   = 0x01
	SWITCH_CAP_VCC = 0x02

	CONTROL_CMD_SINGLE = 0x00 // Control byte for single command (Co = 0, D/C# = 0)
	CONTROL_CMD_STREAM = 0x40 // Control byte for command stream (Co = 0, D/C# = 1)
)
