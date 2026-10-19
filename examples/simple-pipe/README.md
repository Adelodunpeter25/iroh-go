# simple-pipe

How to set up a simple directional byte stream that reads (receiver) / writes
(sender) over an iroh connection.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./simple-pipe.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./simple-pipe.go
```

Options: 

* `-secret <string>` uses the provided string to derive a secret keypair (will
  be the same each run).
* `-v` print program logs, sharing program state. Use this as a receiver to get
  out the endpoint id.
* `-endpoint <endpoint-id>` used as a sender to connect to the receiver

## Terminal 1 (receiver)

```
./simple-pipe -v -secret receiver

// outputs:
 simple-pipe -endpoint 15a6e990467a876abe7c91075f05e4b0452dd56e7167bfc269615f772da359bf
```

## Terminal 2 (sender)

```
// pipe the source code of this program over iroh to the receiver
./simple-pipe -endpoint <endpoint-id> < ./simple-pipe.go

// for example:
simple-pipe -endpoint 15a6e990467a876abe7c91075f05e4b0452dd56e7167bfc269615f772da359bf < ./simple-pipe.go
```

You should see the source code for the program appear in terminal 1.
