# PadRelay

PadRelay 是面向双机直播和录制场景的远程手柄可视化工具。它将游戏电脑上的手柄输入实时传送到推流电脑，并通过浏览器页面展示操作状态，可直接用作 OBS 浏览器源。

```text
手柄 -> 游戏电脑（客户端） -> 推流电脑（服务端） -> OBS 浏览器源
```

客户端支持 Linux 和 Windows，支持手柄热插拔，并会根据手柄型号自动选择对应皮肤。

## 使用客户端

从项目的 Releases 页面下载对应系统的客户端，并在连接手柄的游戏电脑上运行。

服务端运行在同一台电脑时，直接启动：

```bash
./pad-relay-client
```

Windows：

```powershell
.\pad-relay-client.exe
```

双机使用时，通过 `-url` 指定推流电脑上的服务端地址：

```bash
./pad-relay-client -url ws://192.168.1.10:8000/ws/controller
```

使用 HTTPS 域名部署的服务端应使用 `wss://`：

```bash
./pad-relay-client -url wss://pad.example.com/ws/controller
```

客户端连接后会自动检测手柄。运行期间可以直接拔出或更换手柄，无需重启。

在浏览器中打开服务端地址即可查看手柄状态：

```text
http://192.168.1.10:8000
```

在 OBS 中添加“浏览器”源，并将上述地址填写为 URL，即可将手柄画面加入直播或录制场景。

如果 OBS 与服务端运行在同一台推流电脑上，浏览器源 URL 可以直接填写 `http://127.0.0.1:8000`。

遇到未识别或按键映射异常的手柄时，可以查看原始 SDL 输入：

```bash
./pad-relay-client -diagnose
```

在 Windows 上，如果客户端在后台无法捕获手柄操作，可以尝试右键 `pad-relay-client.exe` 并选择“以管理员身份运行”，尤其是游戏或 OBS 本身也以管理员身份运行时。

## 部署服务端

在推流电脑上安装 Python 3.13 和 [uv](https://docs.astral.sh/uv/)。进入项目目录并安装锁定的依赖：

```bash
uv sync --frozen
```

完成后启动服务端：

```bash
uv run python main.py
```

服务端默认监听 `0.0.0.0:8000`，局域网内的其他设备可以直接访问。通过 `PORT` 修改端口：

```bash
PORT=9000 uv run python main.py
```

### systemd

在 Linux 上可以使用 systemd 保持服务常驻。假设项目位于 `/opt/pad-relay`，创建 `/etc/systemd/system/pad-relay.service`。将 `YOUR_USER` 替换为实际用户，并通过 `command -v uv` 确认 `uv` 的安装路径：

```ini
[Unit]
Description=PadRelay server
After=network.target

[Service]
Type=simple
User=YOUR_USER
WorkingDirectory=/opt/pad-relay
Environment=PORT=8000
ExecStart=/usr/local/bin/uv run --frozen python main.py
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now pad-relay
sudo systemctl status pad-relay
```

更新服务端后，重新同步依赖并重启服务：

```bash
uv sync --frozen
sudo systemctl restart pad-relay
```

### HTTPS 反向代理

需要通过域名访问时，可以使用 Caddy 提供 HTTPS。Caddy 会自动代理 WebSocket，无需额外配置：

```caddyfile
pad.example.com {
    reverse_proxy 127.0.0.1:8000
}
```

浏览器访问 `https://pad.example.com`，客户端连接 `wss://pad.example.com/ws/controller`。

## 安全说明

PadRelay 当前没有身份认证，并按单个手柄客户端场景设计。建议仅部署在可信局域网或 VPN 内，不要将服务端端口直接暴露到公网，也不要同时连接多个客户端。

## 许可证

PadRelay 基于 [MIT License](LICENSE) 发布，Copyright (c) 2026 Azusa-mikan。

浏览器端手柄视图基于 [e7d/gamepad-viewer](https://github.com/e7d/gamepad-viewer) 修改，原项目同样采用 MIT License，Copyright (c) 2017-2020 Michaël "e7d" Ferrand。
