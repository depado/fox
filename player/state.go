package player

import (
	"fmt"
	"sync"
)

// State stores the various state of the player. It is also used by the queue
// to determine some actions related to currently playing tracks.
type State struct {
	sync.RWMutex
	Playing bool
	Stopped bool
	Paused  bool
	Volume  int
}

// NewPlayerState will return a new player state
func NewState() *State {
	return &State{
		Volume: 256, // 256 is the normal volume
	}
}

func (p *Player) Playing() bool {
	p.state.RLock()
	defer p.state.RUnlock()
	return p.state.Playing
}

func (p *Player) Stopped() bool {
	p.state.RLock()
	defer p.state.RUnlock()
	return p.state.Stopped
}

func (p *Player) Paused() bool {
	p.state.RLock()
	defer p.state.RUnlock()
	return p.state.Paused
}

func (p *Player) Volume() int {
	p.state.RLock()
	defer p.state.RUnlock()
	return p.state.Volume
}

// Stop will immediately stop the player
func (p *Player) Stop() {
	p.state.Lock()
	defer p.state.Unlock()

	if !p.state.Playing {
		return
	}
	p.state.Stopped = true
	p.stop <- true
}

// Pause will pause an ongoing stream
func (p *Player) Pause() {
	p.state.Lock()
	defer p.state.Unlock()

	if p.stream == nil || !p.state.Playing {
		return
	}
	p.state.Paused = true
	p.stream.SetPaused(true)
}

// Skip will skip the currently playing track
func (p *Player) Skip() {
	p.state.Lock()
	defer p.state.Unlock()

	if !p.state.Playing {
		return
	}
	p.stop <- true
}

// Resume will resume a paused stream
func (p *Player) Resume() {
	p.state.Lock()
	defer p.state.Unlock()

	if p.stream == nil || !p.state.Paused {
		return
	}
	p.state.Paused = false
	p.stream.SetPaused(false)
}

// SetVolume will set the volume for the next track
func (p *Player) SetVolume(v int) error {
	p.state.Lock()
	defer p.state.Unlock()

	// Invalid values
	if v < 0 || v > 512 {
		return fmt.Errorf("invalid volume value")
	}
	p.state.Volume = v
	return nil
}

func (p *Player) SetVolumePercent(v int) error {
	p.state.Lock()
	defer p.state.Unlock()

	if v < 0 || v > 200 {
		return fmt.Errorf("invalid volume percentage")
	}
	vol := 256 * v / 100
	if err := p.SetVolume(vol); err != nil {
		return fmt.Errorf("unable to set volume: %w", err)
	}
	return nil
}
