package main

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn                = []byte("iroh-go-simple-pipe/0")
	frameSize    uint32 = 256
	endpointPeer *string
	verboseFlag  *bool
	secretFlag   *string
)

type SimplePipe struct {
	recv *iroh.RecvStream
	send *iroh.SendStream
	totalBytesRead int
}

func (sp *SimplePipe) setRecv(recv *iroh.RecvStream) {
	sp.recv = recv
}

func (sp *SimplePipe) setSend(send *iroh.SendStream) {
	sp.send = send
}

/* inlining go's docs for io.Reader and io.Writer. i'm not sure i'm entirely fulfilling the implementation
* contract right now but it also works in my smol tests :)*/

/// Read reads up to len(buf) bytes into buf. It returns the number of bytes read (0 <=
/// n <= len(buf)) and any error encountered.
///
/// Even if Read returns n < len(buf), it may use all of buf as scratch space during the
/// call. If some data is available but not len(buf) bytes, Read conventionally returns
/// what is available instead of waiting for more.
///
/// When Read encounters an error or end-of-file condition after successfully
/// reading n > 0 bytes, it returns the number of bytes read. It may return the
/// (non-nil) error from the same call or return the error (and n == 0) from a
/// subsequent call . An instance of this general case is that a Reader returning a
/// non-zero number of bytes at the end of the input stream may return either
/// err == EOF or err == nil.  The next Read should return 0, EOF.
///
/// Callers should always process the n > 0 bytes returned before considering the
/// error err. Doing so correctly handles I/O errors that happen after reading some
/// bytes and also both of the allowed EOF behaviors.
///
/// If len(buf) == 0, Read should always return n == 0. It may return a non-nil error
/// if some error condition is known, such as EOF.
///
/// Implementations of Read are discouraged from returning a zero byte count with a
/// nil error, except when len(buf) == 0. Callers should treat a return of 0 and nil a
/// s indicating that nothing happened; in particular it does not indicate EOF.
///
/// Implementations must not retain buf.
func (sp *SimplePipe) Read(buf []byte) (int, error) {
	retbuf, err := sp.recv.Read(uint32(len(buf)))
	if err != nil {
		return len(retbuf), err
	}
	sp.totalBytesRead += len(retbuf)
	// fmt.Println("retbuf", retbuf, "len(retbuf)", len(retbuf), "err", err)
	// signal end of file
	if sp.totalBytesRead >  0 && len(retbuf) == 0 {
		return 0, io.EOF
	}
	for i, d := range retbuf {
		buf[i] = d
	}
	return len(retbuf), err
}

/// Write writes len(buf) bytes from buf to the underlying data stream.
/// It returns the number of bytes written from buf (0 <= n <= len(buf)) and any error encountered that caused the write to stop early.
/// Write must return a non-nil error if it returns n < len(buf). Write must not modify the slice data, even temporarily.
///
/// Implementations must not retain buf.
func (sp *SimplePipe) Write(buf []byte) (int, error) {
	n64, err := sp.send.Write(buf)
	n := int(n64)
	return n, err
}

// use as:
// var sp SimplePipe
// sp.setRecv(stream.Recv()) or sp.setSend(stream.Send())
//
// io.Copy(*sp, os.Stdin)
// io.Copy(os.Stdout, *sp)

var done = make(chan os.Signal, 1)

func awaitInterrupt() {
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	os.Exit(0)
}

func parseFlags() {
	endpointPeer = flag.String("endpoint", "", "endpoint id of peer")
	verboseFlag = flag.Bool("v", false, "include logging from program")
	secretFlag = flag.String("secret", "", "provide a string to use to derive the endpoint's keypair")
	flag.Parse()
}

func log(s string) {
	if *verboseFlag {
		fmt.Println(s)
	}
}

func send(e *iroh.Endpoint, id string, sp *SimplePipe) {
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

	log("connection established")

	sp.setSend(stream.Send())

	// write initial data to open the connection
	sp.Write([]byte{1})

	// copy stdin directly to opened connection's send stream
	n, err := io.Copy(sp, os.Stdin)
	log(fmt.Sprintf("wrote %d", n))
	if err = iroh.As(err); err != nil {
		panic(err)
	}
	// mark send stream as done
	stream.Send().Finish()
	exit(conn)

}

func exit(conn *iroh.Connection) {
	// wait a bit to cleanly exit
	time.Sleep(1 * time.Second)
	// send an event to the done channel to stop this process via awaitInterrupt
	done <- os.Interrupt
}

func listen(e *iroh.Endpoint, sp *SimplePipe) {
	incoming := *e.AcceptNext()
	handleIncoming(incoming, sp)
}

func handleIncoming(incoming *iroh.Incoming, sp *SimplePipe) {
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

	sp.setRecv(stream.Recv())
	n, err := io.Copy(os.Stdout, sp)
	log(fmt.Sprintf("read n %d", n))
	if errors.Is(err, iroh.TimeoutError) ||
		errors.Is(err, iroh.StreamError) {
		log("lost connection")
	} else if err = iroh.As(err); err != nil {
		panic(err)
	}
	exit(conn)
}

func main() {
	var sp SimplePipe
	parseFlags()

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}

	// if --secret: use the input, and sha256 hash it to derive a keypair
	if *secretFlag != "" {
		digest := sha256.Sum256([]byte(*secretFlag))
		secret := digest[:]
		opts.SecretKey = &secret
	}

	endpoint, err := iroh.EndpointBind(opts)
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	endpoint.Online()

	if *endpointPeer == "" {
		log("this endpoint is online and listening! use this id to connect via a peer:")
		log(fmt.Sprintf("simple-pipe -endpoint %s\n", endpoint.Id()))
		go func() { listen(endpoint, &sp) }()
	} else {
		go func() { send(endpoint, *endpointPeer, &sp) }()
	}

	awaitInterrupt()
}
