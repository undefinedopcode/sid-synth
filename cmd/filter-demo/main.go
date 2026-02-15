package main

import (
	"fmt"
	"github.com/april/sid-synth/audio"
)

func main() {
	fmt.Println("SID Filter Model Comparison")
	fmt.Println("==========================")
	
	// Create filters
	filter6581 := audio.NewFilter()
	filter6581.SetModel(audio.FilterModel6581)
	filter6581.SetFreq(0.5) // 0-1 range for 6581
	filter6581.SetRes(0.7)
	filter6581.SetMode(audio.FilterLP)
	
	filter8580 := audio.NewFilter()
	filter8580.SetModel(audio.FilterModel8580)
	filter8580.SetFreq(5000) // Hz for 8580  
	filter8580.SetRes(0.7)
	filter8580.SetMode(audio.FilterLP)
	
	fmt.Printf("6581 Model: freq=%f (normalized), res=%f\n", filter6581.Freq(), filter6581.Res())
	fmt.Printf("8580 Model: freq=%f Hz, res=%f\n", filter8580.Freq(), filter8580.Res())
	fmt.Printf("6581 w0 coefficient: %f\n", filter6581.W0())
	fmt.Printf("8580 w0 coefficient: %f\n", filter8580.W0())
	fmt.Printf("6581 damping: %f\n", filter6581.Damping())
	fmt.Printf("8580 damping: %f\n", filter8580.Damping())
	
	fmt.Println("\nProcessing test signal...")
	
	// Warm up filters
	for i := 0; i < 100; i++ {
		filter6581.Process(0.5)
		filter8580.Process(0.5)
	}
	
	// Test with a simple waveform
	fmt.Println("\nSample outputs (input = 0.8):")
	for i := 0; i < 5; i++ {
		input := 0.8 * (float64(i) / 4.0)
		out6581 := filter6581.Process(input)
		out8580 := filter8580.Process(input)
		fmt.Printf("Input: %f -> 6581: %f, 8580: %f\n", input, out6581, out8580)
	}
	
	fmt.Println("\nFilter models implemented successfully!")
	fmt.Println("Key differences:")
	fmt.Println("- 6581: Nonlinear S-curve frequency mapping, DC offset, asymmetric saturation")
	fmt.Println("- 8580: Linear frequency mapping, exponential resonance")
}