package admin

import (
	"errors"
	"policy/internal/app"

	"github.com/spf13/cobra"
)

func NewAdminCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "admin",
		Short: "User is a microservices application",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("must provide at least one argument - name of the admin command")
			}
			return app.Execute(cmd.Context(), args[0], args[1:]...)
		},
	}
}
