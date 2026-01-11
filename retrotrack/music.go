package retrotrack

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

// Initialize initializes the internal audio context.
// Call this once at the very start of your game.
func Initialize() {
	mu.Lock()
	defer mu.Unlock()
	if context == nil {
		context = audio.NewContext(sampleRate)
	}
}

// GetContext returns the internal context if you need it for other Ebiten functions
func GetContext() *audio.Context {
	return context
}

var (
	context *audio.Context
	player  *audio.Player
	playing bool
	mu      sync.Mutex
)

func Init(ctx *audio.Context) {
	context = ctx
}

// --- PUBLIC MUSIC API ---

func Start() {
	mu.Lock()
	defer mu.Unlock()
	if playing || context == nil {
		return
	}
	pcm := buildPCM()
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	var err error
	player, err = context.NewPlayer(loop)
	if err == nil {
		player.Play()
		playing = true
	}
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if !playing || player == nil {
		return
	}
	player.Close()
	playing = false
}

// --- INTERNAL GENERATORS ---

func buildPCM() []byte {
	beat := 0.4
	measures := 20
	spb := int(sampleRate * beat)
	spm := spb * 8
	mix := make([]float64, spm*measures)

	for m := 0; m < measures; m++ {
		off := m * spm
		addDrums(mix, off, spb)
		addBass(mix, off, m, spb)
		if m < 8 {
			addArp(mix, off, m, spb)
		} else if m < 16 {
			addLead(mix, off, m, spb)
		}
		// else {
		// 	addEbRiff(mix, off, m, spb)
		// }
	}
	return floatToPCM(mix)
}

func addDrums(buf []float64, off, spb int) {
	for b := 0; b < 8; b++ {
		start := off + b*spb
		if b%4 == 0 {
			for i := 0; i < spb && start+i < len(buf); i++ {
				t := float64(i) / sampleRate
				f := 120.0 * math.Exp(-t*50)
				buf[start+i] += waveform(f, "triangle", i, 0.4, 0)
			}
		}
		if b%4 == 2 {
			for i := 0; i < spb && start+i < len(buf); i++ {
				buf[start+i] += (rand.Float64()*2 - 1) * 0.25 * math.Exp(-float64(i)/1000)
			}
		}
	}
}

func addBass(buf []float64, off, m, spb int) {
	prog := []float64{82.41, 98.00, 73.42, 110.00}
	root := prog[m%4]
	if m >= 16 {
		root = 77.78
	}
	for b := 0; b < 8; b++ {
		start := off + b*spb
		freq := root
		if b%2 == 1 {
			freq *= 0.5
		}
		for i := 0; i < spb && start+i < len(buf); i++ {
			if i > spb-200 {
				continue
			}
			buf[start+i] += waveform(freq, "triangle", i, 0.2, 0)
		}
	}
}

func addArp(buf []float64, off, m, spb int) {
	prog := []float64{82.41, 98.00, 73.42, 110.00} // E, G, D, A
	root := prog[m%4]

	// Classic rolling 8th-note pattern
	arpNotes := []float64{1.0, 1.25, 1.5, 1.25, 2.0, 1.5, 1.25, 1.0}

	for b := 0; b < 8; b++ {
		start := off + b*spb
		freq := root * 2 * arpNotes[b]

		for i := 0; i < spb && start+i < len(buf); i++ {
			// We use 'saw' for that fat 80s warmth
			// v is a tiny bit of pitch drift for 'unstable analog' feel
			drift := 0.002 * math.Sin(2*math.Pi*2.0*float64(i)/sampleRate)

			// A Sin envelope softens the 'pluck', making it less annoying
			env := math.Sin(math.Pi * float64(i) / float64(spb))
			buf[start+i] += waveform(freq, "saw", i, 0.1*env, drift)
		}
	}
}

func addLead(buf []float64, off, m, spb int) {
	prog := []float64{82.41, 98.00, 73.42, 110.00}
	root := prog[m%4]
	for b := 0; b < 8; b++ {
		start := off + b*spb
		for i := 0; i < spb && start+i < len(buf); i++ {
			v := 0.012 * math.Sin(2*math.Pi*6.0*float64(i)/sampleRate)
			env := math.Exp(-float64(i) / 4000)
			buf[start+i] += waveform(root*2, "square", i, 0.12*env, v)
		}
	}
}

func addEbRiff(buf []float64, off, m, spb int) {
	root := 77.78
	for b := 0; b < 8; b++ {
		start := off + b*spb
		f := root * 4
		if b%2 == 1 {
			f *= 1.414
		}
		for i := 0; i < spb && start+i < len(buf); i++ {
			v := 0.04 * math.Sin(2*math.Pi*10.0*float64(i)/sampleRate)
			buf[start+i] += waveform(f, "square", i, 0.15, v)
		}
	}
}

// --- AUDIO UTILITIES ---

func waveform(freq float64, kind string, i int, vol float64, vibrato float64) float64 {
	t := float64(i) / sampleRate
	f := freq + (freq * vibrato)
	switch kind {
	case "saw":
		// This is the classic 80s synth sound
		return (2.0 * (t*f - math.Floor(t*f+0.5))) * vol
	case "square":
		if math.Sin(2*math.Pi*f*t) >= 0 {
			return vol
		}
		return -vol
	case "pulse12":
		if math.Sin(2*math.Pi*f*t) > 0.75 {
			return vol
		}
		return -vol
	case "triangle":
		return (2*math.Abs(2*((f*t)-math.Floor((f*t)+0.5))) - 1) * vol
	}
	return 0
}

func generateNoteBuf(freq float64, duration float64, kind string, vol float64) []float64 {
	samples := int(sampleRate * duration)
	buf := make([]float64, samples)
	for i := 0; i < samples; i++ {
		buf[i] = waveform(freq, kind, i, vol, 0) * (1.0 - (float64(i) / float64(samples)))
	}
	return buf
}

func generateSimpleSFX(freqs []float64, duration float64, kind string, vol float64) []byte {
	samples := int(sampleRate * duration)
	buf := make([]float64, samples)
	for i := 0; i < samples; i++ {
		for _, f := range freqs {
			buf[i] += waveform(f, kind, i, vol, 0)
		}
		buf[i] *= math.Exp(-float64(i) / 3000)
	}
	return floatToPCM(buf)
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
