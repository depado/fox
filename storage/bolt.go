package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Depado/fox/cmd"
	"github.com/Depado/fox/models"
	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/fx"
)

var (
	// ErrGuildNotFound is returned when the guild was not found in the bolt db
	ErrGuildNotFound = errors.New("guild not found")
	// ErrConfNotFound is returned when the guild's conf was not found in the bolt db
	ErrConfNotFound = errors.New("guild conf not found")
)

// Various constant keys and bucket names
const (
	ConfKey      = "conf"
	UsersBucket  = "users"
	GuildsBucket = "guilds"
)

// BoltStorage is
type BoltStorage struct {
	db     *bolt.DB
	log    zerolog.Logger
	users  *bolt.Bucket
	guilds *bolt.Bucket
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
		if bs.users, err = tx.CreateBucketIfNotExists([]byte(UsersBucket)); err != nil {
			return err
		}
		if bs.guilds, err = tx.CreateBucketIfNotExists([]byte(GuildsBucket)); err != nil {
			return err
		}
		return nil
	})

	return bs, err
}

// NewGuildConf will create a new nested bucket with the guild ID, and will put
// the conf in the `conf` key.
func (bs *BoltStorage) NewGuildConf(guildID string) (*models.Conf, error) {
	gc := models.NewConf(guildID)

	err := bs.db.Update(func(t *bolt.Tx) error {
		gb, err := bs.guilds.CreateBucketIfNotExists([]byte(guildID))
		if err != nil {
			return fmt.Errorf("create guild bucket: %w", err)
		}
		if buf, err := json.Marshal(gc); err != nil {
			return fmt.Errorf("marshal guild conf: %w", err)
		} else if err := gb.Put([]byte(ConfKey), buf); err != nil {
			return fmt.Errorf("put guild conf: %w", err)
		}
		return nil
	})

	return gc, err
}

// GetGuildConf will attempt to fetch the guild configuration for a given ID.
func (bs *BoltStorage) GetGuilConf(guildID string) (*models.Conf, error) {
	gc := &models.Conf{ID: guildID}

	err := bs.db.View(func(t *bolt.Tx) error {
		gb := bs.guilds.Bucket([]byte(guildID))
		if gb == nil {
			return ErrGuildNotFound
		}
		raw := gb.Get([]byte(ConfKey))
		if raw == nil {
			return ErrConfNotFound
		}
		if err := json.Unmarshal(raw, gc); err != nil {
			return fmt.Errorf("unmarshal guild conf: %w", err)
		}
		return nil
	})

	return gc, err
}

// SaveGuildConf will save the guild conf to the appropriate bucket
// Note that this can fail in the event of manual tampering with the guild conf,
// for example removing its ID or setting it to empty string before calling this
// method.
func (bs *BoltStorage) SaveGuildConf(gc *models.Conf) error {
	if gc.ID == "" {
		return fmt.Errorf("unable to save conf with no GuildID")
	}
	return bs.db.Update(func(t *bolt.Tx) error {
		gb := bs.guilds.Bucket([]byte(gc.ID))
		if gb == nil {
			return ErrGuildNotFound
		}
		if buf, err := json.Marshal(gc); err != nil {
			return fmt.Errorf("marshal guild conf: %w", err)
		} else if err := gb.Put([]byte(ConfKey), buf); err != nil {
			return fmt.Errorf("put guild conf: %w", err)
		}
		return nil
	})
}
