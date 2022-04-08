package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	"log"
	"sync"
)

type ModbusPowerMeter struct {
	Logger *log.Logger

	mqttDevice mqttDevice

	modbus modbus.Client
	regs   registers
}

func NewModbusPowerMeter(name string, slaveID uint8, regs registers, transporter modbus.Transporter) *ModbusPowerMeter {
	packer := modbus.NewRTUClientHandler(name)
	packer.SlaveId = slaveID

	device := mqttDevice{
		Manufacturer: "Jakub Kaflik",
		Model:        "ha-powermeter-bridge",
		Name:         name,
		Identifiers:  name,
	}

	return &ModbusPowerMeter{
		mqttDevice: device,
		modbus:     modbus.NewClient2(packer, transporter),
		regs:       regs,
	}
}

func (m *ModbusPowerMeter) AutoDiscover(mqtt mqtt.Client) error {
	for name, r := range m.regs {
		if r.s == nil {
			continue
		}

		topic, payload, err := m.encodeSensorDiscovery(name, r.s.stateClass, r.s.unit)
		if err != nil {
			return errors.Wrapf(err, "encoding %s sensor discovery error", name)
		}

		token := mqtt.Publish(topic, 0, true, payload)
		if token.Wait() && token.Error() != nil {
			return errors.Wrapf(token.Error(), "publish %s sensor discovery error", name)
		}
	}

	return nil
}

func (m *ModbusPowerMeter) ReadAllAndPublishTo(mqtt mqtt.Client) {
	data := m.ReadAll()

	for name, value := range data {
		topic := getStateTopic(m.mqttDevice, name)
		payload := fmt.Sprint(value)
		token := mqtt.Publish(topic, 0, true, payload)
		if token.Wait() && token.Error() != nil {
			if m.Logger != nil {
				m.Logger.Printf("publish %s sensor state error: %s", name, token.Error())
			}
			continue
		}
	}
}

func (m *ModbusPowerMeter) ReadAll() map[string]interface{} {
	data := make(map[string]interface{})

	for name, r := range m.regs {
		value, err := m.readRegister(r.reg)
		if err != nil {
			if m.Logger != nil {
				m.Logger.Print(err)
			}
			continue
		}

		data[name] = value
	}

	return data
}

var lock sync.Mutex

func (m *ModbusPowerMeter) readRegister(reg register) (interface{}, error) {
	lock.Lock()
	defer lock.Unlock()

	data, err := m.modbus.ReadHoldingRegisters(reg.addr, reg.size)
	if err != nil {
		return nil, errors.Wrapf(err, "read holding register 0x%X failed", reg.addr)
	}
	if len(data) == 0 {
		return nil, errors.Errorf("empty data return from register 0x%X", reg.addr)
	}

	return reg.converter(data), nil
}

func (m *ModbusPowerMeter) encodeSensorDiscovery(sensorName, stateClass, unit string) (topic string, payload []byte, err error) {
	sensorID := nameToIdentifier(sensorName)
	topic = fmt.Sprintf("homeassistant/sensor/%s/%s/config", m.mqttDevice.Identifiers, sensorID)

	sensor := mqttSensor{
		Name:              sensorName,
		StateTopic:        getStateTopic(m.mqttDevice, sensorName),
		StateClass:        stateClass,
		UnitOfMeasurement: unit,
		DeviceClass:       getDeviceClass(unit),
		UniqueID:          fmt.Sprintf("%s_%s", m.mqttDevice.Identifiers, sensorID),
		Device:            m.mqttDevice,
	}

	payload, err = json.Marshal(sensor)

	return topic, payload, err
}
