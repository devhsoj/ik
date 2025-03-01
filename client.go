package ik

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

type Client struct {
	addr string
	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Reader
	mu   sync.Mutex

	subscribed chan bool
}

// sendEvent makes sure the Client is connected, then sends the passed event and data via sendPacket.
func (c *Client) sendEvent(event string, data []byte) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	return sendPacket(c.w, ProtoVersion, event, data)
}

// Send sends the passed event and data, then returns a response from the Server the Client is connected to.
func (c *Client) Send(event string, data []byte) ([]byte, error) {
	if err := c.sendEvent(event, data); err != nil {
		return nil, err
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	_, _, dataLength, err := readPacketMetadata(c.r)

	if err != nil {
		return nil, err
	}

	buf := make([]byte, dataLength)

	if _, err = io.ReadFull(c.r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// Subscribe sends the passed event and data, locks the client, then via a go-routine, calls the passed handler with the
// responses from the registered EventHandler from the Server the Client is connected to.
func (c *Client) Subscribe(event string, data []byte, handler func(data []byte)) error {
	if err := c.sendEvent(event, data); err != nil {
		return err
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	c.subscribed <- true

	go func() {
		for {
			select {
			case subscribed := <-c.subscribed:
				if !subscribed {
					break
				}
			default:
				_, _, dataLength, err := readPacketMetadata(c.r)

				if err != nil {
					log.Printf("ik: failed to read packet metadata on subscription to '%s': %v", event, err)
				}

				buf := make([]byte, dataLength)

				if _, err = io.ReadFull(c.r, buf); err != nil {
					log.Printf("ik: failed to read packet data on subscription to '%s': %v", event, err)
				}

				handler(buf)
			}
		}
	}()

	return nil
}

type ClientStreamOptions struct {
	// Event is the event to specify this stream for.
	Event string
	// Reader is the reader that will stream to the Server the Client is connected to.
	Reader io.Reader
	// ReadBufferSize is the read buffer size used to created buffered reads for streams. Defaults to 1 KiB.
	ReadBufferSize int
	// Handler is a function that is called when a response is received from the Server the Client is streaming to.
	Handler func(data []byte)
}

// Stream pipes data read from the passed reader to the Server the Client is connected to.
func (c *Client) Stream(opts ClientStreamOptions) error {
	if opts.Reader == nil {
		return errors.New("ik: reader is nil")
	}

	if opts.ReadBufferSize <= 0 {
		opts.ReadBufferSize = 1_024
	}

	buf := make([]byte, opts.ReadBufferSize)

	for {
		n, err := opts.Reader.Read(buf)

		if err != nil && err != io.EOF {
			break
		}

		if n == 0 {
			break
		}

		c.mu.Lock()

		if err = c.sendEvent(opts.Event, buf[:n]); err != nil {
			if err = c.w.Flush(); err != nil {
				log.Printf("ik: failed to flush writer: %s", err)
			}

			return err
		}

		_, _, dataLength, err := readPacketMetadata(c.r)

		if err != nil {
			return err
		}

		res := make([]byte, dataLength)

		if _, err = io.ReadFull(c.r, res); err != nil {
			log.Printf("ik: failed to read stream response: %s\n", err)
		}

		if opts.Handler != nil {
			opts.Handler(res)
		}

		c.mu.Unlock()

		if n < opts.ReadBufferSize {
			break
		}
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	return c.w.Flush()
}

// Connect connects the Client to the configured address via TCP.
func (c *Client) Connect() error {
	if c.conn != nil {
		return nil
	}

	conn, err := net.Dial("tcp", c.addr)

	if err != nil {
		return err
	}

	c.conn = conn
	c.w = bufio.NewWriter(conn)
	c.r = bufio.NewReader(conn)

	return nil
}

// Close closes the current subscription, attempts to lock the client, flushes all data to be written, then closes the
// underlying TCP net.Conn.
func (c *Client) Close() error {
	c.subscribed <- false

	c.mu.Lock()

	defer c.mu.Unlock()

	if c.w != nil {
		if err := c.w.Flush(); err != nil {
			return err
		}
	}

	c.r = nil

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

type ClientOptions struct {
	// Addr is the address of the Server to connect to.
	Addr string
}

func NewClient(opts *ClientOptions) *Client {
	if opts == nil {
		opts = &ClientOptions{}
	}

	if opts.Addr == "" {
		opts.Addr = "localhost:48923"
	}

	return &Client{
		addr:       opts.Addr,
		subscribed: make(chan bool, 1),
	}
}
