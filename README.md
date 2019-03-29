# goredis

In the wild, Go is used as a low-level, though not entirely "systems" language. While it's found in great many places these days, its two mainstays are network programming and cloud. During our first few weeks we're going to work on both, while learning some concepts of Go and AWS. Ultimately, we will try to create a standalone binary that implements (some of) [Redis protocol](https://redis.io/topics/protocol) while using DynamoDB for storage.

## Week 2

In week 2, we are going to make our implementation stateful and give it some decent test coverage. We need to make our application stateful in order to support authorization. Let's see what it looks like when using a "real" Redis, with "bacon" as our password:

```text
# Starting a Redis server in one window:
$  docker run --rm -it -p 6379:6379 redis redis-server --requirepass bacon

# Starting telnet in the other window:
$ telnet localhost 6379
AUTH cabbage
-ERR invalid password
GET bacon
-NOAUTH Authentication required.
AUTH bacon
+OK
```

Let's pass the expected password to our binary via the environment and ensure nothing goes through without previous successful `AUTH` command. Perhaps you will have noticed that I created a `sessionHandler` struct to handle each client session - perhaps it's not a bad place to store authorization state?

The other - more challenging - task will be to provide decent test coverage for your application. You may have noticed that we made `store` an interface, with just one implementation for now (`inMemoryStore`). In the future, we can use it to plug in a DynamoDB implementation. For now, we can use it to test the `sessionHandler` with a mock store that can return errors that the `inMemoryStore` won't.

You may have also noticed that `sessionHandler` takes `io.ReadWriteCloser` rather than a more specialised `net.Conn`. This should allow you to wrap something like `bytes.Buffer` in a closer, and use it for testing purposes.

Please use [`stretchr/testify`](https://github.com/stretchr/testify) for testing - in particular `suite` to run test suites and `mock` for mocking out the `store` interface. Please use [codecov.io](https://codecov.io) to display coverage. I recommend using CircleCI to run your tests like so:

```yaml
- run:
    name: Test (go test)
    command: go test -race -coverprofile=coverage.txt -covermode=atomic ./...

- run:
    name: Upload coverage data
    command: bash <(curl -s https://codecov.io/bash)
```

When testing, please **make sure you're testing your code, not the standard library or third party dependencies**.

## Week 1

In week 1, we are only going to create a small TCP server (no TLS yet) that we can talk to from `telnet`. Ideally, we want to be able to set and get keys using the human-readable [inline commands](https://redis.io/topics/protocol#inline-commands). For now, let's keep keys and values in memory, as a map.

The basic interaction we're looking for this week is this:

```text
$ telnet localhost 6379
Trying ::1...
Connected to localhost.
Escape character is '^]'.
SET key value
+OK
GET key
$5
value
^]
telnet> Connection closed.
```

On a Mac you will probably need to install `telnet`. If you're using `brew`, the command is just:

```
$ brew install telnet
```

Resources:

- [Build a concurrent TCP server in Go](https://opensource.com/article/18/5/building-concurrent-tcp-server-go);
- [`net/textproto` library](https://golang.org/pkg/net/textproto/);
- [telnet tutorial](https://www.computerhope.com/unix/utelnet.htm) if you ever need it;

If you're bored and want to do something extra, try implementing key expiration and other commands that strike your fancy.
