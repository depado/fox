package message

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func Delete(s *discordgo.Session, m *discordgo.Message, log zerolog.Logger) {
	if m.GuildID == "" { // Message is a MP
		return
	}
	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		log.Err(err).Msg("unable to delete user message")
	}
}
