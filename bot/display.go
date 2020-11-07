package bot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Depado/soundcloud"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
)

func (b *BotInstance) SendNowPlaying(t soundcloud.Track) {
	b.Player.tracksM.Lock()
	defer b.Player.tracksM.Unlock()

	e := &discordgo.MessageEmbed{
		Title: t.Title,
		URL:   t.PermalinkURL,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: t.User.AvatarURL,
			Name:    t.User.Username,
			URL:     t.User.PermalinkURL,
		},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: t.ArtworkURL},
		Description: "**Now Playing**",
		Color:       0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Plays", Value: strconv.Itoa(t.PlaybackCount), Inline: true},
			{Name: "Likes", Value: strconv.Itoa(t.LikesCount), Inline: true},
			{Name: "Reposts", Value: strconv.Itoa(t.RepostsCount), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(t.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%d tracks left in queue", len(b.Player.tracks)),
		},
	}

	_, err := b.Session.ChannelMessageSendEmbed(b.conf.Bot.Channels.Public, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
}

func (b *BotInstance) DisplayQueue(m *discordgo.MessageCreate) {
	b.Player.tracksM.Lock()
	defer b.Player.tracksM.Unlock()

	var body string
	var tot int
	if len(b.Player.tracks) > 0 {
		for i, t := range b.Player.tracks {
			if i <= 10 {
				body += fmt.Sprintf("[%s - %s](%s)\n", t.Title, t.User.Username, t.PermalinkURL)
			}
			tot += t.Duration
		}
		if len(b.Player.tracks) > 10 {
			body += fmt.Sprintf("\nAnd **%d** other tracks", len(b.Player.tracks)-10)
		}
	} else {
		body = "There is currently no track in queue"
	}

	e := &discordgo.MessageEmbed{
		Title:       "Current Queue",
		Description: body,
		Color:       0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Tracks", Value: strconv.Itoa(len(b.Player.tracks)), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(tot) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
			{Name: "Requested by", Value: fmt.Sprintf("<@%s>", m.Author.ID), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Tip: Add new tracks using '%s add' or '%s next'", b.conf.Bot.Prefix, b.conf.Bot.Prefix),
		},
	}
	_, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
}

func (b *BotInstance) SendPublicMessage(title, body string) {
	b.SendNotice(title, body, "", b.conf.Bot.Channels.Public)
}

func (b *BotInstance) SendControlMessage(title, body string) {
	b.SendNotice(title, body, "", b.conf.Bot.Channels.Control)
}

func (b *BotInstance) SendNotice(title, body, footer string, channel string) {
	e := &discordgo.MessageEmbed{
		Title:       title,
		Description: body,
		Footer:      &discordgo.MessageEmbedFooter{Text: footer},
		Color:       0xff5500,
	}
	_, err := b.Session.ChannelMessageSendEmbed(channel, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
}

func (b *BotInstance) DeleteUserMessage(m *discordgo.MessageCreate) {
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

func (b *BotInstance) SendNamedNotice(m *discordgo.MessageCreate, prefix, title, body, footer string) {
	e := &discordgo.MessageEmbed{
		Title:       title,
		Description: body,
		Fields: []*discordgo.MessageEmbedField{
			{Name: prefix, Value: fmt.Sprintf("<@%s>", m.Author.ID)},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: footer},
		Color:  0xff5500,
	}
	_, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
}

func (b *BotInstance) DisplayTemporaryMessage(m *discordgo.MessageCreate, title, body, footer string) {
	e := &discordgo.MessageEmbed{
		Title:       title,
		Description: body,
		Footer:      &discordgo.MessageEmbedFooter{Text: footer},
		Color:       0xff5500,
	}
	mess, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}

	go func(mess *discordgo.Message) {
		time.Sleep(5 * time.Second)
		if err := b.Session.ChannelMessageDelete(mess.ChannelID, mess.ID); err != nil {
			b.log.Err(err).Msg("unable to delete message")
		}
	}(mess)
}
