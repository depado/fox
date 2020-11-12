package player

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (p *Player) GenerateNowPlayingEmbed() *discordgo.MessageEmbed {
	if !p.State.Playing {
		return nil
	}

	t := p.Queue.Get()
	if t == nil {
		return nil
	}

	e := t.Embed()
	e.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf(
			"%d tracks left in queue - %s",
			p.Queue.Len(), p.Queue.DurationString(),
		),
	}

	return e
}

func (p *Player) SendNotice(title, body, footer string) *discordgo.Message {
	e := &discordgo.MessageEmbed{
		Title:       title,
		Description: body,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
		Color: 0xff5500,
	}

	m, err := p.session.ChannelMessageSendEmbed(p.conf.Bot.Channels.Text, e)
	if err != nil {
		p.log.Err(err).Msg("unable to send embed")
	}
	return m
}
