package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	pb "github.com/evmar/smash/proto"
	"github.com/golang/protobuf/proto"
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
			if err == io.EOF {
				return nil
			}
			return err
		}

		msg := pb.OutputResponse{
			Text: string(buf[:n]),
		}
		data, err := proto.Marshal(&msg)
		if err != nil {
			return err
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
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
		msg := pb.RunRequest{}
		if err := proto.Unmarshal(buf, &msg); err != nil {
			return err
		}

		if err := runCmd(conn, string(msg.Command)); err != nil {
			return err
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
