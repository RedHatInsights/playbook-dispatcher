package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "pd",
		Short: "Playbook Dispatcher",
	}
)

const (
	moduleApi              = "api"
	moduleResponseConsumer = "response-consumer"
)

func init() {
	runCommand := &cobra.Command{
		Use:   "run",
		Short: "Run playbook-dispatcher",
		RunE:  run,
	}

	runCommand.Flags().StringSliceP("module", "m", []string{moduleApi, moduleResponseConsumer}, "module(s) to run")
	rootCmd.AddCommand(runCommand)

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
	}

	rootCmd.AddCommand(migrateCmd)

	migrateCmd.AddCommand(&cobra.Command{
		Use:   migrationActionUp,
		Short: "Run database migrations",
		RunE:  migrate,
	})

	migrateCmd.AddCommand(&cobra.Command{
		Use:   migrationActionDown,
		Short: "Undo last database migration",
		RunE:  migrate,
	})

	migrateCmd.AddCommand(&cobra.Command{
		Use:   migrationActionDownAll,
		Short: "Undo all database migration",
		RunE:  migrate,
	})
}

func Execute() error {
	return rootCmd.Execute()
}
