package midi

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sort"
	"time"
)

// MIDI event types
const (
	NoteOff         = 0x80
	NoteOn          = 0x90
	Aftertouch      = 0xA0
	ControlChange   = 0xB0
	ProgramChange   = 0xC0
	ChannelPressure = 0xD0
	PitchBend       = 0xE0
)

// Meta event types
const (
	MetaTempo         = 0x51
	MetaEndOfTrack    = 0x2F
	MetaTimeSignature = 0x58
	MetaKeySignature  = 0x59
	MetaTrackName     = 0x03
)

// Event represents a MIDI event
type Event struct {
	Tick    int64 // Absolute tick position
	Channel int   // 0-15 (not used for meta events)
	Type    int   // Event type (e.g., NoteOn, NoteOff, etc.)
	Data1   int   // First data byte (note number, CC number, etc.)
	Data2   int   // Second data byte (velocity, CC value, etc.)
	Data3   int   // Third data byte (for meta events like tempo)
	IsMeta  bool  // True for meta events
}

// Track represents a MIDI track
type Track struct {
	Events []*Event
}

// File represents a parsed MIDI file
type File struct {
	Format      int // 0, 1, or 2
	NumTracks   int // Number of tracks
	Division    int // Ticks per quarter note (= ticksPerBeat)
	Tracks      []*Track
	Tempo       int // Initial tempo: microseconds per quarter note (default 500000)
	AllEvents   []*Event // All events merged from all tracks, sorted by tick
}

// ParseFile parses a MIDI file from a filename
func ParseFile(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

// Parse parses a MIDI file from an io.Reader
func Parse(r io.Reader) (*File, error) {
	mf := &File{}

	// Read header
	header := make([]byte, 14)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}

	// Check header magic "MThd"
	if string(header[0:4]) != "MThd" {
		return nil, errors.New("not a MIDI file")
	}

	// Header length (should be 6)
	headerLen := int(binary.BigEndian.Uint32(header[4:8]))
	if headerLen != 6 {
		return nil, errors.New("invalid header length")
	}

	// Format, tracks, division
	mf.Format = int(binary.BigEndian.Uint16(header[8:10]))
	mf.NumTracks = int(binary.BigEndian.Uint16(header[10:12]))
	mf.Division = int(binary.BigEndian.Uint16(header[12:14]) & 0x7FFF)

	// Default tempo (120 BPM = 500000 microseconds per quarter)
	mf.Tempo = 500000

	// Parse tracks
	mf.Tracks = make([]*Track, mf.NumTracks)

	for trackNum := 0; trackNum < mf.NumTracks; trackNum++ {
		track, err := parseTrack(r)
		if err != nil {
			return nil, err
		}
		mf.Tracks[trackNum] = track
	}

	// Merge all tracks into a single sorted event list (matching JS behavior)
	mf.AllEvents = mf.mergeEvents()

	// Find initial tempo from the merged events
	for _, evt := range mf.AllEvents {
		if evt.IsMeta && evt.Type == MetaTempo {
			mf.Tempo = (evt.Data1 << 16) | (evt.Data2 << 8) | evt.Data3
			break
		}
	}

	return mf, nil
}

// mergeEvents flattens all tracks into a single event list sorted by tick
// This matches the JS parser which does: allEvents.sort((a, b) => a.tick - b.tick)
func (mf *File) mergeEvents() []*Event {
	var all []*Event
	for _, track := range mf.Tracks {
		all = append(all, track.Events...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Tick < all[j].Tick
	})
	return all
}

// parseTrack parses a single MIDI track
func parseTrack(r io.Reader) (*Track, error) {
	track := &Track{}

	// Read track header
	header := make([]byte, 8)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}

	// Check track magic "MTrk"
	if string(header[0:4]) != "MTrk" {
		return nil, errors.New("invalid track header")
	}

	trackLen := int(binary.BigEndian.Uint32(header[4:8]))

	// Read track data
	data := make([]byte, trackLen)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	// Parse events
	track.Events = parseTrackEvents(data)

	return track, nil
}

