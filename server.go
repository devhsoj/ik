package ik

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	l net.Listener
	e EventHandlerMap
}

func (s *Server) Register(event string, handler EventHandler) {
	s.e.Register(event, handler)
}

func (s *Server) handleConn(conn net.Conn) error {
	var event string
	var dataLength int

	var n int
	var err error

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		_, event, dataLength, err = readPacketMetadata(r)

		if err != nil {
			break
		}

		data := make([]byte, dataLength)

		if n, err = io.ReadFull(r, data); err != nil {
			break
		}

		if n != dataLength {
			err = errors.New("ik: invalid data sent")
			break
		}

		handler, ok := s.e[event]

		if !ok {
			err = errors.New(fmt.Sprintf("ik: event: '%s' not registered", event))
			break
		}

		res := handler(data)

		if err = sendPacket(w, ProtoVersion, event, res); err != nil {
			break
		}
	}

	return err
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

func NewServer() *Server {
	return &Server{
		e: make(EventHandlerMap),
	}
}
