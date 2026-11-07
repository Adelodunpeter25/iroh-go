> **Fork notice.** This is a fork of [decentral1se/iroh-go](https://git.coopcloud.tech/decentral1se/iroh-go)
> (Apache-2.0 / MIT, licenses kept in `licenses/`). It adds macOS and Windows
> static libraries built by `.github/workflows/libs.yml`, and is maintained
> independently. Module path: `github.com/Adelodunpeter25/iroh-go`.
> Credit for the bindings goes to the original author.

# `iroh-go`

[![Build Status](https://build.coopcloud.tech/api/badges/decentral1se/iroh-go/status.svg?ref=refs/heads/main)](https://build.coopcloud.tech/decentral1se/iroh-go)

> **WARNING** Highly Experimental ™️

Go FFI bindings for [`iroh`](https://iroh.computer) via
[`iroh-ffi`](https://github.com/n0-computer/iroh-ffi).

## Motivation

The goal is to provide, as far as possible, seamless "just `go build` it"
support for Go iroh hackers without requiring them to deal with shared
libraries, the Rust toolchain or a deepdive in how FFI works. The current
approach is to embed `libiroh` as a `staticlib` built with Rust for each
platform. We use [`musl`](https://musl.libc.org) to avoid `glibc`
incompatibilities on different linux systems. It's then possible to build a
static binary with `go build` or [with
Zig](https://dev.to/kristoff/zig-makes-go-cross-compilation-just-work-29ho).
`iroh-go` does require `CGO_ENABLED=1`.

## Approach

`iroh-go` is generated directly from
[`iroh-ffi`](https://github.com/n0-computer/iroh-ffi) guaranteeing
compatibility and long-term alignment with upstream. Some API wrappers are
implemented to smooth over some of the unavoidable rough edges of FFI binding
generation. This is a different approach than e.g.
[`go-iroh`](https://github.com/tmc/go-iroh) which is a custom port.

## Version support

* `iroh`: [`v1.0.0`](https://github.com/n0-computer/iroh/releases/tag/v1.0.0)
* `iroh-ffi`: [`v1.2.0`](https://github.com/n0-computer/iroh-ffi/releases/tag/v1.2.0)

## Platform support

> **NOTE** More platform can be provided on request and pending some support to
> build binaries for specific architectures. If you would like to help out,
> please let us know! See [`#8`](https://git.coopcloud.tech/decentral1se/iroh-go/issues/8)
> and [`#9`](https://git.coopcloud.tech/decentral1se/iroh-go/issues/8) for more.

* `x86_64-unknown-linux-musl`
* `aarch64-unknown-linux-musl`

## Documentation

* Original Rust API: [`docs.rs`](https://docs.rs/iroh/1.0.0/iroh/)
* Go bindings: [`godocs.io`](https://godocs.io/git.coopcloud.tech/decentral1se/iroh-go)
* Community bindings docs: [`docs.iroh.computer`](https://docs.iroh.computer/languages/go)

## Examples

See [`examples`](./examples).

## Run your own `iroh-relay`

Each [`iroh` release](https://github.com/n0-computer/iroh/releases) features a
relay binary. You can also make use of the upstream provided docker image:
`n0computer/iroh-relay`.

### Local hacking

First, run `iroh-relay` in development mode with noisy logging.

```
RUST_LOG=info,iroh_relay=debug ./iroh-relay --dev
```

You can then use `http://localhost:3340` as the default location of your own
relay in the construction of your endpoint as follows. Once you create an
endpoint, you should see your relay respond locally with corresponding logs.

```go
preset := iroh.PresetN0DisableRelay()
relay, _ := iroh.RelayModeCustomFromUrls([]string{"http://localhost:3340"})
opts := iroh.EndpointOptions{
    Preset:    &preset,
    RelayMode: &relay,
    // ...moar config...
}
```

## Generate bindings

See [`iroh-go`](./iroh-go) for the Rust toolchain configuration. Run
`./scripts/generate.sh` to build a new set of libraries which end up vendored
in [`libs`](./libs).

## Who is using it?

> **NOTE** Please add your own if you're using `iroh-go` :pray:

* [`iroh-tracker`](https://git.coopcloud.tech/cblgh/iroh-tracker)

## Known issues

* We patched the upstream `iroh-ffi` `endpoint.rs` to fix an incompatibility
  with
  [`uniffi-bindgen-go`](https://github.com/NordSecurity/uniffi-bindgen-go). See
  [this
  commit](https://github.com/decentral1se/iroh-ffi/commit/3e102569e1c348480416a9b05cf020760f607dfe)
  for more.
* We patched the upstream `iroh-ffi` `Cargo.toml` to remove the `crate-type`
  configuration when runnings builds here. See [this
  commit](https://github.com/decentral1se/iroh-ffi/commit/207e5d7cd7dee6eab350d3aad85029431a63ca3f)
  for more.
* Error handling is not super ergonomic due to
  [`iroh-ffi#263`](https://github.com/n0-computer/iroh-ffi/issues/263). We have
  created an error handling wrapper API which is documented in
  [`./examples/errors`](./examples/errors/) to improve on the situation.

## Licenses

Apache 2.0 & MIT, following the upstream licensing choices in
[`iroh-ffi`](https://github.com/n0-computer/iroh-ffi).

## LLM Policy

We adopt the [Strict No LLM / No AI
Policy](https://ziglang.org/code-of-conduct/#strict-no-llm-no-ai-policy) of
[Zig](https://ziglang.org).
