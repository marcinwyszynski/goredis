package main

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type sessionHandler struct {
	conn   io.ReadWriteCloser
	logger *logrus.Entry
	store  store
}

func (s *sessionHandler) handle() {
	defer s.logger.Info("Closed connection")
	defer s.conn.Close()

	buffer := bufio.NewReader(s.conn)
	for {
		line, err := textproto.NewReader(buffer).ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			s.logger.Errorf("Could not read command: %v", err)
			break
		}

		command := strings.Split(line, " ")
		switch strings.ToLower(command[0]) {
		case "get":
			err = s.handleGet(command[1:])
		case "ping":
			err = s.handlePing(command[1:])
		case "set":
			err = s.handleSet(command[1:])
		default:
			err = s.handleUnknown(command)
		}

		if err == nil {
			continue
		}

		if err != io.EOF {
			s.logger.Errorf("Could not handle command %s: %v", line, err)
		}

		break
	}
}

func (s *sessionHandler) handleGet(args []string) error {
	if len(args) < 1 {
		_, err := fmt.Fprintln(s.conn, "-ERR wrong number of arguments for 'get' command")
		return err
	}

	key, args, err := consumeQuotes(args)
	if err != nil {
		return err
	}

	value, found, err := s.store.get(key)
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
	response := "PONG"

	if len(args) > 0 {
		response = strings.Trim(strings.Join(args, " "), "\"")
	}

	_, err := fmt.Fprintf(s.conn, "%q\n", response)
	return err
}

func (s *sessionHandler) handleSet(args []string) error {
	if len(args) < 2 {
		_, err := fmt.Fprintln(s.conn, "-ERR wrong number of arguments for 'set' command")
		return err
	}

	key, args, err := consumeQuotes(args)
	if err != nil {
		return err
	}

	value, _, err := consumeQuotes(args)
	if err != nil {
		return err
	}

	if err := s.store.set(key, value); err != nil {
		return errors.Wrap(err, "could not write to the store")
	}

	_, err = fmt.Fprintln(s.conn, "+OK")
	return err
}

func (s *sessionHandler) handleUnknown(args []string) error {
	_, err := fmt.Fprintf(s.conn, "-ERR unknown command `%s`, with args beginning with %s\n", args[0], args[1:])
	return err
}

func consumeQuotes(args []string) (value string, rest []string, err error) {
	if !strings.HasPrefix(args[0], "\"") {
		return args[0], args[1:], nil
	}

	var valueArgs []string
	for index, arg := range args {
		valueArgs = append(valueArgs, arg)
		rest = args[index+1:]
		if strings.HasSuffix(arg, "\"") {
			value = strings.Trim(strings.Join(valueArgs, " "), "\"")
			return
		}
	}

	err = errors.New("unbalanced quotes")
	return
}
