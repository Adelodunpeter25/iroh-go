# http-dialer

Demonstrates an http client in go that dials an http server, also in go, over Iroh using
from the standard library `net/http` (e.g. `http.Client`)

This example builds on `echo-endpoint` and does away with any external program dependencies;
everything is done in go.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./http-dialer.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./http-dialer.go
```

## Terminal 1 (http server)

```
go build ./http-dialer.go
./http-dialer 
// outputs, among other things, an endpoint ticket you need to provide in terminal 2
```

## Terminal 2 (http client)

We send a HTTP POST request to our http server over iroh by using the program `http-client`
from `client/` subdirectory of this example:

```
go build ./client/http-client.go
./http-client --address <endpoint from terminal1> hello hello
```

You should see the string "hello hello" echoed back out to you, prefixed with the
public key of the iroh endpoint establishing the connection in Terminal 1. 

If you run the program again, you will see another prefixed public key. This is because
`http-client` creates a new Iroh identity each time it runs.
