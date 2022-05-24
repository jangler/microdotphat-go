package microdotphat

import (
	"log"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	// Width is the width of the display in pixels.
	Width = 30

	// Height is the height of the display in pixels.
	Height = 7

	// width of an individual display matrix
	matrixWidth = 5

	// I2C command codes
	cmdMatrix1    = 0x01
	cmdMode       = 0x09
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

// SetPixel sets the buffer pixel at (x,y) to lit or unlit.
// Panics if (x,y) is outside the bounds of the buffer.
func SetPixel(x, y int, lit bool) {
	buf[x][y] = lit
}

// Show outputs the buffer to the display.
func Show() error {
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

func setMatrixPixel(x, y int) {
	i := x / matrixWidth
	if i % 2 == 0 {
		matrices[i][y] |= (1 << (x % matrixWidth))
	} else {
		matrices[i][x % matrixWidth] |= (1 << y)
	}
}
