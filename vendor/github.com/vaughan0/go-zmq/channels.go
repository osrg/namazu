package zmq

import (
	"sync"
)

// Channels provides a method for using Go channels to send and receive messages on a Socket. This is useful not only
// because it allows one to use select for sockets, but also because Sockets by themselves are not thread-safe (ie. one
// should not Send and Recv on the same socket from different threads).
type Channels struct {
	stopch   chan bool
	wg       sync.WaitGroup
	socket   *Socket       // Target socket
	insock   *Socket       // Read-end of outgoing messages socket
	outsock  *Socket       // Write-end of outgoing messages socket
	closein  *Socket       // Read-end of closing socket
	closeout *Socket       // Write-end of closing socket
	in       chan [][]byte // Incoming messages
	out      chan [][]byte // Outgoing messages
	errors   chan error    // Error notification channel
}

// Creates a new Channels object with the given channel buffer size.
func (s *Socket) ChannelsBuffer(chanbuf int) (c *Channels) {
	s.SetSendTimeout(0)
	s.SetRecvTimeout(0)
	c = &Channels{
		stopch: make(chan bool),
		socket: s,
		in:     make(chan [][]byte, chanbuf),
		out:    make(chan [][]byte, chanbuf),
		errors: make(chan error, 2),
	}
	c.insock, c.outsock = s.ctx.MakePair()
	c.closein, c.closeout = s.ctx.MakePair()
	c.wg.Add(2)
	go c.processOutgoing()
	go c.processSockets()
	return
}

// Creates a new Channels object with the default channel buffer size (zero).
func (s *Socket) Channels() *Channels {
	return s.ChannelsBuffer(0)
}

// Closes the Channels object. This will ensure that a number of internal sockets are closed, and that worker goroutines
// are stopped cleanly.
func (c *Channels) Close() {
	close(c.stopch)
	c.wg.Wait()
}

func (c *Channels) In() <-chan [][]byte {
	return c.in
}
func (c *Channels) Out() chan<- [][]byte {
	return c.out
}
func (c *Channels) Errors() <-chan error {
	return c.errors
}

func (c *Channels) processOutgoing() {
	defer c.wg.Done()
	defer c.outsock.Close()
	defer func() {
		c.closeout.SendPart([]byte{}, false)
		c.closeout.Close()
	}()
	for {
		select {
		case <-c.stopch:
			return
		case msg := <-c.out:
			if err := c.outsock.Send(msg); err != nil {
				c.errors <- err
				goto Error
			}
		}
	}
Error:
	for {
		select {
		case <-c.stopch:
			return
		case _ = <-c.out:
			/* discard outgoing messages */
		}
	}
}

func (c *Channels) processSockets() {
	defer c.wg.Done()
	defer c.insock.Close()
	defer c.closein.Close()
	defer close(c.in)

	var poller PollSet
	poller.Socket(c.socket, In)
	poller.Socket(c.insock, In)
	poller.Socket(c.closein, In)
	var sending [][]byte

	for {

		if sending == nil {
			poller.Update(0, In) // Don't monitor main socket for send events
			poller.Update(1, In) // Monitor the outgoing messages socket
		} else {
			poller.Update(0, In|Out) // Monitor the main socket for send events
			poller.Update(1, None)   // Don't monitor the outgoing messages socket, we're waiting for sending to go through
		}
		_, err := poller.Poll(-1)
		if err != nil {
			c.errors <- err
			goto Error
		}

		if poller.Events(0).CanRecv() {
			// Receive a new incoming message
			incoming, err := c.socket.Recv()
			if err != nil {
				if err != ErrTimeout {
					c.errors <- err
					goto Error
				}
			} else {
				select {
				case c.in <- incoming:
				case <-c.stopch:
				}
			}
		}

		if poller.Events(0).CanSend() {
			// Send the outgoing message
			if sending == nil {
				panic("sending is nil")
			}
			if err := c.socket.Send(sending); err != nil {
				if err != ErrTimeout {
					c.errors <- err
					goto Error
				}
			} else {
				sending = nil
			}
		}

		if poller.Events(1).CanRecv() {
			// Receive a new outgoing message
			outgoing, err := c.insock.Recv()
			if err != nil {
				c.errors <- err
				goto Error
			}
			if sending != nil {
				panic("sending should be nil")
			}
			sending = outgoing
		}

		if poller.Events(2).CanRecv() {
			// Check for close message
			_, err := c.closein.Recv()
			if err != nil && err != ErrTimeout {
				c.errors <- err
				goto Error
			} else if err == nil {
				// We're done
				if sending != nil {
					c.sendFinal(sending)
				}
				return
			}
		}

	}
	return

Error:
	poller.Update(0, In)
	poller.Update(1, In)
	poller.Update(2, In)

	for {
		_, err := poller.Poll(-1)
		if err != nil {
			return
		}
		if poller.Events(0).CanRecv() {
			// Discard new incoming message
			if _, err = c.socket.Recv(); err != nil && err != ErrTimeout {
				return
			}
		}
		if poller.Events(1).CanRecv() {
			// Discard outgoing message
			if _, err = c.insock.Recv(); err != nil && err != ErrTimeout {
				return
			}
		}
		if poller.Events(2).CanRecv() {
			_, err = c.closein.Recv()
			if err != nil && err != ErrTimeout {
				return
			} else if err == nil {
				return
			}
		}
	}

}

func (c *Channels) sendFinal(msg [][]byte) {
	linger := c.socket.GetLinger()
	c.socket.SetSendTimeout(linger)
	c.socket.Send(msg)
}
