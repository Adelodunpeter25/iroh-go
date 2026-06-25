# connect-desktop-example

A version of [`../connect`](../connect) as a desktop application. On running the desktop
application, an Iroh endpoint id will be displayed. This id can be pluggedinto a terminal
running the [`../connect`](../connect) example as a sender. What the sender types in the
terminal will appear in the connect-desktop application window.

## Dependencies

* Wails v2.10.2 (old version, haven't tried on a newer yet - hopefully works there too!)
* Vite v3.0.7
* Go v1.18
* npm v10.2+
* node v20.11+

**Getting setup**

* Install node and golang

Install Wails:

```
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2 // for tested version
// OR: go install github.com/wailsapp/wails/v2/cmd/wails@latest // for latest

// then install frontend dependency
cd frontend 
npm i
```

## Running in development mode

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

Run [`../connect`](../connect) in a terminal and connect to the endpoint id displayed by the
listener.
```
./connect -endpoint <endpoint-id>
```

You'll see a connection notice if the two endpoints manage to connect to each other. You can
type messages in the sender terminal and they will be shown on the screen of the desktop app.
