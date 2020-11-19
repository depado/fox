package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/guild"
	"github.com/Depado/fox/message"
	"github.com/Depado/fox/player"
	"github.com/Depado/fox/storage"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type setup struct {
	BaseCommand
	Storage *storage.StormDB
}

func (c *setup) handleVoiceChannel(s *discordgo.Session, m *discordgo.Message, gstate *guild.State, value string) {
	if err := gstate.SetChannel(s, value, true); err != nil {
		if errors.Is(err, guild.ChannelNotFoundError) {
			message.SendShortTimedNotice(s, m, "I couldn't find any vocal channel named like this", c.log)
			return
		}
		c.log.Err(err).Msg("unable to set voice channel")
		return
	}
	message.SendShortTimedNotice(s, m, "Alright, I'll stream the music to this channel from now on", c.log)
}

func (c *setup) handleTextChannel(s *discordgo.Session, m *discordgo.Message, gstate *guild.State, value string) {
	if err := gstate.SetChannel(s, value, false); err != nil {
		if errors.Is(err, guild.ChannelNotFoundError) {
			message.SendShortTimedNotice(s, m, "I couldn't find any text channel named like this", c.log)
			return
		}
		c.log.Err(err).Msg("unable to set voice channel")
		return
	}
	message.SendShortTimedNotice(s, m, fmt.Sprintf("Noted, the music channel is now <#%s>", gstate.TextChannel), c.log)
}

func (c *setup) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	var err error
	var gstate *guild.State

	if gstate, err = c.Storage.GetGuildState(m.GuildID); err != nil {
		c.log.Err(err).Msg("unable to fetch guild state")
		return
	}

	if len(args) < 1 {
		return
	}

	setup := strings.Split(args[0], "=")
	if len(setup) != 2 {
		c.log.Error().Str("arg", args[0]).Msg("unable to parse param")
		return
	}
	param, value := setup[0], strings.Trim(setup[1], `"`)
	c.log.Debug().Str("param", param).Str("value", value).Msg("got that")

	switch param {
	case "voice":
		c.handleVoiceChannel(s, m, gstate, value)
	case "text":
		c.handleTextChannel(s, m, gstate, value)
	default:
		message.SendShortTimedNotice(s, m, "Unknwon parameter", c.log)
		return
	}

	if err := c.Storage.SaveGuildState(gstate); err != nil {
		c.log.Err(err).Msg("unable to save guild state")
	}
}

func NewSetupCommand(p *player.Players, log *zerolog.Logger, storage *storage.StormDB) Command {
	cmd := "setup"
	return &setup{
		BaseCommand: BaseCommand{
			ChannelRestriction: acl.Anywhere,
			RoleRestriction:    acl.Admin,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long: cmd,
			Help: Help{
				Usage:       cmd,
				ShortDesc:   "Setup the bot",
				Description: "This commands allows to setup the various bits of the bot.",
				Examples: []Example{
					{Command: `setup voice="My Vocal Channel"`, Explanation: "Setup the vocal channel of the bot"},
					{Command: `setup text="fox-radio"`, Explanation: "Setup the text channel of the bot"},
					{Command: `setup djrole="DJ"`, Explanation: "Setup the privileged DJ role"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
		Storage: storage,
	}
}
