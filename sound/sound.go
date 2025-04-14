package sound

import (
	"math"

	"github.com/hajimehoshi/oto"
)

type Player struct {
	p *oto.Player
}

func InitSound() *Player {
	const sampleRate = 44100
	const channelNum = 1
	const bitDepthInBytes = 2

	var err error
	ctx, _ := oto.NewContext(sampleRate, channelNum, bitDepthInBytes, 1024)
	player := ctx.NewPlayer()
	if err != nil {
		return nil
	}
	return &Player{p: player}
}

func (player *Player) PlayBeep() {
	go func() {
		const freq = 440.0
		const durationSeconds = 0.2
		const sampleRate = 44100
		const volume = 0.2

		numSamples := int(sampleRate * durationSeconds)
		buf := make([]byte, numSamples*2)

		for i := 0; i < numSamples; i++ {
			t := float64(i) / sampleRate
			v := int16(volume * math.Sin(2*math.Pi*freq*t) * 32767)
			buf[2*i] = byte(v)
			buf[2*i+1] = byte(v >> 8)
		}

		player.p.Write(buf)
	}()
}
