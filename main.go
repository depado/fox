package main

import (
	"github.com/Depado/soundcloud"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/bot"
	"github.com/depado/fox/cmd"
	"github.com/depado/fox/commands"
	"github.com/depado/fox/player"
	sp "github.com/depado/fox/soundcloud"
	"github.com/depado/fox/storage"
)

// Main command that will be run when no other command is provided on the
// command-line
var rootCmd = &cobra.Command{
	Use: "fox",
	Run: func(com *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			fx.Provide(
				cmd.NewConf, cmd.NewLogger, acl.NewACL, player.NewPlayers, storage.NewBoltStorage,
				soundcloud.NewAutoIDClient, sp.NewSoundCloudProvider,
				commands.InitializeAllCommands,
				bot.NewBot,
			),
			fx.Invoke(bot.Run),
		).Run()
	},
}

func main() {
	cmd.AddAllFlags(rootCmd)
	rootCmd.AddCommand(cmd.VersionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("unable to execute root command")
	}
}
