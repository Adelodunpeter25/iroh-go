package main

import (
	"context"
	"fmt"
	"strings"
	iroh "git.coopcloud.tech/decentral1se/iroh-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"

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

// App struct
type App struct {
	ctx context.Context
	endpoint *iroh.Endpoint
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	preset := iroh.PresetN0()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{alpn},
	}

	endpoint, err := iroh.EndpointBind(opts)
	check(err)
	endpoint.Online()
	a.endpoint = endpoint

	listen(a.ctx, a.endpoint)
}

func (a *App) GetEndpointIDString() string {
	return a.endpoint.Id().String()
}

func listen(ctx context.Context, e *iroh.Endpoint) {
	for {
		incoming := *e.AcceptNext()
		go handleIncoming(ctx, incoming)
	}
}

func handleIncoming (ctx context.Context, incoming *iroh.Incoming) {
	accepting, err := incoming.Accept()
	check(err)

	conn, err := accepting.Connect()
	check(err)

	stream, err := conn.AcceptBi()
	check(err)

	recv := stream.Recv()
	for {
		frame, err := recv.Read(frameSize)
		// handle timeout
		if err != nil {
			irohErr := err.(*iroh.IrohError)
			errMsg := irohErr.Message()
			if strings.HasPrefix(errMsg, "ConnectionLost") {
				fmt.Println("lost connection")
				break
			} else {
				check(err)
			}
		}
		runtime.EventsEmit(ctx, "data", map[string]interface{}{
			"data": string(frame),
		})
	}
}
