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

type stats struct {
	BaseCommand
}

func (c *stats) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	st := c.Player.Stats()
	if st == nil {
		if err := message.SendTimedReply(s, m, "", "There is no encoding session", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed response")
		}
		return
	}

	e := &discordgo.MessageEmbed{
		Title: "📈 Stream & encoding stats",
		Color: 0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Playback", Value: st.PlaybackPosition.String(), Inline: true},
			{Name: "Encoded", Value: st.Duration.String(), Inline: true},
			{Name: "Size", Value: fmt.Sprintf("%5d kB", st.Size), Inline: true},
			{Name: "Bitrate", Value: fmt.Sprintf("%6.2f kB/s", st.Bitrate), Inline: true},
			{Name: "Speed", Value: fmt.Sprintf("%5.1fx", st.Speed), Inline: true},
		},
	}

	mess, err := s.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		c.log.Err(err).Msg("unable to send embed")
	}
	go func() {
		time.Sleep(10 * time.Second)
		if err := s.ChannelMessageDelete(mess.ChannelID, mess.ID); err != nil {
			c.log.Err(err).Msg("unable to delete stats message")
		}
	}()
}

func NewStatsCommand(p *player.Player, log *zerolog.Logger) Command {
	cmd := "stats"
	return &stats{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Privileged,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"s"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "View encoding and streaming instant stats",
				Description: "This command will display the encoding and " +
					"streaming stats if a stream is ongoing.",
			},
			Player: p,
			log:    log.With().Str("command", cmd).Logger(),
		},
	}
}
