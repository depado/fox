package acl

import "fmt"

// RoleRestrictionString returns a user-friendly representation of the role
// restriction
func RoleRestrictionString(r RoleRestriction) string {
	var rr string

	switch r {
	case Admin:
		rr = "🔐 Admin only"
	case Privileged:
		rr = "🔒 Admin or DJ"
	case Anyone:
		rr = "🔓 No role restriction"
	}

	return rr
}

// ChannelRestrictionString returns a user-friendly representation of the
// channel restriction
func ChannelRestrictionString(c ChannelRestriction) string {
	var cr string

	switch c {
	case Music:
		cr = "🎶 Music text channel only"
	case Anywhere:
		cr = "🌍 No channel restriction"
	}

	return cr
}

// RestrictionString returns a user-friendly representation of an restriction
// pair
func RestrictionString(c ChannelRestriction, r RoleRestriction) string {
	return fmt.Sprintf("%s\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0%s",
		ChannelRestrictionString(c), RoleRestrictionString(r))
}
