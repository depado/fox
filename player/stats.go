package player

import "time"

type Stats struct {
	PlaybackPosition time.Duration
	Duration         time.Duration
	Bitrate          float32
	Size             int
	Speed            float32
}

func (p *Player) Stats() *Stats {
	if p.stream == nil || p.encode == nil || !p.State.Playing {
		return nil
	}
	d := p.stream.PlaybackPosition()
	s := p.encode.Stats()
	return &Stats{
		PlaybackPosition: d,
		Bitrate:          s.Bitrate,
		Duration:         s.Duration,
		Size:             s.Size,
		Speed:            s.Speed,
	}
}
