# pairchat

This is a more complicated version of [`../connect`](../connect) which allows
the listener to also write back to the sender. The result is a basic pairwise
(only 2 peers can connect at once) chat program.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./pairchat.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./pairchat.go
```

## Terminal 1 (listener)

```
./pairchat
```

## Terminal 2 (sender)

```
./pairchat -endpoint <endpoint-id>
```

You'll see a connection notice if the two endpoints manage to connect to each other. You can
type messages on both sides of the connection, and they will appear on the other end. So
Terminal 1 will see what Terminal 2 types, and Terminal 2 will see what Terminal 1 types.
