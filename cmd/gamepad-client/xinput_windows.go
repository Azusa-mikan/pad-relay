//go:build windows

package main

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	xinputDPadUp        = 0x0001
	xinputDPadDown      = 0x0002
	xinputDPadLeft      = 0x0004
	xinputDPadRight     = 0x0008
	xinputStart         = 0x0010
	xinputBack          = 0x0020
	xinputLeftThumb     = 0x0040
	xinputRightThumb    = 0x0080
	xinputLeftShoulder  = 0x0100
	xinputRightShoulder = 0x0200
	xinputA             = 0x1000
	xinputB             = 0x2000
	xinputX             = 0x4000
	xinputY             = 0x8000
)

type xinputGamepad struct {
	Buttons      uint16
	LeftTrigger  uint8
	RightTrigger uint8
	ThumbLX      int16
	ThumbLY      int16
	ThumbRX      int16
	ThumbRY      int16
}

type xinputState struct {
	PacketNumber uint32
	Gamepad      xinputGamepad
}

var (
	xinputOnce     sync.Once
	xinputGetState *syscall.LazyProc
)

func loadXInput() {
	for _, name := range [...]string{"xinput1_4.dll", "xinput1_3.dll", "xinput9_1_0.dll"} {
		dll := syscall.NewLazyDLL(name)
		if err := dll.Load(); err != nil {
			continue
		}
		proc := dll.NewProc("XInputGetState")
		if err := proc.Find(); err == nil {
			xinputGetState = proc
			return
		}
	}
}

func readXInputSnapshot(id string) (snapshot, bool) {
	xinputOnce.Do(loadXInput)
	if xinputGetState == nil {
		return snapshot{}, false
	}

	for index := uintptr(0); index < 4; index++ {
		var raw xinputState
		result, _, _ := xinputGetState.Call(index, uintptr(unsafe.Pointer(&raw)))
		if result == 0 {
			return snapshotFromXInput(id, raw.Gamepad), true
		}
	}
	return snapshot{}, false
}

func snapshotFromXInput(id string, gamepad xinputGamepad) snapshot {
	state := snapshot{ID: id}
	setXInputButton(&state, 0, gamepad.Buttons, xinputA)
	setXInputButton(&state, 1, gamepad.Buttons, xinputB)
	setXInputButton(&state, 2, gamepad.Buttons, xinputX)
	setXInputButton(&state, 3, gamepad.Buttons, xinputY)
	setXInputButton(&state, 4, gamepad.Buttons, xinputLeftShoulder)
	setXInputButton(&state, 5, gamepad.Buttons, xinputRightShoulder)
	setXInputButton(&state, 8, gamepad.Buttons, xinputBack)
	setXInputButton(&state, 9, gamepad.Buttons, xinputStart)
	setXInputButton(&state, 10, gamepad.Buttons, xinputLeftThumb)
	setXInputButton(&state, 11, gamepad.Buttons, xinputRightThumb)
	setXInputButton(&state, 12, gamepad.Buttons, xinputDPadUp)
	setXInputButton(&state, 13, gamepad.Buttons, xinputDPadDown)
	setXInputButton(&state, 14, gamepad.Buttons, xinputDPadLeft)
	setXInputButton(&state, 15, gamepad.Buttons, xinputDPadRight)

	state.Axes[0] = normalizeAxis(gamepad.ThumbLX)
	state.Axes[1] = round4(-normalizeAxis(gamepad.ThumbLY))
	state.Axes[2] = normalizeAxis(gamepad.ThumbRX)
	state.Axes[3] = round4(-normalizeAxis(gamepad.ThumbRY))
	setXInputTrigger(&state, 6, gamepad.LeftTrigger)
	setXInputTrigger(&state, 7, gamepad.RightTrigger)
	return state
}

func setXInputButton(state *snapshot, index int, buttons, mask uint16) {
	pressed := buttons&mask != 0
	value := 0.0
	if pressed {
		value = 1
	}
	state.Buttons[index] = buttonState{Pressed: pressed, Value: value}
}

func setXInputTrigger(state *snapshot, index int, raw uint8) {
	value := round4(float64(raw) / 255)
	state.Buttons[index] = buttonState{Pressed: raw != 0, Value: value}
}
