# http-dialer

This example demonstrates a Go stdlib implementation of a HTTP client which
dials a HTTP server over Iroh.

It builds on the `echo-endpoint` example and does away with any external
program dependencies; everything is implemented in standard Go.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./client/http-client.go
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./http-dialer.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./http-dialer.go

CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./client/http-client.go
```

It's also possible to build a standard binary quickly with the default `makefile` target.

```
make
```

## Terminal 1 (http server)

```
./http-dialer 
// outputs, amongst other things, an endpoint ticket you need to provide in Terminal 2
```

## Terminal 2 (http client)

Send a HTTP POST request to our HTTP server in Terminal 1 over Iroh using the
program `http-client` from the `client/` sub-directory of this example:

```
./http-client --address <endpoint from Terminal 1> hello hello
```

You should see the string "hello hello" echoed back out to you, prefixed with
the public key of the Iroh endpoint establishing the connection in Terminal 1. 

If you run the program again, you will see another prefixed public key. This is
because `http-client` creates a new Iroh identity each time it runs.
