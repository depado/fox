package soundcloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Depado/fox/tracks"
	"github.com/Depado/soundcloud"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog"
)

type SoundCloudProvider struct {
	client *soundcloud.Client
	log    *zerolog.Logger
}

func NewSoundCloudProvider(log *zerolog.Logger, c *soundcloud.Client) *SoundCloudProvider {
	return &SoundCloudProvider{
		client: c,
		log:    log,
	}
}

func (sc *SoundCloudProvider) GetPlaylist(url string, m *discordgo.Message) (tracks.Tracks, *discordgo.MessageEmbed, error) {
	pls, err := sc.client.Playlist().FromURL(url)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve playlist: %w", err)
	}

	pl, err := pls.Get()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve playlist data: %w", err)
	}

	tr := make(tracks.Tracks, len(pl.Tracks))
	for i, t := range pl.Tracks {
		ts, track, err := sc.client.Track().FromTrack(&t, false)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get track service from track: %w", err)
		}
		tr[i] = tracks.SoundcloudTrack{
			Track:        *track,
			TrackService: *ts,
			User:         m.Author.Username + "#" + m.Author.Discriminator,
			AvatarURL:    m.Author.AvatarURL(""),
		}
	}

	e := &discordgo.MessageEmbed{
		Title: pl.Title,
		URL:   pl.PermalinkURL,
		Color: 0xff5500,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: pl.User.AvatarURL,
			Name:    pl.User.Username,
			URL:     pl.User.PermalinkURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Added by", Value: fmt.Sprintf("<@%s>", m.Author.ID), Inline: true},
			{Name: "Tracks", Value: strconv.Itoa(len(pl.Tracks)), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(pl.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: pl.ArtworkURL},
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: m.Author.AvatarURL(""),
			Text:    "Added by " + m.Author.Username + "#" + m.Author.Discriminator,
		},
	}

	return tr, e, nil
}

func (sc *SoundCloudProvider) GetTrack(url string, m *discordgo.Message) (tracks.Track, *discordgo.MessageEmbed, error) {
	ts, t, err := sc.client.Track().FromURL(url)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get track from url: %w", err)
	}
	e := &discordgo.MessageEmbed{
		Title: t.Title,
		URL:   t.PermalinkURL,
		Color: 0xff5500,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: t.User.AvatarURL,
			Name:    t.User.Username,
			URL:     t.User.PermalinkURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Plays", Value: strconv.Itoa(t.PlaybackCount), Inline: true},
			{Name: "Likes", Value: strconv.Itoa(t.LikesCount), Inline: true},
			{Name: "Reposts", Value: strconv.Itoa(t.RepostsCount), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(t.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: t.ArtworkURL},
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: m.Author.AvatarURL(""),
			Text:    "Added by " + m.Author.Username + "#" + m.Author.Discriminator,
		},
	}

	return tracks.SoundcloudTrack{
		Track:        *t,
		TrackService: *ts,
		User:         m.Author.Username + "#" + m.Author.Discriminator,
		AvatarURL:    m.Author.AvatarURL(""),
	}, e, nil
}
