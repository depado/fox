package storage

import (
	"context"
	"fmt"

	"github.com/Depado/fox/cmd"
	"github.com/Depado/fox/guild"
	"github.com/asdine/storm/v3"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type StormDB struct {
	db  *storm.DB
	log zerolog.Logger
}

func NewStormStorage(lc fx.Lifecycle, c *cmd.Conf, l zerolog.Logger) (*StormDB, error) {
	log := l.With().Str("component", "storage").Logger()
	log.Debug().Str("path", c.Database.Path).Msg("opening database")

	db, err := storm.Open(c.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Init(&guild.Conf{}); err != nil {
		return nil, fmt.Errorf("init guild state model: %w", err)
	}

	sdb := &StormDB{
		db:  db,
		log: log,
	}

	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			sdb.log.Debug().Str("lifecycle", "stop").Msg("closing database")
			return db.Close()
		},
	})

	return sdb, nil
}

func (s *StormDB) GetGuilConf(guildID string) (*guild.Conf, error) {
	gstate := &guild.Conf{}
	return gstate, s.db.One("ID", guildID, gstate)
}

func (s *StormDB) SaveGuildState(gs *guild.Conf) error {
	return s.db.Save(gs)
}

func (s *StormDB) NewGuildState(guildID string) (*guild.Conf, error) {
	gs := guild.NewConf(guildID)
	return gs, s.db.Save(gs)
}
