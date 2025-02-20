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
}

func (s *Server) handleConn(conn net.Conn) error {
	var version byte
	var commandName string
	var dataLength int

	var n int
	var err error

	r := bufio.NewReader(conn)

	for {
		version, commandName, dataLength, err = readPacketMetadata(r)

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

		fmt.Println(version, commandName, dataLength == len(data))
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
	return &Server{}
}
