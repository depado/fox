package tracks

import "github.com/bwmarrin/discordgo"

type Track interface {
	StreamURL() (string, error)
	Duration() int
	Embed() *discordgo.MessageEmbed
	MarkdownLink() string
	ListenStatus() string
}

type Tracks []Track
