package guild

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	ChannelNotFoundError = fmt.Errorf("no channel found")
)

// State represents the guild state at a given point.
type State struct {
	ID             string
	VoiceChannel   string `json:"voice"`
	TextChannel    string `json:"text"`
	QueueHistory   int    `json:"history"`
	PrivilegedRole string `json:"privileged_role"`
}

// New creates a new state for a given guild ID
func NewState(id string) *State {
	return &State{
		ID: id,
	}
}

// SetVoiceChannel will cycle through the available channels to check if the
// vocal channel actually exists and set its ID in the state
func (st *State) SetChannel(s *discordgo.Session, value string, voice bool) error {
	var dtype discordgo.ChannelType

	if st.ID == "" {
		return fmt.Errorf("guild state with empty ID")
	}

	chans, err := s.GuildChannels(st.ID)
	if err != nil {
		return err
	}
	if voice {
		dtype = discordgo.ChannelTypeGuildVoice
	} else {
		dtype = discordgo.ChannelTypeGuildText
	}

	var found bool
	for _, c := range chans {
		if (c.Name == value || c.ID == value) && c.Type == dtype {
			if voice {
				st.VoiceChannel = c.ID
			} else {
				st.TextChannel = c.ID
			}
			found = true
		}
	}

	if !found {
		return ChannelNotFoundError
	}

	return nil
}
