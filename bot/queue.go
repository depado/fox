package bot

import (
	"math/rand"
	"sync"

	"github.com/Depado/soundcloud"
	"github.com/jonas747/dca"
)

type Player struct {
	tracks  soundcloud.Tracks
	tracksM sync.RWMutex
	audioM  sync.RWMutex
	stream  *dca.StreamingSession
	session *dca.EncodeSession
	playing bool
	stop    bool
	// loop    bool
}

func (p *Player) Add(t soundcloud.Track) {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	p.tracks = append(p.tracks, t)
}

func (p *Player) Pop() {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	if len(p.tracks) != 0 {
		p.tracks = p.tracks[1:]
	}
}

func (p *Player) Loop() {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	if len(p.tracks) != 0 {
		t := p.tracks[0]
		p.tracks = p.tracks[1:]
		p.tracks = append(p.tracks, t)
	}
}

func (p *Player) Get() *soundcloud.Track {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	if len(p.tracks) != 0 {
		return &p.tracks[0]
	}
	return nil
}

func (p *Player) Shuffle() {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	if len(p.tracks) < 2 {
		return
	}
	t := p.tracks[0]
	ts := p.tracks[1:]
	rand.Shuffle(len(ts), func(i, j int) { ts[i], ts[j] = ts[j], ts[i] })
	p.tracks = append(soundcloud.Tracks{t}, ts...)
}
