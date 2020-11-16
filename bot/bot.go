package bot

import (
	"fmt"
	"sync"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/cmd"
	"github.com/Depado/fox/commands"
	"github.com/Depado/fox/healthcheck"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type Bot struct {
	log         *zerolog.Logger
	session     *discordgo.Session
	allCommands []commands.Command
	commands    *CommandMap
	conf        *cmd.Conf
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

func NewBot(s *discordgo.Session, l *zerolog.Logger, conf *cmd.Conf, a *acl.ACL, cmds []commands.Command) *Bot {
	b := &Bot{
		log:         l,
		conf:        conf,
		session:     s,
		acl:         a,
		allCommands: cmds,
		commands:    &CommandMap{m: make(map[string]commands.Command)},
	}

	for _, cmd := range cmds {
		b.AddCommand(cmd)
	}

	b.session.AddHandler(b.MessageCreatedHandler)

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

func Run(l *zerolog.Logger, b *Bot, c *cmd.Conf, r *healthcheck.HealthCheck) {
	go func() {
		if err := r.Engine.Run(fmt.Sprintf(":%d", c.Port)); err != nil {
			l.Fatal().Err(err).Msg("unable to start healthcheck router")
		}
	}()
	l.Info().Msg("Bot is now running")
}
