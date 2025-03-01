package ik

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
)

var ErrEventNotRegistered = errors.New("ik: event not registered")

type ServerClient struct {
	c net.Conn
	r *bufio.Reader
	w *bufio.Writer
}

func (s *ServerClient) Send(event string, data []byte) error {
	return sendPacket(s.w, ProtoVersion, event, data)
}

func (s *ServerClient) Receive() (event string, data []byte, err error) {
	var dataLength int

	_, event, dataLength, err = readPacketMetadata(s.r)

	if err != nil {
		return "", nil, err
	}

	data = make([]byte, dataLength)

	if _, err = io.ReadFull(s.r, data); err != nil {
		return "", nil, err
	}

	return event, data, nil
}

func NewServerClient(c net.Conn) *ServerClient {
	return &ServerClient{
		c: c,
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
	}
}

type Server struct {
	l net.Listener
	e eventHandlerMap
}

func (s *Server) Register(event string, handler EventHandler) {
	s.e.Register(event, handler)
}

func (s *Server) handleConn(conn net.Conn) error {
	client := NewServerClient(conn)

	for {
		event, data, err := client.Receive()

		if err != nil {
			return err
		}

		handler, ok := s.e[event]

		if !ok || handler == nil {
			return ErrEventNotRegistered
		}

		res := handler(client, data)

		if res == nil {
			continue
		}

		if err = client.Send(event, res); err != nil {
			return err
		}
	}
}

func (s *Server) serve() error {
	for {
		conn, err := s.l.Accept()

		if err != nil {
			return err
		}

		go func() {
			if err = s.handleConn(conn); err != nil && err != io.EOF {
				log.Printf("ik: server failed to properly handle connection: %v", err)
			}
		}()
	}
}

func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	s.l = l

	return s.serve()
}

func (s *Server) Close() error {
	return s.l.Close()
}

func NewServer() *Server {
	return &Server{
		e: make(eventHandlerMap),
	}
}
