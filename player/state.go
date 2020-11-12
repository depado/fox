package player

// State stores the various state of the player. It is also used by the queue
// to determine some actions related to currently playing tracks.
type State struct {
	Playing bool
	Stopped bool
	Paused  bool
	Volume  int
}

// NewPlayerState will return a new player state
func NewState() *State {
	return &State{
		Volume: 256, // 256 is the normal volume
	}
}
