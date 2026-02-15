package audio

import (
	"math"

	"github.com/april/sid-synth/sound"
)

// Bank represents a bank of 12 SID chips (36 voices total)
type Bank struct {
	chips            []*Chip
	voicesPerChip    int
	voices           []*Voice // Flat list of all voices
	channelPatch     [16]int  // Program number for each MIDI channel (0-15)
	channelVolume    [16]float64
	channelPan       [16]float64
	mutedChannels    map[int]bool
	volume           float64
	activeVoiceCount int // Cached for adaptive mixing
}

// NewBank creates a new bank with 12 SID chips
func NewBank() *Bank {
	b := &Bank{
		voicesPerChip: 3,
		mutedChannels: make(map[int]bool),
		volume:        0.7, // Match JavaScript masterVolume
	}

	// Create 12 chips
	b.chips = make([]*Chip, 12)
	for i := 0; i < 12; i++ {
		b.chips[i] = NewChip()
	}

	// Flatten all voices
	b.voices = make([]*Voice, 12*3)
	idx := 0
	for _, chip := range b.chips {
		for _, v := range chip.voices {
			b.voices[idx] = v
			idx++
		}
	}

	// Initialize channel defaults - match JavaScript values
	for i := 0; i < 16; i++ {
		b.channelPatch[i] = 0
		b.channelVolume[i] = 1.0
		b.channelPan[i] = 0.5
	}

	return b
}

// Chip returns the i-th chip
func (b *Bank) Chip(i int) *Chip {
	if i >= 0 && i < len(b.chips) {
		return b.chips[i]
	}
	return nil
}

// AllVoices returns all voices in the bank
func (b *Bank) AllVoices() []*Voice {
	return b.voices
}

// AllChips returns all chips in the bank
func (b *Bank) AllChips() []*Chip {
	return b.chips
}

// findChipForVoice finds which chip a voice belongs to
func (b *Bank) findChipForVoice(voice *Voice) *Chip {
	for _, chip := range b.chips {
		for _, v := range chip.voices {
			if v == voice {
				return chip
			}
		}
	}
	return nil
}

// AllocateVoice finds an available voice for a given MIDI channel
func (b *Bank) AllocateVoice(midiChannel int) *Voice {
	patchNum := b.channelPatch[midiChannel]
	needsFilter := false
	if patchNum >= 0 && patchNum < len(sound.GM_PATCHES) {
		patch := sound.GM_PATCHES[patchNum]
		needsFilter = patch.FilterMode > 0 && patch.FilterOn
	}

	// TIER 1: Find free voice with best compatibility
	bestFree := -1
	bestFreeScore := -10000

	for i, v := range b.voices {
		if !v.active {
			score := 0

			chip := b.findChipForVoice(v)
			if chip != nil {
				chipHasActiveFilter := false
				for _, enabled := range chip.filterEnable {
					if enabled {
						chipHasActiveFilter = true
						break
					}
				}

				if needsFilter {
					if chipHasActiveFilter {
						score += 10
					}
					for _, other := range chip.voices {
						if other != v && other.active && other.midiChannel == midiChannel {
							score += 20
						}
					}
				} else {
					if !chipHasActiveFilter {
						score += 30
					}
				}

				freeCount := 0
				for _, vo := range chip.voices {
					if !vo.active {
						freeCount++
					}
				}
				score += freeCount
			}

			if score > bestFreeScore {
				bestFreeScore = score
				bestFree = i
			}
		}
	}

	if bestFree >= 0 {
		return b.voices[bestFree]
	}

	// TIER 2: Steal voice
	bestCandidate := -1
	bestScore := 10000

	for i, v := range b.voices {
		if v.active {
			score := 0

			if v.EnvPhase() == EnvRelease {
				score -= 10000
			}
			score -= int((1 - v.EnvValue()) * 5000)
			score -= v.Age() * 100

			if v.MidiChannel() == midiChannel {
				score += 2000
			}
			if v.MidiChannel() == 9 {
				score += 3000
			}

			chip := b.findChipForVoice(v)
			if chip != nil {
				chipHasActiveFilter := false
				for _, enabled := range chip.filterEnable {
					if enabled {
						chipHasActiveFilter = true
						break
					}
				}
				if !needsFilter && chipHasActiveFilter {
					score += 1000
				}
				if needsFilter && !chipHasActiveFilter {
					score += 500
				}
			}

			if score < bestScore {
				bestScore = score
				bestCandidate = i
			}
		}
	}

	if bestCandidate >= 0 {
		return b.voices[bestCandidate]
	}

	return b.voices[0]
}

