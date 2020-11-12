package player

import (
	"context"
	"fmt"
	"sync"

	"github.com/Depado/fox/cmd"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// Player is the struct in charge of connecting to voice channels and streaming
// tracks to them.
type Player struct {
	Queue *Queue
	State *State

	log     zerolog.Logger
	conf    *cmd.Conf
	stop    chan bool
	encode  *dca.EncodeSession
	stream  *dca.StreamingSession
	voice   *discordgo.VoiceConnection
	session *discordgo.Session
	audio   sync.RWMutex
}

func NewPlayer(lc fx.Lifecycle, s *discordgo.Session, conf *cmd.Conf, log *zerolog.Logger) *Player {
	st := NewState()
	p := &Player{
		State:   st,
		Queue:   NewQueue(st),
		session: s,
		log:     log.With().Str("component", "player").Logger(),
		conf:    conf,
		stop:    make(chan bool),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.log.Debug().Msg("on stop called, cleaning up")
			p.Stop()
			p.Disconnect() // nolint:errcheck
			return nil
		},
	})

	return p
}

// Disconnect will disconnect the player from the currently connected voice
// channel if any.
func (p *Player) Disconnect() error {
	if p.voice != nil {
		if err := p.voice.Disconnect(); err != nil {
			return fmt.Errorf("disconnect voice channel: %w", err)
		}
		p.voice = nil
	} else {
		return fmt.Errorf("disconnect voice channel: no voice connection active")
	}
	return nil
}

// Connect will connect the player to the voice channel.
func (p *Player) Connect() error {
	if p.session != nil {
		voice, err := p.session.ChannelVoiceJoin(p.conf.Bot.Guild, p.conf.Bot.Channels.Voice, false, true)
		if err != nil {
			return fmt.Errorf("unable to establish connection to vocal channel: %w", err)
		}
		p.voice = voice
	} else {
		return fmt.Errorf("unable to connect to vocal channel: no discordgo session active")
	}
	return nil
}
