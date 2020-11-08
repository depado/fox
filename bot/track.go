package bot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
)

func (b *BotInstance) AddToQueue(m *discordgo.MessageCreate, url string, next bool) {
	pls, err := b.Soundcloud.Playlist().FromURL(url)
	if err == nil {
		pl, err := pls.Get()
		if err != nil {
			b.log.Err(err).Msg("unable to get playlist details")
			return
		}

		if next {
			b.Player.Next(pl.Tracks...)
		} else {
			b.Player.Append(pl.Tracks...)
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
		}

		if next {
			e.Description = fmt.Sprintf("Added **%d** tracks to start of queue", len(pl.Tracks))
		} else {
			e.Description = fmt.Sprintf("Added **%d** tracks to end of queue", len(pl.Tracks))
		}

		if _, err = b.Session.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			b.log.Err(err).Msg("unable to send embed")
		}
		return
	}
	_, t, err := b.Soundcloud.Track().FromURL(url)
	if err == nil {
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
				{Name: "Added by", Value: fmt.Sprintf("<@%s>", m.Author.ID), Inline: false},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: t.ArtworkURL},
		}
		if next {
			b.Player.Next(*t)
			e.Description = "**Track added to start of queue**"
		} else {
			b.Player.Append(*t)
			e.Description = "**Track added to end of queue**"
		}
		if _, err = b.Session.ChannelMessageSendEmbed(m.ChannelID, e); err != nil {
			b.log.Err(err).Msg("unable to send embed")
		}
		return
	}
	b.SendNotice("", "This is not a track or a playlist", "", m.ChannelID)
}
