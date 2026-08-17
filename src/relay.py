"""
WebSocket relay: one controller -> many viewers.
"""

import json
import logging

from fastapi import WebSocket, WebSocketDisconnect
from fastapi.websockets import WebSocketState

logger = logging.getLogger(__name__)


class ConnectionManager:
    def __init__(self) -> None:
        self.controller: WebSocket | None = None
        self.viewers: list[WebSocket] = []

    async def connect_controller(self, ws: WebSocket) -> None:
        await ws.accept()
        if self.controller is not None:
            logger.warning("New controller connected, replacing previous one.")
        self.controller = ws
        logger.info("Controller connected.")

    def disconnect_controller(self, ws: WebSocket) -> None:
        if self.controller is ws:
            self.controller = None
            logger.info("Controller disconnected.")

    async def connect_viewer(self, ws: WebSocket) -> None:
        await ws.accept()
        self.viewers.append(ws)
        logger.info("Viewer connected. Total viewers: %d", len(self.viewers))

    def disconnect_viewer(self, ws: WebSocket) -> None:
        self.viewers = [v for v in self.viewers if v is not ws]
        logger.info("Viewer disconnected. Total viewers: %d", len(self.viewers))

    async def broadcast(self, message: str) -> None:
        dead: list[WebSocket] = []
        for viewer in self.viewers:
            try:
                if viewer.client_state == WebSocketState.CONNECTED:
                    await viewer.send_text(message)
            except Exception:
                dead.append(viewer)
        for ws in dead:
            self.disconnect_viewer(ws)


manager = ConnectionManager()


async def controller_endpoint(ws: WebSocket) -> None:
    await manager.connect_controller(ws)
    try:
        while True:
            data = await ws.receive_text()
            try:
                json.loads(data)
            except json.JSONDecodeError:
                logger.warning("Invalid JSON from controller, skipping.")
                continue
            logger.debug("STATE %s", data)
            await manager.broadcast(data)
    except WebSocketDisconnect:
        manager.disconnect_controller(ws)


async def viewer_endpoint(ws: WebSocket) -> None:
    await manager.connect_viewer(ws)
    try:
        while True:
            await ws.receive_text()
    except WebSocketDisconnect:
        manager.disconnect_viewer(ws)
