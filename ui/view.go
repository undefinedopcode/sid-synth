package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/april/sid-synth/audio"
	"github.com/charmbracelet/lipgloss"
)

// C64 color palette
var (
	C64Blue      = lipgloss.Color("#6c7ec4")
	C64Cyan      = lipgloss.Color("#77d4d4")
	C64Green     = lipgloss.Color("#6ec264")
	C64Yellow    = lipgloss.Color("#e3d65a")
	C64Red       = lipgloss.Color("#c4646c")
	C64White     = lipgloss.Color("#c5c5c5")
	C64DimGray   = lipgloss.Color("#555555")
	C64Border    = lipgloss.Color("#6c7ec4")
	C64DarkBg    = lipgloss.Color("#1a1a2e")

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(C64Border)

	titleStyle = lipgloss.NewStyle().
			Foreground(C64Cyan).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(C64DimGray)

	cyanStyle = lipgloss.NewStyle().
			Foreground(C64Cyan)

	greenStyle = lipgloss.NewStyle().
			Foreground(C64Green)

	yellowStyle = lipgloss.NewStyle().
			Foreground(C64Yellow)

	redStyle = lipgloss.NewStyle().
			Foreground(C64Red)

	whiteStyle = lipgloss.NewStyle().
			Foreground(C64White)
)

// waveformColor returns a lipgloss style for a waveform type
func waveformStyle(wf int) lipgloss.Style {
	switch wf {
	case audio.WaveTriangle:
		return greenStyle
	case audio.WaveSawtooth:
		return yellowStyle
	case audio.WavePulse:
		return cyanStyle
	case audio.WaveNoise:
		return redStyle
	default:
		return dimStyle
	}
}

// View renders the complete TUI
func (m *Model) View() string {
	w := m.width
	if w < 40 {
		w = 64
	}
	inner := w - 4 // border + padding

	var sections []string

	sections = append(sections, m.renderHeader(inner))
	sections = append(sections, m.renderScope(inner))
	sections = append(sections, m.renderVoices(inner))
	sections = append(sections, m.renderChannels(inner))
	sections = append(sections, m.renderProgress(inner))
	sections = append(sections, m.renderHelp(inner))

	if m.err != "" {
		sections = append(sections, redStyle.Render(" Error: "+m.err))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...) + "\n"
}

// renderHeader shows file info, filter model, volume
func (m *Model) renderHeader(width int) string {
	// Left: title + filename
	name := "no file"
	if m.filename != "" {
		name = filepath.Base(m.filename)
	}
	left := titleStyle.Render("SID SYNTH") + dimStyle.Render(" \u2500 ") + whiteStyle.Render(name)

	// Right: filter + volume
	filterLabel := m.filterModel
	if filterLabel == "none" {
		filterLabel = "off"
	}
	filter := dimStyle.Render("Filter: ") + cyanStyle.Render(filterLabel)

	volPct := int(m.volume * 100)
	volFilled := int(m.volume * 8)
	volBar := strings.Repeat("\u2588", volFilled) + strings.Repeat("\u2591", 8-volFilled)
	volume := dimStyle.Render("Vol: ") + cyanStyle.Render(volBar) + dimStyle.Render(fmt.Sprintf(" %d%%", volPct))

	right := filter + dimStyle.Render("  ") + volume

	// Duration
	var durLine string
	if m.midiFile != nil {
		dur := m.midiFile.Duration()
		durLine = dimStyle.Render("Duration: ") + whiteStyle.Render(fmtDuration(dur))
	}

	// Compose
	topLine := left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right))) + right
	if durLine != "" {
		topLine += "\n" + durLine
	}

	return borderStyle.Width(width).Padding(0, 1).Render(topLine)
}

// renderScope draws the braille oscilloscope
func (m *Model) renderScope(width int) string {
	scopeWidth := (width - 2) // chars available inside border
	if scopeWidth < 10 {
		scopeWidth = 10
	}
	scopeHeight := 4 // braille chars tall (= 16 dot rows)

	lines := RenderBraille(m.scopeBuf, scopeWidth, scopeHeight)

	// Color the waveform
	scopeContent := ""
	for i, line := range lines {
		scopeContent += cyanStyle.Render(line)
		if i < len(lines)-1 {
			scopeContent += "\n"
		}
	}

	header := dimStyle.Render("\u2500 SCOPE \u2500")

	return borderStyle.Width(width).Padding(0, 1).Render(header + "\n" + scopeContent)
}

// renderVoices draws the 36-voice activity grid
func (m *Model) renderVoices(width int) string {
	activeCount := 0
	for _, v := range m.voices {
		if v.Active {
			activeCount++
		}
	}

	header := dimStyle.Render(fmt.Sprintf("\u2500 VOICES (%d/36) \u2500", activeCount))

	// 12 columns (chips) x 3 rows (voices per chip)
	// Label row
	labels := ""
	for i := 0; i < 12; i++ {
		label := fmt.Sprintf("C%-3d", i)
		labels += dimStyle.Render(label)
	}

	// Voice indicators: 3 rows
	rows := make([]string, 3)
	for row := 0; row < 3; row++ {
		line := ""
		for col := 0; col < 12; col++ {
			idx := col*3 + row
			vi := m.voices[idx]
			var dot string
			if vi.Active {
				dot = waveformStyle(vi.Waveform).Render("\u25cf")
			} else {
				dot = dimStyle.Render("\u25cb")
			}
			// Pad to 4 chars wide
			line += dot + "   "
		}
		rows[row] = line
	}

	content := header + "\n" + labels + "\n" + strings.Join(rows, "\n")
	return borderStyle.Width(width).Padding(0, 1).Render(content)
}

