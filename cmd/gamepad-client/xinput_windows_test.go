//go:build windows

package main

import "testing"

func TestSnapshotFromXInput(t *testing.T) {
	state := snapshotFromXInput("test", xinputGamepad{
		Buttons:      xinputA | xinputRightShoulder | xinputDPadLeft,
		LeftTrigger:  255,
		RightTrigger: 128,
		ThumbLX:      32767,
		ThumbLY:      32767,
		ThumbRX:      -32768,
		ThumbRY:      -32768,
	})

	for _, index := range []int{0, 5, 14} {
		if !state.Buttons[index].Pressed || state.Buttons[index].Value != 1 {
			t.Fatalf("button %d is not pressed: %+v", index, state.Buttons[index])
		}
	}
	if state.Buttons[1].Pressed {
		t.Fatalf("B button is unexpectedly pressed: %+v", state.Buttons[1])
	}
	if state.Buttons[6].Value != 1 || state.Buttons[7].Value != 0.502 {
		t.Fatalf("trigger values are incorrect: left=%v right=%v", state.Buttons[6].Value, state.Buttons[7].Value)
	}
	wantAxes := [numAxes]float64{1, -1, -1, 1}
	if state.Axes != wantAxes {
		t.Fatalf("axes = %v, want %v", state.Axes, wantAxes)
	}
}
