package run

import (
	"api/internal/app"

	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "API is a microservices application",
		Long:  `...`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.Run(cmd.Context())
		},
	}
}
