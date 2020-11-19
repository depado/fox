package bot

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/cmd"
	"github.com/Depado/fox/commands"
	"github.com/Depado/fox/guild"
	"github.com/Depado/fox/player"
	"github.com/Depado/fox/storage"
	"github.com/asdine/storm/v3"
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Bot struct {
	log         zerolog.Logger
	session     *discordgo.Session
	allCommands []commands.Command
	commands    *CommandMap
	conf        *cmd.Conf
	players     *player.Players
	storage     *storage.StormDB
	acl         *acl.ACL
}

type CommandMap struct {
	sync.RWMutex
	m map[string]commands.Command
}

func (cm *CommandMap) Get(c string) (commands.Command, bool) {
	cm.Lock()
	defer cm.Unlock()
	co, ok := cm.m[c]
	return co, ok
}

func NewBot(s *discordgo.Session, l *zerolog.Logger, c *cmd.Conf, cmds []commands.Command, p *player.Players, storage *storage.StormDB, a *acl.ACL) *Bot {
	b := &Bot{
		log:         l.With().Str("component", "bot").Logger(),
		conf:        c,
		session:     s,
		allCommands: cmds,
		players:     p,
		commands:    &CommandMap{m: make(map[string]commands.Command)},
		storage:     storage,
		acl:         a,
	}

	for _, cmd := range cmds {
		b.AddCommand(cmd)
	}

	for _, g := range b.session.State.Guilds {
		var err error
		var gstate *guild.State

		if gstate, err = b.storage.GetGuildState(g.ID); err != nil {
			if err == storm.ErrNotFound {
				if gstate, err = b.storage.NewGuildState(g.ID); err != nil {
					b.log.Err(err).Msg("unable to instantiate new guild state")
					continue
				}
			} else {
				b.log.Err(err).Msg("unable to fetch guild state")
				continue
			}
		}

		if err := p.Create(s, c, l, g.ID, storage, gstate); err != nil {
			l.Err(err).Msg("unable to handle guild create")
			continue
		}
		l.Debug().Str("guild", g.ID).Str("name", g.Name).Msg("registered new player")
	}

	b.session.AddHandler(b.MessageCreatedHandler)
	b.session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {})
	b.session.AddHandler(func(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {})

	return b
}

func (b *Bot) AddCommand(c commands.Command) {
	b.commands.Lock()
	defer b.commands.Unlock()

	long, aliases := c.Calls()
	if long == "" && len(aliases) == 0 {
		b.log.Error().Msg("unable to add command, no long call or aliases")
		return
	}
	if long != "" {
		aliases = append([]string{long}, aliases...)
	}

	for _, a := range aliases {
		if _, ok := b.commands.m[a]; ok {
			b.log.Error().Str("command", a).Msg("conflicting command, ignoring")
		} else {
			b.commands.m[a] = c
			b.log.Debug().Str("command", a).Msg("registered command")
		}
	}
}

func Run(l *zerolog.Logger, b *Bot, c *cmd.Conf) {
	go func() {
		gin.SetMode("release")
		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		if err := r.Run(fmt.Sprintf(":%d", c.Port)); err != nil {
			l.Fatal().Err(err).Msg("unable to start healthcheck router")
		}
	}()
	l.Info().Msg("Bot is now running")
}
