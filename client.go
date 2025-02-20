package ik

import (
	"bufio"
	"net"
)

type Client struct {
	addr string
	conn net.Conn
	w    *bufio.Writer
}

func (c *Client) Send(event string, data []byte) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	if len(event) >= 65_535 {
		event = event[:65_535]
	}

	buf := craftPacketMetadata(ProtoVersion, event, len(data))

	if _, err := c.w.Write(buf); err != nil {
		return err
	}

	if _, err := c.w.Write(data); err != nil {
		return err
	}

	return c.w.Flush()
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

	return nil
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}
