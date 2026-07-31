# connect-desktop

> **WARNING** Due to Wails requiring some additional system dependencies, we
> are not testing this example regularly on CI/CD. Therefore, this example
> might not build when you try it. Contributions are very welcome!

A version of [`../connect`](../connect) as a desktop application. When running
the desktop application, an Iroh endpoint id will be displayed. This id can be
plugged into a terminal running the [`../connect`](../connect) example as a
sender. What the sender types in the terminal will appear in the
connect-desktop application window.

## Dependencies

This example has been tested on the following versions.

* Wails v2.10.2
* Node v20.11+

Please consult the [Wails
docs](https://wails.io/docs/gettingstarted/installation) for everything you
need to get up and running.

### Wails

```
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2
cd frontend && npm i
```

## Development mode

```
wails dev
```

## Build a binary

```
wails build .
``` 

## Listener (this example)

```
./connect-desktop
```

## Terminal 2 (sender)

Run [`../connect`](../connect) in a terminal and connect to the endpoint id
displayed by the listener.

```
./connect -endpoint <endpoint-id>
```

You'll see a connection notice if the two endpoints manage to connect to each
other. You can type messages in the sender terminal and they will be shown on
the screen of the desktop app.
