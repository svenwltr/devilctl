package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/svenwltr/devilctl/pkg/bll/ticker"
	"github.com/svenwltr/devilctl/pkg/dal/homie"
	"github.com/svenwltr/devilctl/pkg/dal/raumfeld"
	"golang.org/x/sync/errgroup"
)

type HomieBridgeRunner struct {
	broker string
}

func (r *HomieBridgeRunner) Bind(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringVar(
		&r.broker, "broker", "",
		`The broker MQTT URI. ex: tcp://10.10.1.1:1883`)
	return nil
}

func (r *HomieBridgeRunner) Run(ctx context.Context) error {
	homieBroker, err := homie.New(r.broker)
	if err != nil {
		return fmt.Errorf("create homie broker: %w", err)
	}
	defer homieBroker.MustClose()

	bridge := HomieBridge{
		Broker: homieBroker,
	}

	return bridge.Run(ctx)
}

type HomieBridge struct {
	Broker   *homie.Broker
	Speakers map[string]raumfeld.Speaker
}

func (b *HomieBridge) Run(ctx context.Context) error {
	sub, err := raumfeld.NewSubsciptionServer(b)
	if err != nil {
		return fmt.Errorf("create subscription server: %w", err)
	}

	var enableHandler sync.Once

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return sub.Run(ctx)
	})

	group.Go(func() error {
		for range ticker.Every(ctx, 5*time.Minute) {
			speakers, err := raumfeld.Discover(ctx)
			if err != nil {
				return fmt.Errorf("discover speakers: %w", err)
			}
			logrus.Infof("discovered %d devices", len(speakers))
			b.Speakers = speakers

			enableHandler.Do(func() {
				b.Broker.ActionHandler = b.HandleBrokerAction
			})

			err = b.PublishHomieDefinitions(ctx)
			if err != nil {
				return fmt.Errorf("publish homie definitions: %w", err)
			}

			for _, speaker := range speakers {
				err := sub.Subscribe(speaker)
				if err != nil {
					return fmt.Errorf("subscribe: %w", err)
				}
			}

		}

		return nil
	})

	return group.Wait()
}

func (b *HomieBridge) HandleBrokerAction(nodeID, propertyID, value string) error {
	logrus.
		WithField("node-id", nodeID).
		WithField("property-id", propertyID).
		WithField("value", value).
		Info("received new action from broker")
	speaker, found := b.Speakers[nodeID]
	if !found {
		return fmt.Errorf("node %#v not found in cache", nodeID)
	}

	switch propertyID {
	case "volume":
		vol, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}

		return speaker.SetVolumeFloat(context.Background(), vol)

	case "mute":
		return speaker.SetMute(context.Background(), value == "true")

	case "onoff":
		return speaker.SetOnOff(context.Background(), value == "true")

	default:
		return fmt.Errorf("no action for property %#v", propertyID)

	}
}

func (b *HomieBridge) PublishHomieDefinitions(ctx context.Context) error {
	logrus.Infof("publishing homie nodes")

	device := homie.Device{
		Name:           "devilctl raumfeld-bridge",
		Implementation: "github.com/svenwltr/devilctl",
	}

	for nodeID, speaker := range b.Speakers {
		device.NodeIDs = append(device.NodeIDs, nodeID)
		err := errors.Join(
			b.Broker.PublishNode(homie.Node{
				NodeID:      nodeID,
				Name:        speaker.FriendlyName(),
				Type:        "Speaker",
				PropertyIDs: []string{"onoff", "volume", "mute"},
			}),
			b.Broker.PublishProperty(homie.Property{
				NodeID:     nodeID,
				PropertyID: "onoff",
				Name:       "On/Off",
				DataType:   "boolean",
				Retained:   true,
				Settable:   true,
			}),
			b.Broker.PublishProperty(homie.Property{
				NodeID:     nodeID,
				PropertyID: "volume",
				Name:       "Volume",
				DataType:   "float",
				Format:     "0:1",
				Retained:   true,
				Settable:   true,
			}),
			b.Broker.PublishProperty(homie.Property{
				NodeID:     nodeID,
				PropertyID: "mute",
				Name:       "Mute",
				DataType:   "boolean",
				Format:     "0:1",
				Retained:   true,
				Settable:   true,
			}),
		)
		if err != nil {
			return err
		}
	}

	return b.Broker.PublishDevice(device)
}

func (b *HomieBridge) OnVolumeChange(id string, volume int, channel string) {
	logrus.Infof("volume changed on speaker %#v to %#v", id, volume)
	b.Broker.PublishValue(id, "volume", float64(volume)/100.)
}

func (b *HomieBridge) OnMuteChange(id string, muted bool, channel string) {
	logrus.Infof("mute changed on speaker %#v to %#v", id, muted)
	b.Broker.PublishValue(id, "mute", muted)
}

func (b *HomieBridge) OnPowerStateChange(id, state string) {
	logrus.Infof("power state changed on speaker %#v to %#v", id, state)
	b.Broker.PublishValue(id, "onoff", state != "MANUAL_STANDBY")
}
