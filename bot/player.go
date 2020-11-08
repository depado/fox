package bot

import (
	"errors"
	"fmt"
	"io"

	"github.com/Depado/soundcloud"
	"github.com/jonas747/dca"
)

func (b *BotInstance) PlayQueue() {
	if b.Player.playing {
		return
	}
	go func() {
		b.Player.audioM.Lock()
		defer func() {
			b.Player.audioM.Unlock()
			b.Player.playing = false
			b.Session.UpdateStatus(0, "") // nolint:errcheck
		}()
		for {
			tracklen := b.Player.QueueSize()
			b.log.Debug().Int("length", tracklen).Msg("starting playing queue")
			b.Vote.Reset()
			if tracklen == 0 {
				b.log.Debug().Int("length", tracklen).Msg("track length")
				b.SendPublicMessage("Nothing left to play!", fmt.Sprintf("You can give me more by using the %s command!", b.conf.Bot.Prefix), "")
				return
			}
			b.log.Debug().Msg("getting track")
			t := b.Player.Get()
			if t == nil {
				b.Player.Pop()
				b.log.Error().Msg("queue isn't empty but a nil track was returned")
				continue
			}
			b.log.Debug().Msg("track was found, fetching URL")
			ts, _, _ := b.Soundcloud.Track().FromTrack(t, false)
			url, err := ts.Stream(soundcloud.Opus)
			if err != nil {
				b.log.Debug().Msg("opus format not found, fallback on MP3")
				if url, err = ts.Stream(soundcloud.ProgressiveMP3); err != nil {
					b.Player.Pop()
					b.log.Error().Err(err).Msg("unable to get stream url")
					continue
				}
			}
			b.Player.playing = true
			b.Player.stop = false
			b.SendNowPlaying(*t)
			b.Session.UpdateStatus(0, fmt.Sprintf("%s - %s", t.Title, t.User.Username)) // nolint:errcheck
			if err := b.Play(url); err != nil {
				b.log.Err(err).Msg("unable to play")
				continue
			}
			b.Player.Pop()
			if b.Player.stop {
				return
			}
		}
	}()
}

func (b *BotInstance) Play(url string) error {
	if err := b.Voice.Speaking(true); err != nil {
		return fmt.Errorf("failed setting voice to speaking: %w", err)
	}
	defer b.Voice.Speaking(false) // nolint:errcheck
	b.log.Debug().Msg("voice setup done")

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120

	encodeSession, err := dca.EncodeFile(url, opts)
	if err != nil {
		return fmt.Errorf("failed creating the encoding session: %w", err)
	}
	defer encodeSession.Cleanup()
	b.Player.session = encodeSession
	b.log.Debug().Msg("encoding session setup")

	done := make(chan error)
	stream := dca.NewStream(encodeSession, b.Voice, done)
	b.Player.stream = stream

	for { // nolint:gosimple
		select {
		case err := <-done:
			if err != nil && err == io.EOF {
				return nil
			}
			if errors.Is(err, dca.ErrVoiceConnClosed) {
				b.log.Info().Err(err).Msg("voice connection closed, attempting reconnection")
				if err = b.SalvageVoice(); err != nil {
					b.log.Err(err).Msg("unable to reconnect voice")
					return err
				}
				b.log.Info().Msg("voice reconnected, recreating stream")
				stream = dca.NewStream(encodeSession, b.Voice, done)
				b.Player.stream = stream
				continue
			}
			return fmt.Errorf("reading stream: %w", err)
		}
	}
}
