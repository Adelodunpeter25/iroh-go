package main

import (
	"errors"

	iroh "git.coopcloud.tech/decentral1se/iroh-go"
)

func main() {
	preset := iroh.PresetN0DisableRelay()
	opts := iroh.EndpointOptions{
		Preset: &preset,
		Alpns:  &[][]byte{[]byte("iroh-go-errors/0")},
	}

	endpoint, err := iroh.EndpointBind(opts)

	// BEGIN API PREVIEW ZONE

	// 1. convert error to IrohError with As
	//    bring IrohError.Message/DebugMessage back
	//    maintain `err != nil` compatibility
	if err := iroh.As(err); err != nil {
		panic(err)
	}

	// 1. no As machinery required
	//    sentinel check with errors.Is support
	//    works like usual `err`
	if err != nil && errors.Is(err, iroh.BindError) {
		panic(err)
	}

	// END API PREVIEW ZONE

	endpoint.Online()
}
