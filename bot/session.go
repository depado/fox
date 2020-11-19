package bot

import (
	"github.com/Depado/fox/cmd"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func NewDiscordSession(conf *cmd.Conf, l *zerolog.Logger) *discordgo.Session {
	dg, err := discordgo.New("Bot " + conf.Bot.Token)
	if err != nil {
		l.Fatal().Err(err).Msg("unable to open connection")
	}

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
	if err := dg.Open(); err != nil {
		l.Fatal().Err(err).Msg("unable to open")
	}

	return dg
}
