package microdotphat

import (
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	// Width is the width of the display in pixels.
	Width = 45

	// Height is the height of the display in pixels.
	Height = 7

	// width of an individual display matrix in pixels
	matrixWidth = 8

	// I2C command codes
	cmdMode       = 0x00
	cmdMatrix1    = 0x01
	cmdUpdate     = 0x0C
	cmdOpts       = 0x0D
	cmdMatrix2    = 0x0E
	cmdBrightness = 0x19

	// arguments for I2C commands
	initialMode       = 0b00011000
	initialOpts       = 0b00001110
	initialBrightness = 127
)

var (
	// display buffer vars
	buf      = make([][]bool, Width)
	scrollX  int
	scrollY  int
	decimal  = make([]byte, 6)
	matrices = make([][]byte, 6)

	// I2C vars
	bus   i2c.BusCloser
	addrs = []uint16{0x63, 0x62, 0x61}
)

// called on import
func init() {
	for x := range buf {
		buf[x] = make([]bool, Height)
	}
	for y := range matrices {
		matrices[y] = make([]byte, Height+1)
	}
}

// Open initializes the library, opening a connection to the I2C bus with the
// given name. If left blank, the default bus is used, which is usually
// sufficient.
func Open(name string) error {
	_, err := host.Init()
	if err != nil {
		return err
	}
	bus, err = i2creg.Open(name)
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		if err := bus.Tx(addr, []byte{cmdMode, initialMode}, nil); err != nil {
			return err
		}
		if err := bus.Tx(addr, []byte{cmdOpts, initialOpts}, nil); err != nil {
			return err
		}
		if err := bus.Tx(addr, []byte{cmdBrightness, initialBrightness}, nil); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the connection to the I2C bus.
func Close() error {
	if bus != nil {
		return bus.Close()
	}
	return nil
}

// Clear clears the buffer.
func Clear() {
	for i := range decimal {
		decimal[i] = 0
	}
	for x := range buf {
		for y := range buf[x] {
			buf[x][y] = false
		}
	}
	scrollX, scrollY = 0, 0
}

// DrawTiny draws tiny numbers to the buffer. Display is the zero-based index
// of the display to draw numbers on; s is a string containing only digits.
// Panics if the display index is out of range. Non-digit characters are
// ignored.
func DrawTiny(display int, s string) {
	// TODO
	panic("DrawTiny NYI")
}

// Fill fills the buffer either lit or unlit.
func Fill(lit bool) {
	for x := range buf {
		for y := range buf[x] {
			buf[x][y] = lit
		}
	}
}

// Scroll scrolls the buffer.
func Scroll(dx, dy int) {
	// TODO
	panic("Scroll NYI")
}

// ScrollTo scrolls the buffer to a specific position.
func ScrollTo(x, y int) {
	// TODO
	panic("ScrollTo NYI")
}

// SetBrightness sets the display brightness in the range [0.0, 1.0].
func SetBrightness(brightness float64) {
	// TODO
	panic("SetBrightness NYI")
}

// SetCol sets a whole column of the buffer (only useful when not scrolling
// vertically). The 7 least significant bits of col correspond to each row of
// the column.
func SetCol(x int, col byte) {
	// TODO
	panic("SetCol NYI")
}

// SetDecimal sets the state of a decimal point on a zero-indexed display.
// Panics if the display index is out of range.
func SetDecimal(display int, lit bool) {
	// TODO
	panic("SetDecimal NYI")
}

// SetMirror sets whether the display should be flipped horizontally.
func SetMirror(mirror bool) {
	// TODO
	panic("SetMirror NYI")
}

// SetPixel sets the buffer pixel at (x,y) to lit or unlit.
func SetPixel(x, y int, lit bool) {
	// TODO grow buffer automatically
	buf[x][y] = lit
}

// SetRotate sets whether the display should be rotated 180 degrees.
func SetRotate180(rotate bool) {
	// TODO
	panic("SetRotate180 NYI")
}

// Show outputs the buffer to the display.
func Show() error {
	// TODO scrolling
	// TODO decimal points

	// update matrix buffers
	for _, matrix := range matrices {
		for y := range matrix {
			matrix[y] = 0
		}
	}
	for x := range buf {
		for y := range buf[x] {
			if buf[x][y] {
				setMatrixPixel(x, y)
			}
		}
	}

	// send matrix buffers and update commands
	for i, addr := range addrs {
		if err := bus.Tx(addr,
			append([]byte{cmdMatrix1}, matrices[i*2]...), nil); err != nil {
			return err
		}
		if err := bus.Tx(addr,
			append([]byte{cmdMatrix2}, matrices[i*2+1]...), nil); err != nil {
			return err
		}
		if err := bus.Tx(addr, []byte{cmdUpdate, 1}, nil); err != nil {
			return err
		}
	}
	return nil
}

// helper function for Show
func setMatrixPixel(x, y int) {
	i := x / matrixWidth
	if i % 2 == 0 {
		matrices[i][y] |= (1 << (x % matrixWidth))
	} else {
		matrices[i][x % matrixWidth] |= (1 << y)
	}
}

// WriteChar writes a single character to the buffer at the specified position.
func WriteChar(char rune, x, y int) {
	// TODO
	panic("WriteChar NYI")
}

// WriteString writes a string to the buffer at the specified position. If kern
// is true, characters will be written closely together, ideal for scrolling
// displays. If kern is false, characters are arranged one per display.
func WriteString(s string, x, y int, kern bool) {
	// TODO
	panic("WriteString NYI")
}
