# xm122level
Acconeer XM122 Fluid Level Metering

I use this thing to measure the level of a water body next to my house and publish the results to MQTT.


Further processing is HASS.IO MQTT, Home-Assistent, InfluxDB, Grafana.

This is neither modular nor does it support all XM122 services and detectors but perhaps it's a useful starting point for interacting with XM122 in golang.


# My Setup


excerpt from HA configuration.yaml 
```
sensor Gräfte:
  - platform: mqtt
    state_topic: "xm122level/state"
    name: "Gräfte Pegel"
    unit_of_measurement: "mm"
    qos: 2
    expire_after: 120
  - platform: mqtt
    state_topic: "xm122level/raw"
    name: "Gräfte Pegel Raw"
    unit_of_measurement: "mm"
    qos: 2
    expire_after: 120
```

qweqweq
  
![Holzbrettsensor](/holzbrettsensor.jpg)