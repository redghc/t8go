# t8go - Tiny 8bit Graphics Go

[![Go Version](https://img.shields.io/badge/go-1.24.5+-blue.svg)](https://golang.org/doc/devel/release.html)
[![TinyGo Compatible](https://img.shields.io/badge/tinygo-compatible-brightgreen.svg)](https://tinygo.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Graphics library for monochrome displays

t8go is a lightweight, high-performance graphics library designed for embedded displays and microcontrollers. Built for Go and TinyGo, it provides comprehensive 2D drawing capabilities optimized for resource-constrained environments.

## Features

### Core Drawing Functions

- **Pixel Operations**: Individual pixel drawing and reading
- **Lines**: Horizontal, vertical, and arbitrary lines with Bresenham's algorithm
- **Rectangles**: Outlined and filled rectangles with optional rounded corners
- **Circles & Ellipses**: Perfect circles and ellipses with quadrant selection
- **Arcs**: Partial circles with configurable start/end angles
- **Triangles**: Outlined and filled triangles

### Display Support

- **Generic Interface**: Works with any display implementing the `Display` interface
- **SSD1306 Driver**: Ready-to-use I2C driver for OLED displays
- **Buffer Management**: Efficient display buffer operations

## Installation

```bash
go get github.com/redghc/t8go
```

## Quick Start

```go
package main

import (
    "machine"
    "github.com/redghc/t8go"
    "github.com/redghc/t8go/drivers/ssd1306"
)

func main() {
    // Initialize I2C
    machine.I2C0.Configure(machine.I2CConfig{
        Frequency: machine.TWI_FREQ_400KHZ,
    })

    // Create SSD1306 display (128x64)
    display, err := ssd1306.NewI2C(
        machine.I2C0,
        ssd1306.ADDRESS_GND, // or ADDRESS_VCC
        ssd1306.Config{
            Width:   128,
            Height:  64,
            VCCMode: ssd1306.VCC_SWITCH_CAP,
        },
    )
    if err != nil {
        panic(err)
    }

    // Create T8Go graphics context
    graphics := t8go.New(display)

    // Clear screen
    graphics.ClearDisplay()

    // Draw some shapes
    graphics.DrawLine(0, 0, 127, 63)           // Diagonal line
    graphics.DrawBox(10, 10, 50, 30)           // Rectangle
    graphics.DrawCircle(64, 32, 20, t8go.DrawAll) // Circle

    // Update display
    graphics.Display()
}
```

## API Reference

### Core Methods

#### Display Management

```go
func New(display Display) *T8Go
func (t *T8Go) Display() error
func (t *T8Go) ClearBuffer()
func (t *T8Go) ClearDisplay()
func (t *T8Go) Size() (width, height uint16)
```

#### Pixel Operations

```go
func (t *T8Go) DrawPixel(x, y int16)
func (t *T8Go) SetPixel(x, y int16, on bool)
func (t *T8Go) GetPixel(x, y uint8) bool
```

#### Lines

```go
func (t *T8Go) DrawLine(startX, startY, endX, endY int16)
func (t *T8Go) DrawHLine(originX, originY, length int16)
func (t *T8Go) DrawVLine(originX, originY, length int16)
```

#### Rectangles

```go
func (t *T8Go) DrawBox(originX, originY, width, height int16)
func (t *T8Go) DrawBoxCoords(startX, startY, endX, endY int16)
func (t *T8Go) DrawBoxFill(originX, originY, width, height int16)
func (t *T8Go) DrawBoxFillCoords(startX, startY, endX, endY int16)
func (t *T8Go) DrawRoundBox(originX, originY, width, height, cornerRadius int16)
func (t *T8Go) DrawRoundBoxFill(originX, originY, width, height, cornerRadius int16)
```

#### Circles & Ellipses

```go
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants)
func (t *T8Go) DrawCircleFill(centerX, centerY, radius int16, mask DrawQuadrants)
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)
func (t *T8Go) DrawEllipseFill(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)
```

#### Arcs

```go
func (t *T8Go) DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8)
func (t *T8Go) DrawArcFill(centerX, centerY, radius int16, angleStart, angleEnd uint8)
```

#### Triangles

```go
func (t *T8Go) DrawTriangle(x1, y1, x2, y2, x3, y3 int16)
func (t *T8Go) DrawTriangleFill(x1, y1, x2, y2, x3, y3 int16)
```

### Quadrant Constants

```go
const (
    DrawNone        DrawQuadrants = 0      // Draw entire shape
    DrawTopLeft     DrawQuadrants = 1 << 0
    DrawTopRight    DrawQuadrants = 1 << 1
    DrawBottomRight DrawQuadrants = 1 << 2
    DrawBottomLeft  DrawQuadrants = 1 << 3
    DrawAll                       = DrawTopLeft | DrawTopRight | DrawBottomRight | DrawBottomLeft
)
```

### Angle System

Arcs use a 0-255 angle system:

- `0` = 0° (right/east)
- `64` = 90° (up/north)
- `128` = 180° (left/west)
- `192` = 270° (down/south)
- `255` = 360° (wraps to 0°)

## Supported Hardware

### Displays

- **SSD1306**: 128x64, 128x32 OLED displays via I2C
- **Generic**: Any display implementing the `Display` interface

### Microcontrollers

- **Arduino-compatible**: ESP32, ESP8266, Arduino Uno, etc.
- **ARM Cortex-M**: STM32, nRF52, RP2040, SAMD21/51
- **RISC-V**: ESP32-C3, CH32V, BL602
- **Desktop**: Standard Go runtime for testing/simulation

## Display Interface

To add support for new displays, implement the `Display` interface:

```go
type Display interface {
    Size() (width, height uint16)
    BufferSize() int
    Buffer() []byte

    ClearBuffer()
    ClearDisplay()
    Command(cmd byte) error
    Display() error
    SetPixel(x, y int16, on bool)
    GetPixel(x, y uint8) bool
}
```

## License

This project is licensed under the MIT License.
See the [LICENSE](LICENSE) file for details.

## Inspiration and Design Philosophy

This project is inspired by [u8g2](https://github.com/olikraus/u8g2)
Copyright (c) 2016 olikraus@gmail.com
Licensed under the BSD 2-Clause License:
https://opensource.org/licenses/BSD-2-Clause

t8go is inspired by the [u8g2](https://github.com/olikraus/u8g2) library, a popular graphics library for microcontrollers in C/C++. However, it differs in several key points, being an adapted implementation designed specifically for the TinyGo ecosystem and Go methodologies.

t8go does not include any font implementations or font files from the U8g2 project.
Therefore, no additional font-specific licenses from U8g2 apply to this software.

### Key differences with u8g2:

- **Idiomatic Go Design**: Uses Go-native patterns and conventions
- **TinyGo Integration**: Optimized for TinyGo compiler and microcontrollers
- **Interface-based**: Interface-driven architecture for greater flexibility
- **Memory efficiency**: Memory management adapted to embedded Go constraints

## Acknowledgments

- Inspired by [U8g2 library](https://github.com/olikraus/u8g2) for C/C++
- Bresenham and midpoint algorithms from computer graphics literature
- TinyGo community for embedded Go development

---

**t8go** - Tiny 8bit Graphics Go • _Graphics that fit in tiny spaces_ 🎨
