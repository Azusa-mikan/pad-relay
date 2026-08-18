package main

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

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

var buttonMappings = [...]struct {
	SDL   sdl.GameControllerButton
	Index int
}{
	{sdl.CONTROLLER_BUTTON_A, 0},
	{sdl.CONTROLLER_BUTTON_B, 1},
	{sdl.CONTROLLER_BUTTON_X, 2},
	{sdl.CONTROLLER_BUTTON_Y, 3},
	{sdl.CONTROLLER_BUTTON_LEFTSHOULDER, 4},
	{sdl.CONTROLLER_BUTTON_RIGHTSHOULDER, 5},
	{sdl.CONTROLLER_BUTTON_BACK, 8},
	{sdl.CONTROLLER_BUTTON_START, 9},
	{sdl.CONTROLLER_BUTTON_LEFTSTICK, 10},
	{sdl.CONTROLLER_BUTTON_RIGHTSTICK, 11},
	{sdl.CONTROLLER_BUTTON_DPAD_UP, 12},
	{sdl.CONTROLLER_BUTTON_DPAD_DOWN, 13},
	{sdl.CONTROLLER_BUTTON_DPAD_LEFT, 14},
	{sdl.CONTROLLER_BUTTON_DPAD_RIGHT, 15},
	{sdl.CONTROLLER_BUTTON_GUIDE, 16},
}

func readSnapshot(controller *sdl.GameController, id string, useXInput bool) snapshot {
	if useXInput {
		if state, ok := readXInputSnapshot(id); ok {
			return state
		}
	}

	state := snapshot{ID: id}

	for _, mapping := range buttonMappings {
		pressed := controller.Button(mapping.SDL) != 0
		value := 0.0
		if pressed {
			value = 1
		}
		state.Buttons[mapping.Index] = buttonState{Pressed: pressed, Value: value}
	}

	state.Axes[0] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_LEFTX))
	state.Axes[1] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_LEFTY))
	state.Axes[2] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_RIGHTX))
	state.Axes[3] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_RIGHTY))

	setTrigger(&state, 6, controller.Axis(sdl.CONTROLLER_AXIS_TRIGGERLEFT))
	setTrigger(&state, 7, controller.Axis(sdl.CONTROLLER_AXIS_TRIGGERRIGHT))
	return state
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
