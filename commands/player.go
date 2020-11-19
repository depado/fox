package commands

import (
	"fmt"
	"time"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/message"
	"github.com/Depado/fox/player"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type play struct {
	BaseCommand
}

func (c *play) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if p.State.Playing && p.State.Paused {
		p.Resume()
		msg := fmt.Sprintf("⏯️ Resumed by <@%s>", m.Author.ID)
		if err := message.SendReply(s, m, "", msg, ""); err != nil {
			c.log.Err(err).Msg("unable to ")
		}
		return
	}
	if p.State.Playing {
		message.SendShortTimedNotice(s, m, "Already playing", c.log)
		return
	}

	if p.Queue.Len() == 0 {
		message.SendShortTimedNotice(s, m, "The queue is empty", c.log)
		return
	}

	msg := fmt.Sprintf("▶️ Started playing for <@%s>", m.Author.ID)
	if len(args) > 0 && args[0] == "ambient" {
		if err := p.SetVolumePercent(50); err != nil {
			c.log.Err(err).Msg("unable to set volume")
		} else {
			msg += " in ambient mode"
		}
	}
	p.Play()
	if err := message.SendReply(s, m, "", msg, ""); err != nil {
		c.log.Err(err).Msg("unable to send reply")
	}
}

func NewPlayCommand(p *player.Players, log *zerolog.Logger) Command {
	cmd := "play"
	return &play{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long: cmd,
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Start playing the queue",
				Description: "This command will start playing the queue. " +
					"It has no effect if the player is already " +
					"active.\nThe bot will join the vocal channel when playing " +
					"starts.\n\n`ambient` can be passed as an extra argument " +
					"to play in ambient mode with a lower volume.",
				Examples: []Example{
					{Command: "play", Explanation: "Start playing"},
					{Command: "play ambient", Explanation: "Start playing in ambient mode"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}

type stop struct {
	BaseCommand
}

func (c *stop) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if !p.State.Playing {
		if err := message.SendTimedReply(s, m, "", "Nothing to do", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
		}
		return
	}

	p.Stop()
	msg := fmt.Sprintf("⏹️ Stopped by <@%s>", m.Author.ID)
	if err := message.SendReply(s, m, "", msg, ""); err != nil {
		c.log.Err(err).Msg("unable to send reply")
	}
}

func NewStopCommand(p *player.Players, log *zerolog.Logger) Command {
	cmd := "stop"
	return &stop{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Privileged,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long: cmd,
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Stop the player",
				Description: "This command will stop the player. If the player " +
					"is already stopped, this command has no effect.",
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}

type pause struct {
	BaseCommand
}

func (c *pause) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if p.State.Paused {
		if err := message.SendTimedReply(s, m, "", "Already paused", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
		}
		return
	}
	p.Pause()
	msg := fmt.Sprintf("⏸️ Paused by <@%s>", m.Author.ID)
	if err := message.SendReply(s, m, "", msg, ""); err != nil {
		c.log.Err(err).Msg("unable to ")
	}
}

func NewPauseCommand(p *player.Players, log *zerolog.Logger) Command {
	cmd := "pause"
	return &pause{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long: cmd,
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Pause the player",
				Description: "This command will pause the player, the current " +
					"track won't be skipped and will keep its current playback position. " +
					"If the player is already paused this command has no effect.",
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}

type skip struct {
	BaseCommand
}

func (c *skip) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	p.Skip()
	msg := fmt.Sprintf("⏭️ <@%s> skipped the currently playing track", m.Author.ID)
	if err := message.SendReply(s, m, "", msg, ""); err != nil {
		c.log.Err(err).Msg("unable to send reply")
	}
}

func NewSkipCommand(p *player.Players, log *zerolog.Logger) Command {
	cmd := "skip"
	return &skip{
		BaseCommand{
			ChannelRestriction: acl.Anywhere,
			RoleRestriction:    acl.Privileged,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long: cmd,
			Help: Help{
				Usage:       cmd,
				ShortDesc:   "Skip the currently playing track",
				Description: "This command can be used to skip tracks at will.",
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}

type np struct {
	BaseCommand
}

func (c *np) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if !p.State.Playing {
		if err := message.SendTimedReply(s, m, "", "No track is currently playing", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
		}
		return
	}

	short := true
	if len(args) > 0 && (args[0] == "full" || args[0] == "f") {
		short = false
	}

	e := p.GenerateNowPlayingEmbed(short)
	if e == nil {
		if err := message.SendTimedReply(s, m, "", "No track is currently playing", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
		}
		return
	}

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
		c.log.Err(err).Msg("unable to send embed")
	}
}

func NewNowPlayingCommand(p *player.Players, log *zerolog.Logger) Command {
	cmd := "nowplaying"
	return &np{
		BaseCommand{
			ChannelRestriction: acl.Anywhere,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"np"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Display the currently playing track",
				Description: "This command displays the track that is currently " +
					"being played. This command has no effect if the player isn't running.",
				Examples: []Example{
					{"nowplaying", "Display the current track"},
					{"np", "Short notation"},
					{"nowplaying full", "Display the current track with all the info"},
					{"np f", "Short notation"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}
