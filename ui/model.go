package ui

import (
	"time"

	"github.com/april/sid-synth/audio"
	"github.com/april/sid-synth/midi"
	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg drives periodic UI refresh
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model represents the application state
type Model struct {
	engine   *audio.Engine
	midiFile *midi.File
	player   *midi.Player

	filename    string
	playing     bool
	paused      bool
	progress    float64
	currentTick int64
	totalTicks  int64

	voices   [36]audio.VoiceInfo
	scopeBuf []float64

	volume      float64
	filterModel string

	channelMuted    [16]bool
	channelPatchNum [16]int
	channelPatchName [16]string

	width, height int

	autoPlay bool
	err      string
}

// New creates a new model
func New(filename string) (*Model, error) {
	engine, err := audio.NewEngine()
	if err != nil {
		return nil, err
	}

	m := &Model{
		engine:      engine,
		filename:    filename,
		volume:      0.7,
		filterModel: audio.FilterModel8580,
		scopeBuf:    make([]float64, 512),
	}

	engine.Bank().SetMasterVolume(m.volume)

	if filename != "" {
		if err := m.loadMIDI(filename); err != nil {
			m.err = err.Error()
		} else {
			m.autoPlay = true
		}
	}

	return m, nil
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	if m.autoPlay {
		m.play()
		m.autoPlay = false
	}
	return tickCmd()
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.refreshState()
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// refreshState reads latest data from audio engine
func (m *Model) refreshState() {
	bank := m.engine.Bank()

	// Voice snapshot
	m.voices = bank.VoiceSnapshot()

	// Oscilloscope
	m.engine.ReadScope(m.scopeBuf)

	// Volume and filter
	m.volume = bank.GetMasterVolume()
	m.filterModel = bank.GetFilterModel()

	// Channel state
	for i := 0; i < 16; i++ {
		m.channelMuted[i] = bank.IsChannelMuted(i)
		m.channelPatchNum[i] = bank.ChannelPatchNum(i)
		m.channelPatchName[i] = bank.PatchName(m.channelPatchNum[i])
	}

	// Playback progress
	if m.player != nil {
		m.progress = m.player.Progress()
		m.currentTick = m.player.CurrentTick()
		if m.playing && !m.player.IsPlaying() && !m.player.IsPaused() {
			// Playback finished
			m.playing = false
			m.paused = false
		}
	}
}

// handleKey dispatches keyboard input
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		m.cleanup()
		return m, tea.Quit

	case " ":
		if m.midiFile != nil {
			if m.playing {
				m.pause()
			} else {
				m.play()
			}
		}

	case "s":
		m.stop()

	case "r":
		if m.midiFile != nil {
			m.restart()
		}

	case "f":
		m.cycleFilter()

	case "+", "=":
		m.adjustVolume(0.05)

	case "-":
		m.adjustVolume(-0.05)

	case "left":
		m.seek(-5)

	case "right":
		m.seek(5)

	// Channel mute: 1-9,0 = ch 1-10 (MIDI 0-9)
	case "1":
		m.engine.Bank().ToggleMuteChannel(0)
	case "2":
		m.engine.Bank().ToggleMuteChannel(1)
	case "3":
		m.engine.Bank().ToggleMuteChannel(2)
	case "4":
		m.engine.Bank().ToggleMuteChannel(3)
	case "5":
		m.engine.Bank().ToggleMuteChannel(4)
	case "6":
		m.engine.Bank().ToggleMuteChannel(5)
	case "7":
		m.engine.Bank().ToggleMuteChannel(6)
	case "8":
		m.engine.Bank().ToggleMuteChannel(7)
	case "9":
		m.engine.Bank().ToggleMuteChannel(8)
	case "0":
		m.engine.Bank().ToggleMuteChannel(9)

	// Channel mute: shift+1 through shift+6 = ch 11-16 (MIDI 10-15)
	case "!":
		m.engine.Bank().ToggleMuteChannel(10)
	case "@":
		m.engine.Bank().ToggleMuteChannel(11)
	case "#":
		m.engine.Bank().ToggleMuteChannel(12)
	case "$":
		m.engine.Bank().ToggleMuteChannel(13)
	case "%":
		m.engine.Bank().ToggleMuteChannel(14)
	case "^":
		m.engine.Bank().ToggleMuteChannel(15)
	}

	return m, nil
}

