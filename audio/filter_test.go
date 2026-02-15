package audio

import (
	"fmt"
	"math"
	"testing"
)

func TestFilterModels(t *testing.T) {
	// Test 6581 model
	filter6581 := NewFilter()
	filter6581.SetModel(FilterModel6581)
	filter6581.SetFreq(0.5) // 0-1 range for 6581
	filter6581.SetRes(0.5)
	filter6581.SetMode(FilterLP)
	
	// Test 8580 model
	filter8580 := NewFilter()
	filter8580.SetModel(FilterModel8580)
	filter8580.SetFreq(5000) // Hz for 8580
	filter8580.SetRes(0.5)
	filter8580.SetMode(FilterLP)
	
	// Warm up the filters with some samples
	for i := 0; i < 100; i++ {
		filter6581.Process(math.Sin(float64(i) * 0.1))
		filter8580.Process(math.Sin(float64(i) * 0.1))
	}
	
	// Test input signal
	input := 0.8
	
	// Process through both filters
	output6581 := filter6581.Process(input)
	output8580 := filter8580.Process(input)
	
	fmt.Printf("6581 output: %f\n", output6581)
	fmt.Printf("8580 output: %f\n", output8580)
	
	// Verify they produce different outputs (different characteristics)
	if math.Abs(output6581-output8580) < 0.001 {
		t.Error("6581 and 8580 models should produce different outputs")
	}
	
	// Test cutoff table
	if filter6581.cutoffTable[0] == 0 {
		t.Error("Cutoff table should be initialized")
	}
	
	// Test coefficient calculation
	if filter6581.w0 <= 0 || filter6581.w0 > 0.95 {
		t.Error("w0 coefficient should be in valid range")
	}
	
	// Test resonance mapping
	if filter6581.damping <= 0 {
		t.Error("Damping should be positive")
	}
	
	fmt.Println("Filter model tests passed!")
}

func TestFilterModes(t *testing.T) {
	filter := NewFilter()
	filter.SetModel(FilterModel8580)
	filter.SetFreq(2000)
	filter.SetRes(0.3)
	
	// Test different filter modes
	filter.SetMode(FilterLP)
	lpOutput := filter.Process(0.5)
	
	filter.SetMode(FilterBP)
	bpOutput := filter.Process(0.5)
	
	filter.SetMode(FilterHP)
	hpOutput := filter.Process(0.5)
	
	fmt.Printf("LP: %f, BP: %f, HP: %f\n", lpOutput, bpOutput, hpOutput)
	
	// Different modes should produce different outputs
	if math.Abs(lpOutput-bpOutput) < 0.001 && math.Abs(lpOutput-hpOutput) < 0.001 {
		t.Error("Different filter modes should produce different outputs")
	}
	
	fmt.Println("Filter mode tests passed!")
}

func TestFilterNoneModel(t *testing.T) {
	filter := NewFilter()
	filter.SetModel(FilterModelNone)
	
	input := 0.8
	output := filter.Process(input)
	
	// None model should pass through input unchanged
	if math.Abs(output-input) > 0.0001 {
		t.Errorf("None model should pass through input, got %f expected %f", output, input)
	}
	
	fmt.Println("Filter none model test passed!")
}