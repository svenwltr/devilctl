package raumfeld

import "encoding/xml"

type xmlUPNPPropertySet struct {
	Properties []xmlUPNPProperty `xml:"property"`
}

type xmlUPNPProperty struct {
	LastChange string `xml:"LastChange"`
}

type xmlRaumfeldEvent struct {
	XMLName  xml.Name            `xml:"Event"`
	Instance xmlRaumfeldInstance `xml:"InstanceID"`
}

type xmlRaumfeldInstance struct {
	Volume     *xmlRaumfeldVolume     `xml:"Volume,omitempty"`
	Mute       *xmlRaumfeldMute       `xml:"Mute,omitempty"`
	PowerState *xmlRaumfeldPowerState `xml:"PowerState,omitempty"`
}

type xmlRaumfeldVolume struct {
	Channel string `xml:"Channel,attr"`
	Value   int    `xml:"val,attr"`
}

type xmlRaumfeldMute struct {
	Channel string `xml:"Channel,attr"`
	Value   int    `xml:"val,attr"`
}

type xmlRaumfeldPowerState struct {
	Value string `xml:"val,attr"`
}
