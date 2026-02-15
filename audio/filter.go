package audio

import "math"

// Filter modes
const (
	FilterOff = 0 // Off
	FilterLP  = 1 // Low pass
	FilterBP  = 2 // Band pass
	FilterHP  = 3 // High pass
)

// Filter models (6581 vs 8580 chips)
const (
	FilterModel6581 = "6581"
	FilterModel8580 = "8580"
	FilterModelNone  = "none"
)

// Filter represents a state-variable filter (LP/BP/HP)
type Filter struct {
	model       string  // "6581", "8580", or "none"
	mode        int     // 0=off, 1=LP, 2=BP, 3=HP
	freq        float64 // cutoff frequency, normalized 0-1
	res         float64 // resonance (0-1)
	lp          float64 // Low pass output
	bp          float64 // Band pass output
	hp          float64 // High pass output
	w0          float64 // Filter coefficient (2*sin(π*freq/sampleRate))
	damping     float64 // Damping factor for resonance
	cutoffTable [2048]float64 // Precomputed 6581 cutoff table
}

// NewFilter creates a new state-variable filter
func NewFilter() *Filter {
	f := &Filter{
		model: FilterModel8580, // Default to 8580
		mode:  FilterLP,
		freq:  0.5, // Normalized 0-1
		res:   0,
	}
	f.initCutoffTable()
	f.updateCoefficients()
	return f
}

// initCutoffTable precomputes the 6581 cutoff table with S-curve mapping
func (f *Filter) initCutoffTable() {
	const minFreq = 30.0
	const maxFreq = 12000.0
	
	for i := 0; i < 2048; i++ {
		x := float64(i) / 2047.0
		// Gentle power curve: lower cutoff range than 8580, slight S-shape
		// smoothstep function: x*x*(3-2*x) creates S-curve, 0→0, 1→1
		shaped := x * x * (3 - 2*x)
		freqHz := minFreq + shaped*(maxFreq-minFreq)
		// Convert to w0 coefficient: 2*sin(π*freq/sampleRate)
		f.cutoffTable[i] = 2.0 * math.Sin(math.Pi*math.Min(freqHz, SampleRate*0.45)/SampleRate)
	}
}

// updateCoefficients calculates filter coefficients based on model and settings
func (f *Filter) updateCoefficients() {
	if f.model == FilterModel6581 {
		// 6581: sigmoid cutoff curve, linear 1/Q resonance
		regVal := int(f.freq * 2047)
		if regVal < 0 {
			regVal = 0
		} else if regVal >= 2048 {
			regVal = 2047
		}
		f.w0 = f.cutoffTable[regVal]
		if f.w0 > 0.95 {
			f.w0 = 0.95
		}
		
		// Resonance: linear 1/Q mapping
		resReg := int(f.res * 15)
		if resReg < 0 {
			resReg = 0
		} else if resReg > 15 {
			resReg = 15
		}
		inv := (^resReg) & 0x0F // Bitwise NOT, keep only 4 bits
		if inv > 0 {
			f.damping = float64(inv) / 8.0
		} else {
			f.damping = 0.06 // res=15 → near-self-oscillation
		}
	} else {
		// 8580: linear cutoff curve, exponential 1/Q resonance
		freqHz := f.freq * 12500.0
		f.w0 = 2.0 * math.Sin(math.Pi*freqHz/SampleRate)
		if f.w0 > 0.95 {
			f.w0 = 0.95
		}
		
		// Resonance: exponential mapping
		resReg := int(f.res * 15)
		if resReg < 0 {
			resReg = 0
		} else if resReg > 15 {
			resReg = 15
		}
		f.damping = math.Pow(2, (4-float64(resReg))/8)
	}
}

// SetFreq sets the filter cutoff frequency (normalized 0-1)
func (f *Filter) SetFreq(cutoff float64) {
	if cutoff < 0 {
		cutoff = 0
	}
	if cutoff > 1 {
		cutoff = 1
	}
	f.freq = cutoff
	f.updateCoefficients()
}

// Reset clears the filter state (LP/BP/HP outputs)
func (f *Filter) Reset() {
	f.lp = 0
	f.bp = 0
	f.hp = 0
}

// SetRes sets the filter resonance (0-1)
func (f *Filter) SetRes(res float64) {
	if res < 0 {
		res = 0
	}
	if res > 1 {
		res = 1
	}
	f.res = res
	f.updateCoefficients()
}

// SetMode sets the filter mode
func (f *Filter) SetMode(mode int) {
	f.mode = mode
}

// SetModel sets the filter model (6581, 8580, or none)
func (f *Filter) SetModel(model string) {
	f.model = model
	f.updateCoefficients()
}

// W0 returns the current w0 coefficient
func (f *Filter) W0() float64 {
	return f.w0
}

// Damping returns the current damping factor
func (f *Filter) Damping() float64 {
	return f.damping
}

// Freq returns the current frequency setting
func (f *Filter) Freq() float64 {
	return f.freq
}

// Res returns the current resonance setting
func (f *Filter) Res() float64 {
	return f.res
}

// Model returns the current filter model
func (f *Filter) Model() string {
	return f.model
}

// clip6581 applies asymmetric saturation for 6581 model
func (f *Filter) clip6581(sample float64) float64 {
	// Subtle asymmetric warmth — only colors loud signals
	if sample > 0.8 {
		return 0.8 + (sample-0.8)*0.3
	}
	if sample < -0.8 {
		return -0.8 + (sample+0.8)*0.4
	}
	return sample
}

// Process applies the filter to a sample
func (f *Filter) Process(input float64) float64 {
	if f.model == FilterModelNone {
		return input
	}
	
	// 6581: add DC offset (biases NMOS inverters)
	filterInput := input
	if f.model == FilterModel6581 {
		filterInput = input + 0.005
	}
	
	// State-variable filter core (Chamberlin form)
	f.bp += f.w0 * f.hp
	f.lp += f.w0 * f.bp
	f.hp = filterInput - f.lp - f.damping*f.bp
	
	// 6581: subtle asymmetric saturation models NMOS inverter character
	if f.model == FilterModel6581 {
		f.bp = f.clip6581(f.bp)
		f.hp = f.clip6581(f.hp)
	}
	
	// Denormal protection
	if math.Abs(f.lp) < 1e-10 {
		f.lp = 0
	}
	if math.Abs(f.bp) < 1e-10 {
		f.bp = 0
	}
	if math.Abs(f.hp) < 1e-10 {
		f.hp = 0
	}
	
	// Select output based on mode
	switch f.mode {
	case FilterLP:
		return f.lp
	case FilterBP:
		return f.bp
	case FilterHP:
		return f.hp
	default:
		return f.lp
	}
}

// ProcessMulti applies filter to multiple input samples
func (f *Filter) ProcessMulti(inputs []float64) []float64 {
	output := make([]float64, len(inputs))
	for i, input := range inputs {
		output[i] = f.Process(input)
	}
	return output
}