package microdotphat

import (
	"periph.io/x/conn/v3/i2c"
)

const (
	mode = 0b00011000
	opts = 0b00001110

	cmdBrightness = 0x19
	cmdMode       = 0x09
	cmdUpdate     = 0x0C
	cmdOptions    = 0x0D

	cmdMatrix1 = 0x01
	cmdMatrix2 = 0x0E

	matrix1 = 0
	matrix2 = 1
)

type matrix struct {
	addr       uint16
	brightness byte
	bus        i2c.Bus
	buf        [][]byte
}

func newMatrix(bus i2c.Bus, addr uint16) (*matrix, error) {
	brightness := byte(127)
	if err := bus.Tx(addr, []byte{cmdMode, mode}, nil); err != nil {
		return nil, err
	}
	if err := bus.Tx(addr, []byte{cmdOptions, opts}, nil); err != nil {
		return nil, err
	}
	if err := bus.Tx(addr, []byte{cmdBrightness, brightness}, nil); err != nil {
		return nil, err
	}
	return &matrix{
		addr:       addr,
		brightness: brightness,
		bus:        bus,
		buf: [][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	}, nil
}

func (m *matrix) setPixel(i, x, y int, lit bool) {
	if i == 0 {
		if lit {
			m.buf[i][y] |= (1 << x)
		} else {
			m.buf[i][y] &= ^(1 << x)
		}
	} else {
		if lit {
			m.buf[i][x] |= (1 << y)
		} else {
			m.buf[i][x] &= ^(1 << y)
		}
	}
}

func (m *matrix) clear() error {
	for i := range m.buf {
		for j := range m.buf[i] {
			m.buf[i][j] = 0
		}
	}
	return m.update()
}

func (m *matrix) update() error {
	if err := m.bus.Tx(m.addr, append([]byte{cmdMatrix1}, m.buf[0]...), nil); err != nil {
		return err
	}
	if err := m.bus.Tx(m.addr, append([]byte{cmdMatrix2}, m.buf[1]...), nil); err != nil {
		return err
	}
	if err := m.bus.Tx(m.addr, []byte{cmdUpdate, 1}, nil); err != nil {
		return err
	}
	return nil
}
