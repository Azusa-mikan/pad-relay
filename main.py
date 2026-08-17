"""
gamepad-remote server
- /ws/controller  — gamepad client connects here
- /ws/viewer      — browser viewer connects here
- /               — serves the viewer frontend

Environment variables:
  PORT  — listening port (default: 8000)
"""

import logging
import os
from pathlib import Path

import uvicorn
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from src.relay import controller_endpoint, viewer_endpoint

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

STATIC_DIR = Path(__file__).parent / "src" / "static"
PORT = int(os.environ.get("PORT", 8000))

app = FastAPI(title="gamepad-remote")

app.add_api_websocket_route("/ws/controller", controller_endpoint)
app.add_api_websocket_route("/ws/viewer",     viewer_endpoint)

app.mount("/", StaticFiles(directory=STATIC_DIR, html=True), name="static")


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=PORT)
