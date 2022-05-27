// Package microdotphat implements a pure Go interface to the Micro Dot pHAT
// LED matrix display board. The API is based on that of the official
// microdotphat Python library, and features an automatically resized
// scrollable on/off pixel buffer and built-in text drawing capabilities.
//
// Passing negative coordinates to any function will cause a panic.
package microdotphat

import (
	"errors"
	"strings"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	// Width is the width of the full display in pixels.
	Width = 45

	// Height is the height of the full display in pixels.
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
	ErrNoConn = errors.New("I2C bus connection not open")

	// display buffer vars
	buf      = make([][]bool, Width)
	scrollX  int
	scrollY  int
	decimal  = make([]bool, 6)
	matrices = make([][]byte, 6)
	mirrorX  bool
	mirrorY  bool

	// used only for debugging
	brightness_ byte
	runeBuf     = make([][]rune, Width)

	// I2C vars
	bus   i2c.BusCloser
	addrs = []uint16{0x63, 0x62, 0x61}
)

// called on import
func init() {
	for x := range buf {
		buf[x] = make([]bool, Height)
		runeBuf[x] = make([]rune, Height+1)
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
	if bus == nil {
		return ErrNoConn
	}
	return bus.Close()
}

// Clear clears the buffer and resets its bounds and scroll state.
func Clear() {
	for i := range decimal {
		decimal[i] = false
	}
	for x := range buf {
		for y := range buf[x] {
			buf[x][y] = false
		}
	}
	for x := range buf {
		buf[x] = buf[x][:Height]
	}
	buf = buf[:Width]
	scrollX, scrollY = 0, 0
}

// DrawTiny draws tiny numbers to the buffer. Display is the zero-based index
// of the display to draw numbers on; s is a string containing digits
// (non-digit characters are discarded). This function is not designed for use
// with scrolled buffers.
func DrawTiny(display int, s string) {
	x, y := display*matrixWidth, Height-1
	for _, char := range s {
		if glyph, ok := tinyNumbers[char]; ok {
			for _, row := range glyph {
				if y < 0 {
					return
				}
				for i := 0; i < fontWidth; i++ {
					SetPixel(x+fontWidth-(i+1), y, row&(1<<i) > 0)
				}
				y--
			}
			y--
		}
	}
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
	scrollX = posMod(scrollX+dx, len(buf))
	scrollY = posMod(scrollY+dy, len(buf[0]))
}

// ScrollTo scrolls the buffer to a specific position.
func ScrollTo(x, y int) {
	scrollX = posMod(x, len(buf))
	scrollY = posMod(y, len(buf[0]))
}

// SetBrightness sets the display brightness in the range [0.0, 1.0].
func SetBrightness(brightness float64) error {
	if brightness > 1 {
		brightness = 1
	} else if brightness < 0 {
		brightness = 0
	}
	brightness_ = byte(brightness * 127)
	if bus == nil {
		return ErrNoConn
	}
	for _, addr := range addrs {
		if err := bus.Tx(addr, []byte{cmdBrightness, brightness_}, nil); err != nil {
			return err
		}
	}
	return nil
}

// SetCol sets a whole column of the buffer (only useful when not scrolling
// vertically). The 7 least significant bits of col correspond to each row of
// the column, with the least significant bit on top.
func SetCol(x int, col byte) {
	for y := 0; y < Height; y++ {
		SetPixel(x, y, col&(1<<y) > 0)
	}
}

// SetDecimal sets the state of a decimal point on a zero-indexed display.
// Panics if the display index is out of range.
func SetDecimal(display int, lit bool) {
	decimal[display] = lit
}

// SetMirror sets whether the display should be flipped horizontally and/or
// vertically. To rotate the display 180 degrees, flip both x and y.
func SetMirror(x, y bool) {
	mirrorX, mirrorY = x, y
}

// SetPixel sets the buffer pixel at (x,y) to lit or unlit.
func SetPixel(x, y int, lit bool) {
	for x >= len(buf) {
		buf = append(buf, make([]bool, len(buf[0])))
	}
	if y >= len(buf[0]) {
		for x := range buf {
			buf[x] = append(buf[x], make([]bool, len(buf[x])-(y-1))...)
		}
	}
	buf[x][y] = lit
}

// Show outputs the buffer to the display.
func Show() error {
	updateMatrices()

	// send matrix buffers and update commands
	if bus == nil {
		return ErrNoConn
	}
	for i, addr := range addrs {
		if err := bus.Tx(addr,
			append([]byte{cmdMatrix1}, matrices[i*2+1]...), nil); err != nil {
			return err
		}
		if err := bus.Tx(addr,
			append([]byte{cmdMatrix2}, matrices[i*2]...), nil); err != nil {
			return err
		}
		if err := bus.Tx(addr, []byte{cmdUpdate, 1}, nil); err != nil {
			return err
		}
	}
	return nil
}

// helper function for Show; updates matrices to match dipslay vars but does
// not communicate with the I2C bus
func updateMatrices() {
	// clear matrices
	for _, matrix := range matrices {
		for y := range matrix {
			matrix[y] = 0
		}
	}
	for x := range runeBuf {
		for y := range runeBuf[x] {
			if (y < Height && x%matrixWidth < 5) || x%matrixWidth == 0 {
				runeBuf[x][y] = '.'
			} else {
				runeBuf[x][y] = ' '
			}
		}
	}

	// set pixels based on buffer
	for x := 0; x < Width; x++ {
		for y := 0; y < Height; y++ {
			tx, ty := translateCoords(x, y)
			if buf[tx][ty] {
				setMatrixPixel(x, y)
			}
		}
	}

	// set decimal points
	for i, lit := range decimal {
		if lit {
			if i%2 == 1 {
				matrices[i][6] |= 0b10000000
			} else {
				matrices[i][7] |= 0b01000000
			}
			runeBuf[i*matrixWidth][7] = '#'
		}
	}
}

// helper function for updateMatrices; translates matrix coordinates to buffer
// coordinates based on current display vars
func translateCoords(x, y int) (int, int) {
	if mirrorX {
		x = (Width - 1) - x
	}
	if mirrorY {
		y = (Height - 1) - y
	}
	x = posMod(x+scrollX, len(buf))
	y = posMod(y+scrollY, len(buf[0]))
	return x, y
}

// helper function for updateMatrices; sets an individual pixel of the matrix,
// not accounting for scrolling. ignores out-of-bounds (x,y) values.
func setMatrixPixel(x, y int) {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return
	}
	i := x / matrixWidth
	if i%2 == 1 {
		matrices[i][y] |= (1 << (x % matrixWidth))
	} else {
		matrices[i][x%matrixWidth] |= (1 << y)
	}
	if x%matrixWidth < 5 {
		runeBuf[x][y] = '#'
	}
}

