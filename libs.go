package iroh_ffi

// #cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/x86_64-unknown-linux-musl -liroh -lm
// #cgo linux,arm64 LDFLAGS: -L${SRCDIR}/libs/aarch64-unknown-linux-musl -liroh -lm
import "C"
