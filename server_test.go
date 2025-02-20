package ik_test

import (
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

	server.Register("echo", func(data []byte) []byte {
		return data
	})

	if err := server.Listen(addr); err != nil {
		t.Error(err)
	}
}
