package audio

import (
	"encoding/binary"
	"os"
)

// WAVWriter writes WAV audio file
type WAVWriter struct {
	file           *os.File
	sampleRate     int
	channels       int
	bitsPerSample  int
	dataWritten    int
}

// NewWAVWriter creates a new WAV file writer
func NewWAVWriter(filename string, sampleRate, channels, bitsPerSample int) (*WAVWriter, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	wav := &WAVWriter{
		file:          f,
		sampleRate:    sampleRate,
		channels:      channels,
		bitsPerSample: bitsPerSample,
		dataWritten:   0,
	}

	// Write placeholder header (we'll update it at the end)
	header := make([]byte, 44)
	f.Write(header)

	return wav, nil
}

// WriteSamples writes float64 samples to the WAV file
func (w *WAVWriter) WriteSamples(samples []float64) error {
	for _, s := range samples {
		// Convert to int16
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}

		sample := int16(s * 32767)

		// Write as little-endian
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(sample))

		_, err := w.file.Write(buf)
		if err != nil {
			return err
		}
		w.dataWritten++
	}

	return nil
}

// Close finalizes the WAV file
func (w *WAVWriter) Close() error {
	// Seek to beginning to write header
	_, err := w.file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Write RIFF header
	chunkSize := uint32(36 + w.dataWritten*2) // 2 bytes per sample (16-bit mono)
	binary.Write(w.file, binary.LittleEndian, []byte("RIFF"))
	binary.Write(w.file, binary.LittleEndian, chunkSize)
	binary.Write(w.file, binary.LittleEndian, []byte("WAVE"))

	// Write fmt chunk
	binary.Write(w.file, binary.LittleEndian, []byte("fmt "))
	binary.Write(w.file, binary.LittleEndian, uint32(16)) // Subchunk1Size
	binary.Write(w.file, binary.LittleEndian, uint16(1))  // AudioFormat (PCM)
	binary.Write(w.file, binary.LittleEndian, uint16(w.channels))
	binary.Write(w.file, binary.LittleEndian, uint32(w.sampleRate))
	binary.Write(w.file, binary.LittleEndian, uint32(w.sampleRate*w.channels*w.bitsPerSample/8)) // ByteRate
	binary.Write(w.file, binary.LittleEndian, uint16(w.channels*w.bitsPerSample/8)) // BlockAlign
	binary.Write(w.file, binary.LittleEndian, uint16(w.bitsPerSample))

	// Write data chunk header
	binary.Write(w.file, binary.LittleEndian, []byte("data"))
	binary.Write(w.file, binary.LittleEndian, uint32(w.dataWritten*2))

	return w.file.Close()
}
