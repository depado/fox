package message

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func base(title, body, footer string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: body,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
		Color: 0xff5500,
	}
}

func SendReply(s *discordgo.Session, m *discordgo.Message, title, body, footer string) error {
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, base(title, body, footer))
	if err != nil {
		return fmt.Errorf("unable to send embed: %w", err)
	}
	return nil
}

func SendTimedReply(s *discordgo.Session, m *discordgo.Message, title, body, footer string, t time.Duration) error {
	mess, err := s.ChannelMessageSendEmbed(m.ChannelID, base(title, body, footer))
	if err != nil {
		return fmt.Errorf("unable to send embed: %w", err)
	}
	go func() {
		time.Sleep(t)
		s.ChannelMessageDelete(mess.ChannelID, mess.ID) // nolint:errcheck
	}()
	return nil
}

func SendTimedReplyLog(s *discordgo.Session, m *discordgo.Message, title, body, footer string, t time.Duration, l zerolog.Logger) {
	if err := SendTimedReply(s, m, title, body, footer, t); err != nil {
		l.Err(err).Msg("unable to send timed reply")
	}
}

func SendShortTimedNotice(s *discordgo.Session, m *discordgo.Message, body string, l zerolog.Logger) {
	if err := SendTimedReply(s, m, "", body, "", 5*time.Second); err != nil {
		l.Err(err).Msg("unable to send timed reply")
	}
}
