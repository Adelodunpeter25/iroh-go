package main

import (
	"flag"
	"fmt"
	"time"
	"strings"
	"os"
	"os/signal"
	"syscall"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn                = []byte("iroh-go-timeserve/0")
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

func receiver(e *iroh.Endpoint, id string) {
	remote, err := iroh.EndpointIdFromString(id)
	check(err)

	addr := iroh.NewEndpointAddr(remote, nil, nil)
	conn, err := e.Connect(addr, alpn)
	check(err)

	// from docs: 
	// "The peer that calls open_bi must write to its SendStream before the peer Connection is able to accept the stream
	// using accept_bi()"
	// https://docs.rs/iroh/latest/iroh/endpoint/struct.Connection.html#method.accept_bi
	stream, err := conn.OpenBi()
	check(err)
	// NOTE: for the other peer to realize the bidi stream is open, you have to send some data! the data sent cannot be
	// the empty string ""; empty string == no data
	err = stream.Send().WriteAll([]byte{1})

	check(err)
	handleReading(stream.Recv())
}

func handleReading(recv *iroh.RecvStream) {
	for {
		frame, err := recv.Read(frameSize)
		if checkForTimeout(err) {
			break
		}
		fmt.Println(string(frame))
		check(err)
	}
}

func checkForTimeout(err error) bool {
	// handle timeout
	if err != nil {
		irohErr := err.(*iroh.IrohError)
		errMsg := irohErr.Message()
		if strings.HasPrefix(errMsg, "ConnectionLost") {
			fmt.Println("lost connection")
			return true
		} else {
			// panics and exits bc unhandled err
			check(err)
		}
	}
	return false
}


func listen(e *iroh.Endpoint) {
	for {
		incoming := *e.AcceptNext()
		fmt.Println("new connection")
		go handleIncoming(incoming)
	}
}

func handleIncoming (incoming *iroh.Incoming) {
	accepting, err := incoming.Accept()
	check(err)

	conn, err := accepting.Connect()
	check(err)

	stream, err := conn.AcceptBi()
	check(err)

	for {
		err := stream.Send().WriteAll([]byte(time.Now().Format(time.DateTime)))
		if checkForTimeout(err) {
			break
		}
		// prints every two seconds (not for any technical reason, 1/sec just feels hectic!)
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
	check(err)
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
