package main

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn                = []byte("iroh-go-allowlist/0")
	frameSize    uint32 = 256
	endpointPeer *string
	allowedPeer  *string
	peerSecret   *string
	ticketArg    *string
)

func awaitInterrupt() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	os.Exit(0)
}

func parseFlags() {
	endpointPeer = flag.String("endpoint", "", "endpoint id of peer")
	allowedPeer = flag.String("allow", "", "endpoint id of allowlisted peer")
	peerSecret = flag.String("secret", "", "string that will be hashed to construct secret from which to derive iroh keypair")
	ticketArg = flag.String("ticket", "", "endpoint ticket address")
	flag.Parse()
}

func receiver(e *iroh.Endpoint, id, ticketStr string) {
	var addr *iroh.EndpointAddr
	if ticketStr != "" {
		ticket, err := iroh.EndpointTicketFromString(ticketStr)
		if err := iroh.As(err); err != nil {
			panic(err)
		}
		addr = ticket.EndpointAddr()
	} else {
		remote, err := iroh.EndpointIdFromString(id)
		if err := iroh.As(err); err != nil {
			panic(err)
		}
		addr = iroh.NewEndpointAddr(remote, nil, []string{"127.0.0.1:8992"})
	}

	conn, err := e.Connect(addr, alpn)
	if err != nil && errors.Is(err, iroh.BindError) {
		fmt.Println("experienced a bind error when dialing endpoint")
		os.Exit(1)
	}
	if err != nil && errors.Is(err, iroh.ConnectError) {
		fmt.Println("experienced a connect error when dialing endpoint")
		fmt.Println("(expected behaviour for this allowlist example!:)")
		os.Exit(1)
	}

	if err := iroh.As(err); err != nil {
		panic(err)
	}

	// from docs: "The peer that calls open_bi must write to its SendStream
	// before the peer Connection is able to accept the stream using accept_bi()"
	// https://docs.rs/iroh/latest/iroh/endpoint/struct.Connection.html#method.accept_bi
	stream, err := conn.OpenBi()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	// NOTE: for the other peer to realize the bidi stream is open, you have to
	// send some data! the data sent cannot be the empty string ""; empty string
	// == no data
	err = stream.Send().WriteAll([]byte{1})
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	handleReading(stream.Recv())
}

func handleReading(recv *iroh.RecvStream) {
	firstLoop := true
	for {
		frame, err := recv.Read(frameSize)
		if err != nil && errors.Is(err, iroh.TimeoutError) {
			fmt.Println("lost connection")
			break
		}

		// NOTE (2026-10-13): when the receiver has a direct connection to the listener (the direct connection mode, not
		// via relay), and listener closes the connection during the connecting phase,
		// the error we get back does not seem to watch any of the existing iroh-go error sentinels
		// (i tried all!)
		if err := iroh.As(err); err != nil {
			// only want to emit this friendly error message on the first loop of reading, which
			// is when we would detect that a direct connection has been closed
			if firstLoop {
				fmt.Println("experienced an error after having dialed endpoint")
				fmt.Println("(very likely expected behaviour for this allowlist example!:)")
				os.Exit(1)
			} else {
				panic(err)
			}
		}
		firstLoop = false
		fmt.Println(string(frame))
	}
}

func listen(e *iroh.Endpoint) {
	for {
		incoming := *e.AcceptNext()
		fmt.Println("new connection")
		go handleIncoming(incoming)
	}
}

func isAllowed(remoteId *iroh.EndpointId) bool {
	if *allowedPeer == "" {
		return true
	}
	if remoteId.String() != *allowedPeer {
		return false
	}
	return true
}

