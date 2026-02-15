package audio

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/oto/v2"
)

const scopeBufSize = 4096

// MidiProcessor is called during audio generation for sample-accurate MIDI dispatch
type MidiProcessor interface {
	ProcessBlock(numSamples int)
}

// Engine manages real-time audio playback using Oto
type Engine struct {
	bank   *Bank
	player oto.Player
	ctx    *oto.Context
	mu     sync.Mutex

	sampleRate int
	channels   int
	format     int

	// Buffer for generating samples
	bufferSize int

	// MIDI processor (called during audio generation)
	midiProcessor MidiProcessor

	// Oscilloscope ring buffer
	scopeBuf   [scopeBufSize]float64
	scopeWrite uint64 // atomic write index

	// Control
	playing bool
}

// NewEngine creates a new audio engine
func NewEngine() (*Engine, error) {
	e := &Engine{
		sampleRate: 44100,
		channels:   1, // Mono output
		format:     oto.FormatSignedInt16LE,
		bufferSize: 2048,
	}

	// Create the bank
	e.bank = NewBank()

	// Initialize Oto with options
	ctx, ready, err := oto.NewContextWithOptions(&oto.NewContextOptions{
		SampleRate:   e.sampleRate,
		ChannelCount: e.channels,
		Format:       e.format,
		BufferSize:   0, // Use default
	})
	if err != nil {
		return nil, err
	}

	// Wait for the context to be ready
	<-ready
	e.ctx = ctx

	return e, nil
}

// Bank returns the SID bank
func (e *Engine) Bank() *Bank {
	return e.bank
}

// SetMidiProcessor sets the MIDI processor called during audio generation
func (e *Engine) SetMidiProcessor(mp MidiProcessor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.midiProcessor = mp
}

// Start starts audio playback
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.player != nil {
		return nil // Already playing
	}

	// Create a pipe for streaming samples
	pr, pw := io.Pipe()

	// Create player that reads from the pipe
	e.player = e.ctx.NewPlayer(pr)
	e.playing = true

	// Start the sample generator in a goroutine
	go e.generateAndWrite(pw)

	// Start playing
	e.player.Play()

	return nil
}

// generateAndWrite continuously generates samples and writes to the pipe
func (e *Engine) generateAndWrite(pw *io.PipeWriter) {
	defer pw.Close()

	buffer := make([]byte, e.bufferSize*2) // 2 bytes per sample (int16)

	for e.playing {
		// Process MIDI events for this buffer (sample-synchronized)
		if e.midiProcessor != nil {
			e.midiProcessor.ProcessBlock(e.bufferSize)
		}

		// Generate audio samples
		samples := e.bank.GenerateSamples(e.bufferSize)

		// Copy samples into oscilloscope ring buffer
		writeIdx := atomic.LoadUint64(&e.scopeWrite)
		for _, s := range samples {
			e.scopeBuf[writeIdx%scopeBufSize] = s
			writeIdx++
		}
		atomic.StoreUint64(&e.scopeWrite, writeIdx)

		// Convert to int16 and write to buffer
		for i, s := range samples {
			// Clamp to -1..1
			if s > 1 {
				s = 1
			}
			if s < -1 {
				s = -1
			}

			// Convert to int16
			sample := int16(s * 32767)

			// Write as little-endian
			buffer[i*2] = byte(sample)
			buffer[i*2+1] = byte(sample >> 8)
		}

		// Write to pipe
		_, err := pw.Write(buffer)
		if err != nil {
			break
		}
	}
}

// ReadScope copies the most recent samples from the oscilloscope ring buffer into dst.
// Returns the number of samples copied.
func (e *Engine) ReadScope(dst []float64) int {
	n := len(dst)
	if n > scopeBufSize {
		n = scopeBufSize
	}
	writeIdx := atomic.LoadUint64(&e.scopeWrite)
	for i := 0; i < n; i++ {
		idx := (writeIdx - uint64(n) + uint64(i)) % scopeBufSize
		dst[i] = e.scopeBuf[idx]
	}
	return n
}

// Stop stops audio playback
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.playing = false

	if e.player != nil {
		e.player.Pause()
		e.player = nil
	}
}

// IsPlaying returns true if audio is playing
func (e *Engine) IsPlaying() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.playing && e.player != nil && e.player.IsPlaying()
}

// Close closes the engine
func (e *Engine) Close() error {
	e.Stop()
	return nil
}
