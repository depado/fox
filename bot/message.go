package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Delete will delete the provided message immediately
func (b *BotInstance) Delete(m *discordgo.Message) {
	if m == nil {
		return
	}
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete message")
	}
}

// DeleteAfter will delete the provided message after the given duration
func (b *BotInstance) DeleteAfter(m *discordgo.Message, t time.Duration) {
	if m == nil {
		return
	}
	go func() {
		time.Sleep(t)
		if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete message")
		}
	}()
}

func (b *BotInstance) SendPublicMessage(title, body, footer string) {
	b.SendNotice(title, body, footer, b.conf.Bot.Channels.Public)
}

func (b *BotInstance) SendTimedNotice(title, body, footer, channel string, t time.Duration) {
	m := b.SendNotice(title, body, footer, channel)
	b.DeleteAfter(m, t)
}

func (b *BotInstance) SendNotice(title, body, footer, channel string) *discordgo.Message {
	e := &discordgo.MessageEmbed{
		Title:       title,
		Description: body,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
		Color: 0xff5500,
	}

	m, err := b.Session.ChannelMessageSendEmbed(channel, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
	return m
}
