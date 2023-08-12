package cmd

import (
	"context"
	"fmt"

	"github.com/svenwltr/devilctl/pkg/dal/raumfeld"
	"golang.org/x/sync/errgroup"
)

func DiscoverRunner(ctx context.Context) error {
	speakers, err := raumfeld.Discover(ctx)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)

	for usn, speaker := range speakers {
		fmt.Printf("---------\n")
		fmt.Printf("Name:             %v\n", speaker.FriendlyName())
		fmt.Printf("ID:               %v\n", usn)
		fmt.Printf("Location:         %v\n", speaker.TryMDNSLocation().String())
		fmt.Printf("Discovered From:  %v\n", speaker.LocalAddr())

		var (
			waitVolume     = make(chan struct{})
			waitPowerState = make(chan struct{})
			waitMuted      = make(chan struct{})
		)

		handler := raumfeld.SubscribeHandlerFuncs{
			VolumeChange: func(s raumfeld.Speaker, volume int, channel string) {
				fmt.Printf("Volume:           %v\n", volume)
				close(waitVolume)
			},
			MuteChange: func(s raumfeld.Speaker, muted bool, channel string) {
				fmt.Printf("Muted:            %v\n", muted)
				close(waitMuted)
			},
			PowerStateChange: func(s raumfeld.Speaker, state string) {
				fmt.Printf("PowerState:       %v\n", state)
				close(waitPowerState)
			},
		}

		go func() {
			fmt.Println(speaker.Subscribe(ctx, handler))
		}()

		<-waitVolume
		<-waitPowerState
		<-waitMuted
	}

	return eg.Wait()
}
