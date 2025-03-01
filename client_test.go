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

var opts = ik.ClientOptions{
	Addr:             addr,
	StreamBufferSize: 1_024,
}

func TestClient_Connect(t *testing.T) {
	client := ik.NewClient(opts)

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Send(t *testing.T) {
	client := ik.NewClient(opts)

	res, err := client.Send(event, data)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(res))
}

func TestClient_Subscribe(t *testing.T) {
	client := ik.NewClient(opts)

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

func TestClient_Stream(t *testing.T) {
	client := ik.NewClient(opts)

	r := bytes.NewBuffer(bytes.Repeat([]byte("Hello, world!\n"), 1_024))

	if err := client.Stream("stream", r); err != nil {
		t.Fatal(err)
	}

	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkClient_Send(b *testing.B) {
	client := ik.NewClient(opts)

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
