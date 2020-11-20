package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/player"
	"github.com/Depado/fox/soundcloud"
	"github.com/Depado/fox/storage"
)

func InitializeAllCommands(p *player.Players, l zerolog.Logger, sp *soundcloud.SoundCloudProvider, st *storage.StormDB) []Command {
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
		NewSetupCommand(p, l, st),
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
	DMCapability      bool
}

type SubCommand struct {
	Long        string
	Aliases     []string
	Arg         string
	Description string
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
	SubCommands []SubCommand
	Help        Help
	Players     *player.Players
	log         zerolog.Logger
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
	if len(c.SubCommands) > 0 {
		desc += "\n\n__**Subcommands**__"
		for _, sc := range c.SubCommands {
			desc += fmt.Sprintf("\n\n`%s", sc.Long)
			for _, a := range sc.Aliases {
				desc += fmt.Sprintf("/%s", a)
			}
			if sc.Arg != "" {
				desc += fmt.Sprintf(" <%s>", sc.Arg)
			}
			desc += "`"
			if sc.Description != "" {
				desc += "\n" + sc.Description
			}
		}
	}
	desc += "\n\n__**Restrictions**__\n\n"
	desc += fmt.Sprintf(
		"**%s\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0%s**",
		acl.ChannelRestrictionString(c.ChannelRestriction), acl.RoleRestrictionString(c.RoleRestriction),
	)

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
		desc += "\n\n__**Examples**__"
		fields := []*discordgo.MessageEmbedField{}
		for _, e := range c.Help.Examples {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("`%s %s`", prefix, e.Command), Value: e.Explanation, Inline: false,
			})
		}
		e.Fields = fields
	}
	e.Description = desc

	uc, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		c.log.Err(err).Msg("unable to get channel to DM user")
		return
	}

	if _, err := s.ChannelMessageSendEmbed(uc.ID, e); err != nil {
		c.log.Err(err).Msg("unable to send embed")
	}
}
