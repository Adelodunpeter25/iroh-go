# echo-endpoint

Demonstrates how to set up a simple HTTP echo server that is served only* over iroh.

This example is more complex than `echosrv`. It ties the incoming endpoint ID to the HTTP
request that is handled by the http server. In effect, the underlying http server can gate
resources and responses depending on which iroh keypair is connecting and sending requests.

To run this example, you will need to use [cURL](https://curl.se/) (a multitool of a program for anything HTTP) and also download Number0's program [`dumbpipe`](https://www.dumbpipe.dev/) (Iroh forwarding tool). For dumbpipe binaries as of 2026-10, see [n0-computer/dumbpipe/releases](https://github.com/n0-computer/dumbpipe/releases).

We will use dumbpipe together with cURL to achieve sending http client traffic over Iroh to our Iroh-available HTTP server.

\* *technically, the http server is also available locally on the opened port*

## Build a binary

```
go build --ldflags '-linkmode external -extldflags=-static -s -w' ./echo-endpoint.go
```

If you run into `glibc` incompatibility, you can compile with `zig`/`musl`.

```
CC="zig cc -target x86_64-linux-musl -lunwind" \
CGO_ENABLED=1 \
CGO_LDFLAGS="-static" \
GOOS=linux GOARCH=amd64 \
go build -v -a -ldflags '-extldflags "-static"' ./echo-endpoint.go
```

## Terminal 1 (http server)

```
./echo-endpoint 
// outputs, among other things, an endpoint ticket you need to provide in terminal 2
```

## Terminal 2 (iroh client forwarding)

You can install `dumbpipe` [here](https://www.dumbpipe.dev).

```
./dumbpipe connect-tcp \
  --addr 127.0.0.1:8099 \
  --custom-alpn "utf8:iroh-go-echo-endpoint/0" \
  endpoint<YOURENDPOINT-HERE>
```

A few notes:

* `--addr` - the ip and port has to be same as what you will use in Terminal 3.
  It is completely unrelated to the ip and port use by the http server in
  terminal 1 (we're connecting to that over Iroh!!).
* Note the use of `--custom-alpn`. Specifically, the prefix "utf8:" allows us
  to use an utf8 string, otherwise the alpn would be interpreted differnetly
  and not work.
* `endpoint<YOURENDPOINT-HERE>` should be the ticket output from Terminal 1.

## Terminal 3 (http client)

Finally, we send a HTTP POST request to our http server through the `dumbpipe`
proxied iroh connection in Terminal 2. Take care to use the same ip and port as
used in Terminal 2.

```
curl 127.0.0.1:8099/  --data "hello"
```
You should see the string "hello" echoed back out to you, prefixed with the public key of the
iroh endpoint establishing the connection in Terminal 2. 

If you restart the process running in Terminal 2, you should see a different key being
returned as the prefix in the echo server's response.
