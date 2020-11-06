package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func (b *BotInstance) ack(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return strings.HasPrefix(m.Content, b.conf.Bot.Prefix) &&
		(m.ChannelID == b.conf.Bot.Channels.Public || m.ChannelID == b.conf.Bot.Channels.Control) &&
		m.Author.ID != s.State.User.ID
}

func (b *BotInstance) MessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !b.ack(s, m) {
		return
	}

	fields := strings.Fields(m.Content)

	if len(fields) < 2 {
		if _, err := s.ChannelMessageSend(m.ChannelID, "TODO:usage"); err != nil {
			log.Err(err).Msg("unable to send usage message")
		}
		return
	}

	switch fields[1] {
	case "join", "j":
		b.JoinHandler(m)
	case "leave", "l":
		b.LeaveHandler(m)
	case "queue", "q":
		b.QueueHandler(m, fields[2:])
	case "add", "a":
		b.AddHandler(m, fields[2:])
	case "shuffle", "s":
		b.ShuffleHandler(m)
	case "pause":
		b.PauseHandler()
	case "play":
		b.PlayHandler()
	case "skip":
		b.SkipHandler(m)
	case "resume":
		b.ResumeHandler()
	case "stop":
		b.StopHandler()
	case "info", "i":
		b.InfoHandler(m, fields[2:])
	}
}

func (b *BotInstance) ShuffleHandler(m *discordgo.MessageCreate) {
	b.Player.Shuffle()
	b.SendNamedNotice(m, "Requested by", "🎲 Shuffle!", fmt.Sprintf("I shuffled %d tracks for you.", len(b.Player.tracks)), "")
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

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

	b.AddToQueue(m, url)
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

func (b *BotInstance) QueueHandler(m *discordgo.MessageCreate, args []string) {
	b.DisplayQueue(m)
	if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		b.log.Err(err).Msg("unable to delete user message")
	}
}

func (b *BotInstance) PlayHandler() {
	b.PlayQueue()
}

func (b *BotInstance) PauseHandler() {
	if b.Player.playing {
		b.Player.stream.SetPaused(true)
	}
}

func (b *BotInstance) ResumeHandler() {
	if b.Player.playing {
		b.Player.stream.SetPaused(false)
	}
}

func (b *BotInstance) StopHandler() {
	if b.Player.playing {
		b.Player.session.Stop() // nolint:errcheck
		b.Player.stop = true
	}
}

func (b *BotInstance) SkipHandler(m *discordgo.MessageCreate) {
	if b.Player.playing {
		b.Player.session.Stop() // nolint:errcheck
		b.SendNamedNotice(m, "Requested by", "⏭️ Skip", "The currently playing track has been skipped", "Note: This can take a few seconds")
		b.Session.ChannelMessageDelete(m.ChannelID, m.ID)
	} else {
		b.Player.Pop()
		b.SendNamedNotice(m, "Requested by", "⏭️ Skip", "The next track in queue has been skipped", "")
		b.Session.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func (b *BotInstance) InfoHandler(m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		//TODO:Print usage
	}
	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		if _, err := b.Session.ChannelMessageSend(m.ChannelID, "This doesn't look like a Soundcloud URL"); err != nil {
			log.Err(err).Msg("unable to send usage message")
		}
		return
	}
	b.handleURL(b.Session, m, url)
}

func (b *BotInstance) JoinHandler(m *discordgo.MessageCreate) {
	b.log.Debug().Str("user", m.Author.Username).Str("method", "join").Msg("called")
	if b.Voice == nil {
		voice, err := b.Session.ChannelVoiceJoin(b.conf.Bot.Guild, b.conf.Bot.Channels.Voice, false, true)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to initiate voice connection")
		}
		b.Voice = voice
		b.log.Debug().Str("user", m.Author.Username).Str("method", "join").Msg("bot joined vocal channel")
	}
}

func (b *BotInstance) LeaveHandler(m *discordgo.MessageCreate) {
	b.log.Debug().Str("user", m.Author.Username).Str("method", "leave").Msg("called")
	if b.Voice != nil {
		if err := b.Voice.Disconnect(); err != nil {
			b.log.Error().Err(err).Msg("unable to disconnect from voice channel")
			return
		}
		b.Voice = nil
		b.log.Debug().Str("user", m.Author.Username).Str("method", "leave").Msg("bot left vocal channel")
		return
	}
}
