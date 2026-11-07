package iroh_ffi

// #cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/x86_64-unknown-linux-musl -liroh -lm
// #cgo linux,arm64 LDFLAGS: -L${SRCDIR}/libs/aarch64-unknown-linux-musl -liroh -lm
// #cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/libs/x86_64-apple-darwin -liroh -lm -framework Security -framework CoreFoundation -framework SystemConfiguration
// #cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/libs/aarch64-apple-darwin -liroh -lm -framework Security -framework CoreFoundation -framework SystemConfiguration
// #cgo windows,amd64 LDFLAGS: -L${SRCDIR}/libs/x86_64-pc-windows-gnu -liroh -lws2_32 -luserenv -lbcrypt -lntdll -lcrypt32 -lsecur32 -lncrypt -ladvapi32
import "C"
