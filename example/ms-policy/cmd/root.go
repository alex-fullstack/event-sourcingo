package cmd

import (
	"policy/cmd/admin"
	"policy/cmd/run"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() int {
	var rootCmd = &cobra.Command{
		Use:   "policy",
		Short: "Policy is a microservices application",
		Long:  `...`,
	}
	viper.AutomaticEnv()
	rootCmd.AddCommand(run.NewRunCmd())
	rootCmd.AddCommand(admin.NewAdminCmd())
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
