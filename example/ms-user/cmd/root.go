package cmd

import (
	"user/cmd/run"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() int {
	var rootCmd = &cobra.Command{
		Use:   "user",
		Short: "User is a microservices application",
		Long:  `...`,
	}
	viper.AutomaticEnv()
	rootCmd.AddCommand(run.NewRunCmd())
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
