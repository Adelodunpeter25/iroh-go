# timeserve

This example demonstrates that an endpoint can support multiple live connections at once. The
endpoint aka listener writes the time, once every 2 seconds, to every connected peer.

Try connecting to a running listener with 3 (or more!) receiver peers at once.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./timeserve.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./timeserve.go
```

## Terminal 1 (listener)

```
./timeserve
```

## Terminal 2 (receiver)

```
./timeserve -endpoint <endpoint-id>
```

## Terminal 3 (receiver)

```
./timeserve -endpoint <endpoint-id>
```

You'll see a connection notice if a receiver peer manages to connect to the listener.
Receivers (Terminal 2 and Terminal 3) will display the time sent by the listener (Terminal 1).
