package lib

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

// SessionHandler handles a single client connection.
type SessionHandler struct {
	buffer *textproto.Reader
	conn   io.ReadWriteCloser
	logger *logrus.Entry
	store  Store
}

// NewSessionHandler builds a fully usable SessionHandler.
func NewSessionHandler(conn io.ReadWriteCloser, logger *logrus.Entry, store Store) *SessionHandler {
	return &SessionHandler{
		buffer: textproto.NewReader(bufio.NewReader(conn)),
		conn:   conn,
		logger: logger,
		store:  store,
	}
}

func (s *SessionHandler) badArgs(command string) error {
	_, err := fmt.Fprintf(s.conn, "-ERR wrong number of arguments for '%s' command\n", command)
	return err
}

// Handle handles session connection as a request-response loop.
func (s *SessionHandler) Handle() {
	defer s.logger.Info("Closed connection")
	defer s.conn.Close()

	for {
		if !s.handleLine() {
			break
		}
	}
}

func (s *SessionHandler) handleLine() (keepOpen bool) {
	command, err := s.buffer.ReadLine()
	if err == io.EOF {
		return false
	} else if err != nil {
		s.logger.Errorf("Could not read command: %v", err)
		return false
	}

	err = s.handleCommand(command)
	if err == nil {
		return true
	}

	if err != io.EOF {
		s.logger.Errorf("Could not handle command %s: %v", command, err)
	}

	return false
}

func (s *SessionHandler) handleCommand(command string) error {
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

func (s *SessionHandler) handleGet(args []string) error {
	if len(args) != 1 {
		return s.badArgs("get")
	}

	value, found, err := s.store.Get(args[0])
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

func (s *SessionHandler) handlePing(args []string) error {
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

func (s *SessionHandler) handleSet(args []string) error {
	if len(args) < 2 {
		return s.badArgs("set")
	}

	if err := s.store.Set(args[0], args[1]); err != nil {
		return errors.Wrap(err, "could not write to the store")
	}

	_, err := fmt.Fprintln(s.conn, "+OK")
	return err
}

func (s *SessionHandler) handleUnknown(args []string) error {
	_, err := fmt.Fprintf(s.conn, "-ERR unknown command `%s`, with args beginning with %s\n", args[0], args[1:])
	return err
}
