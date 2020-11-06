package bot

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Depado/fox/cmd"
	"github.com/Depado/soundcloud"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type BotInstance struct {
	conf       *cmd.Conf
	log        *zerolog.Logger
	Session    *discordgo.Session
	Soundcloud *soundcloud.Client
	Voice      *discordgo.VoiceConnection
	Player     *Player
}

func NewBotInstance(lc fx.Lifecycle, c *cmd.Conf, log *zerolog.Logger, sc *soundcloud.Client) *BotInstance {
	rand.Seed(time.Now().UnixNano())
	dg, err := discordgo.New("Bot " + c.Bot.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to init")
	}
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
	if err := dg.Open(); err != nil {
		log.Fatal().Err(err).Msg("unable to open")
	}

	voice, err := dg.ChannelVoiceJoin(c.Bot.Guild, c.Bot.Channels.Voice, false, true)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to initiate voice connection")
	}

	b := &BotInstance{
		conf:       c,
		log:        log,
		Soundcloud: sc,
		Session:    dg,
		Voice:      voice,
		Player:     &Player{tracks: soundcloud.Tracks{}},
	}
	b.Session.AddHandler(b.MessageCreated)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			b.log.Debug().Msg("cleanup")
			if b.Player.playing {
				b.Player.stop = true
				b.Player.session.Stop() // nolint:errcheck
				b.Player.session.Cleanup()
			}
			b.Voice.Close()
			b.Session.Close()
			return nil
		},
	})

	return b
}

// Start is a simple function invoked by fx to bootstrap the dependecy chain
func Start(l *zerolog.Logger, b *BotInstance) {
	l.Info().Msg("Bot is now running")
}

// SalvageVoice will attempt to disconnect and reconnect to the vocal channel
func (b *BotInstance) SalvageVoice() error {
	if b.Voice != nil {
		if err := b.Voice.Disconnect(); err != nil {
			return fmt.Errorf("unable to disonncet vocal channel: %w", err)
		}
	}

	voice, err := b.Session.ChannelVoiceJoin(b.conf.Bot.Guild, b.conf.Bot.Channels.Voice, false, true)
	if err != nil {
		return fmt.Errorf("unable to establish connection to vocal channel: %w", err)
	}
	b.Voice = voice
	return nil
}
