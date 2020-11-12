package message

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
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
