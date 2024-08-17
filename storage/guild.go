package storage

import (
	"encoding/json"
	"fmt"

	"github.com/depado/fox/models"
	"github.com/bwmarrin/discordgo"

	bolt "go.etcd.io/bbolt"
)

// Various constant keys and bucket names
const (
	ConfKey      = "conf"
	InfoKey      = "info"
	UsersBucket  = "users"
	GuildsBucket = "guilds"
)

func (bs *BoltStorage) NewGuild(g *discordgo.GuildCreate) (*models.Conf, error) {
	c, err := bs.NewGuildConf(g.ID)
	if err != nil {
		return nil, fmt.Errorf("new guild conf: %w", err)
	}
	if err := bs.UpdateGuildInfo(g); err != nil {
		return nil, fmt.Errorf("update guild info: %w", err)
	}

	return c, err
}

// NewGuildConf will create a new nested bucket with the guild ID, and will put
// the conf in the `conf` key.
func (bs *BoltStorage) NewGuildConf(guildID string) (*models.Conf, error) {
	gc := models.NewConf(guildID)

	err := bs.db.Update(func(t *bolt.Tx) error {
		guilds := t.Bucket([]byte(GuildsBucket))
		if guilds == nil {
			return ErrGuildsBucketdNotFound
		}
		gb, err := guilds.CreateBucketIfNotExists([]byte(guildID))
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

func (bs *BoltStorage) UpdateGuildInfo(g *discordgo.GuildCreate) error {
	err := bs.db.Update(func(t *bolt.Tx) error {
		guilds := t.Bucket([]byte(GuildsBucket))
		if guilds == nil {
			return ErrGuildsBucketdNotFound
		}
		gb := guilds.Bucket([]byte(g.ID))
		if gb == nil {
			return ErrGuildNotFound
		}

		info := &models.Info{
			Name:     g.Name,
			JoinedAt: g.JoinedAt,
			Members:  g.MemberCount,
		}
		raw, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("marshal info: %w", err)
		}
		if err := gb.Put([]byte(InfoKey), raw); err != nil {
			return fmt.Errorf("put info: %w", err)
		}
		return nil
	})
	return err
}

// GetGuildConf will attempt to fetch the guild configuration for a given ID.
func (bs *BoltStorage) GetGuildConf(guildID string) (*models.Conf, error) {
	gc := &models.Conf{ID: guildID}

	err := bs.db.View(func(t *bolt.Tx) error {
		guilds := t.Bucket([]byte(GuildsBucket))
		if guilds == nil {
			return ErrGuildsBucketdNotFound
		}
		gb := guilds.Bucket([]byte(guildID))
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
		guilds := t.Bucket([]byte(GuildsBucket))
		if guilds == nil {
			return ErrGuildsBucketdNotFound
		}
		gb := guilds.Bucket([]byte(gc.ID))
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
