package ssd1306

// SSD1306 command constants.
const (
	// 1. Fundamental Command.
	SetContrast        byte = 0x81
	DisplayAllOnResume byte = 0xA4
	DisplayAllOn       byte = 0xA5
	SetNormalDisplay   byte = 0xA6
	SetInvertDisplay   byte = 0xA7
	SetDisplayOff      byte = 0xAE
	SetDisplayOn       byte = 0xAF

	// 2. Scrolling Command.
	RightHorizontalScroll            byte = 0x26
	LeftHorizontalScroll             byte = 0x27
	VerticalAndRightHorizontalScroll byte = 0x29
	VerticalAndLeftHorizontalScroll  byte = 0x2A
	DeactivateScroll                 byte = 0x2E
	ActivateScroll                   byte = 0x2F
	SetVerticalScrollArea            byte = 0xA3

	// 3. Addressing Setting Command.
	SetLowerColumn          byte = 0x00
	SetHigherColumn         byte = 0x10
	SetMemoryAddressingMode byte = 0x20
	SetColumnAddress        byte = 0x21
	SetPageAddress          byte = 0x22

	// 4. Hardware Configuration (Panel resolution & layout related) Command.
	SetStartLine                 byte = 0x40
	SetSegmentRemapReset         byte = 0xA0
	SetSegmentRemap              byte = 0xA1
	SetMultiplexRatio            byte = 0xA8
	SetCOMOutputScanDirectionInc byte = 0xC0
	SetCOMOutputScanDirectionDec byte = 0xC8
	SetDisplayOffset             byte = 0xD3
	SetCOMPins                   byte = 0xDA

	// 5. Timing & Driving Scheme Setting Command.
	SetDisplayClockDivideRatio byte = 0xD5
	SetPreChargePeriod         byte = 0xD9
	SetVCOMDeselectLevel       byte = 0xDB
	NOP                        byte = 0xE3

	// 6. Charge Pump Command.
	ChargePumpSetting    byte = 0x8D
	ChargePumpSettingOn  byte = 0x14
	ChargePumpSettingOff byte = 0x10
)
