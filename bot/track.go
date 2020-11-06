package bot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Depado/soundcloud"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog/log"
)

func handlePlaylist(s *discordgo.Session, m *discordgo.MessageCreate, pls *soundcloud.PlaylistService) {
	pl, err := pls.Get()
	if err != nil {
		log.Err(err).Msg("unable to fetch playlist info")
		return
	}

	msg := "```"
	for _, t := range pl.Tracks {
		msg += fmt.Sprintf("\n!add <%s>", t.PermalinkURL)
	}
	msg += "\n```"
	s.ChannelMessageSend(m.ChannelID, msg)
}

func (b *BotInstance) handleTrack(s *discordgo.Session, m *discordgo.MessageCreate, ts *soundcloud.TrackService, t *soundcloud.Track) {
	e := &discordgo.MessageEmbed{
		Title: t.Title,
		URL:   t.PermalinkURL,
		Image: &discordgo.MessageEmbedImage{URL: t.ArtworkURL},
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: t.User.AvatarURL,
			Name:    t.User.Username,
			URL:     t.User.PermalinkURL,
		},
		Description: t.Description,
		Color:       0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Plays", Value: strconv.Itoa(t.PlaybackCount), Inline: true},
			{Name: "Likes", Value: strconv.Itoa(t.LikesCount), Inline: true},
			{Name: "Reposts", Value: strconv.Itoa(t.RepostsCount), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(t.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: t.ReleaseDate.String(),
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		log.Err(err).Msg("unable to send embed")
	}

	// s.MessageReactionAdd(m.ChannelID, msg.ID, "⏏️")
	// s.MessageReactionAdd(m.ChannelID, msg.ID, "▶️")
}

func (b *BotInstance) handleURL(s *discordgo.Session, m *discordgo.MessageCreate, url string) {
	pls, err := b.Soundcloud.Playlist().FromURL(url)
	if err == nil {
		handlePlaylist(s, m, pls)
		return
	}
	ts, t, err := b.Soundcloud.Track().FromURL(url)
	if err == nil {
		b.handleTrack(s, m, ts, t)
		return
	}
	if _, err := s.ChannelMessageSend(m.ChannelID, "This is not a track or a playlist"); err != nil {
		log.Err(err).Msg("unable to send usage message")
	}
}

func (b *BotInstance) AddToQueue(m *discordgo.MessageCreate, url string) {
	pls, err := b.Soundcloud.Playlist().FromURL(url)
	if err == nil {
		pl, err := pls.Get()
		if err != nil {
			b.log.Err(err).Msg("unable to get playlist details")
			return
		}
		for _, t := range pl.Tracks {
			b.Player.Add(t)
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
				{Name: "Added by", Value: fmt.Sprintf("<@%s>", m.Author.ID), Inline: false},
				{Name: "Tracks", Value: strconv.Itoa(len(pl.Tracks)), Inline: true},
				{Name: "Duration", Value: durafmt.Parse(time.Duration(pl.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
			},
			Description: fmt.Sprintf("Added **%d** tracks to queue", len(pl.Tracks)),
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: pl.ArtworkURL},
		}

		if _, err = b.Session.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			log.Err(err).Msg("unable to send embed")
		}
		return
	}
	_, t, err := b.Soundcloud.Track().FromURL(url)
	if err == nil {
		b.Player.Add(*t)
		b.SendNotice("", "Added one track to queue", "", m.ChannelID)
		return
	}
	b.SendNotice("", "This is not a track or a playlist", "", m.ChannelID)
}
