package bot

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.uber.org/fx"

	"github.com/depado/fox/acl"
	"github.com/depado/fox/cmd"
	"github.com/depado/fox/commands"
	"github.com/depado/fox/player"
	"github.com/depado/fox/storage"
)

type Bot struct {
	log         zerolog.Logger
	session     *discordgo.Session
	allCommands []commands.Command
	commands    *CommandMap
	conf        *cmd.Conf
	players     *player.Players
	storage     *storage.BoltStorage
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

func NewBot(lc fx.Lifecycle, l zerolog.Logger, c *cmd.Conf, cmds []commands.Command, p *player.Players, storage *storage.BoltStorage, a *acl.ACL) *Bot {
	log := l.With().Str("component", "bot").Logger()
	dg, err := discordgo.New("Bot " + c.Bot.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open connection")
	}
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	b := &Bot{
		log:         log,
		conf:        c,
		session:     dg,
		allCommands: cmds,
		players:     p,
		commands:    &CommandMap{m: make(map[string]commands.Command)},
		storage:     storage,
		acl:         a,
	}

	for _, cmd := range cmds {
		b.AddCommand(cmd)
	}

	b.session.AddHandler(b.MessageCreatedHandler)
	b.session.AddHandler(b.GuildCreatedHandler)
	b.session.AddHandler(func(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {})

	if err := dg.Open(); err != nil {
		log.Fatal().Err(err).Msg("unable to open")
	}

	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			b.log.Debug().Str("lifecycle", "stop").Msg("killing players")
			b.players.Kill()
			b.session.Close()
			return nil
		},
	})

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

func Run(l zerolog.Logger, b *Bot, c *cmd.Conf) {
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
