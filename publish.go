package main

import (
	"bytes"
	"encoding/json"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
	"time"
)

// Home Assitant MQTT Discovery support
type MqttDiscoveryMsg struct {
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	UnitOfMeasurement string `json:"unit_of_measurement"`
	UniqueId          string `json:"unique_id"`
	ExpireAfter       int    `json:"expire_after"`
	Qos               int    `json:"qos"`
	//SwVersion	    string `json:"sw_version"`
}

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	log.Debug("TOPIC/MSG", msg.Topic(), "/", msg.Payload())
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
	opts.SetUsername(username)
	opts.SetKeepAlive(120 * time.Second)
	opts.SetPingTimeout(20 * time.Second)
	opts.SetConnectionLostHandler(onLost)
	opts.SetOrderMatters(false)
	opts.SetOnConnectHandler(onConnect)
	opts.SetMaxReconnectInterval(5 * 60 * time.Second)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Panic("failed to connect to MQTT Broker, bailing out ", token.Error())
		os.Exit(-1)
	}

	return c
}

type OpenSenseMapJSON struct {
	Sensor string `json:"sensor"`
	Value  string `json:"value"`
}

//Register sensor with Home Assistant via MQTT
//requires HA MQTT Discovey to be enabled
func registerHaDiscovery(cl MQTT.Client, sensName string, configRoot string, rootTopic string) *MqttDiscoveryMsg {
	var newDisco = &MqttDiscoveryMsg{}
	newDisco.Name = "Gräfte"
	newDisco.UnitOfMeasurement = "mm"
	newDisco.Qos = 2

	saneName := strings.Join(strings.Fields(sensName), "_")
	newDisco.StateTopic = rootTopic + "/" + saneName + "/state"
	configTopic := configRoot + "/" + saneName + "/config"
	newDisco.ExpireAfter = 5 * 60 //seconds

	discoJson, err := json.Marshal(newDisco)
	if err != nil {
		log.Panic(err)
	}
	pub(cl, configTopic, string(discoJson))

	return newDisco
}

func pub(cl MQTT.Client, topic string, payload string) error {
	log.Debug("MQTT: ", topic, " <- ", payload)
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

func sanitizeParamName(paramName string) string {
	return strings.Join(strings.Fields(paramName), "_")
}
