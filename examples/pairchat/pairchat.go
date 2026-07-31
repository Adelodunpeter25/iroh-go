package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn                = []byte("iroh-go-pairchat/0")
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

func send(e *iroh.Endpoint, id string) {
	remote, err := iroh.EndpointIdFromString(id)
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	addr := iroh.NewEndpointAddr(remote, nil, nil)
	conn, err := e.Connect(addr, alpn)
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	stream, err := conn.OpenBi()
	if err := iroh.As(err); err != nil {
		panic(err)
	}
	fmt.Println("connection established")
	recv := stream.Recv()
	done := make(chan struct{})
	go handleInput(stream, done)
	go handleReading(recv, done)
}

func listen(e *iroh.Endpoint) {
	for {
		incoming := *e.AcceptNext()
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

	recv := stream.Recv()
	done := make(chan struct{})
	go handleInput(stream, done)
	go handleReading(recv, done)
}

func handleInput(stream *iroh.BiStream, done chan struct{}) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			err := stream.Send().WriteAll([]byte(line))
			if err := iroh.As(err); err != nil {
				panic(err)
			}
		}

		select {
		// exit if channel done is closed
		case <-done:
			return
		default:
			// do nothing
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func handleReading(recv *iroh.RecvStream, done chan struct{}) {
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
	close(done)
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
		fmt.Printf("pairchat -endpoint %s\n", endpoint.Id())
		go func() { listen(endpoint) }()
	} else {
		go func() { send(endpoint, *endpointPeer) }()
	}

	awaitInterrupt()
}
