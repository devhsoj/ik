package main

import (
	"fmt"
	"github.com/devhsoj/ik"
	"os"
)

func main() {
	client := ik.NewClient(&ik.ClientOptions{
		Addr: "localhost:3000",
	})

	if err := client.Connect(); err != nil {
		fmt.Printf("failed to connect client: %s\n", err)
		os.Exit(1)
	}

	if err := client.Stream("stdout-stream", os.Stdin); err != nil {
		fmt.Printf("failed to stream: %s\n", err)
		os.Exit(1)
	}

	if err := client.Close(); err != nil {
		fmt.Printf("failed to close client: %s\n", err)
		os.Exit(1)
	}
}
