package bot

import (
	"errors"
	"fmt"
	"io"

	"github.com/Depado/soundcloud"
	"github.com/jonas747/dca"
	"github.com/rs/zerolog/log"
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
		b.log.Debug().Int("length", len(b.Player.tracks)).Msg("starting playing queue")
		for {
			if len(b.Player.tracks) == 0 {
				b.log.Debug().Int("length", len(b.Player.tracks)).Msg("track length")
				b.SendPublicMessage("Nothing left to play!", fmt.Sprintf("You can give me more by using the %s command!", b.conf.Bot.Prefix))
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
				b.Player.Pop()
				b.log.Error().Err(err).Msg("unable to get stream url")
				continue
			}
			b.Player.playing = true
			b.Player.stop = false
			b.SendNowPlaying(*t)
			b.Session.UpdateStatus(0, fmt.Sprintf("%s - %s", t.Title, t.User.Username)) // nolint:errcheck
			if err := b.Play(url); err != nil && errors.Is(err, dca.ErrVoiceConnClosed) {
				if err := b.SalvageVoice(); err != nil {
					b.SendControlMessage("Voice connection lost", "I tried fixing it by myself but it didn't work.")
					return
				}
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
	err := b.Voice.Speaking(true)
	if err != nil {
		log.Err(err).Msg("Failed setting speaking")
		return err
	}
	defer b.Voice.Speaking(false) // nolint:errcheck

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120

	encodeSession, err := dca.EncodeFile(url, opts)
	if err != nil {
		log.Err(err).Msg("failed creating an encoding session")
	}
	b.Player.session = encodeSession

	done := make(chan error)
	stream := dca.NewStream(encodeSession, b.Voice, done)
	b.Player.stream = stream

	for { // nolint:gosimple
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				log.Err(err).Msg("error occured during playback")
			} else {
				err = nil
			}
			encodeSession.Cleanup()
			return err
		}
	}
}
