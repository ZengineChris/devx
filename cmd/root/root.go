package root

import (
	"fmt"
	"log"

	"github.com/zenginechris/devx/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "devx",
	Short:   "opiniated local develepment",
	Version: config.AppVersion().Version,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Warn("here")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println(cmd.Name())

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
