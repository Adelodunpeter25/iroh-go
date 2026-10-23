package main

import (
	"errors"
	"bytes"
	"strings"
	"time"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"gomod.cblgh.org/cerca/util/eout"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn  = []byte("iroh-go-http-dialer/0")
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

func (sp *SimplePipe) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	retbuf, err := sp.recv.Read(uint32(len(buf)))

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

func (sp *SimplePipe) Write(buf []byte) (int, error) {
	n64, err := sp.send.Write(buf)
	n := int(n64)

	if n < len(buf) {
		return n, errors.New("SimplePipe.Write exited before writing all bytes")
	}

	return n, err
}

var verbose bool
var httpServerAddress string
var additionalInput string

func log(s string) {
	if verbose {
		fmt.Println(s)
	}
}

// IrohConn implements net.Conn (Read,Write,Close + Local/RemoteAddr + Deadline methods)
type IrohConn struct {
	sp *SimplePipe
	addrLocal *IrohAddr
	addrRemote *IrohAddr
}

type IrohAddr struct {
	internalAddr string
}
func (addr *IrohAddr) String() string {
	return addr.internalAddr
}
func (addr *IrohAddr) Network() string {
	return "tcp"
}

func newIrohConn(remoteEndpoint, localEndpoint string) *IrohConn {
	ic := IrohConn{
		sp: &SimplePipe{},
		addrLocal: &IrohAddr{localEndpoint},
		addrRemote: &IrohAddr{remoteEndpoint},
	}
	return &ic
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (ic *IrohConn) Read(b []byte) (n int, err error) {
	return ic.sp.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (ic *IrohConn) Write(b []byte) (n int, err error) {
	return ic.sp.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
// Close may or may not block until any buffered data is sent;
// for TCP connections see [*TCPConn.SetLinger].
func (ic *IrohConn) Close() error {
	return fmt.Errorf("TODO Close()")
}

// LocalAddr returns the local network address, if known.
func (ic *IrohConn) LocalAddr() net.Addr {
	return ic.addrLocal
}

// RemoteAddr returns the remote network address, if known.
func (ic *IrohConn) RemoteAddr() net.Addr {
	return ic.addrRemote
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (ic *IrohConn) SetDeadline(t time.Time) error {
	return fmt.Errorf("TODO SetDeadline")
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (ic *IrohConn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("TODO SetReadDeadline")
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (ic *IrohConn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("TODO SetWriteDeadline")
}

// Receives addressing information and sets up an iroh connection to the endpoint described by the addressing
// information, and returns the receive/send streams in a format that golang's net.http ecosystem will accept.
func DialIroh(network, address string) (net.Conn, error) {
	log(fmt.Sprintf("DialIroh: network %q address %q\n", network, address))
	remoteEndpointString := strings.Split(address, ":")[0]

	var localEndpointString string

	var remoteAddr *iroh.EndpointAddr
	// first try to parse as ticket
	if ticket, err := iroh.EndpointTicketFromString(remoteEndpointString); err == nil {
		remoteAddr = ticket.EndpointAddr()
	} else {
		// fallback to interpret as endpoint id
		remote, err := iroh.EndpointIdFromString(remoteEndpointString)
		if err := iroh.As(err); err != nil {
			panic(err)
		}
		remoteAddr = iroh.NewEndpointAddr(remote, nil, nil)
	}

	log(fmt.Sprintf("DialIroh: endpoint id %q ", remoteAddr.Id().String()))

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}

	e, err := iroh.EndpointBind(opts)
	if err := iroh.As(err); err != nil {
		panic(err)
	}
	localEndpointString = e.Id().String()

	// NOTE (2026-10-22): pretty sure program execution stops here if you can't 
	// get a connection to a relay (e.g. if you are rate-limited for too many tries)
	e.Online()

	log(fmt.Sprintf("online as %s\n", localEndpointString))

	conn, err := e.Connect(remoteAddr, alpn)
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	stream, err := conn.OpenBi()
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	irohConn := newIrohConn(remoteEndpointString, localEndpointString)

	irohConn.sp.setRecv(stream.Recv())
	irohConn.sp.setSend(stream.Send())

	return irohConn, nil
}

func parseFlags() {
	flag.BoolVar(&verbose, "v", false, "output more log messages")
	flag.StringVar(&httpServerAddress, "address", "", "iroh endpoint serving an http server")
	flag.Parse()
	additionalInput = strings.Join(flag.Args(), " ")
}

func main() {
	parseFlags()
	if httpServerAddress == "" {
		fmt.Println("must provide --address")
		os.Exit(1)
	}
	// setup our dialer
	transport := http.Transport{
		Dial: DialIroh,
	}
	client := &http.Client{
		Transport: &transport,
	}
	createUrl := func (endpoint, urlpath string) string { 
		return fmt.Sprintf("http://%s%s", endpoint, urlpath) 
	}
	// try to dial and send a POST request
	resp, err := client.Post(createUrl(httpServerAddress, "/"), "text/plain", bytes.NewBufferString(additionalInput))
	eout.Check(err, "client.Post")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	eout.Check(err, "read body")
	fmt.Println(string(body))
}
