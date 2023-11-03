package main

import (
	"github.com/aioncore/p2p/pkg/cmd"
	"github.com/aioncore/p2p/pkg/config"
	"github.com/aioncore/p2p/pkg/log"
	"github.com/aioncore/p2p/pkg/service/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootPath := utils.GetRootPath()
	log.InitLogger(rootPath, "p2p")
	config.InitConfig(rootPath, "p2p")
	rootCmd := &cobra.Command{
		Use:   "p2p",
		Short: "p2p",
	}
	rootCmd.AddCommand(
		cmd.NewStartCmd(),
	)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
