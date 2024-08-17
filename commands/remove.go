package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/message"
	"github.com/depado/fox/player"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type remove struct {
	BaseCommand
}

func (c *remove) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if len(args) > 0 && (args[0] == "all" || args[0] == "a" || args[0] == "-a") {
		p.Queue.Clear()
		msg := fmt.Sprintf("🚮 The queue was reset by <@%s>", m.Author.ID)
		if err := message.SendReply(s, m, "", msg, ""); err != nil {
			c.log.Err(err).Msg("unable to send reply")
		}
		return
	}

	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		if err := message.SendTimedReply(s, m, "", "The argument is invalid", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
		}
		return
	}

	p.Queue.RemoveN(n)
	msg := fmt.Sprintf("🚮 The next %d tracks in queue were removed by <@%s>", n, m.Author.ID)
	if err := message.SendReply(s, m, "", msg, ""); err != nil {
		c.log.Err(err).Msg("unable to send reply")
	}
}

func NewRemoveCommand(p *player.Players, log zerolog.Logger) Command {
	cmd := "remove"
	return &remove{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Privileged,
			Options: Options{
				ArgsRequired:      true,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"rm"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Clear the queue",
				Description: "This command can be used to remove all the " +
					"tracks or a certain number of tracks in queue.",
				Examples: []Example{
					{Command: "remove all", Explanation: "Remove all tracks in queue"},
					{Command: "rm -a", Explanation: "Remove all tracks in queue"},
					{Command: "rm 10", Explanation: "Remove the next 10 tracks in queue"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}
