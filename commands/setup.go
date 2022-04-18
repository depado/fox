package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/message"
	"github.com/Depado/fox/models"
	"github.com/Depado/fox/player"
	"github.com/Depado/fox/storage"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type setup struct {
	BaseCommand
	Storage *storage.BoltStorage
}

func (c *setup) handleVoiceChannel(s *discordgo.Session, m *discordgo.Message, gconf *models.Conf, value string) {
	if err := gconf.SetChannel(s, value, true); err != nil {
		if errors.Is(err, models.ChannelNotFoundError) {
			message.SendShortTimedNotice(s, m, "I couldn't find any vocal channel named like this", c.log)
			return
		}
		c.log.Err(err).Msg("unable to set voice channel")
		return
	}
	message.SendShortTimedNotice(s, m, "Alright, I'll stream the music to this channel from now on", c.log)
}

func (c *setup) handleTextChannel(s *discordgo.Session, m *discordgo.Message, gconf *models.Conf, value string) {
	if err := gconf.SetChannel(s, value, false); err != nil {
		if errors.Is(err, models.ChannelNotFoundError) {
			message.SendShortTimedNotice(s, m, "I couldn't find any text channel named like this", c.log)
			return
		}
		c.log.Err(err).Msg("unable to set voice channel")
		return
	}
	message.SendShortTimedNotice(s, m, fmt.Sprintf("Noted, the music channel is now <#%s>", gconf.TextChannel), c.log)
}

func (c *setup) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	var err error
	var gconf *models.Conf

	if gconf, err = c.Storage.GetGuildConf(m.GuildID); err != nil {
		c.log.Err(err).Msg("unable to fetch guild conf")
		return
	}

	if len(args) < 2 {
		message.SendShortTimedNotice(s, m, "This command requires two arguments", c.log)
		return
	}

	v := strings.Join(args[1:], " ")
	param, value := args[0], strings.Trim(v, `"`)
	switch param {
	case "voice":
		c.handleVoiceChannel(s, m, gconf, value)
		if pl := c.Players.GetPlayer(m.GuildID); pl != nil {
			pl.UpdateConf(gconf)
		}
	case "text":
		c.handleTextChannel(s, m, gconf, value)
	default:
		message.SendShortTimedNotice(s, m, "Unknwon parameter", c.log)
		return
	}

	if err := c.Storage.SaveGuildConf(gconf); err != nil {
		c.log.Err(err).Msg("unable to save guild state")
	}
}

func NewSetupCommand(p *player.Players, log zerolog.Logger, storage *storage.BoltStorage) Command {
	cmd := "setup"
	return &setup{
		BaseCommand: BaseCommand{
			ChannelRestriction: acl.Anywhere,
			RoleRestriction:    acl.Admin,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			SubCommands: []SubCommand{
				{Long: "voice", Arg: "voice channel name", Description: "Setup the voice channel"},
				{Long: "text", Arg: "text channel name", Description: "Setup the text channel"},
			},
			Long: cmd,
			Help: Help{
				Usage:       cmd,
				ShortDesc:   "Setup the bot",
				Description: "This commands allows to setup the various bits of the bot.",
				Examples: []Example{
					{Command: `setup voice "My Vocal Channel"`, Explanation: "Setup the vocal channel of the bot"},
					{Command: `setup text fox-radio`, Explanation: "Setup the text channel of the bot"},
					{Command: `setup djrole DJ`, Explanation: "Setup the privileged DJ role"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
		Storage: storage,
	}
}
