package bot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Depado/soundcloud"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
)

func (b *BotInstance) TrackEmbed(t soundcloud.Track, queue bool) *discordgo.MessageEmbed {
	e := &discordgo.MessageEmbed{
		Title: t.Title,
		URL:   t.PermalinkURL,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: t.User.AvatarURL,
			Name:    t.User.Username,
			URL:     t.User.PermalinkURL,
		},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: t.ArtworkURL},
		Description: "**Now Playing**",
		Color:       0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Plays", Value: strconv.Itoa(t.PlaybackCount), Inline: true},
			{Name: "Likes", Value: strconv.Itoa(t.LikesCount), Inline: true},
			{Name: "Reposts", Value: strconv.Itoa(t.RepostsCount), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(t.Duration) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
		},
	}

	if t.Playlist != nil {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:   "In Playlist",
			Value:  fmt.Sprintf("[%s](%s) by [%s](%s)", t.Playlist.Title, t.Playlist.PermalinkURL, t.Playlist.User.Username, t.Playlist.User.PermalinkURL),
			Inline: false,
		})
	}

	if queue {
		e.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf(
				"%d tracks left in queue - %s",
				b.Player.QueueSize(), b.Player.QueueDurationString(),
			),
		}
	}
	return e
}

// SendNowPlaying will send an embed in the public channel displaying the
// details of the track being currently played
func (b *BotInstance) SendNowPlaying(t soundcloud.Track) {
	_, err := b.Session.ChannelMessageSendEmbed(b.conf.Bot.Channels.Public, b.TrackEmbed(t, true))
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
}

func (b *BotInstance) DisplayQueue(m *discordgo.MessageCreate) {
	b.Player.tracksM.Lock()
	defer b.Player.tracksM.Unlock()

	var body string
	var tot int
	if len(b.Player.tracks) > 0 {
		for i, t := range b.Player.tracks {
			if i <= 10 {
				body += fmt.Sprintf("[%s - %s](%s)\n", t.Title, t.User.Username, t.PermalinkURL)
			}
			tot += t.Duration
		}
		if len(b.Player.tracks) > 10 {
			body += fmt.Sprintf("\nAnd **%d** other tracks", len(b.Player.tracks)-10)
		}
	} else {
		body = "There is currently no track in queue"
	}

	e := &discordgo.MessageEmbed{
		Title:       "Current Queue",
		Description: body,
		Color:       0xff5500,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Tracks", Value: strconv.Itoa(len(b.Player.tracks)), Inline: true},
			{Name: "Duration", Value: durafmt.Parse(time.Duration(tot) * time.Millisecond).LimitFirstN(2).String(), Inline: true},
			{Name: "Requested by", Value: fmt.Sprintf("<@%s>", m.Author.ID), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Tip: Add new tracks using '%s add' or '%s next'", b.conf.Bot.Prefix, b.conf.Bot.Prefix),
		},
	}
	_, err := b.Session.ChannelMessageSendEmbed(m.ChannelID, e)
	if err != nil {
		b.log.Err(err).Msg("unable to send embed")
	}
}
