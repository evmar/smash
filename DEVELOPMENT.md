# Development notes

While developing:

```sh
$ make web   # auto-update the JS
$ make cli   # run the go server
```

Now reloading the page reloads the content.

## Formatter

```sh
$ ./fmt.sh --write
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
