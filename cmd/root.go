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
}

func Execute() error {
	return rootCmd.Execute()
}
