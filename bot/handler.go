package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/message"
	"github.com/bwmarrin/discordgo"
)

// InitialCheck will perform basic checks, unrelated to commands
// It will check if the prefix is present in the message, whether or not the
// sender is a bot, or if the sender is itself
// If this method returns true then it is safe to proceed
func (b *Bot) InitialCheck(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return strings.HasPrefix(m.Content, b.conf.Bot.Prefix) && m.Author.ID != s.State.User.ID && !m.Author.Bot
}

func (b *Bot) MessageCreatedHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Quick check for prefix and to not react to itself
	if !b.InitialCheck(s, m) {
		return
	}

	// Check for well formed command
	fields := strings.Fields(m.Content)
	if len(fields) < 2 {
		if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
		return
	}
	args := fields[2:]

	if fields[1] == "help" || fields[1] == "h" {
		defer func() {
			if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				b.log.Err(err).Msg("unable to delete user message")
			}
		}()

		if len(args) < 1 {
			b.DisplayGlobalHelp(s, m)
		} else {
			b.DisplayCommandHelp(s, m, args[0])
		}
		return
	}

	// Retrieve the associated command
	c, ok := b.commands.Get(fields[1])
	if !ok {
		err := message.SendTimedReply(s, m.Message, "", "Unknown command", "", 5*time.Second)
		if err != nil {
			b.log.Err(err).Msg("unable to send timed reply")
		}
		if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
		return
	}

	// Check permissions
	cr, rr := c.ACL()
	ok, err := b.acl.Check(s, m.Message, rr, cr)
	if err != nil {
		b.log.Err(err).Msg("unable to check acl")
		return
	}
	if !ok {
		msg := fmt.Sprintf("You do not have permission to do that.\n%s\n%s", acl.RoleRestrictionString(rr), acl.ChannelRestrictionString(cr))
		err := message.SendTimedReply(s, m.Message, "", msg, "", 5*time.Second)
		if err != nil {
			b.log.Err(err).Msg("unable to send timed reply")
		}
		if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
		return
	}

	// Act on command options
	opts := c.Opts()
	if opts.ArgsRequired && len(args) == 0 {
		msg := fmt.Sprintf(
			"The `%s` command requires additional arguments.\nType `%s help %s` to view this command's help page",
			c.GetHelp().Usage,
			b.conf.Bot.Prefix,
			c.GetHelp().Usage,
		)
		if err := message.SendTimedReply(s, m.Message, "", msg, "", 5*time.Second); err != nil {
			b.log.Err(err).Msg("unable to send timed reply")
		}
		if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			b.log.Err(err).Msg("unable to delete user message")
		}
		return
	}
	if opts.DeleteUserMessage {
		defer func() {
			if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				b.log.Err(err).Msg("unable to delete user message")
			}
		}()
	}
	c.Handler(s, m.Message, args)
}
