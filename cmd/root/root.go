package root

import (
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zenginechris/devx/config"
)

var rootCmd = &cobra.Command{
	Use:     "devx",
	Short:   "Opiniated local development",
	Version: config.AppVersion().Version,
	Run:     func(cmd *cobra.Command, args []string) {},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initLog(); err != nil {
			return err
		}
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return nil
	},
}

func Cmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func initLog() error {
	log.SetOutput(logrus.StandardLogger().Writer())
	log.SetFlags(0)
	return nil
}
