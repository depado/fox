package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/message"
	"github.com/depado/fox/player"
	"github.com/depado/fox/soundcloud"
)

type add struct {
	BaseCommand
	sp *soundcloud.SoundCloudProvider
}

func (c *add) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		if err := message.SendTimedReply(s, m, "", "This doesn't look like a SoundCloud URL", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
			return
		}
		return
	}

	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	tr, e, err := c.sp.GetPlaylist(url, m)
	if err == nil {
		p.Queue.Append(tr...)
		e.Description = fmt.Sprintf("Added **%d** tracks to end of queue", len(tr))
		if _, err = s.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			c.log.Err(err).Msg("unable to send embed")
		}
		return
	}

	t, e, err := c.sp.GetTrack(url, m)
	if err == nil {
		p.Queue.Append(t)
		e.Description = "Added one tracks to end of queue"
		if _, err = s.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			c.log.Err(err).Msg("unable to send embed")
		}
		return
	}

	if err := message.SendTimedReply(s, m, "", "This is neither a playlist nor a track", "", 5*time.Second); err != nil {
		c.log.Err(err).Msg("unable to send timed reply")
	}
}

func NewAddCommand(p *player.Players, log zerolog.Logger, sp *soundcloud.SoundCloudProvider) Command {
	cmd := "add"
	return &add{
		sp: sp,
		BaseCommand: BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      true,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"a"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Add a track or playlist to the end of queue",
				Description: "This command can be used to add tracks and " +
					"complete playlists to the end of the queue. " +
					"It currently only suppports soundcloud URLs.",
				Examples: []Example{
					{Command: "add <url>", Explanation: "Add the track to the end of queue"},
					{Command: "a <url>", Explanation: "Add the track using the alias"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}

type next struct {
	BaseCommand
	sp *soundcloud.SoundCloudProvider
}

func (c *next) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	url := args[0]
	url = strings.Trim(url, "<>")
	if !strings.HasPrefix(url, "https://soundcloud.com") {
		if err := message.SendTimedReply(s, m, "", "This doesn't look like a SoundCloud URL", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
			return
		}
		return
	}

	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	tr, e, err := c.sp.GetPlaylist(url, m)
	if err == nil {
		p.Queue.Prepend(tr...)
		e.Description = fmt.Sprintf("Added **%d** tracks to start of queue", len(tr))
		if _, err = s.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			c.log.Err(err).Msg("unable to send embed")
		}
		return
	}

	t, e, err := c.sp.GetTrack(url, m)
	if err == nil {
		p.Queue.Prepend(t)
		e.Description = "Added one tracks to start of queue"
		if _, err = s.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			c.log.Err(err).Msg("unable to send embed")
		}
		return
	}

	if err := message.SendTimedReply(s, m, "", "This is neither a playlist nor a track", "", 5*time.Second); err != nil {
		c.log.Err(err).Msg("unable to send timed reply")
	}
}

func NewNextCommand(p *player.Players, log zerolog.Logger, sp *soundcloud.SoundCloudProvider) Command {
	cmd := "next"
	return &next{
		sp: sp,
		BaseCommand: BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      true,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"n"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Add a track or playlist at the start of queue",
				Description: "This command can be used to add tracks and " +
					"complete playlists at the start of the queue. " +
					"It currently only suppports soundcloud URLs.",
				Examples: []Example{
					{Command: "next <url>", Explanation: "Add the track to the start of queue"},
					{Command: "n <url>", Explanation: "Add the track using the alias"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}
