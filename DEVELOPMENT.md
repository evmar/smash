# Development notes

While developing, in three separate terminals:

```sh
$ make tsc   # run TypeScript compiler in watch mode
$ make web   # auto-update the JS
$ make cli   # run the go server
```

Now reloading the page reloads the content.

## Formatter

```sh
$ make fmt
```

To run prettier, which is checked on presubmit.

## Protocol changes

```sh
$ make proto
```

Regenerates the protobuf APIs.

## Testing

HTML/JS-only tests are in `web/src/test.ts`, driving a headless Chrome:

```sh
$ cd web; npm run test
```

Go tests use the Go test runner:

```sh
$ cd cli; go test ./...
```

## Chrome PWA

PWAs only work on https or localhost. For one of these on ChromeOS,
the best option seems to be connection forwarding using [Connection
Forwarder](https://chrome.google.com/webstore/detail/connection-forwarder/ahaijnonphgkgnkbklchdhclailflinn) to forward localhost into the crostini IP.

Update: digging in the Chrome source suggests that on ChromeOS specifically,
Chrome also treats penguin.linux.test as a trusted domain. However, I've
never been able to make the Chrome PWA bits work on ChromeOS (localhost
or penguin.linux.test) so I'll leave the previous paragraph here until
I'm confident of the resolution.

## The icon

```sh
$ convert -size 32x32 -gravity center -background white -fill black label:">" icon.png
```

## vt100

Run `script` then the command to capture raw terminal output.

Run `infocmp -L` to understand what the terminal outputs mean.

## bash

To experiment with the bash completion support, run:

```
$ cd cli && go run ./bash/demo
```
