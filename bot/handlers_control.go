package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *BotInstance) PlayHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if b.Player.playing {
		b.SendTimedNotice("", "▶️ Play: Nothing to do", "", m.ChannelID, 5*time.Second)
		return
	}
	b.PlayQueue()
	b.SendNotice("", fmt.Sprintf("▶️ Started playing for <@%s>", m.Author.ID), "", m.ChannelID)
}

func (b *BotInstance) PauseHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.Player.playing {
		b.SendTimedNotice("", "⏸️ Pause: Nothing to do", "", m.ChannelID, 5*time.Second)
		return
	}

	if !b.Player.pause {
		b.Player.stream.SetPaused(true)
		b.Player.pause = true
		b.SendNotice("", fmt.Sprintf("⏸️ Paused by <@%s>", m.Author.ID), "", m.ChannelID)
	} else {
		b.SendTimedNotice("", "⏸️ Pause: Nothing to do", "", m.ChannelID, 5*time.Second)
	}
}

func (b *BotInstance) ResumeHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.Player.playing {
		b.SendNotice("", "▶️ Resume: Nothing to do", "", m.ChannelID)
		return
	}
	if b.Player.pause {
		b.Player.stream.SetPaused(false)
		b.Player.pause = false
		b.SendNotice("", fmt.Sprintf("⏯️ Resumed by <@%s>", m.Author.ID), "", m.ChannelID)
	} else {
		m := b.SendNotice("", "▶️ Resume: Nothing to do", "", m.ChannelID)
		b.DeleteAfter(m, 5*time.Second)
	}
}

func (b *BotInstance) StopHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.restricted(m) {
		b.SendTimedNotice("", "You do not have permission to stop the player", "", m.ChannelID, 5*time.Second)
		return
	}

	if !b.Player.playing {
		b.SendTimedNotice("", "⏹️ Stop: Nothing to do", "", m.ChannelID, 5*time.Second)
		return
	}

	b.Player.session.Stop() // nolint:errcheck
	b.Player.stop = true
	b.SendNotice("", fmt.Sprintf("⏹️ Stopped by <@%s>", m.Author.ID), "", m.ChannelID)
}

func (b *BotInstance) SkipHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.restricted(m) {
		b.SendTimedNotice(
			"", "You do not have permission to arbitrarily skip a track",
			`Tip: Start a vote using "!fox vote"`, m.ChannelID, 10*time.Second,
		)
		return
	}

	if b.Player.playing {
		b.Player.session.Stop() // nolint:errcheck
		b.SendNotice(
			"", fmt.Sprintf("⏭️ <@%s> skipped the currently playing track", m.Author.ID),
			"Note: This can take a few seconds", m.ChannelID,
		)
	} else {
		b.Player.Pop()
		b.SendNotice("", fmt.Sprintf("⏭️ <@%s> skipped the next track in queue", m.Author.ID), "", m.ChannelID)
	}
}

func (b *BotInstance) JoinHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.restricted(m) {
		b.SendTimedNotice("", "You do not have this permission", "Tip: Only admins and DJs can do that", m.ChannelID, 10*time.Second)
		return
	}

	b.log.Debug().Str("user", m.Author.Username).Str("method", "join").Msg("called")
	if b.Voice == nil {
		voice, err := b.Session.ChannelVoiceJoin(b.conf.Bot.Guild, b.conf.Bot.Channels.Voice, false, true)
		if err != nil {
			b.SendNotice("", fmt.Sprintf("Unable to join voice channel as instructed by <@%s>", m.Author.ID), "Error was: "+err.Error(), m.ChannelID)
			b.log.Error().Err(err).Msg("unable to connect to voice channel")
			return
		}
		b.Voice = voice
		b.log.Debug().Str("user", m.Author.Username).Str("method", "join").Msg("bot joined vocal channel")
	}
}

func (b *BotInstance) LeaveHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.restricted(m) {
		b.SendTimedNotice("", "You do not have this permission", "Tip: Only admins and DJs can do that", m.ChannelID, 5*time.Second)
		return
	}

	b.log.Debug().Str("user", m.Author.Username).Str("method", "leave").Msg("called")
	if b.Voice != nil {
		if b.Player.playing {
			b.Player.session.Stop() // nolint:errcheck
			b.Player.stop = true
		}
		if err := b.Voice.Disconnect(); err != nil {
			b.SendNotice("", fmt.Sprintf("Unable to leave voice channel as instructed by <@%s>", m.Author.ID), "Error was: "+err.Error(), m.ChannelID)
			b.log.Error().Err(err).Msg("unable to disconnect from voice channel")
			return
		}
		b.Voice = nil
		b.log.Debug().Str("user", m.Author.Username).Str("method", "leave").Msg("bot left vocal channel")
		return
	}
}

func (b *BotInstance) StatsHandler(m *discordgo.MessageCreate) {
	defer b.Delete(m.Message)
	if !b.restricted(m) {
		b.SendTimedNotice("", "You do not have this permission", "Tip: Only admins and DJs can do that", m.ChannelID, 5*time.Second)
		return
	}

	if !b.Player.playing || b.Player.stream == nil || b.Player.session == nil {
		b.SendTimedNotice("", "There is currently no stream", "", m.ChannelID, 5*time.Second)
		return
	}
	if off, _ := b.Player.stream.Finished(); off {
		b.SendTimedNotice("", "There is currently no stream", "", m.ChannelID, 5*time.Second)
	}

	s := b.Player.session.Stats()
	e := &discordgo.MessageEmbed{
		Title: "Stream & encoding stats",
		Color: 0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Playback", Value: b.Player.stream.PlaybackPosition().String(), Inline: true},
			{Name: "Encoded", Value: s.Duration.String(), Inline: true},
			{Name: "Size", Value: fmt.Sprintf("%5d kB", s.Size), Inline: true},
			{Name: "Bitrate", Value: fmt.Sprintf("%6.2f kB/s", s.Bitrate), Inline: true},
			{Name: "Speed", Value: fmt.Sprintf("%5.1fx", s.Speed), Inline: true},
		},
	}

	mess, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
		return
	}
	b.DeleteAfter(mess, 10*time.Second)
}
