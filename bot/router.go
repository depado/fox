package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// ack tells whether or not the bot should react to a message
// Basically it checks whether the command was issued in the public or control
// channel, if the message contains the defined prefix, and if it didn't react
// to its own message
func (b *BotInstance) ack(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return strings.HasPrefix(m.Content, b.conf.Bot.Prefix) &&
		(m.ChannelID == b.conf.Bot.Channels.Public || m.ChannelID == b.conf.Bot.Channels.Control) &&
		m.Author.ID != s.State.User.ID
}

// restricted will return true if the message was posted in the control channel
func (b *BotInstance) restricted(m *discordgo.MessageCreate) bool {
	return m.ChannelID == b.conf.Bot.Channels.Control
}

// MessageCreated is the main handler and will act as a router for all the
// commands
func (b *BotInstance) MessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !b.ack(s, m) {
		return
	}

	fields := strings.Fields(m.Content)
	if len(fields) < 2 {
		b.HelpHandler(m)
	}
	args := fields[2:]

	switch fields[1] {
	case "next", "n":
		b.NextHandler(m, args)
	case "help", "h":
		b.HelpHandler(m)
	case "join", "j":
		b.JoinHandler(m)
	case "leave", "l":
		b.LeaveHandler(m)
	case "queue", "q":
		b.QueueHandler(m, args)
	case "add", "a":
		b.AddHandler(m, args)
	case "pause":
		b.PauseHandler(m)
	case "play":
		b.PlayHandler(m)
	case "skip":
		b.SkipHandler(m)
	case "resume":
		b.ResumeHandler(m)
	case "stop":
		b.StopHandler(m)
	case "vote":
		b.VoteHandler(m)
	}
}
