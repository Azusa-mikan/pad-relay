"""
WebSocket relay: one controller -> many viewers.
"""

import json
import logging

from fastapi import WebSocket, WebSocketDisconnect

logger = logging.getLogger(__name__)


class ConnectionManager:
    def __init__(self) -> None:
        self.controller: WebSocket | None = None
        self.viewers: list[WebSocket] = []
        self.gamepad_connected = False
        self.gamepad_id: str | None = None

    async def connect_controller(self, ws: WebSocket) -> None:
        await ws.accept()
        previous = self.controller
        if previous is not None and previous is not ws:
            logger.warning("New controller connected, replacing previous one.")
            previous_gamepad_id = self.gamepad_id
            self.controller = ws
            self.gamepad_connected = False
            self.gamepad_id = None
            try:
                await previous.close(code=1000, reason="replaced by a new controller")
            except Exception:
                logger.debug("Failed to close the previous controller.", exc_info=True)
            if previous_gamepad_id is not None:
                await self.broadcast(
                    json.dumps({"id": previous_gamepad_id, "connected": False})
                )
        else:
            self.controller = ws
            self.gamepad_connected = False
            self.gamepad_id = None
        logger.info("Controller connected.")

    def disconnect_controller(self, ws: WebSocket) -> str | None:
        if self.controller is ws:
            # Controller WebSocket closure is itself the lifecycle signal.
            # Keep the last identity even when the controller had already
            # reported a physical disconnect, so a process exit is handled
            # independently of controller-side state messages.
            disconnected_id = self.gamepad_id
            self.controller = None
            self.gamepad_connected = False
            self.gamepad_id = None
            logger.debug("Controller disconnected.")
            return disconnected_id
        return None

    def update_gamepad_state(self, data: str) -> bool:
        try:
            payload = json.loads(data)
        except json.JSONDecodeError:
            logger.warning("Invalid JSON from controller, skipping.")
            return False

        if not isinstance(payload, dict):
            logger.warning("Controller message must be a JSON object, skipping.")
            return False

        gamepad_id = payload.get("id")
        connected = payload.get("connected")
        if connected is None:
            # Accept messages from older clients that used an empty id to
            # represent disconnection.
            connected = isinstance(gamepad_id, str) and gamepad_id != ""
        if not isinstance(connected, bool):
            logger.warning("Invalid connected field from controller, skipping.")
            return False
        if connected and (not isinstance(gamepad_id, str) or gamepad_id == ""):
            logger.warning("Connected controller message has no gamepad id, skipping.")
            return False

        previous_connected = self.gamepad_connected
        previous_id = self.gamepad_id
        self.gamepad_connected = connected
        if connected:
            self.gamepad_id = gamepad_id
        elif isinstance(gamepad_id, str) and gamepad_id:
            # Keep the identity available for a subsequent controller
            # disconnect notification, even after an explicit false event.
            self.gamepad_id = gamepad_id
        if connected != previous_connected or self.gamepad_id != previous_id:
            logger.debug(
                "Gamepad %s%s.",
                "connected" if connected else "disconnected",
                f": {gamepad_id}" if connected else "",
            )
        return True

    async def connect_viewer(self, ws: WebSocket) -> None:
        await ws.accept()
        self.viewers.append(ws)
        logger.info("Viewer connected. Total viewers: %d", len(self.viewers))

    def disconnect_viewer(self, ws: WebSocket) -> None:
        self.viewers = [v for v in self.viewers if v is not ws]
        logger.info("Viewer disconnected. Total viewers: %d", len(self.viewers))

    async def broadcast(self, message: str) -> None:
        dead: list[WebSocket] = []
        sent = 0
        for viewer in self.viewers:
            try:
                # Viewers only receive relay messages and may never send a
                # message back, so client_state is not a send-readiness check.
                await viewer.send_text(message)
                sent += 1
            except Exception:
                dead.append(viewer)
        for ws in dead:
            self.disconnect_viewer(ws)
        logger.debug(
            "Broadcasted message to %d/%d viewers; removed %d dead viewers.",
            sent,
            len(self.viewers) + len(dead),
            len(dead),
        )


manager = ConnectionManager()


async def controller_endpoint(ws: WebSocket) -> None:
    await manager.connect_controller(ws)
    try:
        while True:
            data = await ws.receive_text()
            if manager.controller is not ws:
                continue
            if not manager.update_gamepad_state(data):
                continue
            logger.debug("STATE %s", data)
            await manager.broadcast(data)
    except WebSocketDisconnect:
        disconnected_id = manager.disconnect_controller(ws)
        if disconnected_id is not None:
            logger.debug("Gamepad disconnected (controller connection closed).")
            await manager.broadcast(
                json.dumps({"id": disconnected_id, "connected": False})
            )


async def viewer_endpoint(ws: WebSocket) -> None:
    await manager.connect_viewer(ws)
    try:
        while True:
            await ws.receive_text()
    except WebSocketDisconnect:
        manager.disconnect_viewer(ws)
