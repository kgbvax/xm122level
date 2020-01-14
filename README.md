# xm122level
[Acconeer XM122](https://www.acconeer.com/products) Fluid Level Metering

I use this thing to measure the level of a water body next to my house and publish the results to MQTT.

This is neither modular nor does it support all XM122 services and detectors but perhaps it's a useful starting point for interacting with XM122 in golang.

Measurements are taken with a frequency defined by `rate`. When multiple echos occur, only the echo with the strongest signal is considered.
Measurements are posted each `reportSec` seconds using a moving average of the same duration.


```
Reads distance from Acconeer XM122 and publishes to MQTT (Home Assistant)

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
  -d, --debug              Enable debug mode. Env: DEBUG
  -p, --port=PORT          serial port device name
  -b, --broker=BROKER      address of MQTT broker to connect to, e.g. tcp://mqtt.eclipse.org:1883. Env: BROKER
      --mqttUser=MQTTUSER  username for mqtt broker Env: BROKER_USER
      --mqttPassword=MQTTPASSWORD  
                           password for mqtt broker user. Env: BROKER_PW
      --stateTopic="xm122level/state"  
                           MQTT state topic
      --rawTopic=RAWTOPIC  if set, MQTT topic that receives all (unprocessed) measurements
      --rangeStart=300     Start (min) of measurement range in mm
      --rangeEnd=1000      End (max) of measurement range in mm
  -r, --rate=500           Measurement frequency in 1/1000 Hertz
  -o, --offset=0           Sensor level offset, subtracted from raw reading (in mm)
      --reportSec=60       report a moving average over <value> seconds every <value> seconds

```

# My Setup

Hardware: XM122 hooked up to a RPi3 via USB, power via POE Hat, all in a simple plastic box.
Further processing is: MQTT -> Home-Assistant -> Node-Red -> InfluxDB -> Grafana.

excerpt from HA configuration.yaml 
```
sensor Gräfte:
  - platform: mqtt
    state_topic: "xm122level/state"
    name: "Gräfte Pegel"
    unit_of_measurement: "mm"
    qos: 2
    expire_after: 240
  - platform: mqtt
    state_topic: "xm122level/raw"
    name: "Gräfte Pegel Raw"
    unit_of_measurement: "mm"
    qos: 2
    expire_after: 120
```


  
![Holzbrettsensor](/holzbrettsensor.jpg)
