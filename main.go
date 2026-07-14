package main

import (
	"context"
	"log"
	"time"

	"github.com/TechnoRed2026/redscope/internal/netmon"
	"github.com/TechnoRed2026/redscope/internal/ui"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor := netmon.NewMonitor()
	if err := ui.Run(ctx, monitor, time.Second); err != nil {
		log.Fatal(err)
	}
}
