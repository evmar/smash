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
	"time"

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

// conn wraps a websocket.Conn with a lock.
type conn struct {
	sync.Mutex
	ws *websocket.Conn
}

func (c *conn) writeMsg(msg *pb.ServerMsg) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	c.Lock()
	defer c.Unlock()
	return c.ws.WriteMessage(websocket.BinaryMessage, data)
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

// command represents a subprocess running on behalf of the user.
// req.Cell has the id of the command for use in protocol messages.
type command struct {
	conn *conn
	// req is the initial request that caused the command to be spawned.
	req *pb.RunRequest
	cmd *exec.Cmd

	// stdin accepts input keys and forwards them to the subprocess.
	stdin chan []byte
}

func newCmd(conn *conn, req *pb.RunRequest) *command {
	cmd := &exec.Cmd{Path: req.Argv[0], Args: req.Argv}
	cmd.Dir = req.Cwd
	return &command{
		conn: conn,
		req:  req,
		cmd:  cmd,
	}
}

func (cmd *command) send(out pb.IsOutput_Output) error {
	return cmd.conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
		Cell:   cmd.req.Cell,
		Output: out,
	}}})
}

func (cmd *command) sendError(msg string) error {
	return cmd.send(&pb.Output_Error{msg})
}

func termLoop(tr *vt100.TermReader, r io.Reader) error {
	br := bufio.NewReader(r)
	for {
		if err := tr.Read(br); err != nil {
			if isPtyEOFError(err) {
				err = io.EOF
			}
			return err
		}
	}
}

func (cmd *command) run() error {
	if filepath.Base(cmd.cmd.Path) == cmd.cmd.Path {
		// TODO: should use shell env $PATH.
		if p, err := exec.LookPath(cmd.cmd.Path); err != nil {
			return err
		} else {
			cmd.cmd.Path = p
		}
	}

	f, err := pty.Start(cmd.cmd)
	if err != nil {
		return err
	}

	cmd.stdin = make(chan []byte)
	go func() {
		for input := range cmd.stdin {
			f.Write(input)
		}
	}()

	var mu sync.Mutex // protects term, drawPending, and done
	wake := sync.NewCond(&mu)
	term := vt100.NewTerminal()
	drawPending := false
	var done error

	var tr *vt100.TermReader
	renderFromDirty := func() {
		allDirty := tr.Dirty.Lines[-1]
		text := &pb.TermText{}
		for row, l := range term.Lines {
			if !(allDirty || tr.Dirty.Lines[row]) {
				continue
			}
			rowSpans := &pb.TermText_RowSpans{
				Row: int32(row),
			}
			text.Rows = append(text.Rows, rowSpans)
			span := &pb.TermText_Span{}
			var attr vt100.Attr
			for _, cell := range l {
				if cell.Attr != attr {
					attr = cell.Attr
					rowSpans.Spans = append(rowSpans.Spans, span)
					span = &pb.TermText_Span{Attr: int32(attr)}
				}
				// TODO: super inefficient.
				span.Text += fmt.Sprintf("%c", cell.Ch)
			}
			if len(span.Text) > 0 {
				rowSpans.Spans = append(rowSpans.Spans, span)
			}
		}

		cmd.conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Output{&pb.Output{
			Cell:   cmd.req.Cell,
			Output: &pb.Output_Text{text},
		}}})
	}

	tr = vt100.NewTermReader(func(f func(t *vt100.Terminal)) {
		// This is called from the 'go termLoop' goroutine,
		// when the vt100 impl wants to update the terminal.
		mu.Lock()
		f(term)
		if !drawPending {
			drawPending = true
			wake.Signal()
		}
		mu.Unlock()
	})

	go func() {
		err := termLoop(tr, f)
		mu.Lock()
		done = err
		wake.Signal()
		mu.Unlock()
	}()

	for {
		mu.Lock()
		for !drawPending && done == nil {
			wake.Wait()
		}

		if done == nil {
			mu.Unlock()
			// Allow more pending paints to enqueue.
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
		}

		renderFromDirty()
		tr.Dirty.Reset()
		drawPending = false
		mu.Unlock()

		if done != nil {
			break
		}
	}

	log.Println(done) // TODO: use error
	exitCode := 0
	if err := cmd.cmd.Wait(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			serr := eerr.Sys().(syscall.WaitStatus)
			exitCode = serr.ExitStatus()
		} else {
			cmd.sendError(err.Error())
			exitCode = 1
		}
	}
	return cmd.send(&pb.Output_ExitCode{int32(exitCode)})
}

func serveWS(w http.ResponseWriter, r *http.Request) error {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	conn := &conn{
		ws: wsConn,
	}
	commands := map[int]*command{}
	for {
		_, buf, err := conn.ws.ReadMessage()
		if err != nil {
			return err
		}
		msg := pb.ClientMessage{}
		if err := proto.Unmarshal(buf, &msg); err != nil {
			return err
		}

		if run := msg.GetRun(); run != nil {
			cmd := newCmd(conn, run)
			commands[int(run.Cell)] = cmd
			go func() {
				err := cmd.run()
				if err != nil {
					log.Println(err)
				}
			}()
		} else if key := msg.GetKey(); key != nil {
			cmd := commands[int(key.Cell)]
			if cmd == nil {
				log.Println("got key msg for unknown command", key.Cell)
				continue
			}
			// TODO: what if cmd failed?
			// TODO: what if pipe is blocked?
			cmd.stdin <- []byte(key.Keys)
		}
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
