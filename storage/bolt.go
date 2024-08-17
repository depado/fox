package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/fx"

	"github.com/depado/fox/cmd"
)

var (
	// ErrGuildNotFound is returned when the guild was not found in the bolt db
	ErrGuildNotFound = errors.New("guild not found")
	// ErrGuildsBucketNotFound is returned when the guilds bucket can't be found
	ErrGuildsBucketdNotFound = errors.New("guilds bucket not found")
	// ErrConfNotFound is returned when the guild's conf was not found in the bolt db
	ErrConfNotFound = errors.New("guild conf not found")
)

// BoltStorage is
type BoltStorage struct {
	db  *bolt.DB
	log zerolog.Logger
}

// NewBoltStorage will initiate a new bolt storage backend with the appropriate
// buckets, path and permissions.
func NewBoltStorage(lc fx.Lifecycle, c *cmd.Conf, l zerolog.Logger) (*BoltStorage, error) {
	log := l.With().Str("component", "storage").Str("type", "bolt").Logger()
	db, err := bolt.Open(c.Database.Path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	// Open database and close it on stop lifecycle
	bs := &BoltStorage{db: db, log: log}
	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			return bs.db.Close()
		},
	})

	// Create the base buckets
	err = bs.db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists([]byte(GuildsBucket)); err != nil {
			return err
		}
		if _, err = tx.CreateBucketIfNotExists([]byte(UsersBucket)); err != nil {
			return err
		}
		return nil
	})

	return bs, err
}
