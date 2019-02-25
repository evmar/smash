package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"syscall"

	pb "github.com/evmar/smash/proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

func writeMsg(conn *websocket.Conn, msg *pb.ServerMsg) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}
	return nil
}

type wsWriter struct {
	conn *websocket.Conn
	cell int32
}

func (w *wsWriter) Write(buf []byte) (int, error) {
	err := writeMsg(w.conn, &pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   w.cell,
		Output: &pb.Output_Text{string(buf[:])},
	}}})
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}

func runCmd(conn *websocket.Conn, req *pb.RunRequest) {
	cmd := exec.Command("/bin/sh", "-c", req.Command)
	w := &wsWriter{conn: conn, cell: req.Cell}
	cmd.Stdout = w
	cmd.Stderr = w
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			serr := err.Sys().(syscall.WaitStatus)
			exitCode = serr.ExitStatus()
		} else {
			fmt.Fprintf(w, "%s\n", err)
		}
	}
	err := writeMsg(w.conn, &pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   w.cell,
		Output: &pb.Output_ExitCode{int32(exitCode)},
	}}})
	if err != nil {
		log.Println(err)
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

		runCmd(conn, &msg)
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
