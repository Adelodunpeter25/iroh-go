package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn                = []byte("iroh-go-hello-world/0")
	frameSize    uint32 = 256
	endpointPeer *string
)

func check(err error) {
	if err != nil {
		irohErr := err.(*iroh.IrohError)
		panic(irohErr.Message())
	}
}

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
	check(err)

	addr := iroh.NewEndpointAddr(remote, nil, nil)
	conn, err := e.Connect(addr, alpn)
	check(err)

	stream, err := conn.OpenBi()
	check(err)

	fmt.Println("connected")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			err = stream.Send().WriteAll([]byte(line))
			check(err)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func listen(e *iroh.Endpoint) {
	incoming := *e.AcceptNext()
	accepting, err := incoming.Accept()
	check(err)

	conn, err := accepting.Connect()
	check(err)

	stream, err := conn.AcceptBi()
	check(err)

	recv := stream.Recv()
	for {
		frame, err := recv.Read(frameSize)
		check(err)
		fmt.Println(string(frame))
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
	check(err)
	endpoint.Online()

	if *endpointPeer == "" {
		fmt.Println("this endpoint is online and listening! use this id to connect via a peer:")
		fmt.Printf("connect -endpoint %s\n", endpoint.Id())
		go func() { listen(endpoint) }()
	} else {
		go func() { send(endpoint, *endpointPeer) }()
	}

	awaitInterrupt()
}
