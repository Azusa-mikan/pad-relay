//go:build !windows

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type sdlController struct {
	controller *sdl.GameController
	instanceID sdl.JoystickID
	name       string
	id         string
	state      snapshot
	hasState   bool
}

var sdlButtonMappings = [...]struct {
	button sdl.GameControllerButton
	index  int
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

func runInput(ctx context.Context, updates chan snapshot, diagnose bool) error {
	sdl.SetHint(sdl.HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS, "1")
	if err := sdl.Init(sdl.INIT_GAMECONTROLLER | sdl.INIT_JOYSTICK | sdl.INIT_EVENTS); err != nil {
		return fmt.Errorf("初始化 SDL: %w", err)
	}
	defer sdl.Quit()

	if diagnose {
		return runSDLDiagnostics(ctx)
	}

	active := openFirstSDLController()
	if active == nil {
		log.Print("未找到受支持的手柄，等待插入")
	} else {
		publishSnapshot(updates, readSDLSnapshot(active.controller, active.id))
	}

	for ctx.Err() == nil {
		event := sdl.WaitEventTimeout(16)
		if event != nil {
			active = handleSDLDeviceEvent(event, active, updates)
			for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				active = handleSDLDeviceEvent(event, active, updates)
			}
		}

		if active == nil {
			continue
		}
		if !active.controller.Attached() {
			log.Printf("手柄已断开: %s", active.name)
			active.controller.Close()
			publishSnapshot(updates, disconnectedSnapshot(active.id))
			active = openFirstSDLController()
			continue
		}

		state := readSDLSnapshot(active.controller, active.id)
		if !active.hasState || state != active.state {
			active.state = state
			active.hasState = true
			publishSnapshot(updates, state)
		}
	}

	if active != nil {
		active.controller.Close()
	}
	return nil
}

func handleSDLDeviceEvent(event sdl.Event, active *sdlController, updates chan snapshot) *sdlController {
	switch event := event.(type) {
	case *sdl.JoyDeviceAddedEvent:
		if active == nil {
			return openSDLController(int(event.Which))
		}
	case *sdl.ControllerDeviceEvent:
		switch event.Type {
		case sdl.CONTROLLERDEVICEADDED:
			if active == nil {
				return openSDLController(int(event.Which))
			}
		case sdl.CONTROLLERDEVICEREMOVED:
			if active != nil && event.Which == active.instanceID {
				log.Printf("手柄已断开: %s", active.name)
				active.controller.Close()
				publishSnapshot(updates, disconnectedSnapshot(active.id))
				return openFirstSDLController()
			}
		}
	case *sdl.JoyDeviceRemovedEvent:
		if active != nil && event.Which == active.instanceID {
			log.Printf("手柄已断开: %s", active.name)
			active.controller.Close()
			publishSnapshot(updates, disconnectedSnapshot(active.id))
			return openFirstSDLController()
		}
	}
	return active
}

func openFirstSDLController() *sdlController {
	for index := 0; index < sdl.NumJoysticks(); index++ {
		if controller := openSDLController(index); controller != nil {
			return controller
		}
	}
	return nil
}

func openSDLController(index int) *sdlController {
	if index < 0 || index >= sdl.NumJoysticks() {
		return nil
	}
	addZikwaySDLMapping(index)
	if !sdl.IsGameController(index) {
		log.Printf("忽略没有 SDL 映射的设备: %s (%04x:%04x)，可使用 -diagnose 查看原始输入",
			sdl.JoystickNameForIndex(index), sdl.JoystickGetDeviceVendor(index), sdl.JoystickGetDeviceProduct(index))
		return nil
	}

	controller := sdl.GameControllerOpen(index)
	if controller == nil {
		log.Printf("无法打开手柄 %s: %v", sdl.JoystickNameForIndex(index), sdl.GetError())
		return nil
	}
	joystick := controller.Joystick()
	if joystick == nil {
		controller.Close()
		return nil
	}

	active := &sdlController{
		controller: controller,
		instanceID: joystick.InstanceID(),
		name:       controller.Name(),
		id:         sdlGamepadID(controller.Name(), controller.Vendor(), controller.Product()),
	}
	log.Printf("已连接手柄: %s (%04x:%04x, instance=%d, backend=SDL)",
		active.name, controller.Vendor(), controller.Product(), active.instanceID)
	return active
}

