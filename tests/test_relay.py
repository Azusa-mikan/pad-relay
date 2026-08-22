from unittest.mock import AsyncMock

import pytest
from fastapi import WebSocketDisconnect

import src.relay as relay
from src.relay import ConnectionManager


@pytest.mark.asyncio
async def test_replacing_controller_closes_previous_connection() -> None:
    manager = ConnectionManager()
    previous = AsyncMock()
    current = AsyncMock()

    await manager.connect_controller(previous)
    await manager.connect_controller(current)

    previous.close.assert_awaited_once_with(
        code=1000, reason="replaced by a new controller"
    )
    assert manager.controller is current


@pytest.mark.asyncio
async def test_replacing_controller_broadcasts_gamepad_disconnect() -> None:
    manager = ConnectionManager()
    previous = AsyncMock()
    current = AsyncMock()
    viewer = AsyncMock()
    manager.viewers.append(viewer)

    await manager.connect_controller(previous)
    manager.gamepad_connected = True
    manager.gamepad_id = "DualSense"
    await manager.connect_controller(current)

    viewer.send_text.assert_awaited_once_with(
        '{"id": "DualSense", "connected": false}'
    )


@pytest.mark.asyncio
async def test_broadcast_sends_to_viewer_that_only_receives_messages() -> None:
    manager = ConnectionManager()
    viewer = AsyncMock()
    manager.viewers.append(viewer)

    await manager.broadcast('{"id":"DualSense","connected":false}')

    viewer.send_text.assert_awaited_once_with(
        '{"id":"DualSense","connected":false}'
    )


@pytest.mark.asyncio
async def test_controller_disconnect_broadcasts_lifecycle_event(monkeypatch) -> None:
    manager = ConnectionManager()
    controller = AsyncMock()
    controller.receive_text.side_effect = [
        '{"id":"DualSense","connected":true}',
        WebSocketDisconnect(),
    ]
    viewer = AsyncMock()
    manager.viewers.append(viewer)
    monkeypatch.setattr(relay, "manager", manager)

    await relay.controller_endpoint(controller)

    assert [call.args[0] for call in viewer.send_text.await_args_list] == [
        '{"id":"DualSense","connected":true}',
        '{"id": "DualSense", "connected": false}',
    ]
    assert manager.controller is None
    assert manager.viewers == [viewer]


def test_update_gamepad_state_tracks_connection() -> None:
    manager = ConnectionManager()

    assert manager.update_gamepad_state(
        '{"id":"DualSense","connected":true}'
    )
    assert manager.gamepad_connected is True
    assert manager.gamepad_id == "DualSense"

    assert manager.update_gamepad_state(
        '{"id":"DualSense","connected":false}'
    )
    assert manager.gamepad_connected is False
    assert manager.gamepad_id == "DualSense"


def test_disconnect_controller_returns_last_connected_gamepad() -> None:
    manager = ConnectionManager()
    controller = AsyncMock()
    manager.controller = controller
    manager.gamepad_connected = True
    manager.gamepad_id = "DualSense"

    assert manager.disconnect_controller(controller) == "DualSense"
    assert manager.controller is None
    assert manager.gamepad_connected is False
    assert manager.gamepad_id is None


def test_disconnect_controller_returns_last_gamepad_even_after_state_disconnect() -> None:
    manager = ConnectionManager()
    controller = AsyncMock()
    manager.controller = controller
    manager.gamepad_connected = False
    manager.gamepad_id = "DualSense"

    assert manager.disconnect_controller(controller) == "DualSense"
