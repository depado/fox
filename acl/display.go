package acl

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
