package audio

// Chip represents a single SID chip with 3 voices and a filter
type Chip struct {
	voices       [3]*Voice
	filter       *Filter
	filterEnable [3]bool // Which voices route through filter
	filterFreq   float64
	filterRes    float64
	filterMode   int
	filterModel  string // "6581", "8580", or "none"
	volume       float64
	muted        bool
	lpOut        float64
	bpOut        float64
	hpOut        float64
}

// NewChip creates a new SID chip with 3 voices
func NewChip() *Chip {
	c := &Chip{
		volume:      1.0,             // Match JavaScript default
		filterMode:  FilterOff,       // Default to off
		filterModel: FilterModel8580, // Default to 8580
	}
	for i := 0; i < 3; i++ {
		c.voices[i] = NewVoice()
		c.filterEnable[i] = false // Match JavaScript: voices not filtered by default
	}
	c.filter = NewFilter()
	return c
}

// Voice returns the i-th voice
func (c *Chip) Voice(i int) *Voice {
	if i >= 0 && i < 3 {
		return c.voices[i]
	}
	return nil
}

// SetFilter configures the filter
func (c *Chip) SetFilter(freq float64, res float64, mode int, model string) {
	c.filterFreq = freq
	c.filterRes = res
	c.filterMode = mode
	c.filterModel = model
	c.filter.SetFreq(freq)
	c.filter.SetRes(res)
	c.filter.SetMode(mode)
	c.filter.SetModel(model)
}

// FilterModel returns the current filter model
func (c *Chip) FilterModel() string {
	return c.filterModel
}

// SetVolume sets the master volume
func (c *Chip) SetVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	c.volume = v
}

// SetMute mutes/unmutes the chip
func (c *Chip) SetMute(m bool) {
	c.muted = m
}

// GenerateSamples generates audio samples for this chip
// No per-chip soft clipping — clipping happens at bank level (matching JS)
func (c *Chip) GenerateSamples(n int) []float64 {
	samples := make([]float64, n)
	bypass := c.filterModel == FilterModelNone

	for i := 0; i < n; i++ {
		var filtered, direct float64

		// Auto-clear filter routing for inactive voices (matching JS)
		for v := 0; v < 3; v++ {
			if !c.voices[v].active && c.filterEnable[v] {
				c.filterEnable[v] = false
			}
		}

		// Mix voices with correct ring mod source mapping (matching JS)
		// JS: voice 0 → voice[2], voice 1 → voice[0], voice 2 → voice[1]
		for v := 0; v < 3; v++ {
			var ringModSrc *Voice
			if v == 0 {
				ringModSrc = c.voices[2]
			} else {
				ringModSrc = c.voices[v-1]
			}

			sample := c.voices[v].GenerateSample(ringModSrc)

			if !bypass && c.filterEnable[v] {
				filtered += sample
			} else {
				direct += sample
			}
		}

		// Apply filter (matching JS: only when not bypassed and filterMode > 0)
		if !bypass && c.filterMode > 0 {
			// Check if any voice is routed through filter
			anyRouted := c.filterEnable[0] || c.filterEnable[1] || c.filterEnable[2]
			if !anyRouted {
				// No routing — clear filter state
				c.filter.Reset()
				samples[i] = (filtered + direct) * c.volume
				continue
			}

			// Always process filter (even when filtered input is 0, filter state needs to decay)
			filterOut := c.filter.Process(filtered)
			samples[i] = (filterOut + direct) * c.volume
		} else {
			samples[i] = (filtered + direct) * c.volume
		}
	}

	return samples
}

// AllVoices returns all voices from this chip
func (c *Chip) AllVoices() []*Voice {
	return c.voices[:]
}

// FilterEnable returns the filter routing for each voice
func (c *Chip) FilterEnable() [3]bool {
	return c.filterEnable
}

// ActiveVoices returns the number of currently active voices
func (c *Chip) ActiveVoices() int {
	count := 0
	for _, v := range c.voices {
		if v.active {
			count++
		}
	}
	return count
}
