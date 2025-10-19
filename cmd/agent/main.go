// Entry point for agent app.
package main

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/internal/buildinfo"
	"github.com/Nikolay961996/metsys/models"
)

const (
	addr = ":8081"
)

func main() {
	buildinfo.PrintHello()
	err := models.Initialize("info")
	if err != nil {
		panic(err)
	}

	c := agent.DefaultConfig()
	c.Parse()

	a := agent.InitAgent()
	go a.Run(&c)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	srv := runBackground()
	gracefulShutdown(&a, srv, sigCh)
}

func runBackground() *http.Server {
	srv := &http.Server{Addr: addr}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	return srv
}

func gracefulShutdown(a *agent.Entity, srv *http.Server, sigCh <-chan os.Signal) {
	<-sigCh
	a.Stop()
	time.Sleep(2 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
}
