package game

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

type Audio struct {
	context *audio.Context
}

func NewAudio() *Audio {
	return &Audio{context: audio.NewContext(sampleRate)}
}

func (a *Audio) Start() {
	a.playTone(330, 60, 0.20)
}

func (a *Audio) Apple() {
	a.playTone(660, 80, 0.25)
}

func (a *Audio) Level() {
	a.playTone(880, 160, 0.25)
}

func (a *Audio) Crash() {
	a.playTone(110, 220, 0.35)
}

func (a *Audio) playTone(freq float64, ms int, volume float64) {
	data := synthWAV(freq, ms, volume)
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return
	}
	player, err := a.context.NewPlayer(stream)
	if err != nil {
		return
	}
	player.Play()
}

func synthWAV(freq float64, ms int, volume float64) []byte {
	samples := sampleRate * ms / 1000
	pcm := make([]int16, samples)
	for i := 0; i < samples; i++ {
		t := float64(i) / sampleRate
		env := 1.0 - float64(i)/float64(samples)
		wave := math.Sin(2 * math.Pi * freq * t)
		pcm[i] = int16(wave * env * volume * math.MaxInt16)
	}

	buf := &bytes.Buffer{}
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, uint32(36+len(pcm)*2))
	buf.WriteString("WAVEfmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate*2))
	_ = binary.Write(buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(pcm)*2))
	for _, sample := range pcm {
		_ = binary.Write(buf, binary.LittleEndian, sample)
	}
	return buf.Bytes()
}
