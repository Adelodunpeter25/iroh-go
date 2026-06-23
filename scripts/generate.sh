#!/usr/bin/env bash

set -euxo pipefail

THIS_SCRIPT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
LIB_NAME="libiroh.a"

RUST_FOLDER="$THIS_SCRIPT_DIR/../iroh-go"
FFI_FOLDER="$THIS_SCRIPT_DIR/../iroh-ffi"
GO_FOLDER="$THIS_SCRIPT_DIR/../"

# TARGETS="x86_64-unknown-linux-musl aarch64-unknown-linux-musl aarch64-apple-darwin x86_64-apple-darwin"
TARGETS="x86_64-unknown-linux-musl"

if ! command -v cross &> /dev/null; then
  echo "▸ cross tool not found, installing"
  cargo install cross --git https://github.com/cross-rs/cross
fi

if ! command -v rustup &> /dev/null; then
  echo "▸ rustup not found, please install it"
  exit 1
else
  rustup target add $TARGETS
fi

export RUSTFLAGS="-C target-feature=+crt-static"
for TARGET in $TARGETS; do
  echo "▸ building for $TARGET"
  cross build --manifest-path "$RUST_FOLDER/Cargo.toml" --target "$TARGET" --lib --locked --release
  mkdir -p "${GO_FOLDER}/libs/${TARGET}"
  cp "${RUST_FOLDER}/target/${TARGET}/release/${LIB_NAME}" "${GO_FOLDER}/libs/${TARGET}/${LIB_NAME}"
done

TARGET="x86_64-unknown-linux-musl"
echo "▸ Generate Go bindings"
cd "$RUST_FOLDER"
uniffi-bindgen-go \
    "${GO_FOLDER}/libs/${TARGET}/${LIB_NAME}" \
    --library \
    --out-dir "${RUST_FOLDER}/target/go"

test -f "${RUST_FOLDER}/target/go/iroh_ffi/iroh_ffi.go"
cp -r "${RUST_FOLDER}/target/go/iroh_ffi/iroh_ffi.go" "${GO_FOLDER}"

test -f "${RUST_FOLDER}/target/go/iroh_ffi/iroh_ffi.h"
cp -r "${RUST_FOLDER}/target/go/iroh_ffi/iroh_ffi.h" "${GO_FOLDER}"
