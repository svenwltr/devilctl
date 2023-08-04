# devilctl

Bridge Teufel Raumfeld speakers to MQTT with [Homie](https://homieiot.github.io/) support.

> **Development Status:** devilctl is built for integrating my Teufel Raumfeld speakers with my [Homey](https://homey.app/) over [Homie](https://homieiot.github.io/) [sic]. There is no effort made that this is usable for others, but because the support for Teufel Raumfeld in the IoT area is quite bad, I am publishing anyway.

> This project is open source, but not open for contribution, since helping contributors to ship a PR usually takes more time than doing it myself. Please drop an issue, if you want something changed. Also of course it is possible to fork the whole project.

## Example

### Discover Speakers in the Network

```
$ devilctl discover
---------
Name:             Wohnzimmer
ID:               0500bb45-f61b-4c44-9565-919a6441c99f
Location:         http://raumfeld-living-room.fritz.box.:54888/0500bb45-f61b-4c44-9565-919a6441c99f.xml
Discovered From:  192.168.242.136
INFO[0002] refeshing subscription for "0500bb45-f61b-4c44-9565-919a6441c99f"
PowerState:       AUTOMATIC_STANDBY
Volume:           9
Muted:            false
---------
Name:             Küche
ID:               cd19c884-dcea-4368-bcb2-fa70d3165631
Location:         http://raumfeld-kitchen.fritz.box.:56838/cd19c884-dcea-4368-bcb2-fa70d3165631.xml
Discovered From:  192.168.242.136
INFO[0002] refeshing subscription for "cd19c884-dcea-4368-bcb2-fa70d3165631"
PowerState:       AUTOMATIC_STANDBY
Volume:           13
Muted:            false
```


### Homie Bridge

```
$ devilctl homie-bridge --broker mqtt://localhost:1883 --location http://raumfeld-kitchen.fritz.box.:56838/cd19c884-dcea-4368-bcb2-fa70d3165631.xml
INFO[0000] connected to "http://raumfeld-kitchen.fritz.box.:56838/cd19c884-dcea-4368-bcb2-fa70d3165631.xml"
INFO[0000] connected to "http://raumfeld-living-room.fritz.box.:54888/0500bb45-f61b-4c44-9565-919a6441c99f.xml"
INFO[0000] publishing homie nodes
INFO[0000] refeshing subscription for "cd19c884-dcea-4368-bcb2-fa70d3165631"
INFO[0000] refeshing subscription for "0500bb45-f61b-4c44-9565-919a6441c99f"
INFO[0000] power state changed on speaker to "AUTOMATIC_STANDBY"
INFO[0000] power state changed on speaker to "AUTOMATIC_STANDBY"
INFO[0000] volume changed on speaker to 9
INFO[0000] mute changed on speaker to false
INFO[0000] volume changed on speaker to 13
INFO[0000] mute changed on speaker to false
INFO[0006] mute changed on speaker to true
INFO[0007] mute changed on speaker to true
INFO[0009] mute changed on speaker to false
INFO[0010] mute changed on speaker to false
INFO[0017] received new action from broker               node-id=cd19c884-dcea-4368-bcb2-fa70d3165631 property-id=onoff value=false
INFO[0017] power state changed on speaker to "MANUAL_STANDBY"
INFO[0019] received new action from broker               node-id=0500bb45-f61b-4c44-9565-919a6441c99f property-id=onoff value=false
INFO[0019] power state changed on speaker to "MANUAL_STANDBY"
```

The corresponding MQTT messages look like this:

```
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/$name] On/Off
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/$format]
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/$settable] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/$unit]
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/$retained] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/$datatype] boolean
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume/$datatype] float
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume/$format] 0:1
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume/$unit]
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume/$settable] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute/$name] Mute
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume/$retained] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute/$datatype] boolean
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute/$format] 0:1
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute/$unit]
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume/$name] Volume
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute/$retained] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute/$settable] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/$name] Küche
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/$properties] mute,onoff,volume
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/$name] On/Off
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/$datatype] boolean
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/$format]
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/$settable] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/$unit]
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/$retained] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume/$name] Volume
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/$type] Speaker
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume/$datatype] float
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume/$format] 0:1
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume/$unit]
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume/$settable] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute/$name] Mute
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume/$retained] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute/$datatype] boolean
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute/$format] 0:1
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute/$unit]
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute/$retained] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute/$settable] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/$name] Wohnzimmer
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/$type] Speaker
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/$properties] mute,onoff,volume
2023/08/04 17:23:56 [homie/raumfeld-bridge/$name] Homie Raumfeld Bridge
2023/08/04 17:23:56 [homie/raumfeld-bridge/$homie] 4.0.0
2023/08/04 17:23:56 [homie/raumfeld-bridge/$state] ready
2023/08/04 17:23:56 [homie/raumfeld-bridge/$implementation] github.com/svenwltr/devilctl
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/$nodes] cd19c884-dcea-4368-bcb2-fa70d3165631,0500bb45-f61b-4c44-9565-919a6441c99f
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff] true
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/volume] 0.09
2023/08/04 17:23:56 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute] false
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/volume] 0.13
2023/08/04 17:23:56 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute] false
2023/08/04 17:24:02 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute] true
2023/08/04 17:24:03 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute] true
2023/08/04 17:24:05 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/mute] false
2023/08/04 17:24:06 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/mute] false
2023/08/04 17:24:14 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff/set] false
2023/08/04 17:24:14 [homie/raumfeld-bridge/cd19c884-dcea-4368-bcb2-fa70d3165631/onoff] false
2023/08/04 17:24:16 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff/set] false
2023/08/04 17:24:16 [homie/raumfeld-bridge/0500bb45-f61b-4c44-9565-919a6441c99f/onoff] false
2023/08/04 17:24:22 [homie/raumfeld-bridge/$state] disconnected
```
