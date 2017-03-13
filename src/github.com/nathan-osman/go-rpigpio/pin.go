package rpi

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Indicate whether the pin is used for input or output.
type Direction int

const (
	IN  = iota // pin is used for input
	OUT        // pin is used for output
)

// Indicate the current state of the pin.
type Value int

const (
	LOW  = iota // pin is low (off)
	HIGH        // pin is high (on)
)

// Determine if a specific pin is exported.
func isPinExported(number int) (bool, error) {
	_, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%d", number))
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// Export or unexport a specific pin.
func setPinExport(number int, export bool) error {
	var filename string
	if export {
		filename = "/sys/class/gpio/export"
	} else {
		filename = "/sys/class/gpio/unexport"
	}
	f, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("%d\n", number)))
	if err != nil {
		return err
	}
	return nil
}

// Set the direction of a pin.
func setPinDirection(number int, direction Direction) error {
	filename := fmt.Sprintf("/sys/class/gpio/gpio%d/direction", number)
	f, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	var data []byte
	switch direction {
	case IN:
		data = []byte("in\n")
	case OUT:
		data = []byte("out\n")
	}
	_, err = f.Write(data)
	return err
}

// An individual GPIO pin.
type Pin struct {
	number int
	value  *os.File
}

// Prepare a pin for input or output.
func OpenPin(number int, direction Direction) (*Pin, error) {
	e, err := isPinExported(number)
	if err != nil {
		return nil, err
	}
	if !e {
		if err := setPinExport(number, true); err != nil {
			return nil, err
		}
	}
	if err := setPinDirection(number, direction); err != nil {
		return nil, err
	}
	var flag int
	switch direction {
	case IN:
		flag = os.O_RDONLY
	case OUT:
		flag = os.O_WRONLY
	}
	f, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", number), flag, 0)
	if err != nil {
		return nil, err
	}
	return &Pin{
		number: number,
		value:  f,
	}, nil
}

// Read the current value of the pin.
func (p *Pin) Read() (Value, error) {
	// seek to beginning of file in case we've read it before
	if _, err := p.value.Seek(0, 0); err != nil {
		return LOW, err
	}

	d, err := ioutil.ReadAll(p.value)
	if err != nil {
		return LOW, err
	}
	value := strings.TrimSpace(string(d))
	switch value {
	case "0":
		return LOW, nil
	case "1":
		return HIGH, nil
	default:
		return 0, fmt.Errorf("unrecognized value '%s'", value)
	}
}

// Set the current value of the pin.
func (p *Pin) Write(value Value) error {
	var data []byte
	switch value {
	case LOW:
		data = []byte("0\n")
	case HIGH:
		data = []byte("1\n")
	}
	_, err := p.value.Write(data)
	return err
}

// Close the pin.
func (p *Pin) Close() error {
	if err := p.value.Close(); err != nil {
		return err
	}
	if err := setPinExport(p.number, false); err != nil {
		return err
	}
	return nil
}
