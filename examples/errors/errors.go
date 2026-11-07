package main

import (
	"errors"

	iroh "github.com/Adelodunpeter25/iroh-go"
)

func main() {
	preset := iroh.PresetN0DisableRelay()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{[]byte("iroh-go-errors/0")},
	}

	endpoint, err := iroh.EndpointBind(opts)

	// NOTE(d1): The main route: convert errors to IrohError with `iroh.As`. This
	// brings the `IrohError.Message`/`DebugMessage` back in a regular error
	// value. This ensures we maintain `err != nil` compatibility
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	// NOTE(d1): If you need more fine grained error catching: no `iroh.As`
	// machinery is required. Sentinel checks with the usual `errors.Is` API are
	// supported. It works as expected with `err` values
	if err != nil && errors.Is(err, iroh.BindError) {
		panic(err)
	}

	endpoint.Online()
}