// play starts MIDI playback
func (m *Model) play() {
	if m.midiFile == nil {
		return
	}

	if m.player == nil {
		m.player = midi.NewPlayer(m.midiFile)

		bank := m.engine.Bank()
		m.player.OnNoteOn = func(channel, note, velocity int) {
			bank.NoteOn(channel, note, velocity)
		}
		m.player.OnNoteOff = func(channel, note int) {
			bank.NoteOff(channel, note)
		}
		m.player.OnProgramChange = func(channel, program int) {
			bank.SetPatch(channel, program)
		}
		m.player.OnControlChange = func(channel, controller, value int) {
			if controller == 7 {
				bank.SetChannelVolume(channel, float64(value)/127.0)
			}
		}

		m.engine.SetMidiProcessor(m.player)
	}

	m.player.Play()
	m.playing = true
	m.paused = false

	if !m.engine.IsPlaying() {
		m.engine.Start()
	}
}

// pause pauses playback
func (m *Model) pause() {
	if m.player != nil {
		m.player.Pause()
	}
	m.playing = false
	m.paused = true
}

// stop stops playback
func (m *Model) stop() {
	if m.player != nil {
		m.player.Stop()
	}
	m.playing = false
	m.paused = false
	m.currentTick = 0
	m.progress = 0
	m.engine.Bank().AllNotesOff()
}

// restart restarts playback from beginning
func (m *Model) restart() {
	m.stop()
	if m.midiFile != nil {
		m.play()
	}
}

// cycleFilter cycles through filter models: 8580 -> 6581 -> off -> 8580
func (m *Model) cycleFilter() {
	bank := m.engine.Bank()
	switch m.filterModel {
	case audio.FilterModel8580:
		bank.SetFilterModel(audio.FilterModel6581)
		m.filterModel = audio.FilterModel6581
	case audio.FilterModel6581:
		bank.SetFilterModel(audio.FilterModelNone)
		m.filterModel = audio.FilterModelNone
	default:
		bank.SetFilterModel(audio.FilterModel8580)
		m.filterModel = audio.FilterModel8580
	}
}

// adjustVolume changes volume by delta, clamped to 0-1
func (m *Model) adjustVolume(delta float64) {
	m.volume += delta
	if m.volume > 1.0 {
		m.volume = 1.0
	}
	if m.volume < 0.0 {
		m.volume = 0.0
	}
	m.engine.Bank().SetMasterVolume(m.volume)
}

// seek moves playback position by seconds
func (m *Model) seek(seconds float64) {
	if m.player == nil || m.midiFile == nil {
		return
	}

	// Estimate ticks from seconds using initial tempo
	ticksPerSec := m.midiFile.TicksPerSecond()
	deltaTicks := int64(seconds * ticksPerSec)
	newTick := m.currentTick + deltaTicks

	if newTick < 0 {
		newTick = 0
	}
	if newTick > m.totalTicks {
		newTick = m.totalTicks
	}

	m.engine.Bank().AllNotesOff()
	m.player.SeekTo(newTick)
	m.currentTick = newTick
}

// loadMIDI loads a MIDI file
func (m *Model) loadMIDI(filename string) error {
	midiFile, err := midi.ParseFile(filename)
	if err != nil {
		return err
	}

	m.stop()
	m.engine.SetMidiProcessor(nil)

	m.midiFile = midiFile
	m.filename = filename
	m.player = nil
	m.playing = false
	m.paused = false
	m.currentTick = 0
	m.totalTicks = midiFile.DurationTicks()
	m.err = ""

	return nil
}

// cleanup stops engine on quit
func (m *Model) cleanup() {
	m.stop()
	m.engine.Close()
}
