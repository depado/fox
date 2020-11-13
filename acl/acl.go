package acl

import (
	"github.com/Depado/fox/cmd"
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
	AdminRoleID    string
	DJRoleID       string
	MusicChannelID string
}

func NewACL(conf *cmd.Conf) *ACL {
	return &ACL{
		AdminRoleID:    conf.Bot.Roles.Admin,
		DJRoleID:       conf.Bot.Roles.DJ,
		MusicChannelID: conf.Bot.Channels.Text,
	}
}

// Check will perform checks for the given RoleRestriction and
// ChannelRestriction.
func (a ACL) Check(r RoleRestriction, c ChannelRestriction, u *discordgo.Member, m *discordgo.Message) bool {
	// Check for user restriction
	switch r {
	case Admin:
		if !a.IsAdmin(u) {
			return false
		}
	case Privileged:
		if !a.IsPrivileged(u) {
			return false
		}
	}

	// Check for channel restriction
	if c == Music {
		if !a.IsMusic(m) {
			return false
		}
	}

	return true
}

func RoleRestrictionString(r RoleRestriction) string {
	var rr string

	switch r {
	case Admin:
		rr = "🔐 Admin"
	case Privileged:
		rr = "🔒 Admin or DJ"
	case Anyone:
		rr = "🔓 No restriction"
	}

	return rr
}

func ChannelRestrictionString(c ChannelRestriction) string {
	var cr string

	switch c {
	case Music:
		cr = "🎶 Music text channel only"
	case Anywhere:
		cr = "🌍 No restriction"
	}

	return cr
}

// IsMusic will check if the provided message was sent to the music channel.
func (a ACL) IsMusic(m *discordgo.Message) bool {
	return m.ChannelID == a.MusicChannelID
}

// IsPrivileged will check if a member is either admin or DJ.
func (a ACL) IsPrivileged(m *discordgo.Member) bool {
	for _, r := range m.Roles {
		if r == a.AdminRoleID || r == a.DJRoleID {
			return true
		}
	}
	return false
}

// IsAdmin will check if a member has the admin role.
func (a ACL) IsAdmin(m *discordgo.Member) bool {
	for _, r := range m.Roles {
		if r == a.AdminRoleID {
			return true
		}
	}
	return false
}
