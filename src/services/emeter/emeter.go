// https://gist.github.com/fiorix/9664255
// https://en.wikipedia.org/wiki/Multicast_address
// https://support.mcommstv.com/hc/en-us/articles/202306226-Choosing-multicast-addresses-and-ports

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/utils"
	"github.com/nDenerserve/SmartPi/utils/multicast"

	log "github.com/sirupsen/logrus"
)

const (
	address = "239.12.255.254:9522"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var appVersion = "No Version Provided"

func main() {

	config := config.NewConfig()

	version := flag.Bool("v", false, "prints current version information")
	flag.Parse()
	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("ERROR", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	log.Info("Emeter started")

	// go ping(config, watcher)
	// multicast.Listen(address, msgHandler)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Debugf("EVENT! %#v\n", event)
				time.Sleep(100 * time.Millisecond)

				ping(config)

				// watch for errors
			case err := <-watcher.Errors:
				log.Fatal("ERROR", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(config.SharedDir + "/" + config.SharedFile); err != nil {
		log.Fatal("ERROR", err)

	}

	<-done

}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Println(n, "bytes read from", src)
	log.Println(hex.Dump(b[:n]))
}

func ping(config *config.Config) {

	var datagram []byte

	datagram = append([]byte("SMA"))
	datagram = append(datagram, []byte{0x00, 0x04}...)                         // Data length: 4 byte (0x00000004)
	datagram = append(datagram, []byte{0x02, 0xA0}...)                         // Tag: "Tag0" (42), version 0
	datagram = append(datagram, []byte{0x00, 0x00, 0x00, 0x01}...)             // Group1 (default group)
	datagram = append(datagram, []byte{0x00, 0x2C}...)                         // Data length: 44 byte (variable)
	datagram = append(datagram, []byte{0x00, 0x10}...)                         // Tag: "SMA Net 2", version 0
	datagram = append(datagram, []byte{0x60, 0x69}...)                         // Protocol ID (energy meter protocol), Data length: 2 byte
	datagram = append(datagram, []byte{0x01, 0x0E, 0x00, 0x00, 0x01, 0x02}...) // Energy meter identifier Data length: 6 byte Susy-ID: 270 (0x10E) SerNo.: 258 (0x102)
	tickerMeasuringTimeByte := []byte(Uint32ToBytes(uint32(time.Now().UnixMilli())))

	for len(tickerMeasuringTimeByte) < 4 {
		tickerMeasuringTimeByte = append([]byte{0x00}, tickerMeasuringTimeByte...)
	}

	datagram = append(datagram, tickerMeasuringTimeByte...)

	file, err := os.Open(config.SharedDir + "/" + config.SharedFile)
	utils.Checklog(err)
	defer file.Close()
	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	// records, err := reader.Read()
	// log.Debugf("%v", records)
	// utils.Checklog(err)
	// if len(records) >= 19 {
	// 	// for i := 1; i < len(records)-1; i++ {
	// 	// 	registervalue = 0
	// 	// 	val, err := strconv.ParseFloat(records[i], 32)
	// 	// 	if err != nil {
	// 	// 		log.Fatal("error converting value", err)
	// 	// 	} else {
	// 	// 		registervalue = math.Float32bits(float32(val))
	// 	// 	}
	// 	// 	serv.HoldingRegisters[2*i-2] = uint16(registervalue >> 16)
	// 	// 	serv.HoldingRegisters[2*i-1] = uint16(registervalue)
	// 	// }
	// } else {
	// 	log.Fatal("Values not written")
	// }

	conn, err := multicast.NewBroadcaster(config.EmeterMulticastAddress + ":" + strconv.Itoa(config.EmeterMulticastPort))

	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte(datagram))

	file.Close()

	// for {
	// 	// conn.Write([]byte("hello, world\n"))
	// 	var datagram []byte
	// 	datagram = append([]byte("SMA"))
	// 	datagram = append(datagram, []byte{0x00, 0x04}...)                         // Data length: 4 byte (0x00000004)
	// 	datagram = append(datagram, []byte{0x02, 0xA0}...)                         // Tag: "Tag0" (42), version 0
	// 	datagram = append(datagram, []byte{0x00, 0x00, 0x00, 0x01}...)             // Group1 (default group)
	// 	datagram = append(datagram, []byte{0x00, 0x2C}...)                         // Data length: 44 byte (variable)
	// 	datagram = append(datagram, []byte{0x00, 0x10}...)                         // Tag: "SMA Net 2", version 0
	// 	datagram = append(datagram, []byte{0x60, 0x69}...)                         // Protocol ID (energy meter protocol), Data length: 2 byte
	// 	datagram = append(datagram, []byte{0x01, 0x0E, 0x00, 0x00, 0x01, 0x02}...) // Energy meter identifier Data length: 6 byte Susy-ID: 270 (0x10E) SerNo.: 258 (0x102)

	// 	conn.Write([]byte(datagram))

	// 	// conn.Write([]byte("SMA0"))
	// 	time.Sleep(1 * time.Second)
	// }
}

func Uint32ToBytes(i uint32) []byte {
	if i > 0 {
		return append(big.NewInt(int64(i)).Bytes(), byte(1))
	}
	return append(big.NewInt(int64(i)).Bytes(), byte(0))
}

func IntToBytes(i int) []byte {
	if i > 0 {
		return append(big.NewInt(int64(i)).Bytes(), byte(1))
	}
	return append(big.NewInt(int64(i)).Bytes(), byte(0))
}
func BytesToInt(b []byte) int {
	if b[len(b)-1] == 0 {
		return -int(big.NewInt(0).SetBytes(b[:len(b)-1]).Int64())
	}
	return int(big.NewInt(0).SetBytes(b[:len(b)-1]).Int64())
}
