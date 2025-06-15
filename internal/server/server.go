package server

import (
	"github.com/Nikolay961996/metsys/internal/server/router"
	"net/http"
)

type Entity struct {
	config *Config
}

func InitServer(c *Config) Entity {
	a := Entity{
		config: c,
	}
	return a
}

func (s *Entity) Run() {
	err := http.ListenAndServe(s.config.RunOnServerAddress, router.MetricsRouter())
	if err != nil {
		panic(err)
	}
}
