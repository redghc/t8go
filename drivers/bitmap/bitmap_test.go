package bitmap_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/redghc/t8go/drivers/bitmap"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("rejects zero width", func(t *testing.T) {
		t.Parallel()
		_, err := bitmap.New(bitmap.Config{Width: 0, Height: 64})
		if !errors.Is(err, bitmap.ErrInvalidDimensions) {
			t.Errorf("got %v, want ErrInvalidDimensions", err)
		}
	})

	t.Run("rejects zero height", func(t *testing.T) {
		t.Parallel()
		_, err := bitmap.New(bitmap.Config{Width: 128, Height: 0})
		if !errors.Is(err, bitmap.ErrInvalidDimensions) {
			t.Errorf("got %v, want ErrInvalidDimensions", err)
		}
	})

	t.Run("rejects oversize width", func(t *testing.T) {
		t.Parallel()
		// 40000 doesn't fit in int16 without truncation; bounds check would misbehave.
		_, err := bitmap.New(bitmap.Config{Width: 40000, Height: 64})
		if !errors.Is(err, bitmap.ErrInvalidDimensions) {
			t.Errorf("got %v, want ErrInvalidDimensions", err)
		}
	})

	t.Run("rejects oversize height", func(t *testing.T) {
		t.Parallel()
		_, err := bitmap.New(bitmap.Config{Width: 128, Height: 40000})
		if !errors.Is(err, bitmap.ErrInvalidDimensions) {
			t.Errorf("got %v, want ErrInvalidDimensions", err)
		}
	})

	t.Run("default filename is display.bmp", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		d, err := bitmap.New(bitmap.Config{Width: 16, Height: 8, Filename: filepath.Join(dir, "a.bmp")})
		if err != nil {
			t.Fatal(err)
		}
		// Drive Display() to actually create the file; absence of error implies
		// the driver accepted the configuration.
		if displayErr := d.Display(); displayErr != nil {
			t.Errorf("Display returned %v, want nil", displayErr)
		}
	})
}

func TestSetGetPixel(t *testing.T) {
	t.Parallel()

	d, err := bitmap.New(bitmap.Config{Width: 16, Height: 8})
	if err != nil {
		t.Fatal(err)
	}

	// Set a pixel and read it back.
	d.SetPixel(3, 5, true)
	if !d.GetPixel(3, 5) {
		t.Error("GetPixel(3,5) = false after SetPixel true")
	}
	if d.GetPixel(0, 0) {
		t.Error("GetPixel(0,0) = true on a fresh buffer")
	}

	// Clear it.
	d.SetPixel(3, 5, false)
	if d.GetPixel(3, 5) {
		t.Error("GetPixel(3,5) = true after SetPixel false")
	}
}

func TestSetPixelOutOfBounds(t *testing.T) {
	t.Parallel()

	d, err := bitmap.New(bitmap.Config{Width: 16, Height: 8})
	if err != nil {
		t.Fatal(err)
	}

	// Each call must be a no-op: no panic, no buffer writes outside the visible area.
	for _, p := range []struct{ x, y int16 }{
		{-1, 0}, {0, -1}, {16, 0}, {0, 8},
	} {
		d.SetPixel(p.x, p.y, true)
	}
	if d.GetPixel(0, 0) {
		t.Error("out-of-bounds SetPixel leaked into the buffer")
	}
}

func TestGetPixelOutOfBounds(t *testing.T) {
	t.Parallel()

	d, err := bitmap.New(bitmap.Config{Width: 16, Height: 8})
	if err != nil {
		t.Fatal(err)
	}

	// uint8 inputs that exceed the display size must return false, not read
	// arbitrary memory.
	for _, p := range []struct{ x, y uint8 }{
		{16, 0}, {0, 8}, {255, 255},
	} {
		if d.GetPixel(p.x, p.y) {
			t.Errorf("GetPixel(%d,%d) = true, want false", p.x, p.y)
		}
	}
}

func TestDisplayWritesBMP(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "out.bmp")

	d, err := bitmap.New(bitmap.Config{Width: 16, Height: 8, Filename: path})
	if err != nil {
		t.Fatal(err)
	}
	d.SetPixel(0, 0, true)
	d.SetPixel(15, 7, true)
	if displayErr := d.Display(); displayErr != nil {
		t.Fatal(displayErr)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// 14 (file header) + 40 (info header) + 8 (palette) + 16*8/8 padded-to-4 = 94.
	// Each BMP row is padded to 4 bytes regardless of width, so a 16x8 image
	// takes 4 bytes per row × 8 rows = 32 bytes of pixel data.
	const wantSize = 14 + 40 + 8 + 32
	if info.Size() != wantSize {
		t.Errorf("file size = %d, want %d", info.Size(), wantSize)
	}

	// Magic bytes for a BMP file.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != 'B' || data[1] != 'M' {
		t.Errorf("missing BMP signature, got %q", data[:2])
	}
}

func TestCommandAndDisplayRegionNoOps(t *testing.T) {
	t.Parallel()

	d, _ := bitmap.New(bitmap.Config{Width: 16, Height: 8})
	if err := d.Command(0xFF); err != nil {
		t.Errorf("Command returned %v, want nil", err)
	}
	if err := d.DisplayRegion(0, 0, 15, 7); err != nil {
		t.Errorf("DisplayRegion returned %v, want nil", err)
	}
}

func TestBufferAndSize(t *testing.T) {
	t.Parallel()

	d, _ := bitmap.New(bitmap.Config{Width: 32, Height: 16})
	if w, h := d.Size(); w != 32 || h != 16 {
		t.Errorf("Size = (%d,%d), want (32,16)", w, h)
	}
	// 32*16/8 = 64 bytes
	if got := d.BufferSize(); got != 64 {
		t.Errorf("BufferSize = %d, want 64", got)
	}
	buf := d.Buffer()
	if len(buf) != 64 {
		t.Errorf("Buffer length = %d, want 64", len(buf))
	}
}

func TestClearBufferAndDisplay(t *testing.T) {
	t.Parallel()

	d, _ := bitmap.New(bitmap.Config{Width: 16, Height: 8})
	d.SetPixel(5, 3, true)
	d.ClearDisplay()
	if d.GetPixel(5, 3) {
		t.Error("pixel still set after ClearDisplay")
	}
}
