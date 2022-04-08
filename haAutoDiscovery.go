package main

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttDevice struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	Name         string `json:"name,omitempty"`
	Identifiers  string `json:"identifiers,omitempty"`
}

type mqttSensor struct {
	Name              string     `json:"name,omitempty"`
	StateTopic        string     `json:"state_topic"`
	StateClass        string     `json:"state_class"`
	AvailabilityTopic string     `json:"availability_topic,omitempty"`
	UnitOfMeasurement string     `json:"unit_of_measurement,omitempty"`
	DeviceClass       string     `json:"device_class,omitempty"`
	ForceUpdate       bool       `json:"force_update,omitempty"`
	ExpireAfter       int        `json:"expire_after,omitempty"`
	UniqueID          string     `json:"unique_id,omitempty"`
	Device            mqttDevice `json:"device"`
}

func getDeviceClass(unit string) string {
	switch unit {
	case "W":
		return "power"
	case "kW":
		return "power"
	case "Wh":
		return "energy"
	case "kWh":
		return "energy"
	case "A":
		return "current"
	}
	return ""
}

func nameToIdentifier(sensorName string) string {
	return strings.ReplaceAll(strings.ToUpper(sensorName), " ", "_")
}

func getStateTopic(device mqttDevice, sensorName string) string {
	return fmt.Sprintf("%s/%s", device.Identifiers, nameToIdentifier(sensorName))
}

func encodeSensor(sensorName, unit, stateClass string, device mqttDevice) (topic string, data []byte, err error) {
	var s mqttSensor
	sensorID := nameToIdentifier(sensorName)
	s.Name = sensorName
	s.StateTopic = getStateTopic(device, sensorName)
	s.StateClass = stateClass
	//s.AvailabilityTopic = config.mqttWillTopic
	s.UnitOfMeasurement = unit
	s.DeviceClass = getDeviceClass(unit)
	s.UniqueID = device.Identifiers + "_" + sensorID
	s.Device = device

	topic = fmt.Sprintf("homeassistant/sensor/%s/%s/config", device.Identifiers, sensorID)
	data, err = json.Marshal(s)

	return topic, data, err
}

func discover(client mqtt.Client, device mqttDevice, name string, rs registerToSensor) error {
	sDef := rs.s

	if sDef == nil {
		return nil
	}

	topic, data, err := encodeSensor(name, sDef.unit, sDef.stateClass, device)
	if err != nil {
		return err
	}

	token := client.Publish(topic, 0, true, data)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func publish(client mqtt.Client, device mqttDevice, name string, data interface{}) error {
	token := client.Publish(getStateTopic(device, name), 0, true, data)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
