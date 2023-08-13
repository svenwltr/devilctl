package raumfeld

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func init() {
	chi.RegisterMethod("NOTIFY")
}

type SubscriptionServer struct {
	handler  SubscribeHandler
	listener net.Listener
}

func NewSubsciptionServer(handler SubscribeHandler) (*SubscriptionServer, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("start tcp listener: %w", err)
	}

	return &SubscriptionServer{
		handler:  handler,
		listener: listener,
	}, nil
}

func (s SubscriptionServer) Run(ctx context.Context) error {
	r := chi.NewRouter()
	r.MethodFunc("NOTIFY", "/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			speakerID := chi.URLParam(r, "id")

			var root xmlUPNPPropertySet
			err := xml.NewDecoder(r.Body).Decode(&root)
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
					s.handler.OnVolumeChange(speakerID, event.Instance.Volume.Value, event.Instance.Volume.Channel)
				}

				if event.Instance.Mute != nil {
					muted := event.Instance.Mute.Value > 0
					s.handler.OnMuteChange(speakerID, muted, event.Instance.Mute.Channel)
				}

				if event.Instance.PowerState != nil {
					s.handler.OnPowerStateChange(speakerID, event.Instance.PowerState.Value)
				}

			}
		})
	server := new(http.Server)
	server.Handler = r

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return server.Serve(s.listener)
}

func (s SubscriptionServer) Subscribe(speaker Speaker) error {
	logrus.Infof("refeshing subscription for %#v", speaker.location.String())

	return errors.Join(
		s.subscribeService(speaker, "urn:upnp-org:serviceId:AVTransport"),
		s.subscribeService(speaker, "urn:upnp-org:serviceId:RenderingControl"),
	)
}

func (s SubscriptionServer) subscribeService(speaker Speaker, service string) error {
	port := s.listener.Addr().(*net.TCPAddr).Port

	subURL := *speaker.location
	subURL.Path = speaker.eventSubURLs[service]

	r, err := http.NewRequest("SUBSCRIBE", subURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create subscribe request: %w", err)
	}
	r.Header.Set("NT", "upnp:event")
	r.Header.Set("Callback", fmt.Sprintf("<http://%s:%d/%s>", speaker.localAddr.String(), port, speaker.id))
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
	OnVolumeChange(id string, volume int, channel string)
	OnPowerStateChange(id string, state string)
	OnMuteChange(id string, muted bool, channel string)
}

type SubscribeHandlerFuncs struct {
	VolumeChange     func(id string, volume int, channel string)
	PowerStateChange func(id string, state string)
	MuteChange       func(id string, muted bool, channel string)
}

func (h SubscribeHandlerFuncs) OnVolumeChange(id string, volume int, channel string) {
	if h.VolumeChange != nil {
		h.VolumeChange(id, volume, channel)
	}
}

func (h SubscribeHandlerFuncs) OnPowerStateChange(id string, state string) {
	if h.PowerStateChange != nil {
		h.PowerStateChange(id, state)
	}
}

func (h SubscribeHandlerFuncs) OnMuteChange(id string, muted bool, channel string) {
	if h.MuteChange != nil {
		h.MuteChange(id, muted, channel)
	}
}
