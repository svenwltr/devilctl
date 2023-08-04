package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/svenwltr/localh0rst/apps/devilctl/pkg/dal/homie"
	"gitlab.com/svenwltr/localh0rst/apps/devilctl/pkg/dal/raumfeld"
	"golang.org/x/sync/errgroup"
)

type HomieBridgeRunner struct {
	broker    string
	locations []string
}

func (r *HomieBridgeRunner) Bind(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringVar(
		&r.broker, "broker", "",
		`The broker MQTT URI. ex: tcp://10.10.1.1:1883`)
	cmd.PersistentFlags().StringSliceVar(
		&r.locations, "location", nil,
		`Pin speakers by their URL. URLs are visible with "devilctl discover"`)
	return nil
}

func (r *HomieBridgeRunner) Run(ctx context.Context) error {
	homieBroker, err := homie.New(r.broker)
	if err != nil {
		return fmt.Errorf("create homie broker: %w", err)
	}
	defer homieBroker.MustClose()

	speakers := map[string]raumfeld.Speaker{}
	if len(r.locations) == 0 {
		speakers, err = raumfeld.Discover(ctx)
		if err != nil {
			return fmt.Errorf("discover speakers: %w", err)
		}
		logrus.Infof("discovered %d devices", len(speakers))

		for _, speaker := range speakers {
			logrus.Warnf("consider pinning %#v with the --location flag",
				speaker.TryMDNSLocation().String())
		}

	} else {
		for _, locationString := range r.locations {
			location, err := url.Parse(locationString)
			if err != nil {
				return fmt.Errorf("parse %#v: %w", location, err)
			}

			speaker, err := raumfeld.New(ctx, location)
			if err != nil {
				return fmt.Errorf("connect to speaker %#v: %w", locationString, err)
			}

			logrus.Infof("connected to %#v", locationString)

			speakers[speaker.ID()] = speaker
		}
	}

	bridge := HomieBridge{
		Broker:   homieBroker,
		Speakers: speakers,
	}

	return bridge.Run(ctx)
}

type HomieBridge struct {
	Broker   *homie.Device
	Speakers map[string]raumfeld.Speaker
}

func (b *HomieBridge) Run(ctx context.Context) error {
	b.Broker.ActionHandler = b.HandleBrokerAction

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		for ctx.Err() == nil {
			err := b.PublishHomieDefinitions(ctx)
			if err != nil {
				return fmt.Errorf("publish homie definitions: %w", err)
			}

			select {
			case <-ctx.Done():
			case <-time.After(time.Minute):
			}
		}
		return nil
	})

	for _, speaker := range b.Speakers {
		speaker := speaker
		group.Go(func() error {
			return speaker.Subscribe(ctx, b)
		})
	}

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

	nodes := homie.Nodes{}

	for usn, speaker := range b.Speakers {
		nodes[usn] = homie.Node{
			Name: speaker.FriendlyName(),
			Type: "Speaker",
			Properties: homie.Properties{
				"onoff": homie.Property{
					Name:     "On/Off",
					DataType: "boolean",
					Retained: true,
					Settable: true,
				},
				"volume": homie.Property{
					Name:     "Volume",
					DataType: "float",
					Format:   "0:1",
					Retained: true,
					Settable: true,
				},
				"mute": homie.Property{
					Name:     "Mute",
					DataType: "boolean",
					Format:   "0:1",
					Retained: true,
					Settable: true,
				},
			},
		}
	}

	b.Broker.Nodes = nodes
	return b.Broker.PublishAll()
}

func (b *HomieBridge) OnVolumeChange(s raumfeld.Speaker, volume int, channel string) {
	logrus.Infof("volume changed on speaker to %#v", volume)
	b.Broker.Value(s.ID(), "volume", float64(volume)/100.)
}

func (b *HomieBridge) OnMuteChange(s raumfeld.Speaker, muted bool, channel string) {
	logrus.Infof("mute changed on speaker to %#v", muted)
	b.Broker.Value(s.ID(), "mute", muted)
}

func (b *HomieBridge) OnPowerStateChange(s raumfeld.Speaker, state string) {
	logrus.Infof("power state changed on speaker to %#v", state)
	b.Broker.Value(s.ID(), "onoff", state != "MANUAL_STANDBY")
}
