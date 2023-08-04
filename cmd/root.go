package cmd

import (
	"context"

	"github.com/rebuy-de/rebuy-go-sdk/v5/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const DeviceType = `urn:schemas-upnp-org:device:MediaRenderer:1`

type RunnerFunc func(ctx context.Context) error

func (r RunnerFunc) Bind(cmd *cobra.Command) error {
	return nil
}

func (r RunnerFunc) Run(ctx context.Context) error {
	return r(ctx)
}

func NewRootCommand() *cobra.Command {
	return cmdutil.New(
		"devilctl", "controlling Teufel Raumfeld boxes",
		cmdutil.WithLogVerboseFlag(),
		cmdutil.WithVersionCommand(),

		cmdutil.WithSubCommand(cmdutil.New(
			"discover", "discover hosts in network",
			cmdutil.WithRunner(RunnerFunc(DiscoverRunner)),
		)),

		cmdutil.WithSubCommand(cmdutil.New(
			"homie-bridge", "Bridge Raumfeld speakers to MQTT via Homie convention",
			cmdutil.WithRunner(new(HomieBridgeRunner)),
		)),
	)
}
