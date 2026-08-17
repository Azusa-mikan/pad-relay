const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

test("WebSocket gamepad identity changes reconnect the virtual gamepad", () => {
    const events = [];
    const mapped = [];
    const sockets = [];

    class MockWebSocket {
        constructor(url) {
            this.url = url;
            sockets.push(this);
        }

        close() {}
    }

    class MockGamepadEvent {
        constructor(type, init) {
            this.type = type;
            this.gamepad = init.gamepad;
        }
    }

    const context = {
        console,
        GamepadEvent: MockGamepadEvent,
        navigator: {},
        performance: { now: () => 123 },
        requestAnimationFrame: (callback) => callback(),
        setTimeout: () => {},
        WebSocket: MockWebSocket,
        window: {
            dispatchEvent: (event) => events.push([event.type, event.gamepad.id]),
            gamepad: { map: (index) => mapped.push(index) },
        },
        location: { host: "localhost:8000" },
    };

    const script = fs.readFileSync(
        path.join(__dirname, "../src/static/js/ws-shim.js"),
        "utf8",
    );
    vm.runInNewContext(script, context);

    const socket = sockets[0];
    const state = {
        buttons: Array(17).fill({ pressed: false, value: 0 }),
        axes: [0, 0, 0, 0],
    };

    socket.onmessage({
        data: JSON.stringify({
            ...state,
            id: "DualSense (Vendor: 054c Product: 0ce6)",
        }),
    });
    assert.equal(context.navigator.getGamepads()[0].id, "DualSense (Vendor: 054c Product: 0ce6)");

    socket.onmessage({
        data: JSON.stringify({
            ...state,
            id: "Zikway HID gamepad (Vendor: 3537 Product: 1041)",
        }),
    });
    assert.deepEqual(events, [
        ["gamepadconnected", "DualSense (Vendor: 054c Product: 0ce6)"],
        ["gamepaddisconnected", "DualSense (Vendor: 054c Product: 0ce6)"],
        ["gamepadconnected", "Zikway HID gamepad (Vendor: 3537 Product: 1041)"],
    ]);
    assert.deepEqual(mapped, [0, 0]);

    socket.onmessage({ data: JSON.stringify({ ...state, id: "" }) });
    assert.equal(context.navigator.getGamepads().length, 0);
});
