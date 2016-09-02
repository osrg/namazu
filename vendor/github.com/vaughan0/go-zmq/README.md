go-zmq
=======

Go (golang) bindings for ZeroMQ (zmq, 0mq), currently built for ZeroMQ 3.2.x. Pull requests welcome.

View the API docs [here](http://godoc.org/github.com/vaughan0/go-zmq).

Basic Usage
-----------

Import go-zmq:

```go
import "github.com/vaughan0/go-zmq"
```

Create a new ZeroMQ Context and Socket:

```go
ctx, err := zmq.NewContext()
if err != nil {
  panic(err)
}
defer ctx.Close()
```

Bind to a local endpoint:

```go
sock, err := ctx.Socket(zmq.Rep)
if err != nil {
  panic(err)
}
defer sock.Close()

if err = sock.Bind("tcp://*:5555"); err != nil {
  panic(err)
}
```

Receive and send messages:

```go
for {
  parts, err := sock.Recv()
  if err != nil {
    panic(err)
  }
  response := fmt.Sprintf("Received %d message parts", len(parts))
  if err = sock.Send([][]byte{
    []byte(response),
  }); err != nil {
    panic(err)
  }
}
```

Using Channels
--------------

ZeroMQ sockets are not thread-safe, which would make any ambitious ZeroMQ-using gopher sad.
Luckily go-zmq provides a (thread-safe) way to use sockets with native Go channels.
This also allows one to use the select construct with ZeroMQ sockets.

Start by using the Channels() method of a socket:

```go
chans := sock.Channels()
defer chans.Close()
```

Now you can send and receive messages using the channels returned by chans.Out() and chans.In(),
respectively. Don't forget to also check chans.Errors() to see if any errors occur.

```go
for {
  select {
  case msg := <-chans.In():
    go func() {
      resp := doSomething(msg)
      chans.Out() <- resp
    }()
  case err := <-chans.Errors():
    panic(err)
  }
}
```
