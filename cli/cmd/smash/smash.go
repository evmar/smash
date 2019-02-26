package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
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
	log.Println("run:", req)
	cmd := &exec.Cmd{Path: req.Argv[0], Args: req.Argv}
	cmd.Dir = req.Cwd
	w := &wsWriter{conn: conn, cell: req.Cell}
	exitCode := 0
	if filepath.Base(cmd.Path) == cmd.Path {
		if p, err := exec.LookPath(cmd.Path); err != nil {
			fmt.Fprintf(w, "ERROR: %s\n", err)
			exitCode = 1
			cmd = nil
		} else {
			cmd.Path = p
		}
	}
	if cmd != nil {
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			if eerr, ok := err.(*exec.ExitError); ok {
				serr := eerr.Sys().(syscall.WaitStatus)
				exitCode = serr.ExitStatus()
			} else {
				fmt.Fprintf(w, "ERROR: %s\n", err)
			}
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
