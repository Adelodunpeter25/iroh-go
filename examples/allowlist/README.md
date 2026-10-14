# allowlist

This example demonstrates a time server where the listener endpoint serving the
time can strictly limit who can connect to it by using an explicit allowlist
(listing which endpoints are allowed to connect, and rejecting all others).

It also demonstrates how to derive a static identity, simplifying the setup of
testing an allowlist.

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./allowlist.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./allowlist.go
```

## Terminal 1 (listener)

Start up a time server with a static identity, determined by `-secret`, and
only allow a particular keypair to connect to it, defined by `-allow`.

```
./allowlist \
  -secret <string> \
  -allow <remote endpoint id>
```

For example, passing `-secret hihi` derives a keypair for the listener which
gives us the following information we can pass to the receiver in terminal 2:

```
-endpoint 608964f7ed721580d2a0a7674da37e79970c4ab60768f27501bc937d9f62c9be 
```

To get a concrete value for the allowlist on the receiver side, we derive a
different identity by passing passing `-secret cool`:

```
-allow 2d6577f847d25ddefca000b026fbf95101dc153950ad70dbdc20f15a80d02e07
```

Combined, we get the following that we can run:

```
./allowlist \
  -secret hihi \
  -allow 2d6577f847d25ddefca000b026fbf95101dc153950ad70dbdc20f15a80d02e07
```

## Terminal 2 (receiver)

Attempt to connect to the listener defined by `-endpoint` and derive an
identity based on `-secret`.

```
./allowlist \
  -endpoint 608964f7ed721580d2a0a7674da37e79970c4ab60768f27501bc937d9f62c9be \
  -secret cool
```

Running the above will cause the time server to work as expected, because we
derive the correct identity allowlisted by the listener. If you want to verify
that another id will not work, change `-secret` or omit it completely:

```
./allowlist \
  -endpoint 608964f7ed721580d2a0a7674da37e79970c4ab60768f27501bc937d9f62c9be \
  -secret not-cool
```

**NOTE**: depending on your circumstances and testing context, you might have
to use the `-ticket` information output by the listener to make a connection
(includes more information on how to make a connection). When connecting using
a ticket, things look like this instead:

```
./allowlist \
  -secret not-cool \
  -ticket endpointxxx..
```

Replace `xxx..` with the actual encoded ticket information, which among other
things includes your observable IP addresses.
