package cmd

import (
	"fmt"
	"github.com/aioncore/p2p/pkg/p2p"
	"github.com/aioncore/p2p/pkg/service/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func AddFlags(cmd *cobra.Command) {

}

func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"node", "run"},
		Short:   "Run the tendermint node",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := p2p.NewP2P()

			log.Info("Started p2p")

			if err := p.Start(); err != nil {
				return fmt.Errorf("failed to start p2p: %w", err)
			}

			// Stop upon receiving SIGTERM or CTRL-C.
			TrapSignal(func() {
				if p.IsRunning() {
					if err := p.Stop(); err != nil {
						log.Error("unable to stop the node")
					}
				}
			})

			// Run forever.
			select {}
		},
	}
	return cmd
}

func TrapSignal(cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for _ = range c {
			log.Info("signal trapped")
			if cb != nil {
				cb()
			}
			os.Exit(0)
		}
	}()
}
