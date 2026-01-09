package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

var audioContext = audio.NewContext(sampleRate)

var (
	raceMusicPlaying bool
	raceMusicPlayer  *audio.Player
	mu               sync.Mutex
)

//
// PUBLIC API
//

func StartRaceMusic() {
	mu.Lock()
	defer mu.Unlock()

	if raceMusicPlaying {
		return
	}

	pcm := buildRaceTrackPCM()
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))

	player, err := audioContext.NewPlayer(loop)
	if err != nil {
		panic(err)
	}

	player.Play()
	raceMusicPlayer = player
	raceMusicPlaying = true
}

func StopRaceMusic() {
	mu.Lock()
	defer mu.Unlock()

	if !raceMusicPlaying {
		return
	}

	raceMusicPlayer.Pause()
	raceMusicPlayer.Close()
	raceMusicPlayer = nil
	raceMusicPlaying = false
}

//
// TRACK
//

func buildRaceTrackPCM() []byte {
	beat := 0.5 // 120 BPM
	measures := 18

	spb := int(sampleRate * beat)
	spm := spb * 4
	total := spm * measures
	mix := make([]float64, total)

	for m := 0; m < measures; m++ {
		offset := m * spm

		// DRUMS ALWAYS
		addKick(mix, offset, spb)
		addSnare(mix, offset, spb)
		addHiHat(mix, offset, spb)

		// BASS (measure 3+)
		if m >= 2 {
			addBass(mix, offset, m, spb)
		}

		// GUITAR (measure 5–8)
		if m >= 4 && m < 8 {
			addGuitar(mix, offset, m, spb)
		}

		// SOLO (measure 9+)
		if m >= 8 {
			addSolo(mix, offset, m-8, spb, m)
		}
	}

	return floatToPCM(mix)
}

//
// HARMONY
//

func rootForBeat(measure int, beat int) float64 {
	h := measure % 8

	switch h {
	case 0, 2, 4:
		return 82.41 // E
	case 1, 3, 5:
		if beat < 2 {
			return 98.00 // G
		}
		return 73.42 // D
	case 6:
		return 110.00 // A
	case 7:
		return 123.47 // B
	}
	return 82.41
}

//
// DRUMS
//

func addKick(buf []float64, offset, spb int) {
	for b := 0; b < 4; b++ {
		start := offset + b*spb
		for i := 0; i < spb/6 && start+i < len(buf); i++ {
			buf[start+i] += waveform(90, "sine", i, 0.35)
		}
	}
}

// func addKick(buf []float64, offset, spb int) {
// 	for b := 0; b < 4; b++ {
// 		start := offset + b*spb

// 		for i := 0; i < spb/3 && start+i < len(buf); i++ {
// 			t := float64(i) / sampleRate

// 			// 808 pitch envelope (fast drop)
// 			freq := 48.0 + 140.0*math.Exp(-t*45)

// 			// Amplitude envelope
// 			env := math.Exp(-t * 9)

// 			// Sub sine body
// 			body := math.Sin(2*math.Pi*freq*t) * env

// 			// Distortion / saturation
// 			drive := 3.5
// 			dist := math.Tanh(body * drive)

// 			// Attack click (punk edge)
// 			click := 0.0
// 			if i < 32 {
// 				click = (rand.Float64()*2 - 1) * 0.25
// 			}

// 			buf[start+i] += (dist*0.85 + click*0.08)
// 		}
// 	}
// }

// 🔥 NOISY SNARE
func addSnare(buf []float64, offset, spb int) {
	for _, b := range []int{1, 3} {
		start := offset + b*spb
		for i := 0; i < spb/4 && start+i < len(buf); i++ {
			noise := (rand.Float64()*2 - 1) * 0.25
			env := math.Exp(-float64(i) / 900) // fast decay
			buf[start+i] += noise * env
		}
	}
}

func addHiHat(buf []float64, offset, spb int) {
	for b := 0; b < 4; b++ {
		start := offset + b*spb
		for i := 0; i < spb/10 && start+i < len(buf); i++ {
			buf[start+i] += waveform(5000, "square", i, 0.06)
		}
	}
}

//
// BASS & GUITAR
//

func addBass(buf []float64, offset, measure, spb int) {
	for i := 0; i < 8; i++ {
		root := rootForBeat(measure, i/2)
		start := offset + i*(spb/2)
		for s := 0; s < spb/3 && start+s < len(buf); s++ {
			buf[start+s] += waveform(root, "triangle", s, 0.22)
		}
	}
}

func addGuitar(buf []float64, offset, measure, spb int) {
	for b := 0; b < 4; b++ {
		root := rootForBeat(measure, b)
		start := offset + b*spb
		for s := 0; s < spb/2 && start+s < len(buf); s++ {
			buf[start+s] += waveform(root*2, "square", s, 0.06)
		}
	}
}

//
// SOLO
//

func addSolo(buf []float64, offset, soloMeasure, spb, songMeasure int) {
	root := rootForBeat(songMeasure, 0)

	switch {

	// ARPEGGIATOR — 6 MEASURES
	case soloMeasure < 6:
		arp := []float64{1, 1.5, 2}
		for i := 0; i < 8; i++ {
			freq := root * arp[i%3] * 4
			start := offset + i*(spb/2)
			for s := 0; s < spb/3 && start+s < len(buf); s++ {
				buf[start+s] += waveform(freq, "sine", s, 0.16)
			}
		}

	// SYNTH DROP — 4 MEASURES
	default:
		progress := float64(soloMeasure-6) / 4.0

		for i := 0; i < spb*4 && offset+i < len(buf); i++ {
			drift := 1.0 - progress*0.35 - float64(i)/float64(spb*12)
			freq := root * 4 * drift

			pulse := 0.5 + 0.5*math.Sin(float64(i)/220)
			vol := (0.22 - progress*0.12) * pulse

			buf[offset+i] += waveform(freq, "sine", i, vol)
		}
	}
}

//
// WAVEFORM & PCM
//

func waveform(freq float64, kind string, i int, vol float64) float64 {
	t := float64(i) / sampleRate
	switch kind {
	case "sine":
		return math.Sin(2*math.Pi*freq*t) * vol
	case "square":
		if math.Sin(2*math.Pi*freq*t) >= 0 {
			return vol
		}
		return -vol
	case "triangle":
		return (2*math.Abs(2*((freq*t)-math.Floor((freq*t)+0.5))) - 1) * vol
	}
	return 0
}

func floatToPCM(buf []float64) []byte {
	pcm := make([]byte, len(buf)*2)
	for i, v := range buf {
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		s := int16(v * math.MaxInt16)
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(s))
	}
	return pcm
}
