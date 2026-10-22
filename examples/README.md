# Examples

The examples in this directory show different parts of making an application
that leverages iroh's functionality.

* [`timeserve`](./timeserve) shows how one listener can serve many clients at
  once.
* [`allowlist`](./allowlist) shows how to set up an allowlist, limiting
  connections to only those explicitly approved. The example also shows how you
  can derive an identity that will stay the same between runs (static iroh
  public keypair) and how to interact with iroh's [tickets][tickets] mechanism
  of exchanging address information.
* [`connect`](./connect) is a barebones example showing how to send and read
  data.
* [`connect-desktop`](./connect-desktop) shows how to build a go-based desktop
  application using iroh.
* [`pairchat`](./pairchat) show how to setup a pairwise connection where both
  parties can read and receive data sent over an iroh connection.
* [`simple-pipe`](./simple-pipe) sets up a directional pipe that can be used à
  la netcat to send files over iroh.
* [`echosrv`](./echosrv) shows how to set up a simple HTTP echo server
  that is served only over iroh.
* [`echo-endpoint`](./echo-endpoint) sets up a HTTP echo server that can gate its HTTP
  routes based on the Iroh public key making requests.
* [`http-dialer`](./http-dialer) shows how to send a POST request over iroh using the `net/http`
  standard library approach



[tickets]: https://docs.iroh.computer/connect-two-endpoints#what-is-a-ticket
