package audio

import (
	"fmt"
	"testing"
	"github.com/april/sid-synth/sound"
)

func TestPatchApplication(t *testing.T) {
	bank := NewBank()
	
	// Test setting a patch
	bank.SetPatch(0, 0) // Acoustic Grand
	patch := bank.GetPatch(0)
	
	if patch == nil {
		t.Error("GetPatch should return a patch")
	}
	
	if patch.Name != "Acoustic Grand" {
		t.Errorf("Expected patch name 'Acoustic Grand', got '%s'", patch.Name)
	}
	
	// Test patch application to voice
	bank.NoteOn(0, 60, 100) // Middle C on channel 0
	
	// Check that at least one voice is active
	activeVoices := 0
	for _, v := range bank.AllVoices() {
		if v.Active() {
			activeVoices++
			// Check that patch settings were applied
			if v.Waveform() != patch.Waveform {
				t.Errorf("Expected waveform %d, got %d", patch.Waveform, v.Waveform())
			}
			if v.PulseWidth() != uint16(patch.PulseWidth) {
				t.Errorf("Expected pulse width %d, got %d", patch.PulseWidth, v.PulseWidth())
			}
		}
	}
	
	if activeVoices == 0 {
		t.Error("NoteOn should create active voices")
	}
	
	// Test drum channel
	bank.NoteOn(9, 36, 100) // Kick drum on drum channel
	
	drumActive := false
	for _, v := range bank.AllVoices() {
		if v.Active() && v.MidiChannel() == 9 {
			drumActive = true
			break
		}
	}
	
	if !drumActive {
		t.Error("Drum channel should create active voices")
	}
	
	// Test patch count
	patchCount := bank.PatchCount()
	if patchCount != len(sound.PATCHES) {
		t.Errorf("Expected %d patches, got %d", len(sound.PATCHES), patchCount)
	}
	
	// Test patch name
	patchName := bank.PatchName(0)
	if patchName != "Acoustic Grand" {
		t.Errorf("Expected patch name 'Acoustic Grand', got '%s'", patchName)
	}
	
	fmt.Println("Patch system tests passed!")
}

func TestVoiceAllocation(t *testing.T) {
	bank := NewBank()
	
	// Set a filtered patch
	bank.SetPatch(0, 10) // Lead 1 uses filter
	
	// Allocate multiple voices on same channel
	for i := 0; i < 5; i++ {
		bank.NoteOn(0, 60+i, 100)
	}
	
	// Check that voices were allocated
	activeCount := 0
	for _, v := range bank.AllVoices() {
		if v.Active() {
			activeCount++
		}
	}
	
	if activeCount != 5 {
		t.Errorf("Expected 5 active voices, got %d", activeCount)
	}
	
	// Test voice stealing
	for i := 0; i < 10; i++ {
		bank.NoteOn(0, 70+i, 100)
	}
	
	// Should still have active voices (some may have been stolen)
	finalActiveCount := 0
	for _, v := range bank.AllVoices() {
		if v.Active() {
			finalActiveCount++
		}
	}
	
	if finalActiveCount == 0 {
		t.Error("Should have some active voices after voice stealing")
	}
	
	fmt.Println("Voice allocation tests passed!")
}