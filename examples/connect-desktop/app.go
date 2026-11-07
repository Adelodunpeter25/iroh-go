package main

import (
	"context"
	"errors"
	"fmt"

	iroh "github.com/Adelodunpeter25/iroh-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	alpn                = []byte("iroh-go-hello-world/0")
	frameSize    uint32 = 256
	endpointPeer *string
)

// App struct
type App struct {
	ctx      context.Context
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
	if err := iroh.As(err); err != nil {
		panic(err)
	}
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

func handleIncoming(ctx context.Context, incoming *iroh.Incoming) {
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
	for {
		frame, err := recv.Read(frameSize)
		if err != nil {
			if errors.Is(err, iroh.TimeoutError) {
				fmt.Println("lost connection")
				break
			}
			if err := iroh.As(err); err != nil {
				panic(err)
			}
		}
		runtime.EventsEmit(ctx, "data", map[string]interface{}{
			"data": string(frame),
		})
	}
}
