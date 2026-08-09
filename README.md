# t8go - Tiny 8bit Graphics Go

[![Go Version](https://img.shields.io/badge/go-1.24.5+-blue.svg)](https://golang.org/doc/devel/release.html)
[![TinyGo Compatible](https://img.shields.io/badge/tinygo-compatible-brightgreen.svg)](https://tinygo.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/redghc/t8go.svg)](https://pkg.go.dev/github.com/redghc/t8go)

## Graphics library for monochrome displays

`t8go` is a lightweight, high-performance 2D graphics library specifically designed for embedded displays and microcontrollers. Built for Go and TinyGo compatibility, it provides comprehensive drawing capabilities optimized for resource-constrained environments with minimal memory footprint.

## Features

### Core Drawing Operations

- **Pixel**: Individual pixel drawing, writing, and reading with coordinate-based addressing.
- **Line**: Horizontal, vertical, angled, and arbitrary lines using optimized Bresenham's algorithm.
- **Rectangle**: Outlined, filled, and rounded-corner rectangles.
- **Geometric Shapes**: Perfect circles and ellipses with selective quadrant masking.
- **Arc**: Partial circles and pie slices with start/end angles in a 0–255 fixed-point system.
- **Triangle**: Outlined and scanline-filled triangles.
- **Polygon**: Outlined and scanline-filled arbitrary polygons via variadic `Point` vertices (even-odd rule).

### Display Architecture & Performance

- **Generic Interfaces**: Clean separation between display hardware (`IDisplay`) and graphics operations (`IDisplayDrawer`).
- **Partial Updates**: Bounding-box dirty tracking flushes only modified display regions via `Display()`.
- **Hardware Acceleration**: `IDisplay` interface methods (`FillHRun`, `FillVRun`, `FillRect`, `DisplayRegion`) enable driver-level batching.
- **SSD1306 Driver**: Production-ready I2C driver for OLED displays (128x64, 128x32).
- **Bitmap Driver**: Local file output (`.bmp`) for testing, validation, and visual debugging without hardware.
- **Integer Arithmetic**: Pure integer math throughout—no floating-point overhead on microcontrollers.
- **TinyGo Ready**: Full compatibility with TinyGo compiler and embedded microcontroller targets.

## Installation

```bash
go get github.com/redghc/t8go
```

## Quick Start

### Basic Setup with SSD1306 OLED Display

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
		ssd1306.AddressGND,
		ssd1306.Config{
			Width:   128,
			Height:  64,
			VCCMode: ssd1306.VCCSwitchCap,
		},
	)
	if err != nil {
		panic(err)
	}

	// Initialize graphics context
	gfx := t8go.New(display)

	// Clear display buffer
	gfx.ClearDisplay()

	// Draw primitive shapes
	gfx.DrawPixel(10, 10)
	gfx.DrawLine(0, 0, 127, 63)
	gfx.DrawBox(20, 20, 40, 30)
	gfx.DrawCircle(64, 32, 20, t8go.DrawAll)

	// Flush dirty region to display
	if err := gfx.Display(); err != nil {
		panic(err)
	}
}
```

### Rendering to a Bitmap File (Testing & Debugging)

```go
package main

import (
	"github.com/redghc/t8go"
	"github.com/redghc/t8go/drivers/bitmap"
)

func main() {
	display, err := bitmap.New(bitmap.Config{
		Width:    128,
		Height:   64,
		Filename: "output.bmp",
	})
	if err != nil {
		panic(err)
	}

	gfx := t8go.New(display)
	gfx.ClearBuffer()

	gfx.DrawBoxFill(10, 10, 50, 30)
	gfx.DrawCircle(64, 32, 20, t8go.DrawAll)

	if err := gfx.DisplayAll(); err != nil {
		panic(err)
	}
}
```

## API Reference

### Core Graphics Context & Display Management

The main `t8go` constructor wraps an `IDisplay` and returns an `IDisplayDrawer`:

```go
// Create new graphics context wrapping an IDisplay
func New(display IDisplay) IDisplayDrawer

// Display management methods on IDisplayDrawer
func (t *T8Go) Size() (width, height uint16)
func (t *T8Go) ClearBuffer()
func (t *T8Go) ClearDisplay()
func (t *T8Go) Display() error
func (t *T8Go) DisplayAll() error
```

#### Display Flush Behavior

- `Display()`: Flushes only the bounding region modified since the last flush (dirty-region tracking). Returns immediately if no changes occurred.
- `DisplayAll()`: Flushes the entire screen buffer unconditionally. Recommended after direct buffer modifications.

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

#### Circles, Arcs & Ellipses

```go
// Circle
func (t *T8Go) DrawCircle(centerX, centerY, radius int16, mask DrawQuadrants)
func (t *T8Go) DrawCircleFill(centerX, centerY, radius int16, mask DrawQuadrants)

// Ellipse
func (t *T8Go) DrawEllipse(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)
func (t *T8Go) DrawEllipseFill(centerX, centerY, radiusX, radiusY int16, mask DrawQuadrants)

// Arc (0-255 angle system)
func (t *T8Go) DrawArc(centerX, centerY, radius int16, angleStart, angleEnd uint8)
func (t *T8Go) DrawArcFill(centerX, centerY, radius int16, angleStart, angleEnd uint8)
```

#### Triangles

```go
func (t *T8Go) DrawTriangle(x1, y1, x2, y2, x3, y3 int16)
func (t *T8Go) DrawTriangleFill(x1, y1, x2, y2, x3, y3 int16)
```

#### Polygons

Polygons accept a variadic list of `Point` vertices and automatically close the path:

```go
type Point struct {
	X, Y int16
}

func (t *T8Go) DrawPolygon(points ...Point)     // boundary of the filled region
func (t *T8Go) DrawPolygonFill(points ...Point) // scanline-filled (even-odd rule)
```

`DrawPolygon` draws the exact boundary of the region filled by `DrawPolygonFill`. Both methods handle concave and self-intersecting shapes via scanline even-odd fill rules.

### Quadrant System

The `DrawQuadrants` bitmask enables selective quadrant rendering for circles and ellipses:

```go
const (
	DrawTopLeft     DrawQuadrants = 1 << 0
	DrawTopRight    DrawQuadrants = 1 << 1
	DrawBottomRight DrawQuadrants = 1 << 2
	DrawBottomLeft  DrawQuadrants = 1 << 3
	DrawAll         DrawQuadrants = DrawTopLeft | DrawTopRight | DrawBottomRight | DrawBottomLeft
)
```

A zero value (`0`) renders nothing; `DrawAll` renders all four quadrants.

### Angle System

Arcs use a fixed-point `uint8` angle system (`0`–`255`) optimized for embedded microcontrollers:

- `0` = 0° (Right / East)
- `64` = 90° (Top / North)
- `128` = 180° (Left / West)
- `192` = 270° (Bottom / South)
- `255` = ~360° (wraps to 0)

## Supported Hardware & Drivers

### Displays

- **SSD1306**: 128x64, 128x32 OLED displays via I2C (`drivers/ssd1306`).
- **Bitmap**: Virtual display rendering directly to `.bmp` image files (`drivers/bitmap`).
- **Generic**: Any display hardware implementing `IDisplay`.

### Driver Interface (`IDisplay`)

Custom hardware drivers implement `IDisplay`:

```go
type IDisplay interface {
	Size() (width, height uint16)
	BufferSize() int
	Buffer() []byte

	ClearBuffer()
	ClearDisplay()
	Command(cmd byte) error
	Display() error
	DisplayRegion(x0, y0, x1, y1 int) error
	SetPixel(x, y int16, on bool)
	GetPixel(x, y uint8) bool

	FillHRun(x0, x1, y int16, on bool)
	FillRect(x0, y0, x1, y1 int16, on bool)
	FillVRun(x, y0, y1 int16, on bool)
}
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgements

- Inspired by [U8g2](https://github.com/olikraus/u8g2) (BSD 2-Clause License).
- Copyright (c) 2016 `olikraus@gmail.com`.
- `t8go` does not include font files or font implementations from U8g2.

---

**t8go** - Tiny 8bit Graphics Go • _Graphics that fit in tiny spaces_ 🎨
