package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

type LogConf struct {
	Level  string `mapstructure:"level"`
	Type   string `mapstructure:"type"`
	Caller bool   `mapstructure:"caller"`
}

type BotConf struct {
	Channels ChannelsConf `mapstructure:"channels"`
	Roles    RolesConf    `mapstructure:"roles"`
	Guild    string       `mapstrcture:"guild"`
	Token    string       `mapstructure:"token"`
	Prefix   string       `mapstructure:"prefix"`
}

type ChannelsConf struct {
	Text  string `mapstructure:"text"`
	Voice string `mapstructure:"voice"`
}

type RolesConf struct {
	Admin string `mapstructure:"admin"`
	DJ    string `mapstructure:"dj"`
}

type DatabaseConf struct {
	Path string `mapstructure:"path"`
}

type Conf struct {
	Port     int          `mapstructure:"port"`
	Log      LogConf      `mapstructure:"log"`
	Bot      BotConf      `mapstructure:"bot"`
	Database DatabaseConf `mapstructure:"database"`
}

// NewLogger will return a new logger
func NewLogger(c *Conf) *zerolog.Logger {
	// Level parsing
	warns := []string{}
	lvl, err := zerolog.ParseLevel(c.Log.Level)
	if err != nil {
		warns = append(warns, fmt.Sprintf("unrecognized log level '%s', fallback to 'info'", c.Log.Level))
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(lvl)
	}

	// Type parsing
	switch c.Log.Type {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	case "json":
		break
	default:
		warns = append(warns, fmt.Sprintf("unrecognized log type '%s', fallback to 'json'", c.Log.Type))
	}

	// Caller
	if c.Log.Caller {
		log.Logger = log.With().Caller().Logger()
	}

	// Log messages with the newly created logger
	for _, w := range warns {
		log.Warn().Msg(w)
	}

	return &log.Logger
}

// NewConf will parse and return the configuration
func NewConf() (*Conf, error) {
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("fox")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Configuration file
	if viper.GetString("conf") != "" {
		viper.SetConfigFile(viper.GetString("conf"))
	} else {
		viper.SetConfigName("conf")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/config/")
	}

	viper.ReadInConfig() // nolint: errcheck
	conf := &Conf{}
	if err := viper.Unmarshal(conf); err != nil {
		return conf, fmt.Errorf("unable to unmarshal conf: %w", err)
	}

	return conf, nil
}
