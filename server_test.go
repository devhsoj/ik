package ik_test

import (
    "fmt"
    "github.com/devhsoj/ik"
    "testing"
)

var addr = "localhost:48923"

func TestServer_Listen(t *testing.T) {
	server := ik.NewServer()

	if err := server.Listen(addr); err != nil {
		t.Error(err)
	}
}

func TestServer_Register(t *testing.T) {
	server := ik.NewServer()

	server.Register("echo", func(c *ik.ServerClient, data []byte) []byte {
		return data
	})

	if err := server.Listen(addr); err != nil {
		t.Error(err)
	}
}

func TestServer_RegisterStream(t *testing.T) {
	server := ik.NewServer()

	server.Register("stream", func(c *ik.ServerClient, data []byte) []byte {
		return []byte(fmt.Sprintf("got %d bytes", len(data)))
	})

	if err := server.Listen(addr); err != nil {
		t.Error(err)
	}
}
