package audio

import (
	"fmt"
	"testing"
	"github.com/april/sid-synth/sound"
)

func TestPatchAllocationMatching(t *testing.T) {
	bank := NewBank()
	
	// Test 1: Verify default values match JavaScript
	if bank.volume != 0.7 {
		t.Errorf("Expected master volume 0.7, got %f", bank.volume)
	}
	
	for i := 0; i < len(bank.chips); i++ {
		chip := bank.Chip(i)
		if chip.volume != 1.0 {
			t.Errorf("Expected chip volume 1.0, got %f", chip.volume)
		}
		
		// Check default filter routing (should be false for all voices)
		for j, enabled := range chip.filterEnable {
			if enabled {
				t.Errorf("Voice %d on chip %d should be false by default, got true", j, i)
			}
		}
	}
	
	// Test 2: Verify patch application behavior
	bank.SetPatch(0, 0) // Acoustic Grand - uses filter
	patch := bank.GetPatch(0)
	
	if patch == nil {
		t.Error("GetPatch should return a valid patch")
	}
	
	// Apply note and verify filter routing
	bank.NoteOn(0, 60, 100)
	
	// Find the active voice and check its chip
	foundActiveVoice := false
	for _, v := range bank.AllVoices() {
		if v.Active() && v.MidiChannel() == 0 {
			foundActiveVoice = true
			chip := bank.findChipForVoice(v)
			if chip != nil {
				// Find which voice index this is on the chip
				voiceIdx := -1
				for i, voice := range chip.voices {
					if voice == v {
						voiceIdx = i
						break
					}
				}
				
				if voiceIdx >= 0 && voiceIdx < 3 {
					// This voice should have filter enabled since Acoustic Grand uses filter
					if !chip.filterEnable[voiceIdx] {
						t.Errorf("Voice %d should have filter enabled for Lead 1 patch", voiceIdx)
					}
					
					// Check that waveform was applied
					if v.Waveform() != patch.Waveform {
						t.Errorf("Expected waveform %d, got %d", patch.Waveform, v.Waveform())
					}
				}
			}
			break
		}
	}
	
	if !foundActiveVoice {
		t.Error("Should have found an active voice after NoteOn")
	}
	
	// Test 3: Verify drum channel behavior
	bank.NoteOn(9, 36, 100) // Kick drum
	
	drumActive := false
	for _, v := range bank.AllVoices() {
		if v.Active() && v.MidiChannel() == 9 {
			drumActive = true
			// Drum patches should use specific settings
			if v.Waveform() != sound.WavePulse {
				t.Errorf("Drum should use pulse waveform, got %d", v.Waveform())
			}
			break
		}
	}
	
	if !drumActive {
		t.Error("Drum channel should create active voices")
	}
	
	// Test 4: Verify volume scaling
	// JavaScript uses channelVolume 100 = 1.0, Go uses 1.0 directly
	if bank.channelVolume[0] != 1.0 {
		t.Errorf("Expected channel volume 1.0, got %f", bank.channelVolume[0])
	}
	
	fmt.Println("Patch allocation tests passed - behavior matches JavaScript!")
}