package bot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/Depado/fox/storage"
)

func (b *Bot) GetInvitingUser(s *discordgo.Session, g *discordgo.GuildCreate) string {
	st, err := s.GuildAuditLog(g.ID, "", "", 28, 50)
	if err != nil {
		b.log.Warn().Err(err).Msg("unable to get audit log, fallback on owner")
		return g.OwnerID
	}
	for _, le := range st.AuditLogEntries {
		if le.TargetID == s.State.User.ID {
			b.log.Debug().Msg("found inviting user in audit log")
			return le.UserID
		}
	}
	b.log.Warn().Err(err).Msg("audit log doens't contain invite, fallback on owner")
	return g.OwnerID
}

func (b *Bot) GuildCreatedHandler(s *discordgo.Session, g *discordgo.GuildCreate) {
	gc, err := b.storage.GetGuildConf(g.ID)
	if err != nil && errors.Is(err, storage.ErrGuildNotFound) {
		if gc, err = b.storage.NewGuild(g); err != nil {
			b.log.Err(err).Msg("unable to instantiate new guild state")
			return
		}
		b.MessageInviter(s, g)
	} else if err != nil {
		b.log.Err(err).Msg("unable to fetch guild state")
		return
	} else {
		if err = b.storage.UpdateGuildInfo(g); err != nil {
			b.log.Err(err).Msg("unable to update guild info")
			return
		}
	}

	if err := b.players.Create(b.session, b.conf, b.log, g.ID, b.storage, gc); err != nil {
		b.log.Err(err).Msg("unable to handle guild create")
		return
	}
	b.log.Debug().Str("guild", g.ID).Str("name", g.Name).Msg("registered new player")
}

func (b *Bot) MessageInviter(s *discordgo.Session, g *discordgo.GuildCreate) {
	u := b.GetInvitingUser(s, g)
	c, err := s.UserChannelCreate(u)
	if err != nil {
		b.log.Err(err).Msg("can't establish channel for inviting user")
		return
	}
	e := &discordgo.MessageEmbed{
		Title: "Fox Manual",
		Color: 0xff5500,
		Description: fmt.Sprintf(
			"Hey <@%s>!\nThank you for inviting me in your server! But before I "+
				"can play some music for you, you need to tell me **where** I "+
				"can play music.\nIn order to do that, please use the\n"+
				"`%s setup voice \"<vocal name or ID>\"`\n command "+
				"in one of your channels.\n\nDon't hesitate to use the `%s help` "+
				"command to see how you can make me play music!"+
				"", u, b.conf.Bot.Prefix, b.conf.Bot.Prefix,
		),
	}
	if _, err := s.ChannelMessageSendEmbed(c.ID, e); err != nil {
		b.log.Err(err).Msg("can't send message to user")
		return
	}
}
