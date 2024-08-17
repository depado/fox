package bot

import (
	"fmt"

	"github.com/depado/fox/message"
	"github.com/bwmarrin/discordgo"
)

// DisplayCommandHelp will display the help section for a given command
func (b *Bot) DisplayCommandHelp(s *discordgo.Session, m *discordgo.MessageCreate, cmd string) {
	c, ok := b.commands.Get(cmd)
	if !ok {
		message.SendShortTimedNotice(s, m.Message, fmt.Sprintf("Unknown command %s", cmd), b.log)
		return
	}
	c.DisplayHelp(s, m.Message, b.conf.Bot.Prefix)
}

// DisplayGlobalHelp will cycle through the available command and display a
// global help
func (b *Bot) DisplayGlobalHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	e := &discordgo.MessageEmbed{
		Title: "❓ Fox Help",
		Description: fmt.Sprintf(
			"Fox is a music bot. To interact with it, use `%s <command>` where "+
				"`command` is one of the following commands.\nTo get more details "+
				"about a given command, you can also use `%s help <command>`",
			b.conf.Bot.Prefix, b.conf.Bot.Prefix),
		Color: 0xff5500,
	}

	fields := []*discordgo.MessageEmbedField{}
	for _, c := range b.allCommands {
		h := c.GetHelp()
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("`%s`", h.Usage),
			Value:  h.ShortDesc,
			Inline: true,
		})
	}
	e.Fields = fields

	uc, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		b.log.Err(err).Msg("unable to get create channel DM for user")
		return
	}
	if _, err := s.ChannelMessageSendEmbed(uc.ID, e); err != nil {
		b.log.Err(err).Msg("unable to send reply")
	}
}
