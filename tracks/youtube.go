package tracks

import "github.com/bwmarrin/discordgo"

type YoutubeTrack struct {
}

func (yt YoutubeTrack) StreamURL() (string, error) {
	panic("not implemented")
}

func (yt YoutubeTrack) Duration() int {
	panic("not implemented")
}

func (yt YoutubeTrack) Embed() *discordgo.MessageEmbed {
	panic("not implemented")
}

func (yt YoutubeTrack) MarkdownLink() string {
	panic("not implemented")
}

func (yt YoutubeTrack) ListenStatus() string {
	panic("not implemented")
}
