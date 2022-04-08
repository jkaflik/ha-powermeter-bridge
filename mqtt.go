package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//
//func onMQTTConnect(mclient mqtt.Client) {
//	mqttPublish(mclient, config.mqttWillTopic, "online", 0)
//	if config.ListenOnly == false {
//		mclient.Subscribe(getCommandTopic("+"), 0, onGenericCommand)
//		mclient.Subscribe(getStateTopic("+/set"), 0, onAquareaCommand)
//	}
//	log.Println("MQTT connected")
//}

func makeMQTTConn() mqtt.Client {
	log.Println("Setting up MQTT...")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%v", "tcp", "10.0.10.10", "1883"))
	opts.SetPassword("panasonic")
	opts.SetUsername("panasonic")
	opts.SetClientID("ha-powermeter-bridge")
	//opts.SetWill(config.mqttWillTopic, "offline", 0, true)
	opts.SetKeepAlive(time.Second * 5)

	opts.SetCleanSession(true)  // don't want to receive entire backlog of setting changes
	opts.SetAutoReconnect(true) // default, but I want it explicit
	opts.SetConnectRetry(true)
	//opts.SetOnConnectHandler(onMQTTConnect)

	// connect to broker
	client := mqtt.NewClient(opts)

	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatalf("Fail to connect broker, %v", token.Error())
		//should not happen - SetConnectRetry=true
	}
	return client
}
