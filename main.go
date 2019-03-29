package main

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type config struct {
	Port int `envconfig:"PORT" default:"6379"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log := logrus.New()
	store := newInMemoryStore()

	port := fmt.Sprintf(":%d", cfg.Port)
	log.Infof("About to start serving on port %s", port)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Could not set up listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Could not accept connection: %v", err)
		}

		logger := log.WithField("remote", conn.RemoteAddr())
		logger.Infoln("Accepted connection")

		go (&sessionHandler{conn, logger, store}).handle()
	}

}
