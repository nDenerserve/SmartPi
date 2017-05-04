## go-rpigpio

[![GoDoc](https://godoc.org/github.com/nathan-osman/go-rpigpio?status.svg)](https://godoc.org/github.com/nathan-osman/go-rpigpio)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

This package provides a really simple interface for interacting with the GPIO pins on a Raspberry Pi.

### Usage

The following example demonstrates writing to GPIO2:

    import "github.com/nathan-osman/go-rpigpio"

    p, err := rpi.OpenPin(2, rpi.OUT)
    if err != nil {
        panic(err)
    }
    defer p.Close()

    // set the pin to high (on)
    p.Write(rpi.HIGH)

    // set the pin to low (off)
    p.Write(rpi.LOW)
