//go:build !windows

package main

func readXInputSnapshot(string) (snapshot, bool) {
	return snapshot{}, false
}
