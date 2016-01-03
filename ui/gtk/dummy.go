// When you run tests with the "headless" build tag, e.g.
//   go test -tags headless ./...
// It complains about this directory because it then has no source files.
// This empty file seems to work around the problem, but it seems there
// ought to be some more satisfactory option.
//
// See also https://github.com/golang/go/issues/8279 .

package gtk
