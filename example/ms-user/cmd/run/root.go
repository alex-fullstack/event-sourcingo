package run

import (
	"user/cmd/migrate"
	"user/internal/app"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "User is a microservices application",
		Long:  `...`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := migrate.NewMigrateCmd().Execute(); err != nil {
				return err
			}
			return app.Run(cmd.Context())
		},
	}
}
