package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type sessionHandlerTestSuite struct {
	suite.Suite

	buffer    *bytes.Buffer
	conn      *mockReadWriteCloser
	logOutput *bytes.Buffer
	store     *mockStore

	sut *SessionHandler
}

func (s *sessionHandlerTestSuite) SetupTest() {
	s.buffer = bytes.NewBuffer(nil)
	s.conn = &mockReadWriteCloser{ReadWriter: s.buffer}
	s.logOutput = bytes.NewBuffer(nil)
	s.store = new(mockStore)

	logger := logrus.New()
	logger.SetOutput(s.logOutput)

	s.sut = NewSessionHandler(
		s.conn,
		logger.WithField("test", true),
		s.store,
	)
}

func (s *sessionHandlerTestSuite) TestHandle_EOF() {
	s.conn.
		On("Read", mock.AnythingOfType("[]byte")).Return(io.EOF, 0).
		On("Close").Return(nil)

	s.sut.Handle()
	s.Contains(s.logOutput.String(), "Closed connection")
}

func (s *sessionHandlerTestSuite) TestGet_Found() {
	fmt.Fprintln(s.conn, `GET bacon`)

	s.store.On("Get", "bacon").Return("tasty", true, nil)

	s.True(s.sut.handleLine())
	s.responded("$5\ntasty")
}

func (s *sessionHandlerTestSuite) TestGet_NotFound() {
	fmt.Fprintln(s.conn, `GET bacon`)

	s.store.On("Get", "bacon").Return("", false, nil)

	s.True(s.sut.handleLine())
	s.responded("$-1")
}

func (s *sessionHandlerTestSuite) TestGet_InvalidArgs() {
	fmt.Fprintln(s.conn, `GET bacon cabbage`)

	s.True(s.sut.handleLine())
	s.responded("-ERR wrong number of arguments for 'get' command")
}

func (s *sessionHandlerTestSuite) TestGet_StoreError() {
	fmt.Fprintln(s.conn, `GET bacon`)

	s.store.On("Get", "bacon").Return("", false, errors.New("store error"))

	s.False(s.sut.handleLine())
	s.loggedError("Could not handle command GET bacon: could not read from the store: store error")
}

func (s *sessionHandlerTestSuite) TestPing_NoArgs() {
	fmt.Fprintln(s.conn, "PING")

	s.True(s.sut.handleLine())
	s.responded(`"PONG"`)
}

func (s *sessionHandlerTestSuite) TestPing_SingleArg() {
	fmt.Fprintln(s.conn, `PING "hello world"`)

	s.True(s.sut.handleLine())
	s.responded(`"hello world"`)
}

func (s *sessionHandlerTestSuite) TestPing_InvalidArgs() {
	fmt.Fprintln(s.conn, "PING hello world")

	s.True(s.sut.handleLine())
	s.responded("-ERR wrong number of arguments for 'ping' command")
}

func (s *sessionHandlerTestSuite) TestSet_OK() {
	fmt.Fprintln(s.conn, "SET bacon tasty")

	s.store.On("Set", "bacon", "tasty").Return(nil)

	s.True(s.sut.handleLine())
	s.responded("+OK")
}

func (s *sessionHandlerTestSuite) TestSet_Error() {
	fmt.Fprintln(s.conn, "SET bacon tasty")

	s.store.On("Set", "bacon", "tasty").Return(errors.New("store error"))

	s.False(s.sut.handleLine())
	s.loggedError("Could not handle command SET bacon tasty: could not write to the store: store error")
}

func (s *sessionHandlerTestSuite) TestSet_InvalidArgs() {
	fmt.Fprintln(s.conn, "SET bacon")

	s.True(s.sut.handleLine())
	s.responded("-ERR wrong number of arguments for 'set' command")
}

func (s *sessionHandlerTestSuite) TestUnknown_WellFormed() {
	fmt.Fprintln(s.conn, "BACON test")

	s.True(s.sut.handleLine())
	s.responded("-ERR unknown command `BACON`, with args beginning with [test]")
}

func (s *sessionHandlerTestSuite) TestUnknown_Malformed() {
	fmt.Fprintln(s.conn, `BACON "test`)

	s.True(s.sut.handleLine())
	s.responded("-ERR malformed line: EOF found when expecting closing quote")
}

func (s *sessionHandlerTestSuite) loggedError(message string) {
	s.Contains(s.logOutput.String(), fmt.Sprintf(`level=error msg="%s"`, message))
}

func (s *sessionHandlerTestSuite) responded(message string) {
	s.Equal(fmt.Sprintf("%s\n", message), s.buffer.String())
}

func TestSessionHandler(t *testing.T) {
	suite.Run(t, new(sessionHandlerTestSuite))
}
