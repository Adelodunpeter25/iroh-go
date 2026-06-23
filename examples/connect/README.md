# connect

A simple connection test where one peer can send and the other can receive. See
[`../pairchat`](../pairchat) for a more complete bidirectional example.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./connect.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./connect.go
```

## Terminal 1 (listener)

```
./connect
```

## Terminal 2 (sender)

```
./connect -endpoint <endpoint-id>
```

You'll see a connection notice if the two endpoints manage to connect to each
other. You can type messages in the sender terminal and they will be shown on
the listener side.
