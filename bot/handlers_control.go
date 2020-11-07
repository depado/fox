package bot

import (
	"github.com/bwmarrin/discordgo"
)

func (b *BotInstance) PlayHandler(m *discordgo.MessageCreate) {
	b.PlayQueue()
	b.SendNamedNotice(m, "Requested by", "⏯️ Play", "", "")
	b.DeleteUserMessage(m)
}

func (b *BotInstance) PauseHandler(m *discordgo.MessageCreate) {
	defer b.DeleteUserMessage(m)
	if !b.Player.playing {
		b.SendNamedNotice(m, "Requested by", "⏯️ Pause", "Nothing to do", "")
		return
	}

	if !b.Player.pause {
		b.Player.stream.SetPaused(true)
		b.Player.pause = true
		b.SendNamedNotice(m, "Requested by", "⏯️ Pause", "", "")
	} else {
		b.SendNamedNotice(m, "Requested by", "⏯️ Pause", "Nothing to do", "")
	}
}

func (b *BotInstance) ResumeHandler(m *discordgo.MessageCreate) {
	defer b.DeleteUserMessage(m)
	if !b.Player.playing {
		b.SendNamedNotice(m, "Requested by", "⏯️ Resume", "Nothing to do", "")
		return
	}
	if b.Player.pause {
		b.Player.stream.SetPaused(false)
		b.Player.pause = false
		b.SendNamedNotice(m, "Requested by", "⏯️ Resume", "", "")
	} else {
		b.SendNamedNotice(m, "Requested by", "⏯️ Resume", "Nothing to do", "")
	}
}

func (b *BotInstance) StopHandler(m *discordgo.MessageCreate) {
	defer b.DeleteUserMessage(m)
	if !b.restricted(m) {
		b.DisplayTemporaryMessage(m, "", "You do not have permission to stop the player", "")
		return
	}
	if !b.Player.playing {
		b.SendNamedNotice(m, "Requested by", "⏹️ Stop", "Nothing to do", "")
		return
	}
	b.Player.session.Stop() // nolint:errcheck
	b.Player.stop = true
	b.SendNamedNotice(m, "Requested by", "⏹️ Stop", "", "")
}

func (b *BotInstance) SkipHandler(m *discordgo.MessageCreate) {
	defer b.DeleteUserMessage(m)
	if !b.restricted(m) {
		b.DisplayTemporaryMessage(m, "", "You do not have permission to arbitrarily skip a track", "Tip: Start a vote using '!fox vote'")
		return
	}

	if b.Player.playing {
		b.Player.session.Stop() // nolint:errcheck
		b.SendNamedNotice(m, "Requested by", "⏭️ Skip", "The currently playing track has been skipped", "Note: This can take a few seconds")
		if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
	} else {
		b.Player.Pop()
		b.SendNamedNotice(m, "Requested by", "⏭️ Skip", "The next track in queue has been skipped", "")
		if err := b.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
	}
}

func (b *BotInstance) JoinHandler(m *discordgo.MessageCreate) {
	defer b.DeleteUserMessage(m)
	if !b.restricted(m) {
		b.DisplayTemporaryMessage(m, "", "Permission denied", "Tip: Only admins and DJs can do that")
		return
	}

	b.log.Debug().Str("user", m.Author.Username).Str("method", "join").Msg("called")
	if b.Voice == nil {
		voice, err := b.Session.ChannelVoiceJoin(b.conf.Bot.Guild, b.conf.Bot.Channels.Voice, false, true)
		if err != nil {
			b.SendNamedNotice(m, "Requested by", "Unable to join voice channel", "", "Error was: "+err.Error())
			b.log.Error().Err(err).Msg("unable to connect to voice channel")
			return
		}
		b.Voice = voice
		b.log.Debug().Str("user", m.Author.Username).Str("method", "join").Msg("bot joined vocal channel")
	}
}

func (b *BotInstance) LeaveHandler(m *discordgo.MessageCreate) {
	defer b.DeleteUserMessage(m)
	if !b.restricted(m) {
		b.DisplayTemporaryMessage(m, "", "Permission denied", "Tip: Only admins and DJs can do that")
		return
	}

	b.log.Debug().Str("user", m.Author.Username).Str("method", "leave").Msg("called")
	if b.Voice != nil {
		if b.Player.playing {
			b.Player.session.Stop() // nolint:errcheck
			b.Player.stop = true
		}
		if err := b.Voice.Disconnect(); err != nil {
			b.SendNamedNotice(m, "Requested by", "Unable to leave voice channel", "", "Error was: "+err.Error())
			b.log.Error().Err(err).Msg("unable to disconnect from voice channel")
			return
		}
		b.Voice = nil
		b.log.Debug().Str("user", m.Author.Username).Str("method", "leave").Msg("bot left vocal channel")
		return
	}
}