// renderChannels shows the 16 MIDI channels with patch info and mute state
func (m *Model) renderChannels(width int) string {
	header := dimStyle.Render("\u2500 CHANNELS \u2500")

	// Channel numbers
	numLine := ""
	patchLine := ""

	for ch := 0; ch < 16; ch++ {
		chNum := fmt.Sprintf("%-4d", ch+1)
		muted := m.channelMuted[ch]

		if muted {
			numLine += dimStyle.Render(chNum)
		} else {
			numLine += whiteStyle.Render(chNum)
		}

		// Patch abbreviation (from snapshot)
		abbr := abbreviate(m.channelPatchName[ch])

		// Check if any voice on this channel is active
		chActive := false
		for _, v := range m.voices {
			if v.Active && v.MidiChannel == ch {
				chActive = true
				break
			}
		}

		padded := fmt.Sprintf("%-4s", abbr)
		if muted {
			patchLine += dimStyle.Render(padded)
		} else if chActive {
			patchLine += greenStyle.Render(padded)
		} else if abbr != "\u00b7" {
			patchLine += whiteStyle.Render(padded)
		} else {
			patchLine += dimStyle.Render(padded)
		}
	}

	content := header + "\n" + numLine + "\n" + patchLine
	return borderStyle.Width(width).Padding(0, 1).Render(content)
}

// renderProgress shows play status, time, and progress bar
func (m *Model) renderProgress(width int) string {
	// Status icon
	var icon, status string
	switch {
	case m.playing:
		icon = greenStyle.Render("\u25b6")
		status = greenStyle.Render(" PLAYING")
	case m.paused:
		icon = yellowStyle.Render("\u23f8")
		status = yellowStyle.Render(" PAUSED")
	default:
		icon = dimStyle.Render("\u25a0")
		status = dimStyle.Render(" STOPPED")
	}

	// Time display
	var elapsed, total time.Duration
	if m.midiFile != nil {
		total = m.midiFile.Duration()
		elapsed = time.Duration(m.progress * float64(total))
	}
	timeStr := dimStyle.Render(fmt.Sprintf("  %s / %s  ", fmtDuration(elapsed), fmtDuration(total)))

	// Progress bar
	barWidth := width - lipgloss.Width(icon) - lipgloss.Width(status) - lipgloss.Width(timeStr) - 8
	if barWidth < 10 {
		barWidth = 10
	}
	filled := int(m.progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := cyanStyle.Render(strings.Repeat("\u2501", filled)) +
		dimStyle.Render(strings.Repeat("\u2500", barWidth-filled))

	pctStr := dimStyle.Render(fmt.Sprintf(" %d%%", int(m.progress*100)))

	return " " + icon + status + timeStr + bar + pctStr
}

// renderHelp shows keybinding hints
func (m *Model) renderHelp(width int) string {
	_ = width
	line1 := dimStyle.Render(" [Space]") + whiteStyle.Render(" Play  ") +
		dimStyle.Render("[S]") + whiteStyle.Render(" Stop  ") +
		dimStyle.Render("[R]") + whiteStyle.Render(" Restart  ") +
		dimStyle.Render("[F]") + whiteStyle.Render(" Filter  ") +
		dimStyle.Render("[+/-]") + whiteStyle.Render(" Vol")

	line2 := dimStyle.Render(" [1-0]") + whiteStyle.Render(" Mute  ") +
		dimStyle.Render("[!-^]") + whiteStyle.Render(" Mute 11-16  ") +
		dimStyle.Render("[\u2190\u2192]") + whiteStyle.Render(" Seek  ") +
		dimStyle.Render("[Q]") + whiteStyle.Render(" Quit")

	return line1 + "\n" + line2
}

// abbreviate shortens a patch name to 3 chars
func abbreviate(name string) string {
	if name == "" || name == "Unknown" {
		return "\u00b7"
	}
	// Common abbreviations
	abbrevMap := map[string]string{
		"Acoustic Grand":  "APn",
		"Bright Acoustic": "BPn",
		"Electric Grand":  "EPn",
		"Electric Piano":  "EP",
		"Honky-Tonk":      "HnT",
		"Harpsichord":     "Hps",
		"Clavinet":        "Clv",
		"Nylon Guitar":    "NGt",
		"Steel Guitar":    "SGt",
		"Jazz Guitar":     "JGt",
		"Clean Guitar":    "CGt",
		"Muted Guitar":    "MGt",
		"Overdrive":       "OvD",
		"Distortion":      "Dst",
		"Acoustic Bass":   "ABs",
		"Finger Bass":     "FBs",
		"Slap Bass":       "SBs",
		"Synth Bass":      "SyB",
		"Violin":          "Vln",
		"Viola":           "Vla",
		"Cello":           "Vcl",
		"Strings":         "Str",
		"Trumpet":         "Tpt",
		"Trombone":        "Tbn",
		"French Horn":     "FHn",
		"Brass":           "Brs",
		"Soprano Sax":     "SSx",
		"Alto Sax":        "ASx",
		"Tenor Sax":       "TSx",
		"Flute":           "Flt",
		"Piccolo":         "Pic",
		"Recorder":        "Rec",
		"Pan Flute":       "PFl",
		"Organ":           "Org",
		"Church Organ":    "COr",
		"Drawbar Organ":   "DOr",
	}

	for key, abbr := range abbrevMap {
		if strings.Contains(name, key) {
			return abbr
		}
	}

	// Fallback: first 3 chars
	r := []rune(name)
	if len(r) > 3 {
		return string(r[:3])
	}
	return name
}

// fmtDuration formats a time.Duration as M:SS
func fmtDuration(d time.Duration) string {
	totalSec := int(d.Seconds())
	if totalSec < 0 {
		totalSec = 0
	}
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}

