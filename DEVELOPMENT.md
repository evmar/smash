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

## Chrome PWA

PWAs only work on https or localhost. For one of these on ChromeOS,
the best option seems to be connection forwarding using [Connection
Forwarder](https://chrome.google.com/webstore/detail/connection-forwarder/ahaijnonphgkgnkbklchdhclailflinn) to forward localhost into the crostini IP.

## The icon

```sh
$ convert -size 32x32 -gravity center -background white -fill black label:">" icon.png
```

## vt100

Run `script` then the command to capture raw terminal output.

Run `infocmp -L` to understand what the terminal outputs mean.
