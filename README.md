# `iroh-go`

> **WARNING** Highly Experimental ™️

Go FFI bindings for [`iroh`](https://iroh.computer) via
[`iroh-ffi`](https://github.com/n0-computer/iroh-ffi).

## Motivation

The goal is to provide, as far as possible, seamless "just `go build` it"
support for Go iroh hackers without requiring them to deal with shared
libraries, the Rust toolchain or a deepdive in how FFI works.

The current approach is to embed `libiroh` as a `staticlib` built with Rust for
each platform. We use [`musl`](https://musl.libc.org) to avoid `glibc`
incompatibilities on different linux systems.

It's then possible to build a static binary with `go build` or [with
Zig](https://dev.to/kristoff/zig-makes-go-cross-compilation-just-work-29ho).
`iroh-go` does require `CGO_ENABLED=1`.

## Platform support

* `x86_64-unknown-linux-musl`

## Documentation

See [`docs.rs`](https://docs.rs/iroh/1.0.0/iroh/).

## Examples

See [`examples`](./examples).

## Known issues

* We need to patch the upstream `iroh-ffi` `endpoint.rs` to fix an
  incompatibility with
  [`uniffi-bindgen-go`](https://github.com/NordSecurity/uniffi-bindgen-go). See
  [this
  commit](https://github.com/decentral1se/iroh-ffi/commit/3e102569e1c348480416a9b05cf020760f607dfe)
  for more.
* We need to patch the upstream `iroh-ffi` `Cargo.toml` to remove the
  `crate-type` configuration when runnings builds here. See [this
  commit](https://github.com/decentral1se/iroh-ffi/commit/207e5d7cd7dee6eab350d3aad85029431a63ca3f)
  for more.