func handleIncoming(incoming *iroh.Incoming) {
	var raddr iroh.IncomingAddrRelay
	v, err := incoming.RemoteAddr()
	if err = iroh.As(err); err != nil {
		panic(err)
	}
	raddr, fromRelay := v.(iroh.IncomingAddrRelay)
	if fromRelay {
		// for reference, here is how the javascript bindings define the interface for value returned by incoming.RemoteAddr()
		//
		// interface IncomingAddr {
		// 	addr?: string;
		// 	description?: string;
		// 	endpointId?: string;
		// 	kind: string;
		// }
		// https://n0-computer.github.io/iroh-ffi/js/interfaces/IncomingAddr.html

		// if the connection is indirect (i.e. relayed), we can at this early stage detect the remote party's iroh
		// public key and stop any other parts of the connection process from happening.
		// at this stage, if detected not to be in the allowlist, we can either silently
		// attempt (incoming.Ignore()) or explicitly refuse it (incoming.Refuse()).
		//
		// Ignore(): the remote party will stay open until a timeout. this has the benefit that there is no information
		// leakage (the other party doesn't know that they have been rejected, or that the endpoint they are trying is
		// valid).
		//
		// NOTE: if you choose Ignore, the remote party's connection process may fire off manhy connection attempts
		// until finally shutting down to a timeout
		//
		// Refuse(): the connection attempt is immediately terminated for the remote party. the remote party knows that
		// the endpoint they tried is valid, but that they were not granted access.
		if !isAllowed(raddr.EndpointId) {
			fmt.Printf("[accepting] rejected a connection not in allowlist (%s)\n", raddr.EndpointId)
			err = incoming.Refuse()
			if err = iroh.As(err); err != nil {
				panic(err)
			}
			return
		}
	} else {
		// a quick check to see the incoming connection is of the other type we expect (i.e. direct IP connection)
		_, ok := v.(iroh.IncomingAddrIp)
		if !ok {
			panic("wasn't IncomingAddrIp either")
		}
	}

	accepting, err := incoming.Accept()
	if err = iroh.As(err); err != nil {
		panic(err)
	}

	conn, err := accepting.Connect()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	// if the connection is direct (i.e. not relayed), we can't --- vis-a-vis the current API --- know what the remote party's iroh
	// public key is until the connection has been established.
	//
	// at this stage, if detected not to be in the allowlist, we close the connection with a reason.
	if !isAllowed(conn.RemoteId()) {
		fmt.Printf("[connecting] rejected a connection not in allowlist (%s)\n", conn.RemoteId())
		err = conn.Close(1, []byte("not in allowlist"))
		if err = iroh.As(err); err != nil {
			panic(err)
		}
		return
	}

	/* peer is in the allowlist: proceed as normal */

	stream, err := conn.AcceptBi()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	for {
		err := stream.Send().WriteAll([]byte(time.Now().Format(time.DateTime)))
		if err != nil && errors.Is(err, iroh.TimeoutError) {
			fmt.Println("lost connection")
			break
		}

		if err != nil && errors.Is(err, iroh.StreamError) {
			fmt.Println("stream error")
			break
		}

		if err := iroh.As(err); err != nil {
			panic(err)
		}

		// prints every two seconds (not for any technical reason, 1/sec just feels
		// hectic!)
		<-time.After(2 * time.Second)
	}
}

func main() {
	parseFlags()
	isListener := *endpointPeer == "" && *ticketArg == ""

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}
	//  if -secret <string> is passed, setup static identity
	if *peerSecret != "" {
		fmt.Println("deriving static secret based on -secret")
		sum := sha256.Sum256([]byte(*peerSecret))
		// quick re-convert from [32]byte to []byte
		secretBytes := sum[:]
		opts.SecretKey = &secretBytes
	}

	endpoint, err := iroh.EndpointBind(opts)
	if err := iroh.As(err); err != nil {
		panic(err)
	}
	fmt.Printf("endpoint's id:\n  %s\n", endpoint.Id())
	endpoint.Online()

	if isListener {
		fmt.Println("this endpoint is online and listening!\n\nconnect using endpoint id or via ticket:\n")
		fmt.Printf("allowlist\n  -endpoint %s\n", endpoint.Id())
		ticket, err := iroh.EndpointTicketFromAddr(endpoint.Addr())

		if err := iroh.As(err); err != nil {
			panic(err)
		}

		fmt.Printf("\nallowlist\n  -ticket %s\n\n", ticket)

		if *allowedPeer == "" {
			fmt.Println("\nallowing all peers to connect")
		} else {
			fmt.Printf("\nallowlist detected: limiting connections to %s\n", *allowedPeer)
		}

		go func() { listen(endpoint) }()
	} else {
		go func() { receiver(endpoint, *endpointPeer, *ticketArg) }()
	}

	awaitInterrupt()
}
