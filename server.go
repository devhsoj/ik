package ik

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

var ErrEventNotRegistered = errors.New("ik: event not registered")

// ServerClient is an abstraction around an underlying TCP net.Conn with a bufio.Reader and bufio.Writer.
type ServerClient struct {
	c  net.Conn
	r  *bufio.Reader
	w  *bufio.Writer
	mu sync.Mutex
}

// Send sends the following event & data to the connected ServerClient.
func (s *ServerClient) Send(event string, data []byte) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	return sendPacket(s.w, ProtoVersion, event, data)
}

// Receive returns the next event and data sent from the client.
func (s *ServerClient) Receive() (event string, data []byte, err error) {
	var dataLength int

	s.mu.Lock()

	defer s.mu.Unlock()

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

// Register registers an event handler func, triggered when a connected client sends this event.
func (s *Server) Register(event string, handler EventHandler) {
	if s.e == nil {
		s.e = make(eventHandlerMap)
	}

	s.e.Register(event, handler)
}

// handleConn handles each client connection by repeatedly reading events and calling the registered event handler.
func (s *Server) handleConn(conn net.Conn) error {
	client := NewServerClient(conn)

	for {
		event, data, err := client.Receive()

		if err != nil {
			return err
		}

		handler, ok := s.e[event]

		if !ok || handler == nil {
			if err = client.Send("ik-error", []byte(ErrEventNotRegistered.Error())); err != nil {
				return err
			}

			return ErrEventNotRegistered
		}

		res := handler(client, data)

		if err = client.Send(event, res); err != nil {
			return err
		}
	}
}

// serve accepts TCP connections and passes them to handleConn.
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

// Listen creates a TCP net.Listener listening on the passed addr and starts the server client connection serve loop.
func (s *Server) Listen(addr string) error {
	if s.e == nil {
		s.e = make(eventHandlerMap)
	}

	l, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	s.l = l

	return s.serve()
}

// Close closes the underlying TCP listener.
func (s *Server) Close() error {
	return s.l.Close()
}

func NewServer() *Server {
	return &Server{
		e: make(eventHandlerMap),
	}
}
