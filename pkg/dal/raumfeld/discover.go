package raumfeld

import (
	"context"
	"fmt"

	"github.com/huin/goupnp"
)

// This URN does not provide any controls, but it help identifying actual
// devices and not some bunch of over virtual speakers.
const RaumfeldTypeURN = `urn:schemas-raumfeld-com:service:RaumfeldGenerator:1`

const (
	ChannelMaster = "Master"
	InstanceID    = 1
)

func Discover(ctx context.Context) (map[string]Speaker, error) {
	devices, err := goupnp.DiscoverDevicesCtx(ctx, RaumfeldTypeURN)
	if err != nil {
		return nil, fmt.Errorf("discover devices: %w", err)
	}

	result := map[string]Speaker{}

	for _, device := range devices {
		if device.Err != nil {
			return nil, fmt.Errorf("read device %#v: %w", device.USN, device.Err)
		}

		speaker, err := newFromRootDevice(ctx, device.Location, device.Root)
		if err != nil {
			return nil, err
		}
		speaker.localAddr = device.LocalAddr

		result[speaker.ID()] = speaker
	}

	return result, nil
}
