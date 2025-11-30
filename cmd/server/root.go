package main

import (
	"github.com/gemyago/golang-backend-boilerplate/internal"
	"github.com/gemyago/golang-backend-boilerplate/internal/config"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

func newRootCmd(container *dig.Container) *cobra.Command {
	logsOutputFile := ""

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Command to start the server",
	}
	cmd.SilenceUsage = true
	cmd.PersistentFlags().StringP("log-level", "l", "", "Produce logs with given level. Default is env specific.")
	cmd.PersistentFlags().StringVar(
		&logsOutputFile,
		"logs-file",
		"",
		"Produce logs to file instead of stdout. Used for tests only.",
	)
	cmd.PersistentFlags().Bool(
		"json-logs",
		false,
		"Indicates if logs should be in JSON format or text (default)",
	)
	cmd.PersistentFlags().StringP(
		"env",
		"e",
		"",
		"Env that the process is running in.",
	)
	cfg := config.New()
	lo.Must0(cfg.BindPFlags(cmd.PersistentFlags()))
	cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		rootCtx := cmd.Context()

		return internal.Setup(
			rootCtx,
			cfg,
			container,
		)
	}
	return cmd
}
