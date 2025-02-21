package ik_test

import (
	"fmt"
	"github.com/devhsoj/ik"
	"testing"
	"time"
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

func TestServer_RegisterSubscription(t *testing.T) {
	server := ik.NewServer()

	server.Register("subscribe", func(c *ik.ServerClient, data []byte) []byte {
		for x := range 10 {
			if err := c.Send("message", []byte(fmt.Sprintf("%d", x))); err != nil {
				t.Error(err)
			}

			time.Sleep(time.Second)
		}

		return nil
	})

	if err := server.Listen(addr); err != nil {
		t.Error(err)
	}
}
