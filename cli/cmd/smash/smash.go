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

type wsWriter struct {
	conn *websocket.Conn
}

func (w *wsWriter) Write(buf []byte) (int, error) {
	msg := pb.OutputResponse{
		Text: string(buf[:]),
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		return 0, err
	}
	if err := w.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return 0, err
	}
	return len(buf), nil
}

func runCmd(conn *websocket.Conn, cmdline string) {
	cmd := exec.Command("/bin/sh", "-c", cmdline)
	w := &wsWriter{conn}
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			// Exit failure, ignore.
		} else {
			fmt.Fprintf(w, "%s\n", err)
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

		runCmd(conn, string(msg.Command))
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
