package main

import (
	"bytes"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	log.WithFields(log.Fields{
		"topic":   msg.Topic(),
		"payload": msg.Payload(),
	}).Debug("MQTT recieved")
	//msg.Ack()
}

func connectMQTT(host string, username string, password string) MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(host)
	//when testing two clients may be running thus we grab a MAC address to create a semi-static machine specific clientID
	clientId := "xm122level-" + getMacAddr()
	log.Info("Connecting as ", clientId)
	opts.SetClientID(clientId)
	opts.SetDefaultPublishHandler(f)
	opts.SetAutoReconnect(true)
	opts.SetPassword(password)
	log.Debug("MQTT username: ", username)
	opts.SetUsername(username)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(20 * time.Second)
	opts.SetConnectionLostHandler(onLost)
	opts.SetOrderMatters(false)
	opts.SetOnConnectHandler(onConnect)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetMessageChannelDepth(500)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Panic("failed to connect to MQTT Broker, bailing out: ", token.Error())
		os.Exit(-1)
	}

	return c
}

/*
type OpenSenseMapJSON struct {
	Sensor string `json:"sensor"`
	Value  string `json:"value"`
} */

func pub(cl MQTT.Client, topic string, payload string) error {
	log.WithFields(log.Fields{
		"topic":   topic,
		"payload": payload,
	}).Debug("MQTT")

	if token := cl.Publish(topic, 1, false, payload); token.Wait() && token.Error() != nil {
		log.Error("failed to publish message to ", topic, " error: ", token.Error())
		return token.Error()
	}
	return nil
}

func getMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				// Don't use random as we have a real address
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}

func onConnect(client MQTT.Client) {
	log.Info("MQTT client connected.")
}

func onLost(client MQTT.Client, err error) {
	log.Warn("MQTT connection lost: ", err)
}
