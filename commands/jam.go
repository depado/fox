package commands

import (
	"fmt"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/player"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type jam struct {
	BaseCommand
}

func (c *jam) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	size := "1x"
	if len(args) > 0 {
		switch args[0] {
		case "small", "s":
			size = "1x"
		case "medium", "m":
			size = "2x"
		case "large", "l":
			size = "3x"
		}
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> is jamming!", m.Author.ID)); err != nil {
		c.log.Err(err).Msg("unable to send message")
	}
	if _, err := s.ChannelMessageSend(
		m.ChannelID,
		fmt.Sprintf("https://cdn.betterttv.net/emote/5f1b0186cf6d2144653d2970/%s.gif", size),
	); err != nil {
		c.log.Err(err).Msg("unable to send message")
	}
}

func NewJamCommand(p *player.Players, log zerolog.Logger) Command {
	cmd := "jam"
	return &jam{
		BaseCommand{
			ChannelRestriction: acl.Anywhere,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long: cmd,
			Help: Help{
				Usage:       cmd,
				ShortDesc:   "CAT. JAM.",
				Description: "CAT. JAM.",
				Examples: []Example{
					{Command: "jam", Explanation: "Normal jam"},
					{Command: "jam large", Explanation: "Large jam"},
					{Command: "jam small", Explanation: "Small jam"},
					{Command: "jam l", Explanation: "Large jam"},
					{Command: "jam s", Explanation: "Small jam"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}
