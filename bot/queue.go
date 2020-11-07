package bot

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Depado/soundcloud"
	"github.com/hako/durafmt"
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
	pause   bool
}

// QueueDuration will return the total duration of the active queue
// This will effectively lock the tracks mutex until done
func (p *Player) QueueDuration() int {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()

	var tot int
	for _, t := range p.tracks {
		tot += t.Duration
	}
	return tot
}

// QueueDurationString will return the total duration of the active queue in
// human readable format
func (p *Player) QueueDurationString() string {
	return durafmt.Parse(time.Duration(p.QueueDuration()) * time.Millisecond).LimitFirstN(2).String()
}

// QueueSize will return the current number of tracks in queue
func (p *Player) QueueSize() int {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	return len(p.tracks)
}

func (p *Player) Next(tr ...soundcloud.Track) {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	if p.playing && len(p.tracks) != 0 {
		tr = append(soundcloud.Tracks{p.tracks[0]}, tr...)
		p.tracks = append(tr, p.tracks[1:]...)
	} else {
		p.tracks = append(tr, p.tracks...)
	}
}

func (p *Player) Append(tr ...soundcloud.Track) {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	p.tracks = append(p.tracks, tr...)
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
	if len(p.tracks) > 1 {
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

func (p *Player) Clear() {
	p.tracksM.Lock()
	defer p.tracksM.Unlock()
	if len(p.tracks) == 0 {
		return
	}
	if p.playing {
		p.tracks = soundcloud.Tracks{p.tracks[0]}
	} else {
		p.tracks = soundcloud.Tracks{}
	}
}
