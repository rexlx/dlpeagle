package main

import (
	"fmt"
	"os"
	"time"
)

type Notification struct {
	ID   string `json:"id"`
	Info string `json:"info"`
	Time string `json:"time"`
}

type SoundBlock struct {
	Duration   time.Duration `json:"duration"`
	Frequency  float64       `json:"frequency"`
	SampleRate int           `json:"sample_rate"`
	Period     time.Duration `json:"period"`
}

func NewSoundBlock(duration, period time.Duration, frequency float64, sampleRate int) *SoundBlock {
	return &SoundBlock{
		Duration:   duration,
		Frequency:  frequency,
		SampleRate: sampleRate,
		Period:     period,
	}
}

func SoundBlockIn440Hz(t time.Duration) *SoundBlock {
	samples := int(float64(440) * t.Seconds())
	fmt.Println(samples, t)
	period := time.Second / time.Duration(samples)
	return NewSoundBlock(t, period, 440, samples)
}

func SoundBlockIn880Hz(t time.Duration) *SoundBlock {
	samples := int(float64(880) * t.Seconds())
	period := time.Second / time.Duration(samples)
	return NewSoundBlock(t, period, 880, samples)
}

func (s *SoundBlock) PlaySound() {
	for i := 0; i < s.SampleRate; i++ {
		if i%2 == 0 {
			os.Stdout.Write([]byte{7}) // ASCII bell character
		}
		time.Sleep(s.Period)
	}
}
