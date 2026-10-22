package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gomod.cblgh.org/cerca/util/eout"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

var (
	alpn  = []byte("iroh-go-http-dialer/0")
	conns sync.Map
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

type ITCP struct {
	LocalListener net.Listener
	Endpoint      *iroh.Endpoint
	ch            chan string
}

// create an iroh endpoint
func newITCP() *ITCP {
	var t ITCP
	t.ch = make(chan string)

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}

	// NOTE (2026-10-19): static endpoint key being set
	digest := sha256.Sum256([]byte("http-dialer"))
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

func getHost() string {
	return fmt.Sprintf(":%d", PORT)
}

func (t *ITCP) Start(ctx context.Context) {
	tcphost := getHost()
	var err error
	log(fmt.Sprintf("setting up local listener on %s", tcphost))
	t.LocalListener, err = net.Listen("tcp", tcphost)
	eout.Check(err, "set up local listener")

	go t.handleIncoming(tcphost)
}

// implements net.Listener
func (t *ITCP) Accept() (net.Conn, error) {
	c, err := t.LocalListener.Accept()
	// receive endpoint id from t.handleIncoming goroutine
	endpointId := <-t.ch
	conns.Store(c, endpointId)
	return c, err
}

// implements net.Listener
func (t *ITCP) Close() error {
	// TODO (2026-10-21)
	return t.LocalListener.Close()
}

// implements net.Listener
func (t *ITCP) Addr() net.Addr {
	return t.LocalListener.Addr()
}

func (t *ITCP) handleIncoming(tcphost string) {
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

		// send endpoint public key to be received by accept handler
		go func(endpointId string) {
			t.ch <- endpointId
		}(conn.RemoteId().String())

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
	tcpServe := newITCP()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log("received req")
		body, err := io.ReadAll(req.Body)
		// access the endpoint id of the remote sending the request via req.Context set by server.ConnContext
		v := req.Context().Value("endpointId")
		// return early, endpointId isn't set
		if v == nil || v == "" {
			log("request did not have endpointId on context")
			res.Write(body)
			return
		}
		log(fmt.Sprintf("request context for key 'endpointId' %q", v))
		eout.Check(err, "read body")
		log(fmt.Sprintf("data %q\n", string(body)))
		res.Write([]byte(fmt.Sprintf("[%s] %s", v, body)))
	})

	server := http.Server{
		Addr: getHost(),
		// provide an extra key `endpointId` to all incoming http.Request's such that each handler can access the key
		// for a request as `req.Context().Value('endpointId')`
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			var s string
			// sync.map 'conns' retrieves incoming net.Conn from map and then deletes it to prevent memory leak
			v, ok := conns.LoadAndDelete(c)
			if ok {
				s = v.(string)
			} else {
				log("sync.map 'conns' could not retrieve incoming net.Conn from map")
			}
			// we set the key 'endpointId' on the context, which means we can retrieve it from any handler's request as
			// `req.Context().Value('endpointId')`
			return context.WithValue(ctx, "endpointId", s)
		},
		Handler: mux,
	}

	tcpServe.Start(context.Background())
	fmt.Println("started iroh http server")
	err := server.Serve(tcpServe)
	eout.Check(err, "http.Serve(tcpServe)")
	awaitInterrupt()
}
