/**
 * ws-shim.js
 * Replaces the browser Gamepad API with WebSocket data from gamepad-remote.
 * Must be loaded BEFORE gamepad.js.
 */
(function () {
    const WS_PROTOCOL = location.protocol === "https:" ? "wss:" : "ws:";
    const WS_URL = `${WS_PROTOCOL}//${location.host}/ws/viewer`;

    function createGamepad(id) {
        return {
            index: 0,
            id,
            buttons: Array(17).fill(null).map(() => ({ pressed: false, value: 0 })),
            axes: [0, 0, 0, 0],
            timestamp: 0,
            connected: true,
            mapping: "standard",
            vibrationActuator: null,
        };
    }

    let fakeGamepad = createGamepad("");
    let wsConnected = false;

    // Override the Gamepad API
    navigator.getGamepads = () => wsConnected ? [fakeGamepad] : [];
    if ("webkitGetGamepads" in navigator) {
        navigator.webkitGetGamepads = navigator.getGamepads;
    }

    function disconnectGamepad() {
        if (!wsConnected) return;
        wsConnected = false;
        fakeGamepad.connected = false;
        window.dispatchEvent(new GamepadEvent("gamepaddisconnected", { gamepad: fakeGamepad }));
    }

    function connectGamepad(data) {
        const identityChanged = wsConnected && fakeGamepad.id !== data.id;
        if (identityChanged) disconnectGamepad();

        if (!wsConnected) {
            fakeGamepad = createGamepad(data.id);
        }
        fakeGamepad.buttons = data.buttons;
        fakeGamepad.axes = data.axes;
        fakeGamepad.timestamp = performance.now();

        if (!wsConnected) {
            wsConnected = true;
            window.dispatchEvent(new GamepadEvent("gamepadconnected", { gamepad: fakeGamepad }));
            requestAnimationFrame(() => {
                if (window.gamepad) window.gamepad.map(0);
            });
        }
    }

    function connect() {
        const ws = new WebSocket(WS_URL);

        ws.onopen = () => {
            console.info("[ws-shim] connected to", WS_URL);
        };

        ws.onmessage = (event) => {
            let data;
            try {
                data = JSON.parse(event.data);
            } catch {
                return;
            }

            if (typeof data.id !== "string" || data.id === "") {
                disconnectGamepad();
                return;
            }
            connectGamepad(data);
        };

        ws.onclose = () => {
            console.warn("[ws-shim] disconnected, reconnecting in 2 s…");
            disconnectGamepad();
            setTimeout(connect, 2000);
        };

        ws.onerror = () => ws.close();
    }

    connect();
})();
