package main

import (
	"testing"
)

func TestNormalizeAxis(t *testing.T) {
	tests := []struct {
		raw  int16
		want float64
	}{
		{-32768, -1},
		{0, 0},
		{16384, 0.5},
		{32767, 1},
	}
	for _, test := range tests {
		if got := normalizeAxis(test.raw); got != test.want {
			t.Fatalf("normalizeAxis(%d) = %v, want %v", test.raw, got, test.want)
		}
	}
}

func TestSetTriggerClampsRange(t *testing.T) {
	var state snapshot
	setTrigger(&state, 6, -1)
	if state.Buttons[6].Value != 0 || state.Buttons[6].Pressed {
		t.Fatalf("negative trigger was not clamped: %+v", state.Buttons[6])
	}
	setTrigger(&state, 6, 32767)
	if state.Buttons[6].Value != 1 || !state.Buttons[6].Pressed {
		t.Fatalf("full trigger is incorrect: %+v", state.Buttons[6])
	}
}

func TestGamepadIDIncludesPortableControllerType(t *testing.T) {
	tests := []struct {
		name    string
		vendor  int
		product int
		want    string
	}{
		{"Wireless Controller", 0x054c, 0x09cc, "DualShock 4 Wireless Controller (Vendor: 054c Product: 09cc)"},
		{"PS5 Controller", 0, 0, "DualSense PS5 Controller (Vendor: 0000 Product: 0000)"},
		{"Xbox Controller", 0, 0, "XInput Xbox Controller (Vendor: 0000 Product: 0000)"},
	}
	for _, test := range tests {
		if got := gamepadID(test.name, test.vendor, test.product); got != test.want {
			t.Fatalf("gamepadID(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}
