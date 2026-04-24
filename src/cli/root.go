package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	
	"gitmirror/internal/logger"
)

var debugMode bool

var rootCmd = &cobra.Command{
	Use:     "gitmirror",
	Short:   "Mirror repositories between providers",
	Long:    `Mirror repositories between providers like GitHub and GitLab. This tool helps you keep your repositories in sync across different platforms.`,
	Version: fmt.Sprintf("%s (%s/%s) built at %s", version, buildOS, buildArch, buildTime),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Init(debugMode, "")
		logger.Debug("Debug mode enabled")
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode")

	rootCmd.AddCommand(syncCmd)
}
