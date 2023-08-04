package raumfeld

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func (s Speaker) Subscribe(ctx context.Context, handler SubscribeHandler) error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("start tcp listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for ctx.Err() == nil {
			logrus.Infof("refeshing subscription for %#v", s.ID())

			err = s.requestSubscribe(port, "urn:upnp-org:serviceId:AVTransport")
			if err != nil {
				logrus.Errorf("subscribe av: %v", err.Error())
			}

			err = s.requestSubscribe(port, "urn:upnp-org:serviceId:RenderingControl")
			if err != nil {
				logrus.Errorf("subscribe rc: %v", err.Error())
			}

			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Minute):
			}

		}
	}()

	server := new(http.Server)
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var root xmlUPNPPropertySet
		err = xml.NewDecoder(r.Body).Decode(&root)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, p := range root.Properties {
			if p.LastChange == "" {
				continue
			}

			var event xmlRaumfeldEvent
			err := xml.Unmarshal([]byte(p.LastChange), &event)
			if err != nil {
				fmt.Println(err)
				return
			}

			if event.Instance.Volume != nil {
				handler.OnVolumeChange(s, event.Instance.Volume.Value, event.Instance.Volume.Channel)
			}

			if event.Instance.Mute != nil {
				muted := event.Instance.Mute.Value > 0
				handler.OnMuteChange(s, muted, event.Instance.Mute.Channel)
			}

			if event.Instance.PowerState != nil {
				handler.OnPowerStateChange(s, event.Instance.PowerState.Value)
			}

		}
	})

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return server.Serve(listener)
}

func (s Speaker) requestSubscribe(port int, service string) error {
	subURL := *s.location
	subURL.Path = s.eventSubURLs[service]

	r, err := http.NewRequest("SUBSCRIBE", subURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create subscribe request: %w", err)
	}
	r.Header.Set("NT", "upnp:event")
	r.Header.Set("Callback", fmt.Sprintf("<http://%s:%d/>", s.localAddr.String(), port))
	r.Header.Set("Timeout", "Second-1800")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("send subscribe request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status from subscribe: %s", resp.Status)
	}

	return nil
}

type SubscribeHandler interface {
	OnVolumeChange(s Speaker, volume int, channel string)
	OnPowerStateChange(s Speaker, state string)
	OnMuteChange(s Speaker, muted bool, channel string)
}

type SubscribeHandlerFuncs struct {
	VolumeChange     func(s Speaker, volume int, channel string)
	PowerStateChange func(s Speaker, state string)
	MuteChange       func(s Speaker, muted bool, channel string)
}

func (h SubscribeHandlerFuncs) OnVolumeChange(s Speaker, volume int, channel string) {
	if h.VolumeChange != nil {
		h.VolumeChange(s, volume, channel)
	}
}

func (h SubscribeHandlerFuncs) OnPowerStateChange(s Speaker, state string) {
	if h.PowerStateChange != nil {
		h.PowerStateChange(s, state)
	}
}

func (h SubscribeHandlerFuncs) OnMuteChange(s Speaker, muted bool, channel string) {
	if h.MuteChange != nil {
		h.MuteChange(s, muted, channel)
	}
}
