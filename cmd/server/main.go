// Entry point for server app.
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nikolay961996/metsys/internal/buildinfo"
	"github.com/Nikolay961996/metsys/internal/server"
	"github.com/Nikolay961996/metsys/models"
)

func main() {
	buildinfo.PrintHello()
	err := models.Initialize("info")
	if err != nil {
		panic(err)
	}

	c := server.DefaultConfig()
	c.Parse()
	entity := server.InitServer(&c)
	entity.Run(c.RunOnServerAddress, c.KeyForSigning, c.CryptoKey)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	gracefulShutdown(&entity, sigCh)
}

func gracefulShutdown(entity *server.MetricServer, sigCh <-chan os.Signal) {
	<-sigCh
	entity.Stop(10 * time.Second)
	time.Sleep(100 * time.Millisecond)
}
