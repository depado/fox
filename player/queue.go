package player

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"

	"github.com/Depado/fox/tracks"
)

type Queue struct {
	sync.RWMutex

	tracks tracks.Tracks
	state  *State
}

func NewQueue(s *State) *Queue {
	return &Queue{state: s}
}

// Duration will return the total duration of the active queue.
func (q *Queue) Duration() int {
	q.Lock()
	defer q.Unlock()

	var tot int
	for _, t := range q.tracks {
		tot += t.Duration()
	}
	return tot
}

// DurationString will return the total duration of the active queue in human
// readable format.
func (q *Queue) DurationString() string {
	return durafmt.Parse(time.Duration(q.Duration()) * time.Millisecond).LimitFirstN(2).String()
}

// Len will return the current number of tracks in queue.
func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()

	return len(q.tracks)
}

// Prepend will add tracks to the start of the queue, either right after the
// currently playing track, or right at the start if the player is stopped.
func (q *Queue) Prepend(t ...tracks.Track) {
	q.Lock()
	defer q.Unlock()

	if q.state.Playing && len(q.tracks) != 0 {
		tr := append(tracks.Tracks{q.tracks[0]}, t...)
		q.tracks = append(tr, q.tracks[1:]...)
	} else {
		q.tracks = append(t, q.tracks...)
	}
}

// Append will append tracks at the end of queue.
func (q *Queue) Append(t ...tracks.Track) {
	q.Lock()
	defer q.Unlock()

	q.tracks = append(q.tracks, t...)
}

// Pop will remove the first track in queue.
func (q *Queue) Pop() {
	q.Lock()
	defer q.Unlock()

	if len(q.tracks) != 0 {
		q.tracks = q.tracks[1:]
	}
}

// Loop will move the first track at the end of the queue if there is more than
// one track in queue. Otherwise it does nothing, leaving the first track in its
// position to be played once more.
func (q *Queue) Loop() {
	q.Lock()
	defer q.Unlock()

	if len(q.tracks) > 1 {
		t := q.tracks[0]
		q.tracks = q.tracks[1:]
		q.tracks = append(q.tracks, t)
	}
}

// Get will return the first track in queue if any. If there is no track in
// queue, nil will be returned.
func (q *Queue) Get() tracks.Track {
	q.Lock()
	defer q.Unlock()

	if len(q.tracks) != 0 {
		return q.tracks[0]
	}
	return nil
}

// Shuffle will shuffle all the tracks in queue, except the first one if it's
// currently being played.
func (q *Queue) Shuffle() {
	q.Lock()
	defer q.Unlock()

	// Either there is one track or none, do nothing
	if len(q.tracks) < 2 {
		return
	}

	// If the first track is currently being played, do not shuffle it
	if q.state.Playing {
		t := q.tracks[0]
		ts := q.tracks[1:]
		rand.Shuffle(len(ts), func(i, j int) { ts[i], ts[j] = ts[j], ts[i] })
		q.tracks = append(tracks.Tracks{t}, ts...)
	} else {
		rand.Shuffle(len(q.tracks), func(i, j int) { q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i] })
	}
}

// Clear will reset the queue, removing all tracks from it except the first one
// if it is currently played.
func (q *Queue) Clear() {
	q.Lock()
	defer q.Unlock()

	if len(q.tracks) == 0 {
		return
	}
	if q.state.Playing {
		q.tracks = tracks.Tracks{q.tracks[0]}
	} else {
		q.tracks = tracks.Tracks{}
	}
}

// RemoveN will remove the next n tracks in queue
func (q *Queue) RemoveN(n int) {
	q.Lock()
	defer q.Unlock()

	if n >= len(q.tracks) || (q.state.Playing && n+1 >= len(q.tracks)) {
		q.Clear()
		return
	}
	if q.state.Playing {
		q.tracks = append(tracks.Tracks{q.tracks[0]}, q.tracks[n+1:]...)
	} else {
		q.tracks = q.tracks[n:]
	}
}

func (q *Queue) GenerateQueueEmbed() *discordgo.MessageEmbed {
	q.Lock()
	defer q.Unlock()

	var body string
	var tot int
	if len(q.tracks) > 0 {
		for i, t := range q.tracks {
			if i <= 10 {
				body += t.MarkdownLink()
			}
			tot += t.Duration()
		}
		if len(q.tracks) > 10 {
			body += fmt.Sprintf("\nAnd **%d** other tracks", len(q.tracks)-10)
		}
	} else {
		body = "There is currently no track in queue"
	}

	e := &discordgo.MessageEmbed{
		Title:       "Current Queue",
		Description: body,
		Color:       0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Tracks", Value: strconv.Itoa(len(q.tracks)), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(tot) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
	}

	return e
}
