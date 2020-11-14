package player

import (
	"errors"
	"fmt"
	"io"

	"github.com/jonas747/dca"
)

// Play will start to play the current queue
func (p *Player) Play() {
	if p.State.Playing {
		return
	}
	go func() {
		p.audio.Lock()
		defer func() {
			p.audio.Unlock()
			if err := p.session.UpdateListeningStatus(""); err != nil {
				p.log.Err(err).Msg("unable to update listening status")
			}
		}()
		for {
			tracklen := p.Queue.Len()
			if tracklen == 0 {
				p.SendNotice("Nothing left to play!", fmt.Sprintf("You can give me more by using the `%s` command!", p.conf.Bot.Prefix), "")
				if err := p.Disconnect(); err != nil {
					p.log.Err(err).Msg("unable to disconnect from voice channel")
				}
				return
			}

			t := p.Queue.Get()
			if t == nil {
				continue
			}

			stream, err := t.StreamURL()
			if err != nil {
				p.log.Err(err).Msg("unable to find a suitable stream")
				p.Queue.Pop()
				continue
			}
			if err := p.session.UpdateListeningStatus(t.ListenStatus()); err != nil {
				p.log.Err(err).Msg("unable to update listening status")
			}

			if err := p.Read(stream); err != nil {
				p.log.Err(err).Msg("unable to read stream")
			}

			if p.State.Stopped {
				if err := p.Disconnect(); err != nil {
					p.log.Err(err).Msg("unable to disconnect from voice channel")
				}
				return
			}
			p.Queue.Pop()
		}
	}()
}

func (p *Player) onReadStart() error {
	if p.voice == nil {
		if err := p.Connect(); err != nil {
			return fmt.Errorf("unable to connect: %w", err)
		}
	}
	if err := p.voice.Speaking(true); err != nil {
		return fmt.Errorf("failed setting voice to speaking: %w", err)
	}
	p.State.Playing = true
	p.State.Stopped = false
	return nil
}

func (p *Player) onReadEnd() {
	if p.voice != nil {
		if err := p.voice.Speaking(false); err != nil {
			p.log.Err(err).Msg("unable to set speaking to false")
		}
	}
	p.State.Playing = false
	p.stream = nil
	p.encode = nil
}

func (p *Player) Read(url string) error {
	var err error

	if err = p.onReadStart(); err != nil {
		return fmt.Errorf("unable to start playing: %w", err)
	}
	defer p.onReadEnd()

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120
	opts.Volume = p.State.Volume

	p.encode, err = dca.EncodeFile(url, opts)
	if err != nil {
		return fmt.Errorf("failed creating the encoding session: %w", err)
	}
	defer p.encode.Cleanup()

	done := make(chan error)
	p.stream = dca.NewStream(p.encode, p.voice, done)

	for {
		select {
		case err := <-done:
			if err != nil && err == io.EOF {
				return nil
			}
			if errors.Is(err, dca.ErrVoiceConnClosed) {
				p.log.Info().Msg("voice connection lost")
				if err = p.Disconnect(); err != nil {
					p.log.Err(err).Msg("unable to disconnect vocal channel")
					return err
				}
				if err = p.Connect(); err != nil {
					p.log.Err(err).Msg("unable to reconnect")
					p.voice = nil
					return err
				}
				p.stream = dca.NewStream(p.encode, p.voice, done)
				p.log.Info().Msg("voice reconnected")
				continue
			}
			return fmt.Errorf("reading stream: %w", err)
		case <-p.stop:
			p.log.Debug().Str("event", "stop").Msg("stopping player")
			return nil
		}
	}
}
