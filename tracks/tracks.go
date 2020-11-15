package tracks

import "github.com/bwmarrin/discordgo"

type Track interface {
	StreamURL() (string, error)
	Duration() int
	Embed(duration bool) *discordgo.MessageEmbed
	MarkdownLink() string
	ListenStatus() string
	GetUser() (string, string)
}

type Tracks []Track
