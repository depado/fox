package player

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func (p *Player) GeneratePlayerString(dur time.Duration) string {
	player := []rune("------------------------------")
	pb := p.stream.PlaybackPosition()
	pos := int(pb*100/dur) * len(player) / 100
	if pos >= len(player) {
		pos = len(player) - 1
	}
	player[pos] = '●'

	return fmt.Sprintf("%s  %s  %s", fmtDuration(pb), string(player), fmtDuration(dur))
}

func (p *Player) GenerateNowPlayingEmbed(short bool) *discordgo.MessageEmbed {
	if !p.State.Playing {
		return nil
	}

	t := p.Queue.Get()
	if t == nil {
		return nil
	}

	tot := time.Duration(t.Duration()) * time.Millisecond
	u, a := t.GetUser()
	e := t.Embed(false)
	e.Footer = &discordgo.MessageEmbedFooter{
		IconURL: a,
		Text:    "Added by " + u,
	}
	if short {
		e.Fields = nil
		e.Description = p.GeneratePlayerString(tot)
	} else {
		e.Description += p.GeneratePlayerString(tot)
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:   "Queue",
			Value:  fmt.Sprintf("%d tracks left in queue - %s", p.Queue.Len(), p.Queue.DurationString()),
			Inline: false,
		})
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

	m, err := p.session.ChannelMessageSendEmbed(p.GuildState.TextChannel, e)
	if err != nil {
		p.log.Err(err).Msg("unable to send embed")
	}
	return m
}
