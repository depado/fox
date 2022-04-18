package models

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	ChannelNotFoundError = fmt.Errorf("no channel found")
)

// Conf represents the guild conf at a given point.
type Conf struct {
	ID             string `json:"-"`
	VoiceChannel   string `json:"voice"`
	TextChannel    string `json:"text"`
	QueueHistory   int    `json:"history"`
	PrivilegedRole string `json:"privileged_role"`
}

type Info struct {
	Name     string    `json:"name"`
	Members  int       `json:"members"`
	JoinedAt time.Time `json:"joined_at"`
}

// New creates a new conf for a given guild ID
func NewConf(id string) *Conf {
	return &Conf{
		ID: id,
	}
}

// SetVoiceChannel will cycle through the available channels to check if the
// vocal channel actually exists and set its ID in the conf
func (c *Conf) SetChannel(s *discordgo.Session, value string, voice bool) error {
	var dtype discordgo.ChannelType

	if c.ID == "" {
		return fmt.Errorf("guild conf with empty ID")
	}

	chans, err := s.GuildChannels(c.ID)
	if err != nil {
		return err
	}
	if voice {
		dtype = discordgo.ChannelTypeGuildVoice
	} else {
		dtype = discordgo.ChannelTypeGuildText
	}

	var found bool
	for _, ch := range chans {
		if (ch.Name == value || c.ID == value) && ch.Type == dtype {
			if voice {
				c.VoiceChannel = ch.ID
			} else {
				c.TextChannel = ch.ID
			}
			found = true
		}
	}

	if !found {
		return ChannelNotFoundError
	}

	return nil
}
