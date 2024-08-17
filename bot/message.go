package bot

import (
	"fmt"
	"strings"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/message"
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
		message.Delete(s, m.Message, b.log)
		return
	}
	args := fields[2:]

	// Check for help command
	if fields[1] == "help" || fields[1] == "h" {
		defer message.Delete(s, m.Message, b.log)

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
		message.SendShortTimedNotice(s, m.Message, "Unknown command", b.log)
		message.Delete(s, m.Message, b.log)
		return
	}
	opts := c.Opts()

	// Check if DM and if so, check if the command has DM capability
	if m.GuildID == "" && !opts.DMCapability {
		err := message.SendReply(s, m.Message, "", "Commands must be executed in a server channel.\nThe only exception is the `help` command.", "")
		if err != nil {
			b.log.Err(err).Msg("unable to send reply")
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
		msg := fmt.Sprintf("You do not have permission to do that.\n**%s**", acl.RestrictionString(cr, rr))
		message.SendShortTimedNotice(s, m.Message, msg, b.log)
		message.Delete(s, m.Message, b.log)
		return
	}

	// Act on command options
	if opts.ArgsRequired && len(args) == 0 {
		msg := fmt.Sprintf(
			"The `%s` command requires additional arguments.\nType `%s help %s` to view this command's help page",
			c.GetHelp().Usage,
			b.conf.Bot.Prefix,
			c.GetHelp().Usage,
		)
		message.SendShortTimedNotice(s, m.Message, msg, b.log)
		message.Delete(s, m.Message, b.log)
		return
	}
	if opts.DeleteUserMessage {
		defer message.Delete(s, m.Message, b.log)
	}
	c.Handler(s, m.Message, args)
}