func addZikwaySDLMapping(index int) {
	if sdl.JoystickGetDeviceVendor(index) != 0x3537 || sdl.JoystickGetDeviceProduct(index) != 0x1041 {
		return
	}

	guid := sdl.JoystickGetGUIDString(sdl.JoystickGetDeviceGUID(index))
	mapping := fmt.Sprintf(
		"%s,Zikway HID gamepad,a:b0,b:b1,x:b2,y:b3,back:b8,guide:b10,start:b9,leftstick:b11,rightstick:b12,leftshoulder:b4,rightshoulder:b5,dpup:h0.1,dpdown:h0.4,dpleft:h0.8,dpright:h0.2,leftx:a0,lefty:a1,rightx:a2,righty:a3,lefttrigger:a5,righttrigger:a4,",
		guid,
	)
	if result := sdl.GameControllerAddMapping(mapping); result < 0 {
		log.Printf("添加 Zikway SDL 映射失败: %v", sdl.GetError())
	}
}

func readSDLSnapshot(controller *sdl.GameController, id string) snapshot {
	state := connectedSnapshot(id)
	for _, mapping := range sdlButtonMappings {
		setDigitalButton(&state, mapping.index, controller.Button(mapping.button) != 0)
	}
	state.Axes[0] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_LEFTX))
	state.Axes[1] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_LEFTY))
	state.Axes[2] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_RIGHTX))
	state.Axes[3] = normalizeAxis(controller.Axis(sdl.CONTROLLER_AXIS_RIGHTY))
	setTrigger(&state, 6, controller.Axis(sdl.CONTROLLER_AXIS_TRIGGERLEFT))
	setTrigger(&state, 7, controller.Axis(sdl.CONTROLLER_AXIS_TRIGGERRIGHT))
	return state
}

func sdlGamepadID(name string, vendor, product int) string {
	lowerName := strings.ToLower(name)
	switch {
	case product == 0x0ce6 || strings.Contains(lowerName, "dualsense") || strings.Contains(lowerName, "ps5"):
		if !strings.Contains(lowerName, "dualsense") {
			name = "DualSense " + name
		}
	case vendor == 0x054c || strings.Contains(lowerName, "dualshock") || strings.Contains(lowerName, "ps4"):
		if !strings.Contains(lowerName, "dualshock") {
			name = "DualShock 4 " + name
		}
	case vendor == 0x045e || strings.Contains(lowerName, "xbox") || strings.Contains(lowerName, "xinput"):
		if !strings.Contains(lowerName, "xinput") {
			name = "XInput " + name
		}
	}
	return fmt.Sprintf("%s (Vendor: %04x Product: %04x)", name, vendor, product)
}

func runSDLDiagnostics(ctx context.Context) error {
	joysticks := make(map[sdl.JoystickID]*sdl.Joystick)
	open := func(index int) {
		instanceID := sdl.JoystickGetDeviceInstanceID(index)
		if _, exists := joysticks[instanceID]; exists {
			return
		}
		joystick := sdl.JoystickOpen(index)
		if joystick == nil {
			log.Printf("无法打开 joystick %d: %v", index, sdl.GetError())
			return
		}
		joysticks[instanceID] = joystick
		log.Printf("诊断设备: %s (%04x:%04x, instance=%d, axes=%d, buttons=%d, hats=%d, backend=SDL)",
			joystick.Name(), joystick.Vendor(), joystick.Product(), joystick.InstanceID(), joystick.NumAxes(), joystick.NumButtons(), joystick.NumHats())
	}

	for index := 0; index < sdl.NumJoysticks(); index++ {
		open(index)
	}
	log.Print("按动手柄控件，按 Ctrl+C 退出")

	for ctx.Err() == nil {
		event := sdl.WaitEventTimeout(100)
		switch event := event.(type) {
		case *sdl.JoyAxisEvent:
			log.Printf("instance=%d axis=%d value=%d", event.Which, event.Axis, event.Value)
		case *sdl.JoyButtonEvent:
			log.Printf("instance=%d button=%d state=%d", event.Which, event.Button, event.State)
		case *sdl.JoyHatEvent:
			log.Printf("instance=%d hat=%d value=%d", event.Which, event.Hat, event.Value)
		case *sdl.JoyDeviceAddedEvent:
			open(int(event.Which))
		case *sdl.JoyDeviceRemovedEvent:
			if joystick := joysticks[event.Which]; joystick != nil {
				log.Printf("诊断设备已断开: %s (instance=%d)", joystick.Name(), event.Which)
				joystick.Close()
				delete(joysticks, event.Which)
			}
		}
	}

	for _, joystick := range joysticks {
		joystick.Close()
	}
	return nil
}
