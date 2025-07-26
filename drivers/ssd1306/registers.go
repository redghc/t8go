package ssd1306

// SSD1306 command constants
const (
	// 1. Fundamental Command
	SET_CONTRAST          byte = 0x81
	DISPLAY_ALL_ON_RESUME byte = 0xA4
	DISPLAY_ALL_ON        byte = 0xA5
	SET_NORMAL_DISPLAY    byte = 0xA6
	SET_INVERT_DISPLAY    byte = 0xA7
	SET_DISPLAY_OFF       byte = 0xAE
	SET_DISPLAY_ON        byte = 0xAF

	// 2. Scrolling Command
	RIGHT_HORIZONTAL_SCROLL              byte = 0x26
	LEFT_HORIZONTAL_SCROLL               byte = 0x27
	VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL byte = 0x29
	VERTICAL_AND_LEFT_HORIZONTAL_SCROLL  byte = 0x2A
	DEACTIVATE_SCROLL                    byte = 0x2E
	ACTIVATE_SCROLL                      byte = 0x2F
	SET_VERTICAL_SCROLL_AREA             byte = 0xA3

	// 3. Addressing Setting Command
	SET_LOWER_COLUMN           byte = 0x00
	SET_HIGHER_COLUMN          byte = 0x10
	SET_MEMORY_ADDRESSING_MODE byte = 0x20
	SET_COLUMN_ADDRESS         byte = 0x21
	SET_PAGE_ADDRESS           byte = 0x22

	// 4. Hardware Configuration (Panel resolution & layout related) Command
	SET_START_LINE                    byte = 0x40
	SET_SEGMENT_REMAP_RESET           byte = 0xA0
	SET_SEGMENT_REMAP                 byte = 0xA1
	SET_MULTIPLEX_RATIO               byte = 0xA8
	SET_COM_OUTPUT_SCAN_DIRECTION_INC byte = 0xC0
	SET_COM_OUTPUT_SCAN_DIRECTION_DEC byte = 0xC8
	SET_DISPLAY_OFFSET                byte = 0xD3
	SET_COM_PINS                      byte = 0xDA

	// 5. Timing & Driving Scheme Setting Command
	SET_DISPLAY_CLOCK_DIVIDE_RATIO byte = 0xD5
	SET_PRE_CHARGE_PERIOD          byte = 0xD9
	SET_VCOM_DESELECT_LEVEL        byte = 0xDB
	NOP                            byte = 0xE3

	// 6. Charge Pump Command
	CHARGE_PUMP_SETTING     byte = 0x8D
	CHARGE_PUMP_SETTING_ON  byte = 0x14
	CHARGE_PUMP_SETTING_OFF byte = 0x10
)
