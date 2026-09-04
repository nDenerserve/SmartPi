package modulesRepository

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	models "github.com/nDenerserve/SmartPi/models/modules"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/utils"
)

// i2cRowHeader matches one data row of `i2cdetect` output, e.g. "20: -- -- 27 ...".
// The header row ("     0  1  2 ...") starts with whitespace, not a hex byte
// followed by a colon, so it never matches and is skipped for free.
var i2cRowHeader = regexp.MustCompile(`^([0-9a-f]{2}):(.*)$`)

// ScanI2C probes every address on the configured I2C bus and reports which
// ones are occupied.
//
// This shells out to `i2cdetect` (package i2c-tools, already a documented
// install dependency, see readme.md) rather than probing the bus directly
// with periph.io like the other modules do, deliberately with the -r flag:
// i2cdetect's default probe mode sends a bare SMBus "quick write" command to
// most addresses, which the tool's own manpage warns can alter state on some
// chips. -r instead uses a read-only "receive byte" probe for every address,
// which matters here because a MCP4725 DAC (analogout420ma module) may be
// live on this same bus, actively driving a 4-20mA loop - a scan must never
// risk writing to it.
func (m ModulesRepository) ScanI2C(conf *config.Moduleconfig) (models.I2CScanStatus, error) {
	var status models.I2CScanStatus
	status.Bus = conf.I2CDevice

	busNum, err := i2cBusNumber(conf.I2CDevice)
	if err != nil {
		return status, err
	}

	out, err := utils.RunCommand("i2cdetect", "-y", "-r", busNum)
	if err != nil {
		return status, fmt.Errorf("running i2cdetect (is the i2c-tools package installed?): %w", err)
	}

	status.Devices = parseI2CDetectOutput(out)
	return status, nil
}

// i2cBusNumber extracts the bus number i2cdetect expects (e.g. "1") from the
// full device path SmartPi's config stores (e.g. "/dev/i2c-1").
func i2cBusNumber(device string) (string, error) {
	const prefix = "/dev/i2c-"
	if !strings.HasPrefix(device, prefix) {
		return "", fmt.Errorf("cannot determine i2c bus number from %q: expected a path like %s<n>", device, prefix)
	}
	return strings.TrimPrefix(device, prefix), nil
}

// parseI2CDetectOutput parses the grid `i2cdetect -y -r <bus>` prints into
// one entry per occupied address. Kept separate from ScanI2C so it can be
// tested directly against canned output, without shelling out.
func parseI2CDetectOutput(output string) []models.I2CDeviceEntry {
	var entries []models.I2CDeviceEntry

	for _, line := range strings.Split(output, "\n") {
		match := i2cRowHeader.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		rowBase, err := strconv.ParseInt(match[1], 16, 16)
		if err != nil {
			continue
		}

		// The row's cells follow immediately after a single fixed space
		// (part of the "%02x: " prefix i2cdetect prints), each exactly 3
		// characters wide ("-- ", "UU ", "27 ", or "   " for an address
		// outside the probed range) - see the man page's example output.
		cells := strings.TrimPrefix(match[2], " ")
		for col := 0; col*3 < len(cells); col++ {
			end := col*3 + 3
			if end > len(cells) {
				end = len(cells)
			}
			cell := strings.TrimSpace(cells[col*3 : end])
			if cell == "" || cell == "--" {
				continue
			}

			addr := int(rowBase) + col
			entries = append(entries, models.I2CDeviceEntry{
				Address: fmt.Sprintf("0x%02x", addr),
				Status:  i2cCellStatus(cell),
				Hint:    i2cAddressHint(addr),
			})
		}
	}

	return entries
}

func i2cCellStatus(cell string) string {
	if cell == "UU" {
		return "in_use"
	}
	return "detected"
}

// i2cAddressHint returns a best-effort, non-confirmed guess at which
// SmartPi-supported chip an address belongs to, based on the address ranges
// those chips ship in. See ScanI2C's comment on why this can never be more
// than a hint: none of these chips exposes a register that positively
// identifies it.
func i2cAddressHint(addr int) string {
	switch {
	case addr >= 0x20 && addr <= 0x27:
		return "MCP23017"
	case addr == 0x38:
		return "ADE7878"
	case addr >= 0x60 && addr <= 0x63:
		return "MCP4725"
	case addr >= 0x68 && addr <= 0x6f:
		return "MCP3424"
	default:
		return ""
	}
}
