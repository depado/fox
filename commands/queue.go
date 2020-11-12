package commands

import (
	"fmt"
	"time"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/message"
	"github.com/Depado/fox/player"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type queue struct {
	BaseCommand
}

func (c *queue) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	if len(args) > 0 && (args[0] == "shuffle" || args[0] == "s") {
		if c.Player.Queue.Len() < 2 {
			err := message.SendTimedReply(s, m, "", "There is not enough tracks to shuffle", "", 5*time.Second)
			if err != nil {
				c.log.Err(err).Msg("unable to send embed")
			}
			return
		}
		c.Player.Queue.Shuffle()
		err := message.SendReply(s, m, "", fmt.Sprintf("🎲 Shuffled **%d** tracks for <@%s>", c.Player.Queue.Len(), m.Author.ID), "")
		if err != nil {
			c.log.Err(err).Msg("unable to send embed")
		}
		return
	}

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, c.Player.Queue.GenerateQueueEmbed()); err != nil {
		c.log.Err(err).Msg("unable to send embed")
	}
}

func NewQueueCommand(p *player.Player, log *zerolog.Logger) Command {
	cmd := "queue"
	return &queue{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"q"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Display or modify the queue",
				Description: "This command will display the current queue. " +
					"It can also shuffle the current queue if the `shuffle` " +
					"argument is passed.",
				Examples: []Example{
					{Command: "queue", Explanation: "Display the queue"},
					{Command: "queue shuffle", Explanation: "Shuffle the queue"},
					{Command: "q", Explanation: "Display the queue with the alias"},
				},
			},
			Player: p,
			log:    log.With().Str("command", cmd).Logger(),
		},
	}
}
