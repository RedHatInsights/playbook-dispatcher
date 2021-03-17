package cmd

import (
	"os"

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
	moduleValidator        = "validator"
)

func init() {
	runCommand := &cobra.Command{
		Use:   "run",
		Short: "Run playbook-dispatcher",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				os.Exit(1)
			}
		},
	}

	runCommand.Flags().StringSliceP("module", "m", []string{moduleApi, moduleResponseConsumer, moduleValidator}, "module(s) to run")
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

	rootCmd.AddCommand(&cobra.Command{
		Use:   "clean",
		Short: "Run database cleanup actions",
		RunE:  clean,
	})
}

func Execute() error {
	return rootCmd.Execute()
}
