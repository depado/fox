package acl

import (
	"fmt"

	"github.com/Depado/fox/models"
	"github.com/Depado/fox/storage"
	"github.com/bwmarrin/discordgo"
)

type RoleRestriction int

const (
	Admin RoleRestriction = iota
	Privileged
	Anyone
)

type ChannelRestriction int

const (
	Music ChannelRestriction = iota
	Anywhere
)

type ACL struct {
	storage *storage.StormDB
}

func NewACL(s *storage.StormDB) *ACL {
	return &ACL{
		storage: s,
	}
}

// Check will perform checks for the given RoleRestriction and
// ChannelRestriction.
func (a ACL) Check(s *discordgo.Session, m *discordgo.Message, r RoleRestriction, c ChannelRestriction) (bool, error) {
	var gc *models.Conf
	var err error
	var rc bool

	// Fetch guild state
	if gc, err = a.storage.GetGuilConf(m.GuildID); err != nil {
		return false, fmt.Errorf("get guild state: %w", err)
	}

	// Check for user restriction
	switch r {
	case Admin:
		if rc, err = a.IsAdmin(s, m); err != nil {
			return false, err
		}
	case Privileged:
		// If no privileged role is defined, automatically refuse unless admin
		if gc.PrivilegedRole == "" {
			if rc, err = a.IsAdmin(s, m); err != nil {
				return false, err
			}
		} else {
			if rc, err = a.IsPrivileged(s, m, gc); err != nil {
				return false, err
			}
		}
	case Anyone:
		rc = true
	}

	if !rc {
		return false, nil
	}

	// Check for channel restriction
	if c == Music {
		// If no text channel defined, automatically approve
		if gc.TextChannel == "" {
			return true, nil
		} else {
			return m.ChannelID == gc.TextChannel, nil
		}
	}

	return true, nil
}

// IsMusic will check if the provided message was sent to the music channel.
func (a ACL) IsMusic(m *discordgo.Message, gc *models.Conf) bool {
	return m.ChannelID == gc.TextChannel
}

// IsPrivileged will check if a member is either admin or DJ.
func (a ACL) IsPrivileged(s *discordgo.Session, m *discordgo.Message, gc *models.Conf) (bool, error) {
	adm, err := a.IsAdmin(s, m)
	if err != nil {
		return false, fmt.Errorf("check admin: %w", err)
	}
	if adm {
		return true, nil
	}

	if gc.PrivilegedRole != "" {
		return a.HasRole(m.Member, gc.PrivilegedRole), nil
	}
	return false, nil
}

// HasRole will check if a guild member has the given role
func (a ACL) HasRole(u *discordgo.Member, r string) bool {
	for _, ur := range u.Roles {
		if ur == r {
			return true
		}
	}
	return false
}

// IsAdmin will check if a member has the admin role.
func (a ACL) IsAdmin(s *discordgo.Session, m *discordgo.Message) (bool, error) {
	g, err := s.Guild(m.GuildID)
	if err != nil {
		return false, fmt.Errorf("get guild: %w", err)
	}

	// Always true for the guild owner
	if m.Author.ID == g.OwnerID {
		return true, nil
	}

	// For every non-managed admin role, check if user has this role
	for _, r := range g.Roles {
		if r.Permissions&discordgo.PermissionAdministrator != 0 && !r.Managed && a.HasRole(m.Member, r.ID) {
			return true, nil
		}
	}

	return false, nil
}
