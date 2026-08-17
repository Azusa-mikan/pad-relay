package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func runSender(ctx context.Context, url string, updates <-chan snapshot) {
	backoff := time.Second
	var latest snapshot
	haveLatest := false

	for ctx.Err() == nil {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			log.Printf("WebSocket 连接失败 (%v)，%s 后重试", err, backoff)
			if !waitForRetry(ctx, backoff) {
				return
			}
			backoff = min(backoff*2, 30*time.Second)
			continue
		}

		log.Printf("已连接到 %s", url)
		backoff = time.Second
		for {
			select {
			case latest = <-updates:
				haveLatest = true
			default:
				goto drained
			}
		}
	drained:
		if haveLatest {
			if err := writeSnapshot(conn, latest); err != nil {
				conn.Close()
				continue
			}
		}

		readDone := make(chan error, 1)
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					readDone <- err
					return
				}
			}
		}()

		ping := time.NewTicker(15 * time.Second)
	connected:
		for {
			select {
			case <-ctx.Done():
				_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
				ping.Stop()
				conn.Close()
				return
			case err := <-readDone:
				log.Printf("WebSocket 连接断开 (%v)", err)
				break connected
			case latest = <-updates:
				haveLatest = true
				if err := writeSnapshot(conn, latest); err != nil {
					log.Printf("WebSocket 发送失败 (%v)", err)
					break connected
				}
			case <-ping.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					log.Printf("WebSocket 心跳失败 (%v)", err)
					break connected
				}
			}
		}
		ping.Stop()
		conn.Close()
	}
}

func writeSnapshot(conn *websocket.Conn, state snapshot) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, payload)
}

func waitForRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
