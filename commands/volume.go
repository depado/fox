package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/message"
	"github.com/depado/fox/player"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type volume struct {
	BaseCommand
}

func (c *volume) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	var v int
	var err error
	var emoji = "🔉"

	p := c.Players.GetPlayer(m.GuildID)
	if p == nil {
		c.log.Error().Msg("no player associated to guild ID")
		return
	}

	if len(args) < 1 {
		v = p.Volume() * 100 / 256
		if v > 100 {
			emoji = "🔊"
		} else if v < 100 {
			emoji = "🔈"
		}
		body := fmt.Sprintf("%s Volume is currently %d%% ", emoji, v)
		if err := message.SendTimedReply(s, m, "", body, "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send reply")
		}
		return
	}

	if args[0] == "reset" {
		v = 100
	} else {
		vol := strings.Trim(args[0], "%")
		if v, err = strconv.Atoi(vol); err != nil {
			c.log.Debug().Err(err).Str("volume", vol).Msg("volume isn't a number")
			if err := message.SendTimedReply(s, m, "", "Invalid volume level", "", 5*time.Second); err != nil {
				c.log.Err(err).Msg("unable to send timed reply")
			}
			return
		}
	}

	if v > 200 || v < 0 {
		if err := message.SendTimedReply(s, m, "", "Invalid volume level (1% → 200%)", "", 5*time.Second); err != nil {
			c.log.Err(err).Msg("unable to send timed reply")
		}
		return
	}

	// Actually set the volume
	if err := p.SetVolumePercent(v); err != nil {
		c.log.Err(err).Msg("unable to set volume percentage")
		return
	}

	// User feedback
	if v > 100 {
		emoji = "🔊"
	} else if v < 100 {
		emoji = "🔈"
	}
	body := fmt.Sprintf("%s Volume set to %d%% by <@%s>", emoji, v, m.Author.ID)
	if err := message.SendReply(s, m, "", body, ""); err != nil {
		c.log.Err(err).Msg("unable to send reply")
	}
}

func NewVolumeCommand(p *player.Players, log zerolog.Logger) Command {
	cmd := "volume"
	return &volume{
		BaseCommand{
			ChannelRestriction: acl.Music,
			RoleRestriction:    acl.Privileged,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"vol"},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Set or see the volume of the player",
				Description: "This command will set the volume for the " +
					"following tracks or display the current volume if no " +
					"argument is provided. The volume change will be applied " +
					"to the next tracks and not to the currently playing one.",
				Examples: []Example{
					{Command: "volume", Explanation: "Display the current volume"},
					{Command: "volume reset", Explanation: "Resets the volume to 100%"},
					{Command: "volume 200%", Explanation: "Sets the volume to the maximum possible"},
					{Command: "vol 50%", Explanation: "Sets the volume to half the normal volume using the alias"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
	}
}
