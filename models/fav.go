package models

type FavTrack struct {
	Title        string `json:"title"`
	PermalinkURL string `json:"permalink_url"`
	Author       string `json:""`
}

type FavList struct {
	ID      string
	UserID  string     `json:"user"`
	GuildID string     `json:"guild"`
	Favs    []FavTrack `json:"favs"`
}
