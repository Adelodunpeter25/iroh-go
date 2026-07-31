package main

import (
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
	alpn                = []byte("iroh-go-timeserve/0")
	frameSize    uint32 = 256
	endpointPeer *string
)

func awaitInterrupt() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	os.Exit(0)
}

func parseFlags() {
	endpointPeer = flag.String("endpoint", "", "endpoint id of peer")
	flag.Parse()
}

func receiver(e *iroh.Endpoint, id string) {
	remote, err := iroh.EndpointIdFromString(id)
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	addr := iroh.NewEndpointAddr(remote, nil, nil)
	conn, err := e.Connect(addr, alpn)
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
	for {
		frame, err := recv.Read(frameSize)
		if errors.Is(err, iroh.TimeoutError) {
			fmt.Println("lost connection")
			break
		}
		if err := iroh.As(err); err != nil {
			panic(err)
		}
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

func handleIncoming(incoming *iroh.Incoming) {
	accepting, err := incoming.Accept()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	conn, err := accepting.Connect()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	stream, err := conn.AcceptBi()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	for {
		err := stream.Send().WriteAll([]byte(time.Now().Format(time.DateTime)))
		if errors.Is(err, iroh.TimeoutError) {
			fmt.Println("lost connection")
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

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}

	endpoint, err := iroh.EndpointBind(opts)
	if err := iroh.As(err); err != nil {
		panic(err)
	}
	endpoint.Online()

	if *endpointPeer == "" {
		fmt.Println("this endpoint is online and listening! use this id to connect via a peer:")
		fmt.Printf("timeserve\n  -endpoint %s\n", endpoint.Id())
		go func() { listen(endpoint) }()
	} else {
		go func() { receiver(endpoint, *endpointPeer) }()
	}

	awaitInterrupt()
}
