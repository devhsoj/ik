package ik_test

import (
	"bytes"
	"fmt"
	"github.com/devhsoj/ik"
	"testing"
	"time"
)

var event = "echo"
var data = []byte("Hello, world!")

func TestClient_Connect(t *testing.T) {
	client := ik.NewClient(addr)

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Send(t *testing.T) {
	client := ik.NewClient(addr)

	res, err := client.Send(event, data)

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res, data) {
		t.Fatal("data mismatch")
	}
}

func TestClient_Subscribe(t *testing.T) {
	client := ik.NewClient(addr)

	if err := client.Subscribe("subscribe", nil, func(data []byte) {
		fmt.Println(string(data))
	}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 10)

	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkClient_Send(b *testing.B) {
	client := ik.NewClient(addr)

	if err := client.Connect(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := client.Send(event, data); err != nil {
			b.Fatal(err)
		}
	}
}
