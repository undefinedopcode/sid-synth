package audio

import "math"

// SID constants matching the original JavaScript
const (
	ClockNTSC   = 1022727
	SampleRate  = 44100
)

// CyclesPerSample must be float to avoid integer truncation (23.1888... not 23)
var CyclesPerSample = float64(ClockNTSC) / float64(SampleRate)

// Waveform types
const (
	WaveTriangle = 0
	WaveSawtooth = 1
	WavePulse    = 2
	WaveNoise    = 3
)

// ADSR envelope phases
const (
	EnvOff     = 0
	EnvAttack  = 1
	EnvDecay   = 2
	EnvSustain = 3
	EnvRelease = 4
)

// Attack times in ms (16 values, 0-15)
var AttackMs = [16]int{2, 8, 16, 24, 38, 56, 68, 80, 100, 250, 500, 800, 1000, 3000, 5000, 8000}

// Decay/Release times in ms
var DecrelMs = [16]int{6, 24, 48, 72, 114, 168, 204, 240, 300, 750, 1500, 2400, 3000, 9000, 15000, 24000}

// NoteToSIDFreq maps MIDI note (0-127) to SID frequency register value
var NoteToSIDFreq [128]uint32

func init() {
	// Precompute MIDI note -> SID frequency register
	for n := 0; n < 128; n++ {
		hz := 440 * math.Pow(2, float64(n-69)/12.0)
		NoteToSIDFreq[n] = uint32(math.Round((hz * 16777216) / ClockNTSC))
	}
}

// Voice represents a single SID oscillator with ADSR envelope
type Voice struct {
	freq         uint32
	pulseWidth   uint16
	waveform     int
	accumulator  uint32
	noisePhase   float64
	lfsr         uint32
	gate         bool
	envPhase     int
	envValue     float64
	attackRate   int
	decayRate    int
	sustainLevel float64
	releaseRate  int
	attackInc    float64
	decayMul     float64
	releaseMul   float64
	active       bool
	midiNote     int
	midiChannel  int
	velocity     float64
	age          int
	ringMod      bool
	sync         bool
}

// NewVoice creates a new SID voice
func NewVoice() *Voice {
	v := &Voice{
		pulseWidth:   2048,
		waveform:     WavePulse,
		sustainLevel: 0.6,
		lfsr:         0x7FFFFF,
	}
	v.computeRates()
	return v
}

// computeRates calculates ADSR rate multipliers from attack/decay/release times
func (v *Voice) computeRates() {
	atkMs := AttackMs[v.attackRate]
	decMs := DecrelMs[v.decayRate]
	relMs := DecrelMs[v.releaseRate]

	atkSamples := float64(atkMs) / 1000.0 * SampleRate
	if atkSamples > 0 {
		v.attackInc = 1.0 / atkSamples
	} else {
		v.attackInc = 1.0
	}

	decSamples := float64(decMs) / 1000.0 * SampleRate
	if decSamples > 0 {
		v.decayMul = math.Pow(0.001, 1.0/decSamples)
	} else {
		v.decayMul = 0
	}

	relSamples := float64(relMs) / 1000.0 * SampleRate
	if relSamples > 0 {
		v.releaseMul = math.Pow(0.001, 1.0/relSamples)
	} else {
		v.releaseMul = 0
	}
}

// SetADSR sets the ADSR parameters (4 nybbles: A, D, S, R)
func (v *Voice) SetADSR(a, d, s, r int) {
	v.attackRate = a & 0xF
	v.decayRate = d & 0xF
	v.sustainLevel = float64(s&0xF) / 15.0
	v.releaseRate = r & 0xF
	v.computeRates()
}

// GateOn triggers the voice with a note
func (v *Voice) GateOn(note, velocity int, origNote int) {
	if note < 0 || note > 127 {
		return
	}
	if origNote < 0 {
		origNote = note
	}
	v.midiNote = origNote
	v.freq = NoteToSIDFreq[note]
	v.velocity = float64(velocity) / 127.0
	v.gate = true
	v.envPhase = EnvAttack
	v.active = true
	v.age = 0
	v.accumulator = 0
}

