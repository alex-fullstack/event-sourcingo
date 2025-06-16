package cmd

import (
	"api/cmd/run"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() int {
	var rootCmd = &cobra.Command{
		Use:   "api",
		Short: "API is a microservices application",
		Long:  `...`,
	}
	viper.AutomaticEnv()
	rootCmd.AddCommand(run.NewRunCmd())
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
