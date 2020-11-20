package player

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rs/zerolog"

	"github.com/Depado/fox/cmd"
	"github.com/Depado/fox/models"
	"github.com/Depado/fox/storage"
)

// Players will hold the guild players
type Players struct {
	sync.RWMutex
	Players map[string]*Player
}

// GetPlayer will get the player associated with the guild ID if any
func (p *Players) GetPlayer(guildID string) *Player {
	p.Lock()
	defer p.Unlock()

	pl, ok := p.Players[guildID]
	if !ok {
		return nil
	}
	return pl
}

// Create will create a new player and associate it with the guild
func (p *Players) Create(s *discordgo.Session, conf *cmd.Conf, log zerolog.Logger, guild string, storage *storage.StormDB, gc *models.Conf) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Players[guild]; ok {
		return fmt.Errorf("guild already has an associated player")
	}
	play, err := NewPlayer(s, conf, log.With().Str("guild", guild).Logger(), guild, storage, gc)
	if err != nil {
		return fmt.Errorf("create player: %w", err)
	}
	p.Players[guild] = play
	return nil
}

// Kill will kill all the players
func (p *Players) Kill() {
	p.Lock()
	defer p.Unlock()

	for _, pl := range p.Players {
		pl.Kill()
	}
}

// NewPlayers instantiates the player map
func NewPlayers() *Players {
	return &Players{
		Players: make(map[string]*Player),
	}
}

type PlaySession struct {
}

// Player is the struct in charge of connecting to voice channels and streaming
// tracks to them.
type Player struct {
	Queue   *Queue
	state   *State
	Guild   string
	Conf    *models.Conf
	Storage *storage.StormDB

	log     zerolog.Logger
	conf    *cmd.Conf
	stop    chan bool
	encode  *dca.EncodeSession
	stream  *dca.StreamingSession
	voice   *discordgo.VoiceConnection
	session *discordgo.Session
	audio   sync.RWMutex
	Stats   *Stats
}

// NewPlayer will create a new player from scratch using the provided
// arguments
func NewPlayer(s *discordgo.Session, conf *cmd.Conf, log zerolog.Logger, guildID string, storage *storage.StormDB, gc *models.Conf) (*Player, error) {
	st := NewState()
	p := &Player{
		state:   st,
		Queue:   NewQueue(st),
		Storage: storage,
		Guild:   guildID,
		Conf:    gc,
		session: s,
		log:     log.With().Str("component", "player").Logger(),
		conf:    conf,
		stop:    make(chan bool),
	}

	return p, nil
}

func (p *Player) Kill() {
	p.log.Debug().Msg("kill called")
	p.Stop()
	p.Disconnect() // nolint:errcheck
}

// Disconnect will disconnect the player from the currently connected voice
// channel if any.
func (p *Player) Disconnect() error {
	if p.voice != nil {
		if err := p.voice.Disconnect(); err != nil {
			return fmt.Errorf("disconnect voice channel: %w", err)
		}
		p.voice = nil
	}
	return nil
}

// Connect will connect the player to the voice channel.
func (p *Player) Connect() error {
	if p.session != nil {
		voice, err := p.session.ChannelVoiceJoin(p.Conf.ID, p.Conf.VoiceChannel, false, true)
		if err != nil {
			return fmt.Errorf("unable to establish connection to vocal channel: %w", err)
		}
		p.voice = voice
	} else {
		return fmt.Errorf("unable to connect to vocal channel: no discordgo session active")
	}
	return nil
}
