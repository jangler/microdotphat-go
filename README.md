# microdotphat-go

A pure Go interface to the Micro Dot pHAT LED matrix display board. The API is
based on that of the official microdotphat Python library, and features an
automatically resized scrollable on/off pixel buffer and built-in text drawing
capabilities.

## Quick links

- Original Python package <https://github.com/pimoroni/microdot-phat>
- API documentation: <https://pkg.go.dev/github.com/jangler/microdotphat-go>

## Differences from the Python microdotphat package

- The Python package clears the display at program exit by default
  (configurable via `set_clear_on_exit`); this package does not. If you want
  this behavior, add this code to your `func main`:
  ```
  defer func() {
  	microdotphat.Clear()
  	microdotphat.Show()
  }()
  ```
- The Python package has three functions for scrolling the buffer (not
  including `scroll_to`). This package collapses those functions into one
  `Scroll` function.
- The Python package has functions `set_mirror` and `set_rotate180`. This
  package collapses those functions into one `SetMirror` function (rotation is
  achieved by flipping both axes).
