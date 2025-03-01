# Streaming Example

This example shows how streaming with ik works. The client will stream from `stdin` to the server through to `stdout`.

**Getting Started:**

```sh
git clone https://github.com/devhsoj/ik.git
cd ik/
```

**Running Server:**
```sh
# from the ik/ directory
go run examples/streaming/server/main.go
```

**Running Client:**
```sh
# from the ik/ directory
cat go.mod | go run examples/streaming/client/main.go

# or for manual input
go run examples/streaming/client/main.go
```