# ik

ik (short for intānettokomando) is a simple, high performance, event-oriented TCP protocol & library.

## Features

* Zero Dependencies!
* Request / Response
* Subscriptions
* Streaming

## Getting Started

```shell
go get github.com/devhsoj/ik
```

**Server Example:**
```go
package main

import "github.com/devhsoj/ik"

func main() {
    server := ik.NewServer()
    
    server.Register("echo", func(c *ik.ServerClient, data []byte) []byte {
        return data
    })

    if err := server.Listen(":3000"); err != nil {
        panic(err)
    }
}
```

**Client Example:**

```go
package main

import (
    "fmt"
    "github.com/devhsoj/ik"
)

func main() {
    client := ik.NewClient(ik.ClientOptions{
        Addr: "localhost:3000",
    })

    res, err := client.Send("echo", []byte("Hello, world!"))

    if err != nil {
        panic(err)
    }

    fmt.Println(string(res)) // Hello, world!
}
```