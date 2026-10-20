package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gomod.cblgh.org/cerca/util/eout"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn = []byte("iroh-go-echosrv/0")
)

type SimplePipe struct {
	recv           *iroh.RecvStream
	send           *iroh.SendStream
	totalBytesRead int
}

func (sp *SimplePipe) setRecv(recv *iroh.RecvStream) {
	sp.recv = recv
}

func (sp *SimplePipe) setSend(send *iroh.SendStream) {
	sp.send = send
}

// Read reads up to len(buf) bytes into buf. It returns the number of bytes read (0 <=
// n <= len(buf)) and any error encountered.
//
// Even if Read returns n < len(buf), it may use all of buf as scratch space during the
// call. If some data is available but not len(buf) bytes, Read conventionally returns
// what is available instead of waiting for more.
//
// When Read encounters an error or end-of-file condition after successfully
// reading n > 0 bytes, it returns the number of bytes read. It may return the
// (non-nil) error from the same call or return the error (and n == 0) from a
// subsequent call . An instance of this general case is that a Reader returning a
// non-zero number of bytes at the end of the input stream may return either
// err == EOF or err == nil.  The next Read should return 0, EOF.
//
// Callers should always process the n > 0 bytes returned before considering the
// error err. Doing so correctly handles I/O errors that happen after reading some
// bytes and also both of the allowed EOF behaviors.
//
// If len(buf) == 0, Read should always return n == 0. It may return a non-nil error
// if some error condition is known, such as EOF.
//
// Implementations of Read are discouraged from returning a zero byte count with a
// nil error, except when len(buf) == 0. Callers should treat a return of 0 and nil a
// s indicating that nothing happened; in particular it does not indicate EOF.
//
// Implementations must not retain buf.
func (sp *SimplePipe) Read(buf []byte) (int, error) {
	// fmt.Println("READ size of buf", len(buf))
	if len(buf) == 0 {
		return 0, nil
	}

	retbuf, err := sp.recv.Read(uint32(len(buf)))

	// fmt.Println("retbuf", retbuf, "len(retbuf)", len(retbuf), "err", err)

	// signal end of file
	if sp.totalBytesRead > 0 && len(retbuf) == 0 {
		return 0, io.EOF
	}

	sp.totalBytesRead += len(retbuf)

	copyBytes := func() {
		for i, d := range retbuf {
			buf[i] = d
		}
	}

	if err != nil {
		copyBytes()
		return len(retbuf), err
	}
	copyBytes()
	return len(retbuf), err
}

// Write writes len(buf) bytes from buf to the underlying data stream. It
// returns the number of bytes written from buf (0 <= n <= len(buf)) and any
// error encountered that caused the write to stop early. Write must return a
// non-nil error if it returns n < len(buf). Write must not modify the slice
// data, even temporarily.
//
// Implementations must not retain buf.
func (sp *SimplePipe) Write(buf []byte) (int, error) {
	// fmt.Println("WRITE size of buf", len(buf))
	n64, err := sp.send.Write(buf)
	n := int(n64)

	if n < len(buf) {
		return n, errors.New("SimplePipe.Write exited before writing all bytes")
	}

	return n, err
}

var done = make(chan os.Signal, 1)

func awaitInterrupt() {
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	os.Exit(0)
}

var PORT int
var verbose bool
func log(s string) {
	if verbose {
		fmt.Println(s)
	}
}

type EchoSrv struct {
	LocalListener net.Listener
	Endpoint      *iroh.Endpoint
}


// create an iroh endpoint
func newEchoSrv() *EchoSrv {
	var t EchoSrv

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}

	// NOTE (2026-10-19): static endpoint key being set
	digest := sha256.Sum256([]byte("echosrv"))
	secret := digest[:]
	opts.SecretKey = &secret

	endpoint, err := iroh.EndpointBind(opts)
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	endpoint.Online()
	fmt.Printf("online as %s\n", endpoint.Id())

	ticket, err := iroh.EndpointTicketFromAddr(endpoint.Addr())
	eout.Check(err, "format ticket")
	fmt.Println("\nticket:")
	fmt.Println(ticket, "\n")

	t.Endpoint = endpoint
	return &t
}

func getHost () string {
	return fmt.Sprintf(":%d", PORT)
}

func (t *EchoSrv) Start(ctx context.Context) {
	tcphost := getHost()
	var err error
	log(fmt.Sprintf("setting up local listener on %s", tcphost))
	t.LocalListener, err = net.Listen("tcp", tcphost)
	eout.Check(err, "set up local listener")

	go t.handleIncoming(tcphost)
}

func (t *EchoSrv) handleIncoming(tcphost string) {
	for {
		incoming := t.Endpoint.AcceptNext()
		log("incoming")
		accepting, err := (*incoming).Accept()
		if err := iroh.As(err); err != nil {
			eout.Check(err, "incoming.Accept")
		}
		log("accepted")
		conn, err := accepting.Connect()
		if err := iroh.As(err); err != nil {
			eout.Check(err, "incoming.Connect")
		}
		log("connected")
		stream, err := conn.AcceptBi()
		if err := iroh.As(err); err != nil {
			eout.Check(err, "incoming.AcceptBi")
			panic(err)
		}
		log("bidi stream")

		recv := stream.Recv()
		send := stream.Send()

		var sp SimplePipe
		sp.setRecv(recv)
		sp.setSend(send)

		log("pre dial")
		dialedConn, err := net.Dial("tcp", tcphost)
		eout.Check(err, "net.dial")
		go io.Copy(dialedConn, &sp)
		go io.Copy(&sp, dialedConn)
		log("post dial")
	}
}

func parseFlags() {
	flag.BoolVar(&verbose, "v", false, "output more log messages")
	flag.IntVar(&PORT, "p", 9899, "port for net.Listener to run on")
	flag.Parse()
}

func main() {
	parseFlags()
	tcpServe := newEchoSrv()
	// NOTE (2026-10-20): a critical improvement (probably for a solution that actually implements net.Conn, net.Dial etc) is for
	// there to be a way to access the endpoint id of the remote sending the request
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log("received req")
		body, err := io.ReadAll(req.Body)
		eout.Check(err, "read body")
		log(fmt.Sprintf("data %q\n", string(body)))
		res.Write(body)
	})
	tcpServe.Start(context.Background())
	fmt.Println("started iroh http server")
	err := http.Serve(tcpServe.LocalListener, nil)
	eout.Check(err, "http.Serve(tcpServe)")
	awaitInterrupt()
}
