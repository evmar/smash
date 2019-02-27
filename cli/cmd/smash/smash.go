package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	pb "github.com/evmar/smash/proto"
	"github.com/evmar/smash/vt100"
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

func (w *wsWriter) WriteText(row int, text *pb.TermText) error {
	return w.conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   w.cell,
		Output: &pb.Output_Text{text},
	}}})
}

func (w *wsWriter) WriteError(msg string) error {
	return w.conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   w.cell,
		Output: &pb.Output_Error{msg},
	}}})
}

// isPtyEOFError tests for a pty close error.
// When a pty closes, you get an EIO error instead of an EOF.
func isPtyEOFError(err error) bool {
	const EIO syscall.Errno = 5
	if perr, ok := err.(*os.PathError); ok {
		if errno, ok := perr.Err.(syscall.Errno); ok && errno == EIO {
			// read /dev/ptmx: input/output error
			return true
		}
	}
	return false
}

func termLoop(tr *vt100.TermReader, r io.Reader) {
	br := bufio.NewReader(r)
	for {
		if err := tr.Read(br); err != nil {
			if isPtyEOFError(err) {
				err = io.EOF
			}
			if err != io.EOF {
				// TODO: w.WriteError(err.Error())
			}
			return
		}
	}
}

func spawn(w *wsWriter, cmd *exec.Cmd) error {
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

	var mu sync.Mutex
	term := vt100.NewTerminal()
	var tr *vt100.TermReader
	tr = vt100.NewTermReader(func(f func(t *vt100.Terminal)) {
		mu.Lock()
		defer mu.Unlock()
		f(term)
		allDirty := tr.Dirty.Lines[-1]
		for row, l := range term.Lines {
			if !(allDirty || tr.Dirty.Lines[row]) {
				continue
			}
			text := &pb.TermText{
				Row: int32(row),
			}
			span := &pb.TermText_Span{}
			var attr vt100.Attr
			for _, cell := range l {
				if cell.Attr != attr {
					attr = cell.Attr
					text.Spans = append(text.Spans, span)
					span = &pb.TermText_Span{Attr: int32(attr)}
				}
				// TODO: super inefficient.
				span.Text += fmt.Sprintf("%c", cell.Ch)
			}
			if len(span.Text) > 0 {
				text.Spans = append(text.Spans, span)
			}
			w.WriteText(row, text)
		}
		tr.Dirty.Reset()
	})

	go termLoop(tr, f)

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
			w.WriteError(err.Error())
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
