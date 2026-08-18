//go:build !windows

package main

import "testing"

func TestSDLGamepadIDIncludesPortableControllerType(t *testing.T) {
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
		if got := sdlGamepadID(test.name, test.vendor, test.product); got != test.want {
			t.Fatalf("sdlGamepadID(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}
