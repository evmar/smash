package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	pb "github.com/evmar/smash/proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

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
	http.Handle("/", http.FileServer(http.Dir("../web/dist")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := serveWS(w, r); err != nil {
			log.Println(err)
		}
	})
	addr := ":8080"
	fmt.Printf("listening on %q\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
