//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"
	"unsafe"
)

type xinputBackend struct {
	getState *syscall.LazyProc
}

func runInput(ctx context.Context, updates chan snapshot, diagnose bool) error {
	backend, err := loadXInputBackend()
	if err != nil {
		return err
	}
	if diagnose {
		return backend.runDiagnostics(ctx)
	}
	return backend.run(ctx, updates)
}

func loadXInputBackend() (*xinputBackend, error) {
	for _, name := range [...]string{"xinput1_4.dll", "xinput1_3.dll", "xinput9_1_0.dll"} {
		dll := syscall.NewLazyDLL(name)
		if err := dll.Load(); err != nil {
			continue
		}
		proc := dll.NewProc("XInputGetState")
		if err := proc.Find(); err == nil {
			return &xinputBackend{getState: proc}, nil
		}
	}
	return nil, fmt.Errorf("无法加载 Windows XInput")
}

func (backend *xinputBackend) run(ctx context.Context, updates chan snapshot) error {
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	active := -1
	var previous snapshot
	hasPrevious := false
	if index, state, ok := backend.firstConnected(); ok {
		active = index
		previous = state
		hasPrevious = true
		log.Printf("已连接手柄: XInput Controller %d (backend=XInput)", index+1)
		publishSnapshot(updates, state)
	} else {
		log.Print("未找到 XInput 手柄，等待插入")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		if active < 0 {
			index, state, ok := backend.firstConnected()
			if !ok {
				continue
			}
			active = index
			previous = state
			hasPrevious = true
			log.Printf("已连接手柄: XInput Controller %d (backend=XInput)", index+1)
			publishSnapshot(updates, state)
			continue
		}

		state, ok := backend.snapshot(active)
		if !ok {
			log.Printf("手柄已断开: XInput Controller %d", active+1)
			publishSnapshot(updates, snapshot{})
			active = -1
			hasPrevious = false
			continue
		}
		if !hasPrevious || state != previous {
			previous = state
			hasPrevious = true
			publishSnapshot(updates, state)
		}
	}
}

func (backend *xinputBackend) firstConnected() (int, snapshot, bool) {
	for index := 0; index < 4; index++ {
		if state, ok := backend.snapshot(index); ok {
			return index, state, true
		}
	}
	return -1, snapshot{}, false
}

func (backend *xinputBackend) snapshot(index int) (snapshot, bool) {
	var raw xinputState
	result, _, _ := backend.getState.Call(uintptr(index), uintptr(unsafe.Pointer(&raw)))
	if result != 0 {
		return snapshot{}, false
	}
	id := fmt.Sprintf("XInput Controller %d", index+1)
	return snapshotFromXInput(id, raw.Gamepad), true
}

func (backend *xinputBackend) runDiagnostics(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var connected [4]bool
	var previous [4]snapshot
	log.Print("XInput 诊断已启动，按动手柄控件，按 Ctrl+C 退出")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
		for index := 0; index < 4; index++ {
			state, ok := backend.snapshot(index)
			if !ok {
				if connected[index] {
					log.Printf("诊断设备已断开: XInput Controller %d", index+1)
					connected[index] = false
				}
				continue
			}
			if !connected[index] {
				log.Printf("诊断设备: XInput Controller %d (backend=XInput)", index+1)
				connected[index] = true
				previous[index] = state
				continue
			}
			for button, value := range state.Buttons {
				if value != previous[index].Buttons[button] {
					log.Printf("controller=%d button=%d pressed=%t value=%.4f", index+1, button, value.Pressed, value.Value)
				}
			}
			for axis, value := range state.Axes {
				if value != previous[index].Axes[axis] {
					log.Printf("controller=%d axis=%d value=%.4f", index+1, axis, value)
				}
			}
			previous[index] = state
		}
	}
}
