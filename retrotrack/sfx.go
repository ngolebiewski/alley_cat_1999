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
	notes := []float64{261.63, 329.63, 392.00, 523.25, 392.00, 587.33}
	noteLen := 0.12
	var combinedBuf []float64
	for _, f := range notes {
		combinedBuf = append(combinedBuf, generateNoteBuf(f, noteLen, "square", 0.2)...)
	}
	p := context.NewPlayerFromBytes(floatToPCM(combinedBuf))
	p.Play()
}

func PlayCityStartSound() {
	if context == nil {
		return
	}
	notes := []float64{261.63, 329.63, 261.63, 392.00, 261.63, 523.25}
	noteLen := 0.12
	var combinedBuf []float64
	for _, f := range notes {
		combinedBuf = append(combinedBuf, generateNoteBuf(f, noteLen, "square", 0.2)...)
	}
	p := context.NewPlayerFromBytes(floatToPCM(combinedBuf))
	p.Play()
}

func PlayManifestSound() {
	if context == nil {
		return
	}
	notes := []float64{130.56, 261.63, 392.00, 523.25}
	noteLen := 0.16
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

func PlayGameOverSound() {
	if context == nil {
		return
	}

	// Frequencies for the 4 steps (C3, Bb2, Ab2, G2-ish)
	steps := []float64{261.63, 233.08, 207.65, 196.00}
	noteLen := 0.4
	var combinedBuf []float64

	for i, startFreq := range steps {
		totalSamples := int(sampleRate * noteLen)
		// For the last note (waaaaa), make it 3x longer
		if i == len(steps)-1 {
			totalSamples = int(sampleRate * noteLen * 3)
		}

		buf := make([]float64, totalSamples)
		targetFreq := startFreq * 0.85 // Bend down by 15%

		for j := 0; j < totalSamples; j++ {
			t := float64(j) / sampleRate
			progress := float64(j) / float64(totalSamples)

			// Slide the frequency down
			currentFreq := startFreq + (targetFreq-startFreq)*progress

			// Envelope: Fade in slightly then out to create the "wah" articulation
			env := 1.0
			if progress < 0.2 {
				env = progress / 0.2 // Quick swell
			} else {
				env = 1.0 - progress // Fade out
			}

			// Triangle wave for a "brass" feel
			val := math.Abs(math.Mod(t*currentFreq, 1.0)*4-2) - 1
			buf[j] = val * 0.2 * env
		}
		combinedBuf = append(combinedBuf, buf...)
	}

	p := context.NewPlayerFromBytes(floatToPCM(combinedBuf))
	p.Play()
}
