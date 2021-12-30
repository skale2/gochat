package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/skale2/gochat/internal/conn"
	"github.com/skale2/gochat/internal/db"
	"github.com/skale2/gochat/internal/log"
	"github.com/skale2/gochat/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Create handler to cancel context on interrupt signal
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	// Initialize application
	log.Initialize(log.FileWriter())
	db.Initialize(db.SQLLiteDB())
	conn.Initialize(ctx)

	srv := server.Create(ctx, cancel, 8080)

	defer func() {
		signal.Stop(interruptChan)
		server.Shutdown(ctx, srv)
		cancel()
		db.Finalize()
		conn.Finalize()
	}()

	go server.Start(srv)
	<-interruptChan
}
