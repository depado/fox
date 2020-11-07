package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// AddHandler is in charge of pushing a track or playlist to the end of the
// current queue
func (b *BotInstance) AddHandler(m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		b.SendNotice("", fmt.Sprintf("Usage: `%s <add|a> <soundcloud URL>`", b.conf.Bot.Prefix), "", m.ChannelID)
		return
	}

	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		b.SendNotice("", "This doesn't look like a SoundCloud URL", "", m.ChannelID)
		return
	}

	b.AddToQueue(m, url, false)
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

// NextHandler is in charge of pushing a track or playlist in front of the rest
// of the queue
func (b *BotInstance) NextHandler(m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		b.SendNotice("", fmt.Sprintf("Usage: `%s <next|n> <soundcloud URL>`", b.conf.Bot.Prefix), "", m.ChannelID)
		return
	}

	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		b.SendNotice("", "This doesn't look like a SoundCloud URL", "", m.ChannelID)
		return
	}

	b.AddToQueue(m, url, true)
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

// HelpHandler will handle incoming requests for help
func (b *BotInstance) HelpHandler(m *discordgo.MessageCreate) {
	var doc = &discordgo.MessageEmbed{
		Title: "Fox Help",
		Color: 0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Help", Value: fmt.Sprintf("%s <help|h>", b.conf.Bot.Prefix), Inline: true},
			{Name: "Add tracks", Value: fmt.Sprintf("%s <add|a> <soundcloud url>", b.conf.Bot.Prefix), Inline: true},
			{Name: "Add next tracks", Value: fmt.Sprintf("%s <next|n> <soundcloud url>", b.conf.Bot.Prefix)},
			{Name: "Channel", Value: fmt.Sprintf("%s <join/leave>", b.conf.Bot.Prefix)},
			{Name: "Display Queue", Value: fmt.Sprintf("%s <queue|q>", b.conf.Bot.Prefix), Inline: true},
			{Name: "Shuffle Queue", Value: fmt.Sprintf("%s <queue|q> shuffle", b.conf.Bot.Prefix), Inline: true},
			{Name: "Clear Queue", Value: fmt.Sprintf("%s <queue|q> clear", b.conf.Bot.Prefix), Inline: true},
			{Name: "Control", Value: fmt.Sprintf("%s <play/pause/resume/stop>", b.conf.Bot.Prefix)},
		},
	}
	_, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, doc)
	if err != nil {
		log.Err(err).Msg("unable to send embed")
	}
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

// QueueHandler is in charge of dealing with queue commands such as displaying
// the current queue, shuffling the queue or in the control channel, clearing it
func (b *BotInstance) QueueHandler(m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		b.DisplayQueue(m)
		if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
		return
	}
	switch args[0] {
	case "shuffle":
		b.Player.Shuffle()
		b.SendNamedNotice(m, "Requested by", "🎲 Shuffle!", fmt.Sprintf("I shuffled %d tracks for you.", len(b.Player.tracks)), "")
		if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
	case "clear": // The clear command is not public and shouldn't be used
		if b.restricted(m) {
			b.Player.Clear()
			b.SendNamedNotice(m, "Requested by", "🚮 Clear", "The queue has been reset", "")
			if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				b.log.Err(err).Msg("unable to delete user message")
			}
		}
	}
}
