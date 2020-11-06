package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AddLoggerFlags adds support to configure the level of the logger.
func AddLoggerFlags(c *cobra.Command) {
	c.PersistentFlags().String("log.level", "info", "one of debug, info, warn, error or fatal")
	c.PersistentFlags().String("log.type", "console", `one of "console" or "json"`)
	c.PersistentFlags().Bool("log.caller", true, "display the file and line where the call was made")
}

func AddBotFlags(c *cobra.Command) {
	c.PersistentFlags().String("bot.prefix", "!fox", "prefix to call the bot")
	c.PersistentFlags().String("bot.guild", "", "guild (server) the bot is connected to")
	c.PersistentFlags().String("bot.channels.public", "", "public channel ID where basic commands can be issued")
	c.PersistentFlags().String("bot.channels.control", "", "private channel ID (set your roles) where the bot can be controled")
	c.PersistentFlags().String("bot.channels.voice", "", "voice channel ID the bot will stream to")
	c.PersistentFlags().String("bot.token", "", "private bot token")
}

func AddDatabaseFlags(c *cobra.Command) {
	c.PersistentFlags().String("database.path", "monit.db", "path to the database file to use")
}

// AddConfigurationFlag adds support to provide a configuration file on the
// command line.
func AddConfigurationFlag(c *cobra.Command) {
	c.PersistentFlags().String("conf", "", "configuration file to use")
}

// AddAllFlags will add all the flags provided in this package to the provided
// command and will bind those flags with viper.
func AddAllFlags(c *cobra.Command) {
	AddConfigurationFlag(c)
	AddBotFlags(c)
	AddLoggerFlags(c)
	AddDatabaseFlags(c)

	if err := viper.BindPFlags(c.PersistentFlags()); err != nil {
		log.Fatal().Err(err).Msg("couldn't bind flags")
	}
}
