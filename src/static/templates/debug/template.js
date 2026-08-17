(() => {
    const gamepad = window.gamepad;
    let activeGamepad = gamepad.getActive();
    if (!activeGamepad) return;

    const $id        = document.querySelector("#info-id .value");
    const $timestamp = document.querySelector("#info-timestamp .value");
    const $index     = document.querySelector("#info-index .value");
    const $mapping   = document.querySelector("#info-mapping .value");
    const $rumble    = document.querySelector("#info-rumble .value");
    const $axes      = document.querySelector(".axes .container");
    const $buttons   = document.querySelector(".buttons .container");

    $id.textContent      = activeGamepad.id;
    $index.textContent   = activeGamepad.index;
    $mapping.textContent = activeGamepad.mapping;
    $rumble.textContent  = activeGamepad.vibrationActuator
        ? activeGamepad.vibrationActuator.type
        : "N/A";
    updateTimestamp();

    for (let i = 0; i < activeGamepad.axes.length; i++) {
        $axes.insertAdjacentHTML("beforeend", `
            <div class="box medium"><div class="content">
                <div class="label">Axis ${i}</div>
                <div class="value" data-axis="${i}"></div>
            </div></div>`);
    }

    for (let i = 0; i < activeGamepad.buttons.length; i++) {
        $buttons.insertAdjacentHTML("beforeend", `
            <div class="box small"><div class="content">
                <div class="label">B${i}</div>
                <div class="value" data-button="${i}"></div>
            </div></div>`);
    }

    gamepad.updateButton = ($button) => updateElem($button);
    gamepad.updateAxis   = ($axis)   => updateElem($axis, 6);
    gamepad.updateFrame  = ()        => updateTimestamp();

    function updateElem($elem, precision = 2) {
        const value = parseFloat($elem.getAttribute("data-value")).toFixed(precision);
        $elem.textContent = value;
        const color = Math.floor(255 * 0.3 + 255 * 0.7 * Math.abs(value));
        $elem.style.color = `rgb(${color}, ${color}, ${color})`;
    }

    function updateTimestamp() {
        activeGamepad = gamepad.getActive();
        if (!activeGamepad) return;
        $timestamp.textContent = parseFloat(activeGamepad.timestamp).toFixed(3);
    }
})();
