package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"github.com/Depado/fox/acl"
	"github.com/Depado/fox/message"
	"github.com/Depado/fox/player"
	"github.com/Depado/fox/storage"
)

type fav struct {
	BaseCommand
	Storage *storage.StormDB
}

func (c *fav) Handler(s *discordgo.Session, m *discordgo.Message, args []string) {
	if len(args) == 0 {
		p := c.Players.GetPlayer(m.GuildID)
		if p == nil {
			c.log.Error().Msg("no player associated to guild ID")
			return
		}
		if !p.Playing() {
			message.SendShortTimedNotice(s, m, "No track is currently playing", c.log)
			return
		}
		// Save current track
	}

	if len(args) > 0 {
		switch args[0] {
		case "show", "s":
			// Show favlist
		case "clear", "c":
			// Clear favlist
		default:

		}
	}
}

func NewFavCommand(p *player.Players, log zerolog.Logger, storage *storage.StormDB) Command {
	cmd := "fav"
	return &fav{
		BaseCommand: BaseCommand{
			ChannelRestriction: acl.Anywhere,
			RoleRestriction:    acl.Anyone,
			Options: Options{
				ArgsRequired:      false,
				DeleteUserMessage: true,
			},
			Long:    cmd,
			Aliases: []string{"f"},
			SubCommands: []SubCommand{
				{Long: "show", Aliases: []string{"s"}, Description: "Send your fav list in DM"},
				{Long: "clear", Aliases: []string{"c"}, Description: "Clear your fav list"},
			},
			Help: Help{
				Usage:     cmd,
				ShortDesc: "Display or modify your favorite sound list",
				Description: "This command allows you to save the currently " +
					"playing track to your own favorite list. You can then access " +
					"your list using the `show/s` subcommand, or clear it up using " +
					"the `clear/c` subcommand.",
				Examples: []Example{
					{Command: "fav", Explanation: "Add the currently playing track to your fav list"},
					{Command: "fav show", Explanation: "Send your fav list in DM"},
					{Command: "fav clear", Explanation: "Clear up your fav list"},
				},
			},
			Players: p,
			log:     log.With().Str("command", cmd).Logger(),
		},
		Storage: storage,
	}
}
