# t8go - Tiny 8bit Graphics Go

[![Go Version](https://img.shields.io/badge/go-1.24.5+-blue.svg)](https://golang.org/doc/devel/release.html)
[![TinyGo Compatible](https://img.shields.io/badge/tinygo-compatible-brightgreen.svg)](https://tinygo.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/redghc/t8go.svg)](https://pkg.go.dev/github.com/redghc/t8go)

## Graphics library for monochrome displays

t8go is a lightweight, high-performance 2D graphics library specifically designed for embedded displays and microcontrollers. Built for Go and TinyGo compatibility, it provides comprehensive drawing capabilities optimized for resource-constrained environments with minimal memory footprint.

## Features

### Core drawing operations

- **Pixel**: Individual pixel drawing and reading with coordinate-based addressing
- **Line**: Horizontal, vertical, and arbitrary lines using optimized Bresenham's algorithm
- **Rectangle**: Outlined and filled rectangles with optional rounded corners
- **Geometric Shapes**: Perfect circles and ellipses with selective quadrant rendering
- **Arc**: Partial circles and pie charts with configurable start/end angles (0-255° system)
- **Triangle**: Both outlined and filled triangles with scanline-based filling

### Display Architecture

- **Generic Interface**: Works with any display implementing the `Display` interface
- **SSD1306 Driver**: Production-ready I2C driver for OLED displays (128x64, 128x32)
- **Bitmap Driver**: File output for testing and development visualization
- **Buffer Management**: Efficient display buffer operations with memory optimization

### Performance Optimizations

- **Integer Arithmetic**: All operations use integer math for embedded system compatibility
- **Memory Efficient**: Minimal allocations with pre-allocated buffers where possible
- **TinyGo Ready**: Full compatibility with TinyGo compiler and microcontroller targets

## Installation

```bash
go get github.com/redghc/t8go
```

## Quick Start

### Basic Setup with SSD1306 Display

```go
package main

import (
    "machine"
    "github.com/redghc/t8go"
    "github.com/redghc/t8go/drivers/ssd1306"
)

func main() {
    // Initialize I2C bus
    machine.I2C0.Configure(machine.I2CConfig{
        Frequency: machine.TWI_FREQ_400KHZ,
    })

    // Create SSD1306 display (128x64)
    display, err := ssd1306.NewI2C(
        machine.I2C0,
        ssd1306.ADDRESS_GND,
        ssd1306.Config{
            Width:   128,
            Height:  64,
            VCCMode: ssd1306.VCC_SWITCH_CAP,
        },
    )
    if err != nil {
        panic(err)
    }

    // Initialize graphics context
    gfx := t8go.New(display)

    // Clear the display
    gfx.ClearDisplay()

    // Draw various shapes
    gfx.DrawPixel(10, 10)
    gfx.DrawLine(0, 0, 127, 63)
    gfx.DrawBox(20, 20, 40, 30)
    gfx.DrawCircle(64, 32, 20, t8go.DrawAll)

    // Update the display
    gfx.Display()
}
```

## API Reference

### Core graphics context

The main `t8go` struct provides all drawing operations:

```go
type T8Go struct {
    // Contains filtered or unexported fields
}

// Create new graphics context
func New(display Display) *T8Go

// Display management
func (t *T8Go) Size() (width, height uint16)
func (t *T8Go) ClearBuffer()
func (t *T8Go) ClearDisplay()
func (t *T8Go) Display() error
```

### Drawing Functions

#### Basic Primitives

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
func (t *T8Go) DrawLineAngle(originX, originY, length int16, angle uint8)
```

#### Rectangles

```go
// Rectangle outlines
func (t *T8Go) DrawBox(originX, originY, width, height int16)
func (t *T8Go) DrawBoxCoords(startX, startY, endX, endY int16)
func (t *T8Go) DrawRoundBox(originX, originY, width, height, cornerRadius int16)

// Filled rectangles
func (t *T8Go) DrawBoxFill(originX, originY, width, height int16)
func (t *T8Go) DrawBoxFillCoords(startX, startY, endX, endY int16)
func (t *T8Go) DrawRoundBoxFill(originX, originY, width, height, cornerRadius int16)
```

#### Circles, arcs & ellipses

```go
// Circle
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants)
func (t *T8Go) DrawCircleFill(centerX, centerY, radius int16, mask DrawQuadrants)

// Ellipse
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)
func (t *T8Go) DrawEllipseFill(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)

// Arc operations (0-255 angle system)
func (t *T8Go) DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8)
func (t *T8Go) DrawArcFill(centerX, centerY, radius int16, angleStart, angleEnd uint8)
```

#### Triangles

```go
func (t *T8Go) DrawTriangle(x1, y1, x2, y2, x3, y3 int16)
func (t *T8Go) DrawTriangleFill(x1, y1, x2, y2, x3, y3 int16)
```

### Quadrant System

The `DrawQuadrants` type allows selective rendering of circle/ellipse portions:

```go
const (
    DrawNone        DrawQuadrants = 0
    DrawTopLeft     DrawQuadrants = 1 << 0
    DrawTopRight    DrawQuadrants = 1 << 1
    DrawBottomRight DrawQuadrants = 1 << 2
    DrawBottomLeft  DrawQuadrants = 1 << 3
    DrawAll                       = DrawTopLeft | DrawTopRight | DrawBottomRight | DrawBottomLeft
)
```

### Angle System

Arc functions use a 0-255 angle system for precise embedded calculations:

- `0` = 0° (East/Right)
- `64` = 90° (North/Up)
- `128` = 180° (West/Left)
- `192` = 270° (South/Down)
- `255` = ~360° (wraps to 0)

## Supported Hardware

### Displays

- **SSD1306**: 128x64, 128x32 OLED displays via I2C
- **bitmap**: Bitmap driver for rendering to BMP files
- **Generic**: Any display implementing the `Display` interface

### Custom Display Driver

Implement the `Display` interface for custom hardware:

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

## Acknowledgements

- This project is inspired by U8g2: https://github.com/olikraus/u8g2
- Copyright (c) 2016 olikraus@gmail.com
- Licensed under the BSD 2-Clause License:
- https://opensource.org/licenses/BSD-2-Clause

Redistribution and use of the original U8g2 source code or binary forms, with
or without modification, are permitted provided that the following conditions
are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE FROM U8G2 IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE
GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT
OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

### Additional Notice

t8go does not include any font implementations or font files from the U8g2
project. Therefore, no additional font-specific licenses from U8g2 apply to
this software.

---

**t8go** - Tiny 8bit Graphics Go • _Graphics that fit in tiny spaces_ 🎨
