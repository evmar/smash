package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	pb "github.com/evmar/smash/proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

type conn struct {
	mu sync.Mutex
	ws *websocket.Conn
}

func (c *conn) writeMsg(msg *pb.ServerMsg) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.WriteMessage(websocket.BinaryMessage, data)
}

type wsWriter struct {
	conn *conn
	cell int32
}

func (w *wsWriter) Write(buf []byte) (int, error) {
	err := w.conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   w.cell,
		Output: &pb.Output_Text{string(buf[:])},
	}}})
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}

func spawn(w io.Writer, cmd *exec.Cmd) error {
	if filepath.Base(cmd.Path) == cmd.Path {
		// TODO: should use shell env $PATH.
		if p, err := exec.LookPath(cmd.Path); err != nil {
			return err
		} else {
			cmd.Path = p
		}
	}
	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	go func() {
		io.Copy(w, f)
	}()
	return cmd.Wait()
}

func runCmd(conn *conn, req *pb.RunRequest) {
	log.Println("run:", req)
	cmd := &exec.Cmd{Path: req.Argv[0], Args: req.Argv}
	cmd.Dir = req.Cwd
	w := &wsWriter{conn: conn, cell: req.Cell}
	exitCode := 0
	if err := spawn(w, cmd); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			serr := eerr.Sys().(syscall.WaitStatus)
			exitCode = serr.ExitStatus()
		} else {
			fmt.Fprintf(w, "ERROR: %s\n", err)
			exitCode = 1
		}
	}
	err := conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   w.cell,
		Output: &pb.Output_ExitCode{int32(exitCode)},
	}}})
	if err != nil {
		log.Println(err)
	}
}

func serveWS(w http.ResponseWriter, r *http.Request) error {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	conn := &conn{
		ws: wsConn,
	}
	for {
		_, buf, err := conn.ws.ReadMessage()
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
