package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	// Register postgres migration tool.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Register file migration tool.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewMigrateCmd() *cobra.Command {
	postgresURL := viper.GetString("postgres_url")
	migrationsPath := viper.GetString("migrations_path")
	return &cobra.Command{
		Use:   "migrate",
		Short: "User is a microservices application",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMigrate(postgresURL, migrationsPath)
		},
	}
}

func runMigrate(postgresURL, migrationsPath string) (err error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		postgresURL,
	)
	if err != nil {
		return err
	}
	defer func() {
		err1, err2 := m.Close()
		if err == nil {
			if err2 != nil {
				err = err2
			} else if err1 != nil {
				err = err1
			}
		}
	}()
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
