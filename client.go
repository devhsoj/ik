package ik

import (
	"bufio"
	"io"
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

func (c *Client) Send(event string, data []byte) ([]byte, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	if len(event) >= 65_535 {
		event = event[:65_535]
	}

	buf := craftPacketMetadata(ProtoVersion, event, len(data))

	if _, err := c.w.Write(buf); err != nil {
		return nil, err
	}

	if _, err := c.w.Write(data); err != nil {
		return nil, err
	}

	if err := c.w.Flush(); err != nil {
		return nil, err
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	_, _, dataLength, err := readPacketMetadata(c.r)

	if err != nil {
		return nil, err
	}

	buf = nil
	buf = make([]byte, dataLength)

	if _, err = io.ReadFull(c.r, buf); err != nil {
		return nil, err
	}

	return buf, nil
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

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}
