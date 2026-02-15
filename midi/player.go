package midi

import "sync"

const sampleRate = 44100

// Player handles MIDI playback synchronized with audio generation.
// Instead of a timer-based approach, the audio engine calls ProcessBlock()
// during sample generation for sample-accurate event dispatch.
type Player struct {
	file *File

	// Playback state
	playing    bool
	paused     bool
	eventIndex int     // Current position in merged event list
	tickPos    float64 // Current fractional tick position
	tempo      int     // Current tempo in microseconds per quarter note
	samplesPerTick float64

	// Event callbacks
	OnNoteOn        func(channel, note, velocity int)
	OnNoteOff       func(channel, note int)
	OnProgramChange func(channel, program int)
	OnControlChange func(channel, controller, value int)
	OnPitchBend     func(channel, value int)
	OnFinished      func()

	mu sync.Mutex
}

// NewPlayer creates a new MIDI player
func NewPlayer(file *File) *Player {
	p := &Player{
		file:  file,
		tempo: file.Tempo,
	}
	p.computeSamplesPerTick()
	return p
}

// computeSamplesPerTick calculates how many audio samples per MIDI tick
// Formula: samplesPerTick = (tempo / 1_000_000) * sampleRate / ticksPerBeat
// This matches JS: ps.samplesPerTick = (ps.tempo / 1000000) * SAMPLE_RATE / midi.ticksPerBeat
func (p *Player) computeSamplesPerTick() {
	if p.file.Division > 0 {
		p.samplesPerTick = (float64(p.tempo) / 1000000.0) * float64(sampleRate) / float64(p.file.Division)
	} else {
		p.samplesPerTick = 1.0
	}
}

// Play starts playback (no goroutine — ProcessBlock drives timing)
func (p *Player) Play() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playing = true
	p.paused = false
}

// Pause pauses playback
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.paused = true
}

// Resume resumes playback
func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.paused = false
}

// Stop stops playback and resets to beginning
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playing = false
	p.paused = false
	p.eventIndex = 0
	p.tickPos = 0
	p.tempo = p.file.Tempo
	p.computeSamplesPerTick()
}

// SeekTo seeks to a tick position
func (p *Player) SeekTo(tick int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tickPos = float64(tick)

	// Find the event index for this tick position
	p.eventIndex = 0
	for p.eventIndex < len(p.file.AllEvents) {
		if p.file.AllEvents[p.eventIndex].Tick > tick {
			break
		}
		// Process program changes and tempo events we're skipping past
		evt := p.file.AllEvents[p.eventIndex]
		if evt.IsMeta && evt.Type == MetaTempo {
			p.tempo = (evt.Data1 << 16) | (evt.Data2 << 8) | evt.Data3
			p.computeSamplesPerTick()
		} else if evt.Type == ProgramChange && p.OnProgramChange != nil {
			p.OnProgramChange(evt.Channel, evt.Data1)
		}
		p.eventIndex++
	}
}

// SetTempo sets the tempo (microseconds per quarter note)
func (p *Player) SetTempo(tempo int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tempo = tempo
	p.computeSamplesPerTick()
}

// IsPlaying returns true if playing
func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.playing && !p.paused
}

// IsPaused returns true if paused
func (p *Player) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.paused
}

// CurrentTick returns the current tick position
func (p *Player) CurrentTick() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	return int64(p.tickPos)
}

// Duration returns the total duration in ticks
func (p *Player) Duration() int64 {
	if p.file == nil {
		return 0
	}
	return p.file.DurationTicks()
}

// Progress returns playback progress as 0-1
func (p *Player) Progress() float64 {
	dur := p.Duration()
	if dur == 0 {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.tickPos / float64(dur)
}

// ProcessBlock advances playback by numSamples audio samples.
// Called by the audio engine during buffer generation.
// Dispatches all MIDI events that fall within this sample range.
func (p *Player) ProcessBlock(numSamples int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.playing || p.paused {
		return
	}

	events := p.file.AllEvents

	for i := 0; i < numSamples; i++ {
		// Dispatch all events at or before the current tick position
		for p.eventIndex < len(events) {
			evt := events[p.eventIndex]
			if float64(evt.Tick) > p.tickPos {
				break
			}
			p.processEvent(evt)
			p.eventIndex++
		}

		// Advance tick position by one sample's worth
		if p.samplesPerTick > 0 {
			p.tickPos += 1.0 / p.samplesPerTick
		}

		// Check if we've reached the end
		if p.eventIndex >= len(events) {
			// Signal finished (caller can check for active voices to let them ring out)
			p.playing = false
			if p.OnFinished != nil {
				p.OnFinished()
			}
			return
		}
	}
}

// processEvent dispatches a single event to the appropriate callback
func (p *Player) processEvent(event *Event) {
	if event.IsMeta {
		if event.Type == MetaTempo {
			// Update tempo mid-song
			p.tempo = (event.Data1 << 16) | (event.Data2 << 8) | event.Data3
			p.computeSamplesPerTick()
		}
		return
	}

	switch event.Type {
	case NoteOn:
		if p.OnNoteOn != nil {
			p.OnNoteOn(event.Channel, event.Data1, event.Data2)
		}
	case NoteOff:
		if p.OnNoteOff != nil {
			p.OnNoteOff(event.Channel, event.Data1)
		}
	case ProgramChange:
		if p.OnProgramChange != nil {
			p.OnProgramChange(event.Channel, event.Data1)
		}
	case ControlChange:
		if p.OnControlChange != nil {
			p.OnControlChange(event.Channel, event.Data1, event.Data2)
		}
	case PitchBend:
		if p.OnPitchBend != nil {
			value := (event.Data2 << 7) | event.Data1
			p.OnPitchBend(event.Channel, value)
		}
	}
}

// ProcessEvent is the exported version for external callers
func (p *Player) ProcessEvent(event *Event) {
	p.processEvent(event)
}
