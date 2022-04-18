package main

import (
	"github.com/Depado/soundcloud"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/bot"
	"github.com/Depado/fox/cmd"
	"github.com/Depado/fox/commands"
	"github.com/Depado/fox/player"
	sp "github.com/Depado/fox/soundcloud"
	"github.com/Depado/fox/storage"
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
