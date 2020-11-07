package bot

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
)

const (
	voteThreshold = 2
)

type VoteHolder struct {
	sync.RWMutex
	Voters map[string]bool
}

func (v *VoteHolder) Reset() {
	v.Lock()
	defer v.Unlock()
	v.Voters = make(map[string]bool)
}

func (b *BotInstance) VoteHandler(m *discordgo.MessageCreate) {
	b.Vote.Lock()
	defer b.Vote.Unlock()

	if _, ok := b.Vote.Voters[m.Author.ID]; ok {
		return
	}
	b.Vote.Voters[m.Author.ID] = true
	e := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("<@%s> voted to skip this track", m.Author.ID),
		Color:       0xff5500,
	}

	if len(b.Vote.Voters) >= 2 {
		if b.Player.playing {
			b.Player.session.Stop() // nolint:errcheck
			e.Description += fmt.Sprintf("\n%d total votes, skipping current track", len(b.Vote.Voters))
			e.Footer = &discordgo.MessageEmbedFooter{
				Text: "Note: This may take a few seconds",
			}
		} else {
			b.Player.Pop()
			e.Description += fmt.Sprintf("\n%d total votes, the next track in queue has been skipped", len(b.Vote.Voters))
		}
		b.Vote.Voters = make(map[string]bool)
	} else {
		if b.Player.playing {
			e.Description += fmt.Sprintf("\n%d/%d votes to skip this track", len(b.Vote.Voters), voteThreshold)
		} else {
			e.Description += fmt.Sprintf("\n%d/%d votes to skip the next track", len(b.Vote.Voters), voteThreshold)
		}
	}

	if _, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}