// parseTrackEvents parses the events in a track
func parseTrackEvents(data []byte) []*Event {
	events := make([]*Event, 0)

	var runningStatus byte
	pos := 0

	for pos < len(data) {
		// Read delta time
		delta, consumed := readVariableLength(data[pos:])
		pos += consumed

		if pos >= len(data) {
			break
		}

		status := data[pos]

		// Check for running status
		if status&0x80 == 0 {
			// Running status - use previous status, don't advance pos
			status = runningStatus
		} else {
			pos++
			if status < 0xF0 {
				runningStatus = status
			}
		}

		eventType := status & 0xF0
		ch := int(status & 0x0F)

		switch {
		case status == 0xFF:
			// Meta event
			if pos >= len(data) {
				break
			}
			metaType := int(data[pos])
			pos++

			metaLen, consumed := readVariableLength(data[pos:])
			pos += consumed

			if pos+metaLen > len(data) {
				pos = len(data)
				break
			}

			if metaType == MetaTempo && metaLen == 3 {
				events = append(events, &Event{
					Tick:   int64(delta),
					IsMeta: true,
					Type:   MetaTempo,
					Data1:  int(data[pos]),
					Data2:  int(data[pos+1]),
					Data3:  int(data[pos+2]),
				})
			}
			// Skip other meta events (but still count them for delta accumulation)
			pos += metaLen

		case status == 0xF0 || status == 0xF7:
			// SysEx
			sysLen, consumed := readVariableLength(data[pos:])
			pos += consumed
			pos += sysLen

		case eventType == NoteOn:
			if pos+2 <= len(data) {
				note := int(data[pos])
				vel := int(data[pos+1])
				pos += 2
				if vel > 0 {
					events = append(events, &Event{
						Tick:    int64(delta),
						Type:    NoteOn,
						Channel: ch,
						Data1:   note,
						Data2:   vel,
					})
				} else {
					// Note on with velocity 0 = note off
					events = append(events, &Event{
						Tick:    int64(delta),
						Type:    NoteOff,
						Channel: ch,
						Data1:   note,
						Data2:   0,
					})
				}
			}

		case eventType == NoteOff:
			if pos+2 <= len(data) {
				events = append(events, &Event{
					Tick:    int64(delta),
					Type:    NoteOff,
					Channel: ch,
					Data1:   int(data[pos]),
					Data2:   int(data[pos+1]),
				})
				pos += 2
			}

		case eventType == ProgramChange:
			if pos+1 <= len(data) {
				events = append(events, &Event{
					Tick:    int64(delta),
					Type:    ProgramChange,
					Channel: ch,
					Data1:   int(data[pos]),
				})
				pos++
			}

		case eventType == ControlChange:
			if pos+2 <= len(data) {
				events = append(events, &Event{
					Tick:    int64(delta),
					Type:    ControlChange,
					Channel: ch,
					Data1:   int(data[pos]),
					Data2:   int(data[pos+1]),
				})
				pos += 2
			}

		case eventType == PitchBend:
			if pos+2 <= len(data) {
				events = append(events, &Event{
					Tick:    int64(delta),
					Type:    PitchBend,
					Channel: ch,
					Data1:   int(data[pos]),
					Data2:   int(data[pos+1]),
				})
				pos += 2
			}

		case eventType == Aftertouch:
			if pos+2 <= len(data) {
				pos += 2
			}

		case eventType == ChannelPressure:
			if pos+1 <= len(data) {
				pos++
			}

		default:
			// Unknown, try to skip
			pos++
		}
	}

	// Convert delta times to absolute ticks
	absTick := int64(0)
	for _, e := range events {
		absTick += e.Tick
		e.Tick = absTick
	}

	return events
}

// readVariableLength reads a variable-length quantity
func readVariableLength(data []byte) (int, int) {
	var value int
	pos := 0

	for pos < len(data) {
		b := data[pos]
		pos++

		value = (value << 7) | int(b&0x7F)

		if b&0x80 == 0 {
			break
		}
	}

	return value, pos
}

// TicksPerSecond returns the ticks per second at a given tempo
func (mf *File) TicksPerSecond() float64 {
	return (1000000.0 / float64(mf.Tempo)) * float64(mf.Division)
}

// Duration returns the total duration of the MIDI file
func (mf *File) Duration() time.Duration {
	maxTick := mf.DurationTicks()
	seconds := float64(maxTick) / mf.TicksPerSecond()
	return time.Duration(seconds * float64(time.Second))
}

// DurationTicks returns the total duration in ticks
func (mf *File) DurationTicks() int64 {
	var maxTick int64
	for _, event := range mf.AllEvents {
		if event.Tick > maxTick {
			maxTick = event.Tick
		}
	}
	return maxTick
}
