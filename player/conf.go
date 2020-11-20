package player

import "github.com/Depado/fox/models"

func (p *Player) UpdateConf(gc *models.Conf) {
	log := p.log.With().Str("action", "conf_update").Logger()
	if p.Conf.VoiceChannel != gc.VoiceChannel {
		log.Debug().Str("old", p.Conf.VoiceChannel).Str("new", gc.VoiceChannel).Msg("voice channel changed")
		if p.Playing() {
			p.Stop()
			p.Play()
		}
	}
	p.Conf = gc
}
