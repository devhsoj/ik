package ik_test

import (
    "bytes"
    "github.com/devhsoj/ik"
    "testing"
)

func TestClient_Connect(t *testing.T) {
    client := ik.NewClient(addr)

    if err := client.Connect(); err != nil {
        t.Fatal(err)
    }
}

func TestClient_Send(t *testing.T) {
    client := ik.NewClient(addr)

    if err := client.Connect(); err != nil {
        t.Fatal(err)
    }

    if err := client.Send("ping", []byte("pong")); err != nil {
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

    var data = bytes.Repeat([]byte{'x'}, 1_024)

    for b.Loop() {
        if err := client.Send("huge", data); err != nil {
            b.Fatal(err)
        }
    }
}
