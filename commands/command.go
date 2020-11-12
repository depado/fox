package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/player"
	"github.com/Depado/fox/soundcloud"
)

func InitializeAllCommands(p *player.Player, l *zerolog.Logger, sp *soundcloud.SoundCloudProvider) []Command {
	return []Command{
		NewPlayCommand(p, l),
		NewPauseCommand(p, l),
		NewStopCommand(p, l),
		NewVolumeCommand(p, l),
		NewNowPlayingCommand(p, l),
		NewQueueCommand(p, l),
		NewAddCommand(p, l, sp),
		NewNextCommand(p, l, sp),
		NewJamCommand(p, l),
		NewSkipCommand(p, l),
		NewRemoveCommand(p, l),
		NewStatsCommand(p, l),
	}
}

type Command interface {
	Handler(s *discordgo.Session, m *discordgo.Message, args []string)
	DisplayHelp(s *discordgo.Session, m *discordgo.Message, prefix string)
	GetHelp() Help
	ACL() (acl.ChannelRestriction, acl.RoleRestriction)
	Calls() (string, []string)
	Opts() Options
}

type Options struct {
	ArgsRequired      bool
	DeleteUserMessage bool
}

type Example struct {
	Command     string
	Explanation string
}

type Help struct {
	Examples    []Example
	Usage       string
	Title       string
	ShortDesc   string
	Description string
}

type BaseCommand struct {
	// Permissions
	ChannelRestriction acl.ChannelRestriction
	RoleRestriction    acl.RoleRestriction

	// Command calls
	Long    string
	Aliases []string

	Options Options

	// Internal fields
	Help   Help
	Player *player.Player
	log    zerolog.Logger
}

func (c BaseCommand) ACL() (acl.ChannelRestriction, acl.RoleRestriction) {
	return c.ChannelRestriction, c.RoleRestriction
}

func (c BaseCommand) Calls() (string, []string) {
	return c.Long, c.Aliases
}

func (c BaseCommand) Opts() Options {
	return c.Options
}

func (c BaseCommand) GetHelp() Help {
	return c.Help
}

func (c BaseCommand) DisplayHelp(s *discordgo.Session, m *discordgo.Message, prefix string) {
	desc := c.Help.Description

	var cr string
	switch c.ChannelRestriction {
	case acl.Music:
		cr = "🎶 Music text channel only"
	case acl.Anywhere:
		cr = "🌍 No restriction"
	}
	desc += fmt.Sprintf("\n\nChannel Restriction\n**%s**", cr)

	var rr string
	switch c.RoleRestriction {
	case acl.Admin:
		rr = "🔐 Admin"
	case acl.Privileged:
		rr = "🔒 Admin or DJ"
	case acl.Anyone:
		rr = "🔓 No restriction"
	}
	desc += fmt.Sprintf("\n\nRole Restriction\n**%s**", rr)

	var aliases string
	if len(c.Aliases) > 0 {
		for _, a := range c.Aliases {
			aliases += fmt.Sprintf("/`%s`", a)
		}
	}
	e := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("❓ Help for `%s`%s command", c.Long, aliases),
		Color: 0xff5500,
	}

	if len(c.Help.Examples) > 0 {
		desc += "\n\n**Examples:**"
		fields := []*discordgo.MessageEmbedField{}
		for _, e := range c.Help.Examples {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("`%s %s`", prefix, e.Command), Value: e.Explanation, Inline: false,
			})
		}
		e.Fields = fields
	}
	e.Description = desc

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
		c.log.Err(err).Msg("unable to send embed")
	}
}
