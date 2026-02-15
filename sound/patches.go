package sound

// Waveform types (same as audio package)
const (
	WaveTriangle = 0
	WaveSawtooth = 1
	WavePulse    = 2
	WaveNoise    = 3
)

// Filter modes
const (
	FilterOff = 0
	FilterLP  = 1
	FilterBP  = 2
	FilterHP  = 3
)

// Patch represents a single instrument preset
type Patch struct {
	Name        string
	Waveform    int
	PulseWidth  int
	Attack      int
	Decay       int
	Sustain     int
	Release     int
	FilterFreq  float64 // Normalized 0-1 (matching JS cutoff values)
	FilterRes   float64 // 0-1
	FilterMode  int
	FilterModel string // "6581", "8580", or "none"
	FilterOn    bool   // Whether filter is enabled
}

// DrumPatch represents a GM percussion sound
type DrumPatch struct {
	Name       string
	Waveform   int
	Attack     int
	Decay      int
	Sustain    int
	Release    int
	Note       int     // Actual note to play (pitch)
	FilterMode int
	Cutoff     float64 // Normalized 0-1
	Res        float64
}

// GM_PATCHES - Complete General MIDI instrument patches (0-127)
// FilterFreq values are normalized 0-1, matching the JavaScript reference exactly
var GM_PATCHES = []Patch{
	// === Piano (0-7) ===
	{Name: "Acoustic Grand", Waveform: 2, PulseWidth: 2400, Attack: 0, Decay: 7, Sustain: 3, Release: 5, FilterFreq: 0.28, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Bright Acoustic", Waveform: 2, PulseWidth: 2048, Attack: 0, Decay: 7, Sustain: 3, Release: 5, FilterFreq: 0.38, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Electric Grand", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 8, Sustain: 4, Release: 6, FilterFreq: 0.35, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Honky-Tonk", Waveform: 2, PulseWidth: 1800, Attack: 0, Decay: 6, Sustain: 2, Release: 4, FilterFreq: 0.42, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Electric Piano 1", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 8, Sustain: 4, Release: 6, FilterFreq: 0.32, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Electric Piano 2", Waveform: 2, PulseWidth: 1600, Attack: 0, Decay: 7, Sustain: 3, Release: 5, FilterFreq: 0.36, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Harpsichord", Waveform: 2, PulseWidth: 600, Attack: 0, Decay: 5, Sustain: 3, Release: 4, FilterFreq: 0.4, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Clavinet", Waveform: 2, PulseWidth: 512, Attack: 0, Decay: 4, Sustain: 2, Release: 3, FilterFreq: 0.45, FilterRes: 0.45, FilterMode: 1, FilterOn: true},
	// === Chromatic Percussion (8-15) ===
	{Name: "Celesta", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 2, Release: 5, FilterFreq: 0.5, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Glockenspiel", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 0, Release: 4, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Music Box", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 7, Sustain: 0, Release: 6, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Vibraphone", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 8, Sustain: 2, Release: 6, FilterFreq: 0.4, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Marimba", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 0, Release: 3, FilterFreq: 0.35, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Xylophone", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 0, Release: 2, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Tubular Bells", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 9, Sustain: 0, Release: 7, FilterFreq: 0.45, FilterRes: 0.35, FilterMode: 2, FilterOn: true},
	{Name: "Dulcimer", Waveform: 2, PulseWidth: 1400, Attack: 0, Decay: 6, Sustain: 0, Release: 5, FilterFreq: 0.38, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	// === Organ (16-23) ===
	{Name: "Drawbar Organ", Waveform: 2, PulseWidth: 2400, Attack: 2, Decay: 6, Sustain: 14, Release: 5, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Percussive Organ", Waveform: 2, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 12, Release: 4, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Rock Organ", Waveform: 2, PulseWidth: 1600, Attack: 1, Decay: 5, Sustain: 13, Release: 4, FilterFreq: 0.4, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Church Organ", Waveform: 2, PulseWidth: 2048, Attack: 3, Decay: 6, Sustain: 14, Release: 5, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Reed Organ", Waveform: 2, PulseWidth: 512, Attack: 3, Decay: 5, Sustain: 12, Release: 5, FilterFreq: 0.3, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Accordion", Waveform: 2, PulseWidth: 1200, Attack: 2, Decay: 5, Sustain: 13, Release: 5, FilterFreq: 0.32, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Harmonica", Waveform: 2, PulseWidth: 1800, Attack: 1, Decay: 5, Sustain: 12, Release: 5, FilterFreq: 0.35, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Tango Accordion", Waveform: 2, PulseWidth: 900, Attack: 2, Decay: 5, Sustain: 13, Release: 5, FilterFreq: 0.3, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	// === Guitar (24-31) ===
	{Name: "Nylon Guitar", Waveform: 2, PulseWidth: 2200, Attack: 0, Decay: 6, Sustain: 4, Release: 4, FilterFreq: 0.25, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Steel Guitar", Waveform: 2, PulseWidth: 1400, Attack: 0, Decay: 5, Sustain: 3, Release: 4, FilterFreq: 0.38, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Jazz Guitar", Waveform: 2, PulseWidth: 2048, Attack: 0, Decay: 7, Sustain: 5, Release: 5, FilterFreq: 0.22, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Clean Electric", Waveform: 2, PulseWidth: 1200, Attack: 0, Decay: 6, Sustain: 4, Release: 4, FilterFreq: 0.4, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Muted Guitar", Waveform: 2, PulseWidth: 1600, Attack: 0, Decay: 2, Sustain: 0, Release: 1, FilterFreq: 0.3, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Overdrive Guitar", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 8, Release: 4, FilterFreq: 0.35, FilterRes: 0.5, FilterMode: 1, FilterOn: true},
	{Name: "Distortion Guitar", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 9, Release: 3, FilterFreq: 0.42, FilterRes: 0.55, FilterMode: 1, FilterOn: true},
	{Name: "Guitar Harmonics", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 7, Sustain: 0, Release: 5, FilterFreq: 0.5, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	// === Bass (32-39) ===
	{Name: "Acoustic Bass", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 4, Release: 4, FilterFreq: 0.2, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Fingered Bass", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 6, Release: 4, FilterFreq: 0.22, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Picked Bass", Waveform: 2, PulseWidth: 1800, Attack: 0, Decay: 4, Sustain: 3, Release: 3, FilterFreq: 0.25, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Fretless Bass", Waveform: 0, PulseWidth: 2048, Attack: 1, Decay: 7, Sustain: 8, Release: 5, FilterFreq: 0.18, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Slap Bass 1", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 3, Sustain: 2, Release: 2, FilterFreq: 0.28, FilterRes: 0.5, FilterMode: 1, FilterOn: true},
	{Name: "Slap Bass 2", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 3, Release: 2, FilterFreq: 0.25, FilterRes: 0.45, FilterMode: 1, FilterOn: true},
	{Name: "Synth Bass 1", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 2, Release: 2, FilterFreq: 0.2, FilterRes: 0.55, FilterMode: 1, FilterOn: true},
	{Name: "Synth Bass 2", Waveform: 2, PulseWidth: 3200, Attack: 0, Decay: 5, Sustain: 5, Release: 3, FilterFreq: 0.18, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	// === Strings (40-47) ===
	{Name: "Violin", Waveform: 1, PulseWidth: 2048, Attack: 3, Decay: 7, Sustain: 12, Release: 5, FilterFreq: 0.35, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Viola", Waveform: 1, PulseWidth: 2048, Attack: 3, Decay: 7, Sustain: 12, Release: 5, FilterFreq: 0.3, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Cello", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 7, Sustain: 12, Release: 6, FilterFreq: 0.25, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Contrabass", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 8, Sustain: 11, Release: 6, FilterFreq: 0.2, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Tremolo Strings", Waveform: 1, PulseWidth: 2048, Attack: 2, Decay: 6, Sustain: 12, Release: 5, FilterFreq: 0.32, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Pizzicato", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 0, Release: 2, FilterFreq: 0.38, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Harp", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 8, Sustain: 0, Release: 6, FilterFreq: 0.42, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Timpani", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 0, Release: 4, FilterFreq: 0.12, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	// === Ensemble (48-55) ===
	{Name: "String Ensemble 1", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 7, Sustain: 12, Release: 6, FilterFreq: 0.3, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "String Ensemble 2", Waveform: 1, PulseWidth: 2048, Attack: 5, Decay: 8, Sustain: 13, Release: 7, FilterFreq: 0.25, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Synth Strings 1", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 7, Sustain: 13, Release: 6, FilterFreq: 0.38, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Synth Strings 2", Waveform: 1, PulseWidth: 2048, Attack: 5, Decay: 8, Sustain: 14, Release: 7, FilterFreq: 0.35, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Choir Aahs", Waveform: 2, PulseWidth: 2400, Attack: 5, Decay: 7, Sustain: 12, Release: 7, FilterFreq: 0.28, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Voice Oohs", Waveform: 2, PulseWidth: 1800, Attack: 4, Decay: 7, Sustain: 11, Release: 6, FilterFreq: 0.32, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Synth Voice", Waveform: 2, PulseWidth: 1600, Attack: 5, Decay: 7, Sustain: 12, Release: 7, FilterFreq: 0.3, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Orchestra Hit", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 0, Release: 3, FilterFreq: 0.4, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	// === Brass (56-63) ===
	{Name: "Trumpet", Waveform: 1, PulseWidth: 2048, Attack: 2, Decay: 6, Sustain: 10, Release: 4, FilterFreq: 0.35, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Trombone", Waveform: 1, PulseWidth: 2048, Attack: 3, Decay: 7, Sustain: 11, Release: 5, FilterFreq: 0.25, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Tuba", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 8, Sustain: 10, Release: 5, FilterFreq: 0.18, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Muted Trumpet", Waveform: 2, PulseWidth: 1800, Attack: 2, Decay: 5, Sustain: 9, Release: 4, FilterFreq: 0.2, FilterRes: 0.5, FilterMode: 1, FilterOn: true},
	{Name: "French Horn", Waveform: 1, PulseWidth: 2048, Attack: 3, Decay: 7, Sustain: 11, Release: 5, FilterFreq: 0.22, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Brass Section", Waveform: 1, PulseWidth: 2048, Attack: 2, Decay: 6, Sustain: 11, Release: 4, FilterFreq: 0.3, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Synth Brass 1", Waveform: 1, PulseWidth: 2048, Attack: 1, Decay: 5, Sustain: 10, Release: 4, FilterFreq: 0.38, FilterRes: 0.45, FilterMode: 1, FilterOn: true},
	{Name: "Synth Brass 2", Waveform: 1, PulseWidth: 2048, Attack: 2, Decay: 6, Sustain: 11, Release: 5, FilterFreq: 0.32, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	// === Reed (64-71) ===
	{Name: "Soprano Sax", Waveform: 2, PulseWidth: 512, Attack: 2, Decay: 5, Sustain: 10, Release: 4, FilterFreq: 0.38, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Alto Sax", Waveform: 2, PulseWidth: 768, Attack: 2, Decay: 5, Sustain: 10, Release: 5, FilterFreq: 0.32, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Tenor Sax", Waveform: 2, PulseWidth: 1024, Attack: 2, Decay: 6, Sustain: 11, Release: 5, FilterFreq: 0.28, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Baritone Sax", Waveform: 2, PulseWidth: 1400, Attack: 3, Decay: 6, Sustain: 10, Release: 5, FilterFreq: 0.22, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Oboe", Waveform: 2, PulseWidth: 384, Attack: 2, Decay: 5, Sustain: 10, Release: 5, FilterFreq: 0.32, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "English Horn", Waveform: 2, PulseWidth: 600, Attack: 2, Decay: 6, Sustain: 10, Release: 5, FilterFreq: 0.26, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Bassoon", Waveform: 2, PulseWidth: 1200, Attack: 3, Decay: 6, Sustain: 10, Release: 5, FilterFreq: 0.2, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Clarinet", Waveform: 2, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 11, Release: 5, FilterFreq: 0.26, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	// === Pipe (72-79) ===
	{Name: "Piccolo", Waveform: 0, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 12, Release: 4, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Flute", Waveform: 0, PulseWidth: 2048, Attack: 3, Decay: 5, Sustain: 12, Release: 5, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Recorder", Waveform: 0, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 12, Release: 5, FilterFreq: 0.38, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Pan Flute", Waveform: 0, PulseWidth: 2048, Attack: 4, Decay: 6, Sustain: 12, Release: 5, FilterFreq: 0.35, FilterRes: 0.1, FilterMode: 1, FilterOn: true},
	{Name: "Blown Bottle", Waveform: 0, PulseWidth: 2048, Attack: 4, Decay: 6, Sustain: 11, Release: 5, FilterFreq: 0.4, FilterRes: 0.25, FilterMode: 2, FilterOn: true},
	{Name: "Shakuhachi", Waveform: 1, PulseWidth: 2048, Attack: 3, Decay: 6, Sustain: 11, Release: 5, FilterFreq: 0.35, FilterRes: 0.3, FilterMode: 2, FilterOn: true},
	{Name: "Whistle", Waveform: 0, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 13, Release: 4, FilterFreq: 0.42, FilterRes: 0.1, FilterMode: 1, FilterOn: true},
	{Name: "Ocarina", Waveform: 0, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 12, Release: 5, FilterFreq: 0.35, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	// === Synth Lead (80-87) ===
	{Name: "Square Lead", Waveform: 2, PulseWidth: 2048, Attack: 2, Decay: 6, Sustain: 10, Release: 5, FilterFreq: 0.25, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Saw Lead", Waveform: 1, PulseWidth: 2048, Attack: 1, Decay: 6, Sustain: 12, Release: 5, FilterFreq: 0.35, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Calliope Lead", Waveform: 0, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 12, Release: 5, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Chiff Lead", Waveform: 2, PulseWidth: 1200, Attack: 0, Decay: 5, Sustain: 9, Release: 4, FilterFreq: 0.38, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Charang Lead", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 7, Sustain: 9, Release: 4, FilterFreq: 0.42, FilterRes: 0.5, FilterMode: 1, FilterOn: true},
	{Name: "Voice Lead", Waveform: 2, PulseWidth: 2400, Attack: 4, Decay: 7, Sustain: 11, Release: 6, FilterFreq: 0.28, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Fifths Lead", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 10, Release: 4, FilterFreq: 0.45, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Bass+Lead", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 10, Release: 4, FilterFreq: 0.3, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	// === Synth Pad (88-95) ===
	{Name: "New Age Pad", Waveform: 0, PulseWidth: 2048, Attack: 5, Decay: 9, Sustain: 14, Release: 8, FilterFreq: 0.35, FilterRes: 0.1, FilterMode: 1, FilterOn: true},
	{Name: "Warm Pad", Waveform: 0, PulseWidth: 2048, Attack: 5, Decay: 9, Sustain: 14, Release: 8, FilterFreq: 0.3, FilterRes: 0.15, FilterMode: 1, FilterOn: true},
	{Name: "Polysynth Pad", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 8, Sustain: 13, Release: 7, FilterFreq: 0.35, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Choir Pad", Waveform: 2, PulseWidth: 2200, Attack: 5, Decay: 7, Sustain: 12, Release: 7, FilterFreq: 0.28, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Bowed Pad", Waveform: 2, PulseWidth: 1600, Attack: 5, Decay: 8, Sustain: 13, Release: 7, FilterFreq: 0.32, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Metallic Pad", Waveform: 0, PulseWidth: 2048, Attack: 6, Decay: 9, Sustain: 13, Release: 8, FilterFreq: 0.4, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Halo Pad", Waveform: 0, PulseWidth: 2048, Attack: 5, Decay: 8, Sustain: 13, Release: 7, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Sweep Pad", Waveform: 1, PulseWidth: 2048, Attack: 5, Decay: 8, Sustain: 13, Release: 7, FilterFreq: 0.28, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	// === Synth Effects (96-103) ===
	{Name: "FX Rain", Waveform: 3, PulseWidth: 2048, Attack: 4, Decay: 6, Sustain: 7, Release: 6, FilterFreq: 0.38, FilterRes: 0.35, FilterMode: 2, FilterOn: true},
	{Name: "FX Soundtrack", Waveform: 1, PulseWidth: 2048, Attack: 5, Decay: 8, Sustain: 12, Release: 7, FilterFreq: 0.25, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "FX Crystal", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 2, Release: 5, FilterFreq: 0.45, FilterRes: 0.3, FilterMode: 3, FilterOn: true},
	{Name: "FX Atmosphere", Waveform: 0, PulseWidth: 2048, Attack: 5, Decay: 8, Sustain: 13, Release: 7, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "FX Brightness", Waveform: 2, PulseWidth: 1200, Attack: 0, Decay: 7, Sustain: 8, Release: 5, FilterFreq: 0.5, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "FX Goblins", Waveform: 2, PulseWidth: 1600, Attack: 3, Decay: 7, Sustain: 9, Release: 6, FilterFreq: 0.35, FilterRes: 0.4, FilterMode: 2, FilterOn: true},
	{Name: "FX Echoes", Waveform: 1, PulseWidth: 2048, Attack: 4, Decay: 8, Sustain: 10, Release: 8, FilterFreq: 0.25, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "FX Sci-Fi", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 8, Release: 4, FilterFreq: 0.4, FilterRes: 0.5, FilterMode: 3, FilterOn: true},
	// === Ethnic (104-111) ===
	{Name: "Sitar", Waveform: 1, PulseWidth: 2048, Attack: 0, Decay: 7, Sustain: 3, Release: 5, FilterFreq: 0.35, FilterRes: 0.5, FilterMode: 2, FilterOn: true},
	{Name: "Banjo", Waveform: 2, PulseWidth: 800, Attack: 0, Decay: 3, Sustain: 0, Release: 2, FilterFreq: 0.45, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Shamisen", Waveform: 2, PulseWidth: 512, Attack: 0, Decay: 4, Sustain: 0, Release: 2, FilterFreq: 0.4, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Koto", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 6, Sustain: 0, Release: 4, FilterFreq: 0.38, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Kalimba", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 0, Release: 3, FilterFreq: 0.4, FilterRes: 0.2, FilterMode: 1, FilterOn: true},
	{Name: "Bagpipe", Waveform: 2, PulseWidth: 1400, Attack: 3, Decay: 6, Sustain: 14, Release: 5, FilterFreq: 0.3, FilterRes: 0.35, FilterMode: 1, FilterOn: true},
	{Name: "Fiddle", Waveform: 1, PulseWidth: 2048, Attack: 2, Decay: 6, Sustain: 11, Release: 5, FilterFreq: 0.35, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Shanai", Waveform: 2, PulseWidth: 384, Attack: 2, Decay: 5, Sustain: 11, Release: 5, FilterFreq: 0.35, FilterRes: 0.45, FilterMode: 2, FilterOn: true},
	// === Percussive (112-119) ===
	{Name: "Tinkle Bell", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 0, Release: 3, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Agogo", Waveform: 2, PulseWidth: 2048, Attack: 0, Decay: 3, Sustain: 0, Release: 2, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Steel Drums", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 1, Release: 3, FilterFreq: 0.42, FilterRes: 0.35, FilterMode: 2, FilterOn: true},
	{Name: "Woodblock", Waveform: 2, PulseWidth: 512, Attack: 0, Decay: 1, Sustain: 0, Release: 0, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Taiko Drum", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 0, Release: 3, FilterFreq: 0.1, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Melodic Tom", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 5, Sustain: 0, Release: 3, FilterFreq: 0.2, FilterRes: 0.25, FilterMode: 1, FilterOn: true},
	{Name: "Synth Drum", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 4, Sustain: 0, Release: 2, FilterFreq: 0.15, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Reverse Cymbal", Waveform: 3, PulseWidth: 2048, Attack: 6, Decay: 4, Sustain: 8, Release: 2, FilterFreq: 0.5, FilterRes: 0.3, FilterMode: 3, FilterOn: true},
	// === Sound FX (120-127) ===
	{Name: "Fret Noise", Waveform: 3, PulseWidth: 2048, Attack: 0, Decay: 3, Sustain: 0, Release: 2, FilterFreq: 0.5, FilterRes: 0.3, FilterMode: 3, FilterOn: true},
	{Name: "Breath Noise", Waveform: 3, PulseWidth: 2048, Attack: 2, Decay: 5, Sustain: 6, Release: 5, FilterFreq: 0.3, FilterRes: 0.3, FilterMode: 1, FilterOn: true},
	{Name: "Seashore", Waveform: 3, PulseWidth: 2048, Attack: 5, Decay: 7, Sustain: 9, Release: 7, FilterFreq: 0.25, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
	{Name: "Bird Tweet", Waveform: 0, PulseWidth: 2048, Attack: 0, Decay: 2, Sustain: 0, Release: 1, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Telephone", Waveform: 2, PulseWidth: 2048, Attack: 0, Decay: 2, Sustain: 10, Release: 1, FilterFreq: 0.5, FilterRes: 0, FilterMode: 0, FilterOn: false},
	{Name: "Helicopter", Waveform: 3, PulseWidth: 2048, Attack: 3, Decay: 4, Sustain: 10, Release: 4, FilterFreq: 0.2, FilterRes: 0.5, FilterMode: 1, FilterOn: true},
	{Name: "Applause", Waveform: 3, PulseWidth: 2048, Attack: 3, Decay: 6, Sustain: 10, Release: 6, FilterFreq: 0.4, FilterRes: 0.2, FilterMode: 2, FilterOn: true},
	{Name: "Gunshot", Waveform: 3, PulseWidth: 2048, Attack: 0, Decay: 2, Sustain: 0, Release: 1, FilterFreq: 0.45, FilterRes: 0.4, FilterMode: 1, FilterOn: true},
}

// PATCHES is an alias for GM_PATCHES for backward compatibility
var PATCHES = GM_PATCHES

// DRUM_PATCHES maps GM percussion note numbers (35-81) to drum sounds
// Matching the JavaScript DRUM_PATCHES exactly
var DRUM_PATCHES = map[int]DrumPatch{
	35: {Name: "Kick2", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 22, FilterMode: 1, Cutoff: 0.08, Res: 0.3},
	36: {Name: "Kick", Waveform: 0, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 24, FilterMode: 1, Cutoff: 0.1, Res: 0.4},
	37: {Name: "Sidestick", Waveform: 3, Attack: 0, Decay: 1, Sustain: 0, Release: 0, Note: 72, FilterMode: 3, Cutoff: 0.6, Res: 0.3},
	38: {Name: "Snare", Waveform: 3, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 60, FilterMode: 2, Cutoff: 0.35, Res: 0.2},
	39: {Name: "Clap", Waveform: 3, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 64, FilterMode: 3, Cutoff: 0.5, Res: 0.3},
	40: {Name: "Snare2", Waveform: 3, Attack: 0, Decay: 5, Sustain: 0, Release: 3, Note: 58, FilterMode: 2, Cutoff: 0.3, Res: 0.2},
	41: {Name: "Low Tom", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 36, FilterMode: 1, Cutoff: 0.2, Res: 0.3},
	42: {Name: "CHH", Waveform: 3, Attack: 0, Decay: 1, Sustain: 0, Release: 0, Note: 90, FilterMode: 3, Cutoff: 0.7, Res: 0.4},
	43: {Name: "Low Tom2", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 38, FilterMode: 1, Cutoff: 0.2, Res: 0.3},
	44: {Name: "PHH", Waveform: 3, Attack: 0, Decay: 2, Sustain: 0, Release: 1, Note: 88, FilterMode: 3, Cutoff: 0.65, Res: 0.4},
	45: {Name: "Mid Tom", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 43, FilterMode: 1, Cutoff: 0.2, Res: 0.3},
	46: {Name: "OHH", Waveform: 3, Attack: 0, Decay: 5, Sustain: 0, Release: 3, Note: 86, FilterMode: 3, Cutoff: 0.6, Res: 0.5},
	47: {Name: "Mid Tom2", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 45, FilterMode: 1, Cutoff: 0.2, Res: 0.3},
	48: {Name: "Hi Tom", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 50, FilterMode: 1, Cutoff: 0.25, Res: 0.3},
	49: {Name: "Crash", Waveform: 3, Attack: 0, Decay: 9, Sustain: 0, Release: 7, Note: 80, FilterMode: 2, Cutoff: 0.5, Res: 0.3},
	50: {Name: "Hi Tom2", Waveform: 0, Attack: 0, Decay: 4, Sustain: 0, Release: 2, Note: 52, FilterMode: 1, Cutoff: 0.25, Res: 0.3},
	51: {Name: "Ride", Waveform: 3, Attack: 0, Decay: 7, Sustain: 2, Release: 5, Note: 84, FilterMode: 3, Cutoff: 0.55, Res: 0.4},
	52: {Name: "China", Waveform: 3, Attack: 0, Decay: 8, Sustain: 0, Release: 6, Note: 78, FilterMode: 2, Cutoff: 0.45, Res: 0.5},
	53: {Name: "Ride Bell", Waveform: 3, Attack: 0, Decay: 6, Sustain: 3, Release: 4, Note: 82, FilterMode: 3, Cutoff: 0.6, Res: 0.5},
	54: {Name: "Tamb", Waveform: 3, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 92, FilterMode: 3, Cutoff: 0.75, Res: 0.3},
	55: {Name: "Splash", Waveform: 3, Attack: 0, Decay: 6, Sustain: 0, Release: 5, Note: 82, FilterMode: 2, Cutoff: 0.55, Res: 0.4},
	56: {Name: "Cowbell", Waveform: 2, Attack: 0, Decay: 3, Sustain: 0, Release: 1, Note: 68, FilterMode: 0, Cutoff: 0.5, Res: 0},
	57: {Name: "Crash2", Waveform: 3, Attack: 0, Decay: 9, Sustain: 0, Release: 7, Note: 76, FilterMode: 2, Cutoff: 0.45, Res: 0.3},
	58: {Name: "Vibraslap", Waveform: 3, Attack: 0, Decay: 5, Sustain: 0, Release: 4, Note: 70, FilterMode: 2, Cutoff: 0.4, Res: 0.6},
	59: {Name: "Ride2", Waveform: 3, Attack: 0, Decay: 7, Sustain: 2, Release: 5, Note: 86, FilterMode: 3, Cutoff: 0.58, Res: 0.4},
	60: {Name: "Hi Bongo", Waveform: 0, Attack: 0, Decay: 2, Sustain: 0, Release: 1, Note: 60, FilterMode: 1, Cutoff: 0.3, Res: 0.4},
	61: {Name: "Lo Bongo", Waveform: 0, Attack: 0, Decay: 3, Sustain: 0, Release: 1, Note: 48, FilterMode: 1, Cutoff: 0.25, Res: 0.4},
	62: {Name: "Mute Conga", Waveform: 0, Attack: 0, Decay: 1, Sustain: 0, Release: 1, Note: 55, FilterMode: 1, Cutoff: 0.2, Res: 0.3},
	63: {Name: "Hi Conga", Waveform: 0, Attack: 0, Decay: 3, Sustain: 0, Release: 1, Note: 57, FilterMode: 1, Cutoff: 0.28, Res: 0.35},
	64: {Name: "Lo Conga", Waveform: 0, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 46, FilterMode: 1, Cutoff: 0.22, Res: 0.35},
	65: {Name: "Hi Timbale", Waveform: 0, Attack: 0, Decay: 3, Sustain: 0, Release: 1, Note: 62, FilterMode: 1, Cutoff: 0.3, Res: 0.3},
	66: {Name: "Lo Timbale", Waveform: 0, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 50, FilterMode: 1, Cutoff: 0.25, Res: 0.3},
	67: {Name: "Hi Agogo", Waveform: 2, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 72, FilterMode: 0, Cutoff: 0.5, Res: 0},
	68: {Name: "Lo Agogo", Waveform: 2, Attack: 0, Decay: 3, Sustain: 0, Release: 2, Note: 64, FilterMode: 0, Cutoff: 0.5, Res: 0},
	69: {Name: "Cabasa", Waveform: 3, Attack: 0, Decay: 1, Sustain: 0, Release: 0, Note: 95, FilterMode: 3, Cutoff: 0.8, Res: 0.3},
	70: {Name: "Maracas", Waveform: 3, Attack: 0, Decay: 1, Sustain: 0, Release: 1, Note: 94, FilterMode: 3, Cutoff: 0.75, Res: 0.35},
	75: {Name: "Claves", Waveform: 2, Attack: 0, Decay: 1, Sustain: 0, Release: 0, Note: 76, FilterMode: 0, Cutoff: 0.5, Res: 0},
	76: {Name: "Hi Woodblk", Waveform: 2, Attack: 0, Decay: 1, Sustain: 0, Release: 0, Note: 74, FilterMode: 0, Cutoff: 0.5, Res: 0},
	77: {Name: "Lo Woodblk", Waveform: 2, Attack: 0, Decay: 2, Sustain: 0, Release: 1, Note: 66, FilterMode: 0, Cutoff: 0.5, Res: 0},
	80: {Name: "Mute Tri", Waveform: 0, Attack: 0, Decay: 2, Sustain: 0, Release: 1, Note: 80, FilterMode: 0, Cutoff: 0.5, Res: 0},
	81: {Name: "Open Tri", Waveform: 0, Attack: 0, Decay: 6, Sustain: 3, Release: 5, Note: 80, FilterMode: 0, Cutoff: 0.5, Res: 0},
}

// GetPatch returns a patch by index
func GetPatch(index int) *Patch {
	if index < 0 || index >= len(GM_PATCHES) {
		return &GM_PATCHES[0]
	}
	return &GM_PATCHES[index]
}

// GetPatchName returns the name of a patch by index
func GetPatchName(index int) string {
	if index < 0 || index >= len(GM_PATCHES) {
		return "Unknown"
	}
	return GM_PATCHES[index].Name
}

// PatchCount returns the total number of patches
func PatchCount() int {
	return len(GM_PATCHES)
}
