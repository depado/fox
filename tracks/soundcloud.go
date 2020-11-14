package tracks

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Depado/soundcloud"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
)

type SoundcloudTrack struct {
	Track        soundcloud.Track
	TrackService soundcloud.TrackService
}

func (t SoundcloudTrack) ListenStatus() string {
	return t.Track.Title + " - " + t.Track.User.Username
}

func (t SoundcloudTrack) MarkdownLink() string {
	return fmt.Sprintf("[%s - %s](%s)\n", t.Track.Title, t.Track.User.Username, t.Track.PermalinkURL)
}

func (t SoundcloudTrack) Duration() int {
	return t.Track.Duration
}

// GetStreamURL will cycle through the known types of SoundCloud streams and
// return the first successful URL
func (t SoundcloudTrack) StreamURL() (string, error) {
	var url string
	var err error

	knowntypes := []soundcloud.StreamType{soundcloud.Opus, soundcloud.HLSMP3, soundcloud.ProgressiveMP3}
	ts, _, _ := t.TrackService.FromTrack(&t.Track, false)

	for _, st := range knowntypes {
		if url, err = ts.Stream(st); err == nil {
			return url, nil
		}
	}

	return url, err
}

func (t SoundcloudTrack) Embed() *discordgo.MessageEmbed {
	e := &discordgo.MessageEmbed{
		Title: t.Track.Title,
		URL:   t.Track.PermalinkURL,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: t.Track.User.AvatarURL,
			Name:    t.Track.User.Username,
			URL:     t.Track.User.PermalinkURL,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: t.Track.ArtworkURL},
		Color:     0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Plays", Value: strconv.Itoa(t.Track.PlaybackCount), Inline: true},
			{Name: "Likes", Value: strconv.Itoa(t.Track.LikesCount), Inline: true},
			{Name: "Reposts", Value: strconv.Itoa(t.Track.RepostsCount), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(t.Track.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
	}

	if t.Track.Playlist != nil {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name: "In Playlist",
			Value: fmt.Sprintf(
				"[%s](%s) by [%s](%s)",
				t.Track.Playlist.Title,
				t.Track.Playlist.PermalinkURL,
				t.Track.Playlist.User.Username,
				t.Track.Playlist.User.PermalinkURL,
			),
			Inline: false,
		})
	}
	return e
}
