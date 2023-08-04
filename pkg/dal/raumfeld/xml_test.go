package raumfeld

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRaumfeldXMLEncode encodes a XML file from the given data. This is not
// used anywhere, but help understanding how the XML structure for the structs
// look like.
func TestRaumfeldXMLEncode(t *testing.T) {
	data := xmlRaumfeldEvent{
		Instance: xmlRaumfeldInstance{
			Volume: &xmlRaumfeldVolume{
				Channel: "Master",
				Value:   12,
			},
			PowerState: &xmlRaumfeldPowerState{
				Value: "ACTIVE",
			},
		},
	}

	payload, err := xml.Marshal(&data)
	require.NoError(t, err)

	want := `<Event><InstanceID><Volume Channel="Master" val="12"></Volume><PowerState val="ACTIVE"></PowerState></InstanceID></Event>`
	require.Equal(t, want, string(payload))
}

func TestRaumfeldXMLDecodeRCS(t *testing.T) {
	payload := `<Event xmlns="urn:schemas-upnp-org:metadata-1-0/RCS/"><InstanceID val="0"><Volume Channel="Master" val="6"/></InstanceID></Event>`

	var have xmlRaumfeldEvent

	err := xml.Unmarshal([]byte(payload), &have)
	require.NoError(t, err)

	t.Logf("%#v", have)
	require.Equal(t, 6, have.Instance.Volume.Value)
	require.Equal(t, "Master", have.Instance.Volume.Channel)

}

func TestRaumfeldXMLDecodeAVT(t *testing.T) {
	payload := `<Event xmlns="urn:schemas-upnp-org:metadata-1-0/AVT/"><InstanceID val="0"><PowerState val="ACTIVE"/></InstanceID></Event>`

	var have xmlRaumfeldEvent

	err := xml.Unmarshal([]byte(payload), &have)
	require.NoError(t, err)

	t.Logf("%#v", have)
	require.Equal(t, "ACTIVE", have.Instance.PowerState.Value)
}
