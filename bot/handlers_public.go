package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// AddHandler is in charge of pushing a track or playlist to the end of the
// current queue
func (b *BotInstance) AddHandler(m *discordgo.MessageCreate, args []string) {
	defer b.Delete(m.Message)
	if len(args) < 1 {
		b.SendNotice("", fmt.Sprintf("Usage: `%s <add|a> <soundcloud URL>`", b.conf.Bot.Prefix), "", m.ChannelID)
		return
	}

	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		b.SendTimedNotice("", "This doesn't look like a SoundCloud URL", "", m.ChannelID, 5*time.Second)
		return
	}

	b.AddToQueue(m, url, false)
}

// NextHandler is in charge of pushing a track or playlist in front of the rest
// of the queue
func (b *BotInstance) NextHandler(m *discordgo.MessageCreate, args []string) {
	defer b.Delete(m.Message)
	if len(args) < 1 {
		b.SendNotice("", fmt.Sprintf("Usage: `%s <next|n> <soundcloud URL>`", b.conf.Bot.Prefix), "", m.ChannelID)
		return
	}

	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		b.SendTimedNotice("", "This doesn't look like a SoundCloud URL", "", m.ChannelID, 5*time.Second)
		return
	}

	b.AddToQueue(m, url, true)
}

// HelpHandler will handle incoming requests for help
func (b *BotInstance) HelpHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	var doc = &discordgo.MessageEmbed{
		Title: "Fox Help",
		Color: 0xff5500,
		Description: fmt.Sprintf("Prefix your commands with `%s`\n\n", b.conf.Bot.Prefix) +
			"**Available commands:**\n" +
			"· `help` - Display this help message\n" +
			"· `add` or `a` - Add tracks to the end of queue\n" +
			"· `next` or `n` - Add tracks to the start of queue\n" +
			"· `queue` or `q` - Display the queue\n" +
			"· `queue shuffle` - Shuffle the queue\n" +
			"· `play/pause/resume` - Control the player\n" +
			"· `vote` - Vote to skip the currently playing track\n\n" +
			"For example, to add a new track to the playlist, you need to send\n" +
			fmt.Sprintf("`%s add <soundcloud URL>`", b.conf.Bot.Prefix),
	}

	if b.restricted(m) {
		doc.Fields = []*discordgo.MessageEmbedField{
			{Name: "Channel", Value: fmt.Sprintf("%s <join/leave>", b.conf.Bot.Prefix)},
			{Name: "Clear Queue", Value: fmt.Sprintf("%s <queue|q> clear", b.conf.Bot.Prefix), Inline: true},
			{Name: "Control", Value: fmt.Sprintf("%s <stop>", b.conf.Bot.Prefix)},
			{Name: "Skip", Value: fmt.Sprintf("%s <skip>", b.conf.Bot.Prefix)},
			{Name: "Stats", Value: fmt.Sprintf("%s <stats>", b.conf.Bot.Prefix)},
		}
	}

	_, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, doc)
	if err != nil {
		log.Err(err).Msg("unable to send embed")
	}
}

// QueueHandler is in charge of dealing with queue commands such as displaying
// the current queue, shuffling the queue or in the control channel, clearing it
func (b *BotInstance) QueueHandler(m *discordgo.MessageCreate, args []string) {
	defer b.Delete(m.Message)

	if len(args) == 0 {
		b.DisplayQueue(m)
		return
	}

	switch args[0] {
	case "shuffle":
		b.Player.Shuffle()
		b.SendNotice("", fmt.Sprintf("🎲 Shuffled **%d** tracks for <@%s>", b.Player.QueueSize(), m.Author.ID), "", m.ChannelID)
	case "clear": // The clear command is not public and shouldn't be used
		if b.restricted(m) {
			b.Player.Clear()
			b.SendNotice("", fmt.Sprintf("🚮 The queue was reset by <@%s>", m.Author.ID), "", m.ChannelID)
		} else {
			b.SendTimedNotice("", "You do not have this permission", "Tip: Only admins and DJs can do that", m.ChannelID, 10*time.Second)
		}
	default:
		b.SendTimedNotice("", "Unknown command", fmt.Sprintf(`Tip: Use "%s help" for a list of commands`, b.conf.Bot.Prefix), m.ChannelID, 10*time.Second)
	}
}
