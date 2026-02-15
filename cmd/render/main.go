package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/april/sid-synth/audio"
	"github.com/april/sid-synth/midi"
)

const sampleRate = 44100

func main() {
	// CLI flags
	outputPath := flag.String("o", "", "output WAV file path (default: input with .wav extension)")
	volume := flag.Float64("vol", 0.7, "master volume 0.0-1.0")
	filterModel := flag.String("filter", "8580", "filter model: 6581, 8580, off")
	releaseTime := flag.Float64("release", 1.0, "seconds of release tail after last event")
	bufSize := flag.Int("buf", 2048, "render buffer size in samples")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: render [flags] <input.mid>\n\nRenders a MIDI file to WAV using SID synthesis.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)

	// Resolve output path
	out := *outputPath
	if out == "" {
		out = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".wav"
	}

	// Resolve filter model
	model := *filterModel
	switch strings.ToLower(model) {
	case "6581":
		model = audio.FilterModel6581
	case "8580":
		model = audio.FilterModel8580
	case "off", "none":
		model = audio.FilterModelNone
	default:
		log.Fatalf("Unknown filter model %q (use 6581, 8580, or off)", model)
	}

	// Clamp volume
	vol := *volume
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}

	// Parse MIDI
	fmt.Printf("Loading: %s\n", inputPath)
	midiFile, err := midi.ParseFile(inputPath)
	if err != nil {
		log.Fatalf("Error parsing MIDI: %v", err)
	}

	durationTicks := midiFile.DurationTicks()
	duration := midiFile.Duration()
	fmt.Printf("Duration: %s (%d ticks)\n", fmtDuration(duration.Seconds()), durationTicks)
	fmt.Printf("Filter: %s  Volume: %.0f%%\n", model, vol*100)

	// Create bank
	bank := audio.NewBank()
	bank.SetMasterVolume(vol)
	bank.SetFilterModel(model)

	// Create player with callbacks (same path as TUI)
	player := midi.NewPlayer(midiFile)
	player.OnNoteOn = func(channel, note, velocity int) {
		bank.NoteOn(channel, note, velocity)
	}
	player.OnNoteOff = func(channel, note int) {
		bank.NoteOff(channel, note)
	}
	player.OnProgramChange = func(channel, program int) {
		bank.SetPatch(channel, program)
	}
	player.OnControlChange = func(channel, controller, value int) {
		if controller == 7 {
			bank.SetChannelVolume(channel, float64(value)/127.0)
		}
	}
	player.Play()

	// Create WAV writer
	fmt.Printf("Output:  %s\n", out)
	wav, err := audio.NewWAVWriter(out, sampleRate, 1, 16)
	if err != nil {
		log.Fatalf("Error creating WAV: %v", err)
	}

	// Render using ProcessBlock (same code path as real-time engine)
	fmt.Print("Rendering...")
	totalSamples := int(duration.Seconds()*sampleRate) + int(*releaseTime*sampleRate)
	rendered := 0
	lastPct := -1

	for rendered < totalSamples {
		n := *bufSize
		if rendered+n > totalSamples {
			n = totalSamples - rendered
		}

		// Advance MIDI events for this buffer
		player.ProcessBlock(n)

		// Generate audio
		samples := bank.GenerateSamples(n)
		wav.WriteSamples(samples)
		rendered += n

		// Progress
		pct := rendered * 100 / totalSamples
		if pct != lastPct {
			fmt.Printf("\rRendering... %d%%", pct)
			lastPct = pct
		}
	}

	wav.Close()
	fmt.Printf("\rDone: %s\n", out)
}

func fmtDuration(sec float64) string {
	m := int(sec) / 60
	s := int(sec) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
