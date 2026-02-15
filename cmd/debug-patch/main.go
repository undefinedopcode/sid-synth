package main

import (
	"fmt"
	"github.com/april/sid-synth/audio"
)

func main() {
	bank := audio.NewBank()
	
	// Set Lead 1 patch (uses filter)
	bank.SetPatch(0, 10)
	patch := bank.GetPatch(0)
	
	fmt.Printf("Patch: %s\n", patch.Name)
	fmt.Printf("FilterMode: %d\n", patch.FilterMode)
	
	// Trigger a note
	bank.NoteOn(0, 60, 100)
	
	// Check all voices
	for i, v := range bank.AllVoices() {
		if v.Active() {
			fmt.Printf("Voice %d is active\n", i)
			fmt.Printf("  MidiChannel: %d\n", v.MidiChannel())
			
			// Find the chip
			for chipIdx, chip := range bank.AllChips() {
				for voiceIdx, voice := range chip.AllVoices() {
					if voice == v {
						fmt.Printf("  Found on chip %d, voice %d\n", chipIdx, voiceIdx)
						fmt.Printf("  Filter enabled: %t\n", chip.FilterEnable()[voiceIdx])
						break
					}
				}
			}
			break
		}
	}
}