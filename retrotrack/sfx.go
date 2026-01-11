package retrotrack

import (
	"math"
	"math/rand"
)

// --- PUBLIC SFX API ---

func PlayHonk() {
	if context == nil {
		return
	}
	// NYC Taxi dissonance: Bb, B, and E
	freqs := []float64{220.0, 247.0, 329.63}
	duration := 0.6

	// Now we call the helper function
	pcm := generateSimpleSFX(freqs, duration, "square", 0.1)

	p := context.NewPlayerFromBytes(pcm)
	p.Play()
}

func PlayStartSound() {
	if context == nil {
		return
	}
	notes := []float64{261.63, 329.63, 392.00, 523.25}
	noteLen := 0.12
	var combinedBuf []float64
	for _, f := range notes {
		combinedBuf = append(combinedBuf, generateNoteBuf(f, noteLen, "square", 0.2)...)
	}
	p := context.NewPlayerFromBytes(floatToPCM(combinedBuf))
	p.Play()
}

func PlayCrash() {
	if context == nil {
		return
	}
	duration := 0.4
	totalSamples := int(sampleRate * duration)
	buf := make([]float64, totalSamples)
	for i := 0; i < totalSamples; i++ {
		env := math.Exp(-float64(i) / 2000)
		buf[i] = (rand.Float64()*2 - 1) * 0.3 * env
	}
	p := context.NewPlayerFromBytes(floatToPCM(buf))
	p.Play()
}
