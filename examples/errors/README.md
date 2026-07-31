# errors

This example shows a brief example of how to deal with error handling in a
convenient way. Unfortunately, the error handling coming down the pipes in the
FFI bindings is not ergonomic by default ends up creating extremely verbose
error handling. We've created an error handling wrapper API which we've tried
to make as intuitive as possible. See
[`#263`](https://github.com/n0-computer/iroh-ffi/issues/263) for the juicy
details.
