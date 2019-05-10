# goredis

In the wild, Go is used as a low-level, though not entirely "systems" language. While it's found in great many places these days, its two mainstays are network programming and cloud. During our first few weeks we're going to work on both, while learning some concepts of Go and AWS. Ultimately, we will try to create a standalone binary that implements (some of) [Redis protocol](https://redis.io/topics/protocol) while using DynamoDB for storage.

## Week 4

In week 3, we created and tested an implementation of `Store` backed by DynamoDB. In order to maintain performance, we also created another implementation of `Store` that uses memory for caching, and DynamoDB as the source of authority. It's not a production-ready implementation because it's pretty heavy on memory and does do any automated expiry - something you'd probably want from a proper implementation like the one [here](https://github.com/bluele/gcache).

Nevertheless, in last week will try to tweak our design in such a way as to support an arbitrarily large cluster of our own Redises, while maintaining the highest possible cache hit ratio. In order to do that, we need to have each of our servers follow the latest changes to DynamoDB.

Luckily, Amazon has a very useful feature called [DynamoDB Streams](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html), which allows to subscribe to database changes in a way similar to Kafka or Kinesis. We will want to link our cache to the stream from our database, and populate it when any new keys are set. You may choose to use a helper library [like this one](https://github.com/urakozz/go-dynamodb-stream-subscriber) or use raw AWS SDK - just note that the latter involves quite a bit of boilerplate.

## Week 3

In week 2 we decided to skip the authentication bit and only focus on testing, which is always a huge topic, especially in Go where there usually isn't one established way of doing things. We will roll over the authentication bit to week 3.

The main topic of week 3 however is DynamoDB integration. You will need to set up a simple table to store key-value combinations, and provide the implementation of the `Store` interface that works with DynamoDB API.

As a reminder, `Store` is defined as:

```go
type Store interface {
    Get(key string) (value string, found bool, err error)
    Set(key string, value string) error
}
```

Unlike in-memory implementation, this one will be capable of returning actual errors!

AWS integration is usually tricky to test. There are two possible approaches - either running a local DynamoDB and connecting to that for your tests, or mocking out AWS calls. We are going to go down the second route, though feel free to experiment with the first one using Docker Compose.

Luckily, for each Go SDK, AWS provides interfaces that are implemented by actual client, and your Dynamo implemenation of the `Store` should use [that interface](https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbiface/#DynamoDBAPI) to abstract away the actual client.

One problem here is that AWS interfaces are giant, so mocking them would technically require writing implementations of methods that are never used. Here, Go composition comes to the rescue. You can do something like this:

```go
type mockSomethingComplex struct {
    mock.Mock
    something.Complex
}

func (m *mockSomethingComplex) MockedMethod() error {
    return m.Called().Error(0)
}
```

In this approach, we will only need to mock out methods that are actually used, and have all others forwarded to `nil` instance of `something.Complex` (which - if ever used - would obviosly segfault).

Another thing to note is that there are multiple ways of setting up AWS session. Credentials can come from explicitly set environment variables, a credentials file, instance metadata (in EC2, ECS, EKS and Lambda), and perhaps a few other places I'm not aware of. The usual way of setting up AWS sessions and clients allows the SDK to figure out where it's running and how it should be set up:

```go
sess := session.Must(session.NewSession())
client := dynamodb.New(sess)
```

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
