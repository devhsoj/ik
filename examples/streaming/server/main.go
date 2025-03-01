package main

import (
	"fmt"
	"github.com/devhsoj/ik"
	"os"
)

func main() {
	server := ik.NewServer()

	server.Register("stdout-stream", func(c *ik.ServerClient, buf []byte) []byte {
		n, err := os.Stdout.Write(buf)

		if err != nil {
			return []byte(fmt.Sprintf("failed to write to stdout: %s\n", err))
		}

		return []byte(fmt.Sprintf("WROTE %d BYTES", n))
	})

	if err := server.Listen(":3000"); err != nil {
		fmt.Printf("failed to listen: %s\n", err)
		os.Exit(1)
	}
}
