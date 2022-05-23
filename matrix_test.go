package microdotphat

import (
	"testing"
	"time"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func TestMatrix(t *testing.T) {
	if _, err := host.Init(); err != nil {
		t.Fatal(err)
	}
	bus, err := i2creg.Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer bus.Close()
	m, err := newMatrix(bus, 0x61)
	if err != nil {
		t.Fatal(err)
	}
	m.setPixel(0, 0, 0, true)
	if err := m.update(); err != nil {
		t.Fatal(err)
	}
}
