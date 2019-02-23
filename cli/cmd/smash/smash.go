package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/evmar/smash/proto"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

func runCmd(conn *websocket.Conn, cmdline string) error {
	cmd := exec.Command(cmdline)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	var buf [64 << 10]byte
	for {
		n, err := out.Read(buf[:])
		if err != nil {
			return err
		}

		b := flatbuffers.NewBuilder(64 << 10)
		c := b.CreateByteString(buf[:n])
		proto.RespOutputStart(b)
		proto.RespOutputAddText(b, c)
		respOutput := proto.RespOutputEnd(b)
		b.Finish(respOutput)
		msg := b.FinishedBytes()

		if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			return err
		}
	}
}

func serveWS(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	for {
		_, buf, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		msg := proto.GetRootAsReqRun(buf, 0)
		log.Printf("%q", msg.Cmd())

		if err := runCmd(conn, string(msg.Cmd())); err != nil {
			return nil
		}
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
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := serveWS(w, r); err != nil {
			log.Println(err)
		}
	})
	addr := ":8080"
	fmt.Printf("listening on %q\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
