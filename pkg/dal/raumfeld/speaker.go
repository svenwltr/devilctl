package raumfeld

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/gosimple/slug"
	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/av1"
	"github.com/sirupsen/logrus"
)

type Speaker struct {
	id           string
	location     *url.URL
	friendlyName string
	localAddr    net.IP
	eventSubURLs map[string]string

	av1 *av1.AVTransport1
	rc1 *av1.RenderingControl1
}

func New(ctx context.Context, location *url.URL) (Speaker, error) {
	root, err := goupnp.DeviceByURLCtx(ctx, location)
	if err != nil {
		return Speaker{}, fmt.Errorf("create device: %w", err)
	}

	speaker, err := newFromRootDevice(ctx, location, root)
	if err != nil {
		return Speaker{}, fmt.Errorf("create speaker: %w", err)
	}

	conn, err := net.Dial("udp", location.Host)
	if err != nil {
		return Speaker{}, fmt.Errorf("guess local addr: %w", err)
	}
	defer conn.Close()

	speaker.localAddr = conn.LocalAddr().(*net.UDPAddr).IP

	return speaker, nil
}

func newFromRootDevice(ctx context.Context, location *url.URL, root *goupnp.RootDevice) (Speaker, error) {
	id, _, _ := strings.Cut(strings.TrimPrefix(root.Device.UDN, "uuid:"), ":")
	id = slug.Make(id)

	av1Clients, err := av1.NewAVTransport1ClientsFromRootDevice(root, location)
	if err != nil {
		return Speaker{}, fmt.Errorf("create AV1 client: %w", err)
	}
	if len(av1Clients) != 1 {
		return Speaker{}, fmt.Errorf("expected exactly one av1 client, but got %d", len(av1Clients))
	}

	rc1Clients, err := av1.NewRenderingControl1ClientsFromRootDevice(root, location)
	if err != nil {
		return Speaker{}, fmt.Errorf("create RC1 client: %w", err)
	}
	if len(rc1Clients) != 1 {
		return Speaker{}, fmt.Errorf("expected exactly one rc1 client, but got %d", len(rc1Clients))
	}

	eventSubURLs := map[string]string{}
	for _, s := range root.Device.Services {
		eventSubURLs[s.ServiceId] = s.EventSubURL.Str
	}

	return Speaker{
		id:           id,
		location:     location,
		friendlyName: strings.TrimPrefix(root.Device.FriendlyName, "Speaker "),
		localAddr:    nil,
		av1:          av1Clients[0],
		rc1:          rc1Clients[0],
		eventSubURLs: eventSubURLs,
	}, nil
}

func (s Speaker) ID() string {
	return s.id
}

func (s Speaker) FriendlyName() string {
	return s.friendlyName
}

func (s Speaker) Location() *url.URL {
	return s.location
}

func (s Speaker) MDNSLocation() (*url.URL, error) {
	u := *s.location

	addr, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("split host port of %#v: %w", u.Host, err)
	}

	hosts, err := net.LookupAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("lookup address %#v: %w", addr, err)
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no mdns found for %#v", hosts)
	}

	u.Host = net.JoinHostPort(hosts[0], port)

	return &u, nil
}

func (s Speaker) TryMDNSLocation() *url.URL {
	u, err := s.MDNSLocation()
	if err != nil {
		logrus.Warn(err)
		return s.location
	}

	return u
}

func (s Speaker) LocalAddr() net.IP {
	return s.localAddr
}

func (s Speaker) SetVolumePercent(ctx context.Context, value uint16) error {
	return s.rc1.SetVolumeCtx(ctx, InstanceID, ChannelMaster, value)
}

func (s Speaker) SetVolumeFloat(ctx context.Context, value float64) error {
	return s.SetVolumePercent(ctx, uint16(value*100.))
}

func (s Speaker) SetMute(ctx context.Context, value bool) error {
	return s.rc1.SetMuteCtx(ctx, InstanceID, ChannelMaster, value)
}

func (s Speaker) SetOnOff(ctx context.Context, on bool) error {
	var request struct {
		InstanceID string
	}

	request.InstanceID = "0"

	var response any

	action := "EnterManualStandby"
	if on {
		action = "LeaveStandby"
	}

	return s.av1.SOAPClient.PerformActionCtx(ctx,
		"urn:upnp-org:serviceId:AVTransport", action,
		&request, &response,
	)
}
