package main

import (
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/kelseyhightower/envconfig"
	"github.com/marcinwyszynski/goredis/lib"
	"github.com/sirupsen/logrus"
)

type config struct {
	DynamoTable string `envconfig:"DYNAMO_TABLE" required:"true"`
	Port        int    `envconfig:"PORT" default:"6379"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log := logrus.New()

	session := session.Must(session.NewSession())

	store := lib.NewCachingStore(
		&lib.DynamoDBStore{API: dynamodb.New(session), TableName: cfg.DynamoTable},
		lib.NewInMemoryStore(),
	)

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

		go (lib.NewSessionHandler(conn, logger, store)).Handle()
	}
}
