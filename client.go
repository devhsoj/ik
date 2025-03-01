package ik

import (
	"bufio"
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

	subscribed       chan bool
	streamBufferSize int
}

func (c *Client) sendPacket(event string, data []byte) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	return sendPacket(c.w, ProtoVersion, event, data)
}

func (c *Client) Send(event string, data []byte) ([]byte, error) {
	if err := c.sendPacket(event, data); err != nil {
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

func (c *Client) Subscribe(event string, data []byte, handler func(data []byte)) error {
	if err := c.sendPacket(event, data); err != nil {
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

func (c *Client) Stream(event string, r io.Reader) error {
	buf := make([]byte, c.streamBufferSize)

	for {
		n, err := r.Read(buf)

		if err != nil && err != io.EOF {
			break
		}

		if n == 0 {
			break
		}

		c.mu.Lock()

		if err = c.sendPacket(event, buf[:n]); err != nil {
			return err
		}

		c.mu.Unlock()

		if n < c.streamBufferSize {
			break
		}
	}

	return nil
}

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
	Addr             string
	StreamBufferSize int
}

func NewClient(opts ClientOptions) *Client {
	if opts.StreamBufferSize <= 0 {
		opts.StreamBufferSize = 1_024
	}

	return &Client{
		addr:             opts.Addr,
		subscribed:       make(chan bool, 1),
		streamBufferSize: opts.StreamBufferSize,
	}
}
