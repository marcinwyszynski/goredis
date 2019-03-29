package main

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/google/shlex"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type sessionHandler struct {
	conn   io.ReadWriteCloser
	logger *logrus.Entry
	store  store
}

func (s *sessionHandler) badArgs(command string) error {
	_, err := fmt.Fprintf(s.conn, "-ERR wrong number of arguments for '%s' command\n", command)
	return err
}

func (s *sessionHandler) handle() {
	defer s.logger.Info("Closed connection")
	defer s.conn.Close()

	buffer := bufio.NewReader(s.conn)
	for {
		command, err := textproto.NewReader(buffer).ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			s.logger.Errorf("Could not read command: %v", err)
			break
		}

		err = s.handleCommand(command)
		if err == nil {
			continue
		}

		if err != io.EOF {
			s.logger.Errorf("Could not handle command %s: %v", command, err)
		}

		break
	}
}

func (s *sessionHandler) handleCommand(command string) error {
	args, err := shlex.Split(command)
	if err != nil {
		_, err := fmt.Fprintf(s.conn, "-ERR malformed line: %v\n", err)
		return err
	}

	switch strings.ToLower(args[0]) {
	case "get":
		return s.handleGet(args[1:])
	case "ping":
		return s.handlePing(args[1:])
	case "set":
		return s.handleSet(args[1:])
	default:
		return s.handleUnknown(args)
	}
}

func (s *sessionHandler) handleGet(args []string) error {
	if len(args) != 1 {
		return s.badArgs("get")
	}

	value, found, err := s.store.get(args[0])
	if err != nil {
		return errors.Wrap(err, "could not read from the store")
	}

	if !found {
		_, err := fmt.Fprintln(s.conn, "$-1")
		return err
	}

	_, err = fmt.Fprintf(s.conn, "$%d\n%s\n", len(value), value)
	return err
}

func (s *sessionHandler) handlePing(args []string) error {
	if len(args) > 1 {
		return s.badArgs("ping")
	}

	response := "PONG"
	if len(args) == 1 {
		response = args[0]
	}

	_, err := fmt.Fprintf(s.conn, "%q\n", response)
	return err
}

func (s *sessionHandler) handleSet(args []string) error {
	if len(args) < 2 {
		return s.badArgs("set")
	}

	if err := s.store.set(args[0], args[1]); err != nil {
		return errors.Wrap(err, "could not write to the store")
	}

	_, err := fmt.Fprintln(s.conn, "+OK")
	return err
}

func (s *sessionHandler) handleUnknown(args []string) error {
	_, err := fmt.Fprintf(s.conn, "-ERR unknown command `%s`, with args beginning with %s\n", args[0], args[1:])
	return err
}
