package player

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jonas747/dca"
)

// Play will start to play the current queue
func (p *Player) Play() {
	if p.Playing() {
		return
	}
	go func() {
		p.audio.Lock()
		defer p.audio.Unlock()

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

			if err := p.Read(stream); err != nil {
				p.log.Err(err).Msg("unable to read stream")
			}

			if p.Stopped() {
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
	p.Stats = &Stats{}
	p.state.Lock()
	defer p.state.Unlock()
	p.state.Playing = true
	p.state.Stopped = false
	return nil
}

func (p *Player) onReadEnd() {
	if p.voice != nil {
		if err := p.voice.Speaking(false); err != nil {
			p.log.Err(err).Msg("unable to set speaking to false")
		}
	}
	p.Stats = nil
	p.state.Lock()
	defer p.state.Unlock()
	p.state.Playing = false
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
	opts.Volume = p.Volume()

	p.encode, err = dca.EncodeFile(url, opts)
	if err != nil {
		return fmt.Errorf("failed creating the encoding session: %w", err)
	}
	defer p.encode.Cleanup()

	done := make(chan error)
	p.stream = dca.NewStream(p.encode, p.voice, done)
	tc := time.NewTicker(5 * time.Second)
	defer tc.Stop()

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
		case <-tc.C:
			d := p.stream.PlaybackPosition()
			s := p.encode.Stats()
			p.Stats.Lock()
			p.Stats.PlaybackPosition = d
			p.Stats.Bitrate = s.Bitrate
			p.Stats.Duration = s.Duration
			p.Stats.Speed = s.Speed
			p.Stats.Size = s.Size
			p.Stats.TimeAxis = append(p.Stats.TimeAxis, float64(s.Duration))
			p.Stats.BiteRateAxis = append(p.Stats.BiteRateAxis, float64(s.Bitrate))
			p.Stats.Unlock()
		}
	}
}
