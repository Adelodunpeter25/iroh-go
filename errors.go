package iroh_ffi

import (
	"errors"
	"fmt"
)

// `As` casts an error to IrohError and back, extracting relevant error message
// information. It is a single argument variant of the typical errors.As API
// because we always expect an IrohError back from the iroh FFI bindings. The
// bindings do not return an IrohError for reasons and will not for the
// forseeable future due to API stability guarantees.
func As(err error) error {
	if err, ok := errors.AsType[IrohError](err); ok {
		return fmt.Errorf("%s: %s", err.Message(), err.DebugMessage())
	}
	return err
}

// `Is` supports matching against the stable iroh error taxonomy using the
// error sentinel construction below. The caller must pass in a valid error
// sentinel specified by this library, e.g. iroh.BindError.
func (e *IrohError) Is(target error) bool {
	if err, ok := errors.AsType[sentinel](target); ok {
		return uint(e.Kind()) == err.kind
	}
	return false
}

type sentinel struct {
	kind    uint
	message string
}

func (s sentinel) Error() string {
	return s.message
}

func (s sentinel) String() string {
	return s.message
}

var InvalidInputError error = sentinel{
	kind:    uint(IrohErrorKindInvalidInput),
	message: "iroh: invalid input supplied by the caller",
}

var BindError error = sentinel{
	kind:    uint(IrohErrorKindBind),
	message: "iroh: failure while binding an endpoint",
}

var ConnectError error = sentinel{
	kind:    uint(IrohErrorKindConnect),
	message: "iroh: failure while initiating or completing an outgoing connection",
}

var ConnectionError error = sentinel{
	kind:    uint(IrohErrorKindConnection),
	message: "iroh: an established connection failed or closed unexpectedly",
}

var AlpnError error = sentinel{
	kind:    uint(IrohErrorKindAlpn),
	message: "iroh: ALPN negotiation or lookup failed",
}

var KeyParsingError error = sentinel{
	kind:    uint(IrohErrorKindKeyParsing),
	message: "iroh: endpoint id / public key parsing failed",
}

var TicketParsingError error = sentinel{
	kind:    uint(IrohErrorKindKeyParsing),
	message: "iroh: ticket parsing failed",
}

var RelayError error = sentinel{
	kind:    uint(IrohErrorKindRelay),
	message: "iroh: relay configuration or relay operation failed",
}

var StreamError error = sentinel{
	kind:    uint(IrohErrorKindStream),
	message: "iroh: stream read/write/control operation failed",
}

var DatagramError error = sentinel{
	kind:    uint(IrohErrorKindDatagram),
	message: "iroh: datagram send/receive operation failed",
}

var FCallbackError error = sentinel{
	kind:    uint(IrohErrorKindCallback),
	message: "iroh: foreign callback failed",
}

var ClosedError error = sentinel{
	kind:    uint(IrohErrorKindClosed),
	message: "iroh: operation was attempted on a closed stream/connection/resource",
}

var TimeoutError error = sentinel{
	kind:    uint(IrohErrorKindTimeout),
	message: "iroh: operation timed out",
}

var InternalError error = sentinel{
	kind:    uint(IrohErrorKindInternal),
	message: "iroh: unclassified internal error",
}
