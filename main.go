package main

import (
	"flag"
	"github.com/goburrow/modbus"
	"log"
	"os"
	"time"
)

var serialAddr string
var verboseFlag bool
var publishFlag bool
var readFrequency time.Duration
var logger *log.Logger

func main() {
	flag.StringVar(&serialAddr, "s", "/dev/ttyUSB0", "serial port")
	flag.BoolVar(&verboseFlag, "v", false, "run in verbose")
	flag.BoolVar(&publishFlag, "p", false, "publish data")
	flag.DurationVar(&readFrequency, "f", time.Second, "read & publish frequency")
	flag.Parse()

	if verboseFlag {
		logger = log.New(os.Stderr, "verbose", log.LstdFlags)
	}

	serial := modbus.NewRTUClientHandler(serialAddr)
	serial.BaudRate = 9600
	serial.DataBits = 8
	serial.StopBits = 1
	serial.Parity = "E"
	serial.Timeout = time.Second / 2
	serial.Logger = logger

	log.Printf("Connecting to %s\n", serialAddr)

	// Connect manually so that multiple requests are handled in one connection session
	err := serial.Connect()
	if err != nil {
		log.Panic(err)
	}

	defer serial.Close()

	mqtt := makeMQTTConn()

	pms := []*ModbusPowerMeter{
		NewModbusPowerMeter("ha-powermeter-bridge", 0x02, registers517, serial),
		NewModbusPowerMeter("ha-powermeter-bridge-heat-pump", 0x01, registers514, serial),
	}

	for _, pm := range pms {
		log.Printf("%s autodiscover\n", pm.mqttDevice.Identifiers)
		pm.Logger = logger
		if err := pm.AutoDiscover(mqtt); err != nil {
			log.Panic(err)
		}
	}

	for {
		for _, pm := range pms {
			if publishFlag {
				log.Printf("%s read and publish\n", pm.mqttDevice.Identifiers)

				pm.ReadAllAndPublishTo(mqtt)
			} else {
				log.Printf("%s read\n", pm.mqttDevice.Identifiers)

				readPowerMeter(pm)
			}
		}

		time.Sleep(readFrequency)
	}
}

func readPowerMeter(pm *ModbusPowerMeter) {
	data := pm.ReadAll()

	for name, value := range data {
		log.Printf("%s = %s\n", name, value)
	}
}
