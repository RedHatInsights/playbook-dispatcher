package cmd

import (
	"context"
	"fmt"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/db"
	"playbook-dispatcher/internal/common/utils"

	goMigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

const (
	migrationActionUp      = "up"
	migrationActionDown    = "down"
	migrationActionDownAll = "down-all"
)

func migrate(cmd *cobra.Command, args []string) error {
	log := utils.GetLoggerOrDie()
	cfg := config.Get()
	ctx := utils.SetLog(context.Background(), log)

	_, sql := db.Connect(ctx, cfg)
	driver, err := postgres.WithInstance(sql, &postgres.Config{})
	utils.DieOnError(err)

	m, err := goMigrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", cfg.GetString("migrations.dir")),
		"postgresql",
		driver)
	utils.DieOnError(err)

	log.Info("Running migrations")

	var fn func() error

	switch action := cmd.CalledAs(); action {
	case migrationActionUp:
		fn = m.Up
	case migrationActionDown:
		fn = func() error {
			return m.Steps(-1)
		}
	case migrationActionDownAll:
		fn = m.Down
	default:
		return fmt.Errorf("unknown command: %s", action)
	}

	if err := fn(); err != nil {
		if err != goMigrate.ErrNoChange {
			log.Error(err)
			return err
		} else {
			log.Info("No change")
		}
	} else {
		log.Info("Migrations applied")
	}

	return nil
}