// GateOff releases the voice
func (v *Voice) GateOff() {
	v.gate = false
	if v.envPhase != EnvOff {
		v.envPhase = EnvRelease
	}
}

// Waveform returns the waveform type
func (v *Voice) Waveform() int {
	return v.waveform
}

// SetWaveform sets the waveform type
func (v *Voice) SetWaveform(w int) {
	if w >= 0 && w <= 3 {
		v.waveform = w
	}
}

// PulseWidth returns the pulse width
func (v *Voice) PulseWidth() uint16 {
	return v.pulseWidth
}

// SetPulseWidth sets the pulse width
func (v *Voice) SetPulseWidth(pw uint16) {
	v.pulseWidth = pw
}

// EnvPhase returns the envelope phase
func (v *Voice) EnvPhase() int {
	return v.envPhase
}

// EnvValue returns the envelope value
func (v *Voice) EnvValue() float64 {
	return v.envValue
}

// Age returns the voice age
func (v *Voice) Age() int {
	return v.age
}

// Active returns whether the voice is active
func (v *Voice) Active() bool {
	return v.active
}

// MidiNote returns the MIDI note number
func (v *Voice) MidiNote() int {
	return v.midiNote
}

// MidiChannel returns the MIDI channel
func (v *Voice) MidiChannel() int {
	return v.midiChannel
}

// GenerateSample produces one audio sample
// ringModSource is used for ring modulation (pass nil if not used)
func (v *Voice) GenerateSample(ringModSource *Voice) float64 {
	if !v.active {
		return 0
	}
	v.age++

	// Advance accumulator
	step := float64(v.freq) * CyclesPerSample
	v.accumulator = (v.accumulator + uint32(step)) % 16777216

	// Generate waveform (12-bit, 0-4095)
	var output uint16
	acc := v.accumulator
	top12 := uint16((acc >> 12) & 0xFFF)

	switch v.waveform {
	case WaveTriangle:
		tri := uint16(top12)
		if acc&0x800000 != 0 {
			tri = 0xFFF - tri
		}
		if v.ringMod && ringModSource != nil {
			if ringModSource.accumulator&0x800000 != 0 {
				tri = 0xFFF - tri
			}
		}
		output = tri
	case WaveSawtooth:
		output = uint16(top12)
	case WavePulse:
		if top12 >= v.pulseWidth {
			output = 4095
		} else {
			output = 0
		}
	case WaveNoise:
		v.noisePhase += step
		for v.noisePhase >= (1 << 20) {
			v.noisePhase -= (1 << 20)
			bit := ((v.lfsr >> 17) ^ (v.lfsr >> 22)) & 1
			v.lfsr = ((v.lfsr << 1) | bit) & 0x7FFFFF
		}
		output = uint16((v.lfsr >> 11) & 0xFFF)
	}

	// Normalize to -1..1
	sample := float64(output)/2048.0 - 1.0

	// Mild noise boost
	if v.waveform == WaveNoise {
		sample *= 1.2
	}

	// ADSR envelope
	switch v.envPhase {
	case EnvAttack:
		v.envValue += v.attackInc
		if v.envValue >= 1.0 {
			v.envValue = 1.0
			v.envPhase = EnvDecay
		}
	case EnvDecay:
		target := v.sustainLevel
		v.envValue = target + (v.envValue-target)*v.decayMul
		if v.envValue <= target+0.001 {
			v.envValue = target
			v.envPhase = EnvSustain
		}
	case EnvSustain:
		v.envValue = v.sustainLevel
		if v.sustainLevel < 0.002 {
			// Zero sustain - voice is silent, free it up
			v.envValue = 0
			v.envPhase = EnvOff
			v.active = false
			v.gate = false
		}
	case EnvRelease:
		v.envValue *= v.releaseMul
		if v.envValue < 0.001 {
			v.envValue = 0
			v.envPhase = EnvOff
			v.active = false
		}
	default:
		v.envValue = 0
		v.active = false
	}

	return sample * v.envValue * v.velocity
}
