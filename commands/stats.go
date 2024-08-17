package commands

import (
	"bytes"
	"fmt"
	"time"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/message"
	"github.com/depado/fox/player"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/wcharczuk/go-chart/v2"
)

type stats struct {
	BaseCommand
}

func (c *stats) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if p.Stats == nil {
		message.SendShortTimedNotice(s, m, "There is no encoding session", c.log)
		return
	}

	p.Stats.RLock()
	e := &discordgo.MessageEmbed{
		Title: "📈 Stream & encoding stats",
		Color: 0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Playback", Value: p.Stats.PlaybackPosition.String(), Inline: true},
			{Name: "Encoded", Value: p.Stats.Duration.String(), Inline: true},
			{Name: "Size", Value: fmt.Sprintf("%5d kB", p.Stats.Size), Inline: true},
			{Name: "Bitrate", Value: fmt.Sprintf("%6.2f kB/s", p.Stats.Bitrate), Inline: true},
			{Name: "Speed", Value: fmt.Sprintf("%5.1fx", p.Stats.Speed), Inline: true},
		},
	}
	p.Stats.RUnlock()

	msg := &discordgo.MessageSend{
		Embed: e,
	}

	if len(args) > 0 {
		switch args[0] {
		case "chart", "c", "graph", "g":
			g := p.Stats.GenerateChart()
			if g != nil {
				buffer := bytes.NewBuffer([]byte{})
				if err := g.Render(chart.PNG, buffer); err != nil {
					c.log.Err(err).Msg("unable to render")
				} else {
					msg.File = &discordgo.File{
						Name:        "graph.png",
						ContentType: "image/png",
						Reader:      buffer,
					}
				}
			}
		}
	}

	mess, err := s.ChannelMessageSendComplex(m.ChannelID, msg)
	if err != nil {
		c.log.Err(err).Msg("unable to send embed")
	}
	go func() {
		time.Sleep(30 * time.Second)
		if err := s.ChannelMessageDelete(mess.ChannelID, mess.ID); err != nil {
			c.log.Err(err).Msg("unable to delete stats message")
		}
	}()
}

func NewStatsCommand(p *player.Players, log zerolog.Logger) Command {
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
			SubCommands: []SubCommand{
				{
					Long:        "graph",
					Aliases:     []string{"chart", "c", "g"},
					Description: "Display an additional graph for kB/s",
				},
			},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "View encoding and streaming instant stats",
				Description: "This command will display the encoding and " +
					"streaming stats if a stream is ongoing. Using the graph " +
					"subcommand will also display a graph of the bitrate over " +
					"the whole encoding session.",
				Examples: []Example{
					{Command: "stats", Explanation: "Display instant stats"},
					{Command: "stats graph", Explanation: "Display instant stats and graph"},
					{Command: "stats g", Explanation: "Same in short notation"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}
