package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/Depado/fox/bot"
	"github.com/Depado/fox/cmd"
	"github.com/Depado/soundcloud"
)

// Build number and versions injected at compile time, set yours
var (
	Version = "unknown"
	Build   = "unknown"
)

// Main command that will be run when no other command is provided on the
// command-line
var rootCmd = &cobra.Command{
	Use: "fox",
	Run: func(cmd *cobra.Command, args []string) { run() },
}

// Version command that will display the build number and version (if any)
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show build and version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Build: %s\nVersion: %s\n", Build, Version)
	},
}

func run() {
	fx.New(
		fx.Provide(cmd.NewConf, cmd.NewLogger, soundcloud.NewAutoIDClient, bot.NewBotInstance),
		fx.Invoke(bot.Start),
	).Run()
}

func main() {
	cmd.AddAllFlags(rootCmd)
	rootCmd.AddCommand(versionCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("unable to execute root command")
	}
}