// NoteOn triggers a note on a specific MIDI channel
func (b *Bank) NoteOn(channel, note, velocity int) {
	if channel < 0 || channel > 15 {
		return
	}
	if b.mutedChannels[channel] {
		return
	}

	// Handle drum channel (channel 9 / 10 in 1-based)
	if channel == 9 {
		b.drumNoteOn(note, velocity)
		return
	}

	// Get the patch for this channel
	patchNum := b.channelPatch[channel]
	var patch *sound.Patch
	if patchNum >= 0 && patchNum < len(sound.GM_PATCHES) {
		patch = &sound.GM_PATCHES[patchNum]
	} else {
		patch = &sound.GM_PATCHES[0]
	}

	voice := b.AllocateVoice(channel)
	voice.midiChannel = channel

	// Apply patch settings to voice
	voice.SetWaveform(patch.Waveform)
	voice.SetPulseWidth(uint16(patch.PulseWidth))
	voice.SetADSR(patch.Attack, patch.Decay, patch.Sustain, patch.Release)
	voice.GateOn(note, velocity, -1)

	// Apply channel volume
	voice.velocity *= b.channelVolume[channel]

	// Apply filter settings
	b.applyFilterToVoice(voice, patch)
}

// drumNoteOn handles drum channel note events using DRUM_PATCHES
func (b *Bank) drumNoteOn(note, velocity int) {
	drum, ok := sound.DRUM_PATCHES[note]
	if !ok {
		// No drum patch for this note, use a default snare-like sound
		drum = sound.DrumPatch{
			Name:       "Default",
			Waveform:   3, // noise
			Attack:     0,
			Decay:      3,
			Sustain:    0,
			Release:    2,
			Note:       note,
			FilterMode: 0,
			Cutoff:     0.5,
			Res:        0,
		}
	}

	voice := b.AllocateVoice(9)
	voice.midiChannel = 9

	voice.SetWaveform(drum.Waveform)
	if drum.Waveform == WavePulse {
		voice.SetPulseWidth(2048)
	}
	voice.SetADSR(drum.Attack, drum.Decay, drum.Sustain, drum.Release)
	voice.GateOn(drum.Note, velocity, note)

	// Apply filter for drum
	chip := b.findChipForVoice(voice)
	if chip != nil {
		voiceIdx := -1
		for i, v := range chip.voices {
			if v == voice {
				voiceIdx = i
				break
			}
		}
		if voiceIdx >= 0 && voiceIdx < 3 {
			needsFilter := drum.FilterMode > 0
			chip.filterEnable[voiceIdx] = needsFilter
			if needsFilter {
				hasOtherFilteredVoices := false
				for i, enabled := range chip.filterEnable {
					if i != voiceIdx && enabled {
						hasOtherFilteredVoices = true
						break
					}
				}
				if !hasOtherFilteredVoices {
					chip.SetFilter(drum.Cutoff, drum.Res, drum.FilterMode, chip.filterModel)
				}
			}
		}
	}
}

// applyFilterToVoice applies a patch's filter settings to the voice's chip
func (b *Bank) applyFilterToVoice(voice *Voice, patch *sound.Patch) {
	chip := b.findChipForVoice(voice)
	if chip == nil {
		return
	}

	voiceIdx := -1
	for i, v := range chip.voices {
		if v == voice {
			voiceIdx = i
			break
		}
	}

	if voiceIdx < 0 || voiceIdx >= 3 {
		return
	}

	needsFilter := patch.FilterMode > 0 && patch.FilterOn
	chip.filterEnable[voiceIdx] = needsFilter

	if needsFilter {
		hasOtherFilteredVoices := false
		for i, enabled := range chip.filterEnable {
			if i != voiceIdx && enabled {
				hasOtherFilteredVoices = true
				break
			}
		}
		if !hasOtherFilteredVoices {
			chip.SetFilter(patch.FilterFreq, patch.FilterRes, patch.FilterMode, chip.filterModel)
		}
	}
}

// NoteOff releases a note on a specific MIDI channel
func (b *Bank) NoteOff(channel, note int) {
	for _, v := range b.voices {
		if v.active && v.midiChannel == channel && v.midiNote == note {
			v.GateOff()
		}
	}
}

// AllNotesOff releases all active notes
func (b *Bank) AllNotesOff() {
	for _, v := range b.voices {
		if v.active {
			v.GateOff()
		}
	}
}

// SetPatch sets the program/patch for a channel
func (b *Bank) SetPatch(channel, patchNum int) {
	if channel < 0 || channel > 15 {
		return
	}
	if patchNum < 0 || patchNum >= len(sound.GM_PATCHES) {
		return
	}
	b.channelPatch[channel] = patchNum
}

