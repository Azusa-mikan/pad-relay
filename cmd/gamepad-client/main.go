package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

const (
	defaultURL         = "ws://localhost:8000/ws/controller"
	clientFeatureLevel = "platform-backends-v1"
)

var clientVersion = "dev"

func main() {
	os.Exit(realMain())
}

func realMain() int {
	url := flag.String("url", defaultURL, "WebSocket 服务端地址")
	diagnose := flag.Bool("diagnose", false, "输出手柄诊断信息")
	showVersion := flag.Bool("version", false, "输出客户端版本")
	flag.Parse()
	if *showVersion {
		printVersion()
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	updates := make(chan snapshot, 1)
	if !*diagnose {
		go runSender(ctx, *url, updates)
	}
	if err := runInput(ctx, updates, *diagnose); err != nil {
		log.Printf("错误: %v", err)
		return 1
	}
	return 0
}

func printVersion() {
	revision := "unknown"
	modified := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				modified = setting.Value
			}
		}
	}
	fmt.Printf("PadRelay client %s (feature=%s, revision=%s, modified=%s)\n",
		clientVersion, clientFeatureLevel, revision, modified)
}

func publishSnapshot(updates chan snapshot, state snapshot) {
	select {
	case updates <- state:
		return
	default:
	}
	select {
	case <-updates:
	default:
	}
	updates <- state
}
