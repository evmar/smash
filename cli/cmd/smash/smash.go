package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func serveHTML(w io.Writer) error {
	_, err := w.Write([]byte(`<!doctype html>
<body></body>
<script>`))
	if err != nil {
		return err
	}
	f, err := os.Open("../web/dist/bundle.js")
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	_, err = w.Write([]byte(`</script>`))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			err := serveHTML(w)
			if err != nil {
				log.Println(err)
			}
			return
		}
		http.NotFound(w, r)
	})
	addr := ":8080"
	fmt.Printf("listening on %q\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