// GetPatch returns the current patch for a channel
func (b *Bank) GetPatch(channel int) *sound.Patch {
	if channel < 0 || channel > 15 {
		return nil
	}
	patchNum := b.channelPatch[channel]
	if patchNum < 0 || patchNum >= len(sound.GM_PATCHES) {
		return nil
	}
	return &sound.GM_PATCHES[patchNum]
}

// PatchCount returns the total number of available patches
func (b *Bank) PatchCount() int {
	return len(sound.GM_PATCHES)
}

// PatchName returns the name of a patch by number
func (b *Bank) PatchName(patchNum int) string {
	if patchNum < 0 || patchNum >= len(sound.GM_PATCHES) {
		return "Unknown"
	}
	return sound.GM_PATCHES[patchNum].Name
}

// SetChannelVolume sets the volume for a channel
func (b *Bank) SetChannelVolume(channel int, volume float64) {
	if channel < 0 || channel > 15 {
		return
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	b.channelVolume[channel] = volume
}

// SetMasterVolume sets the master volume
func (b *Bank) SetMasterVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	b.volume = v
}

// SetMutedChannels sets which channels are muted
func (b *Bank) SetMutedChannels(muted map[int]bool) {
	b.mutedChannels = muted
}

// GenerateSamples generates mixed audio from all chips with adaptive mixing (matching JS)
func (b *Bank) GenerateSamples(n int) []float64 {
	samples := make([]float64, n)

	// Generate from each chip (no per-chip clipping — raw output)
	for _, chip := range b.chips {
		chipSamples := chip.GenerateSamples(n)
		for i := 0; i < n; i++ {
			samples[i] += chipSamples[i]
		}
	}

	// Update active voice count periodically (every buffer, ~46ms at 2048 samples)
	count := 0
	for _, v := range b.voices {
		if v.active {
			count++
		}
	}
	b.activeVoiceCount = count

	// Adaptive mixing: divide by sqrt of active voices (perceptual loudness scaling)
	// With a minimum divisor so a single voice isn't ear-splittingly loud
	divisor := math.Max(2, math.Sqrt(math.Max(1, float64(b.activeVoiceCount))))

	for i := 0; i < n; i++ {
		out := samples[i] / divisor * b.volume

		// Soft clip (true tanh — smooth at all levels, no hard knee)
		out = math.Tanh(out)

		samples[i] = out
	}

	return samples
}

// ActiveVoices returns total number of active voices
func (b *Bank) ActiveVoices() int {
	count := 0
	for _, v := range b.voices {
		if v.active {
			count++
		}
	}
	return count
}

// VoiceInfo holds a snapshot of voice state for UI display
type VoiceInfo struct {
	Active      bool
	Waveform    int
	EnvLevel    float64
	MidiNote    int
	MidiChannel int
}

// VoiceSnapshot reads all 36 voices into a snapshot array
func (b *Bank) VoiceSnapshot() [36]VoiceInfo {
	var snap [36]VoiceInfo
	for i, v := range b.voices {
		if i >= 36 {
			break
		}
		snap[i] = VoiceInfo{
			Active:      v.Active(),
			Waveform:    v.Waveform(),
			EnvLevel:    v.EnvValue(),
			MidiNote:    v.MidiNote(),
			MidiChannel: v.MidiChannel(),
		}
	}
	return snap
}

// ToggleMuteChannel toggles mute for a MIDI channel, gates off active notes.
// Returns the new muted state.
func (b *Bank) ToggleMuteChannel(ch int) bool {
	if ch < 0 || ch > 15 {
		return false
	}
	b.mutedChannels[ch] = !b.mutedChannels[ch]
	if b.mutedChannels[ch] {
		// Gate off all active notes on this channel
		for _, v := range b.voices {
			if v.active && v.midiChannel == ch {
				v.GateOff()
			}
		}
	}
	return b.mutedChannels[ch]
}

// IsChannelMuted returns whether a channel is muted
func (b *Bank) IsChannelMuted(ch int) bool {
	if ch < 0 || ch > 15 {
		return false
	}
	return b.mutedChannels[ch]
}

// SetFilterModel sets the filter model on all 12 chips
func (b *Bank) SetFilterModel(model string) {
	for _, chip := range b.chips {
		chip.SetFilter(chip.filterFreq, chip.filterRes, chip.filterMode, model)
	}
}

// GetFilterModel reads the filter model from chip 0
func (b *Bank) GetFilterModel() string {
	if len(b.chips) > 0 {
		return b.chips[0].FilterModel()
	}
	return FilterModel8580
}

// GetMasterVolume returns the master volume
func (b *Bank) GetMasterVolume() float64 {
	return b.volume
}

// ChannelPatchNum returns the program number for a channel
func (b *Bank) ChannelPatchNum(ch int) int {
	if ch < 0 || ch > 15 {
		return 0
	}
	return b.channelPatch[ch]
}
