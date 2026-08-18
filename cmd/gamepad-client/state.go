package main

import "math"

const (
	numButtons = 17
	numAxes    = 4
)

type buttonState struct {
	Pressed bool    `json:"pressed"`
	Value   float64 `json:"value"`
}

type snapshot struct {
	ID      string                  `json:"id"`
	Buttons [numButtons]buttonState `json:"buttons"`
	Axes    [numAxes]float64        `json:"axes"`
}

func setDigitalButton(state *snapshot, index int, pressed bool) {
	value := 0.0
	if pressed {
		value = 1
	}
	state.Buttons[index] = buttonState{Pressed: pressed, Value: value}
}

func setTrigger(state *snapshot, index int, raw int16) {
	value := float64(raw) / 32767
	value = min(1, max(0, value))
	value = round4(value)
	state.Buttons[index] = buttonState{Pressed: value > 0, Value: value}
}

func normalizeAxis(raw int16) float64 {
	divisor := 32767.0
	if raw < 0 {
		divisor = 32768
	}
	return round4(float64(raw) / divisor)
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}
