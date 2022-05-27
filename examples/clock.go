package main

// this example displays the current time in HH.MM.SS format, updating once per
// second.

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jangler/microdotphat-go"
)

func main() {
	// open I2C connection
	if err := microdotphat.Open(""); err != nil {
		log.Fatal(err)
	}

	// catch ^C
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		log.Print((<-c).String())
		microdotphat.Clear()
		microdotphat.Show()
		time.Sleep(time.Millisecond) // wait for transaction before closing
		microdotphat.Close()
		os.Exit(0)
	}()

	// display time every second
	microdotphat.SetDecimal(2, true)
	microdotphat.SetDecimal(4, true)
	for _ = range time.Tick(time.Second) {
		s := time.Now().Format("150405")
		microdotphat.WriteString(s, 0, 0, false)
		if err := microdotphat.Show(); err != nil {
			log.Fatal(err)
		}
	}
}
