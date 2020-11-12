package player

import "fmt"

// Stop will immediately stop the player
func (p *Player) Stop() {
	if !p.State.Playing {
		return
	}
	p.State.Stopped = true
	p.stop <- true
}

// Pause will pause an ongoing stream
func (p *Player) Pause() {
	if p.stream == nil || !p.State.Playing {
		return
	}
	p.State.Paused = true
	p.stream.SetPaused(true)
}

// Skip will skip the currently playing track
func (p *Player) Skip() {
	if !p.State.Playing {
		return
	}
	p.stop <- true
}

// Resume will resume a paused stream
func (p *Player) Resume() {
	if p.stream == nil || !p.State.Paused {
		return
	}
	p.State.Paused = false
	p.stream.SetPaused(false)
}

// SetVolume will set the volume for the next track
func (p *Player) SetVolume(v int) error {
	// Invalid values
	if v < 0 || v > 512 {
		return fmt.Errorf("invalid volume value")
	}
	p.State.Volume = v
	return nil
}

func (p *Player) SetVolumePercent(v int) error {
	if v < 0 || v > 200 {
		return fmt.Errorf("invalid volume percentage")
	}
	vol := 256 * v / 100
	if err := p.SetVolume(vol); err != nil {
		return fmt.Errorf("unable to set volume: %w", err)
	}
	return nil
}
