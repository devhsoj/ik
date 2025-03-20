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

	_, event, dataLength, err := readPacketMetadata(c.r)

	if err != nil {
		return nil, err
	}

	buf := make([]byte, dataLength)

	if _, err = io.ReadFull(c.r, buf); err != nil {
		return nil, err
	}

	if event == errorEvent {
		return nil, errors.New(string(buf))
	}

	return buf, nil
}

const streamReadBufferSize = 1_024

// Stream pipes data read from the passed reader to the Server the Client is connected to.
func (c *Client) Stream(event string, r io.Reader) error {
	if r == nil {
		return errors.New("ik: reader is nil")
	}

	buf := make([]byte, streamReadBufferSize)

	for {
		n, err := r.Read(buf)

		if err != nil && err != io.EOF {
			break
		}

		if n == 0 {
			break
		}

		c.mu.Lock()

		if err = c.sendEvent(event, buf[:n]); err != nil {
			if err = c.w.Flush(); err != nil {
				log.Printf("ik: failed to flush writer: %s", err)
			}

			c.mu.Unlock()

			return err
		}

		_, event, dataLength, err := readPacketMetadata(c.r)

		if err != nil {
			return err
		}

		res := make([]byte, dataLength)

		if _, err = io.ReadFull(c.r, res); err != nil {
			log.Printf("ik: failed to read stream response: %s\n", err)
		}

		if event == errorEvent {
			c.mu.Unlock()
			return errors.New(string(res))
		}

		c.mu.Unlock()

		if n < streamReadBufferSize {
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
		addr: opts.Addr,
	}
}
