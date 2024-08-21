package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var appVersion = "No Version Provided"

func main() {

	// moduleconfig := config.NewModuleconfig()

	version := flag.Bool("v", false, "prints current version information")

	flag.Parse()
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	// log.SetLevel(moduleconfig.LogLevel)
	log.SetLevel(log.DebugLevel)

	connect("web.vodafone.de", "2345")

}

func connect(apn string, pin string) (bool, error) {

	// Shutdown wwan0
	log.Debug("Shutdown wwan0")
	cmd := exec.Command("sudo", "ip", "link", "set", "wwan0", "down")
	out, err := cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		return false, err
	}
	// Set raw mode
	log.Debug("Set RAW")
	cmd = exec.Command("sudo", "tee", "/sys/class/net/wwan0/qmi/raw_ip")
	stdin, err := cmd.StdinPipe()

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "Y")
	}()

	out, err = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	mess := fmt.Sprintf("%s", out)
	if err != nil {
		return false, err
	}

	if mess != "Y" {
		return false, errors.New(mess)
	}

	// Setup wwan0
	log.Debug("Setup wwan0")
	cmd = exec.Command("sudo", "ip", "link", "set", "wwan0", "up")
	out, err = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		return false, err
	}

	// verify PIN
	log.Debug("Verify PIN")
	cmd = exec.Command("sudo", "qmicli", "-d", "/dev/cdc-wdm0", "--uim-verify-pin=PIN1,"+pin)
	out, err = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		return false, err
	}

	// set APN
	log.Debug("Set APN")
	// cmd = exec.Command("sudo", "qmicli", "-p", "-d", "/dev/cdc-wdm0", "--device-open-net='net-raw-ip|net-no-qos-header'", "--wds-start-network=\"apn='"+apn+"',ip-type=4\"", " --client-no-release-cid")
	cmd = exec.Command("bash", "-c", "sudo qmicli -p -d /dev/cdc-wdm0 --device-open-net='net-raw-ip|net-no-qos-header' --wds-start-network=\"apn='"+apn+"',ip-type=4\" --client-no-release-cid")
	fmt.Println(cmd)
	out, err = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		return false, err
	}

	// assign default ip
	log.Debug("Assign default IP")
	cmd = exec.Command("sudo", "udhcpc", "-q", "-f", "-i", "wwan0")
	out, err = cmd.CombinedOutput()
	log.Debug(fmt.Sprintf("%s", out))
	if err != nil {
		return false, err
	}

	return true, nil
}