// String returns a string representing the expected LED display state.
func String() string {
	var b strings.Builder
	for y := range runeBuf[0] {
		for x := range runeBuf {
			b.WriteRune(runeBuf[x][y])
		}
		if y < len(runeBuf[0])-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

// WriteChar writes a single character to the buffer at the specified position.
// Characters not covered by the built-in font are discarded.
func WriteChar(char rune, x, y int) {
	if glyph, ok := font[char]; ok {
		for gx := 0; gx < fontWidth; gx++ {
			for gy := 0; gy < fontHeight; gy++ {
				lit := glyph[gx]&(1<<gy) > 0
				SetPixel(x+gx, y+gy, lit)
			}
		}
	}
}

// WriteString writes a string to the buffer at the specified position. If kern
// is true, characters will be written closely together, ideal for scrolling
// displays. If kern is false, characters are arranged one per display.
// Characters not covered by the built-in font are discarded.
func WriteString(s string, x, y int, kern bool) {
	for _, char := range s {
		if _, ok := font[char]; !ok {
			continue
		}
		WriteChar(char, x, y)
		if kern {
			x += fontWidth + 1
		} else {
			x += matrixWidth
		}
	}
}

// helper function; modulo that always returns a positive result
func posMod(val, mod int) int {
	val = val % mod
	if val < 0 {
		val += mod
	}
	return val
}
