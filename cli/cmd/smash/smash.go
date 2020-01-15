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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/evmar/smash/bash"
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

// run synchronously runs the subprocess to completion, sending terminal
// updates as it progresses.  It may return errors if the subprocess failed
// to run for whatever reason (e.g. no such path), and otherwise returns
// the subprocess exit code.
func (cmd *command) run() (int, error) {
	if cmd.cmd.Path == "cd" {
		if len(cmd.cmd.Args) != 2 {
			return 0, fmt.Errorf("bad arguments to cd")
		}
		dir := cmd.cmd.Args[1]
		st, err := os.Stat(dir)
		if err != nil {
			return 0, err
		}
		if !st.IsDir() {
			return 0, fmt.Errorf("%s: not a directory", dir)
		}
		return 0, nil
	}

	if filepath.Base(cmd.cmd.Path) == cmd.cmd.Path {
		// TODO: should use shell env $PATH.
		if p, err := exec.LookPath(cmd.cmd.Path); err != nil {
			return 0, err
		} else {
			cmd.cmd.Path = p
		}
	}

	size := pty.Winsize{
		Rows: 24,
		Cols: 80,
	}
	f, err := pty.StartWithSize(cmd.cmd, &size)
	if err != nil {
		return 0, err
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
		// Called with mu held.
		allDirty := tr.Dirty.Lines[-1]
		update := &pb.TermUpdate{}
		if tr.Dirty.Cursor {
			update.Cursor = &pb.TermUpdate_Cursor{
				Row:    int32(term.Row),
				Col:    int32(term.Col),
				Hidden: term.HideCursor,
			}
		}
		for row, l := range term.Lines {
			if !(allDirty || tr.Dirty.Lines[row]) {
				continue
			}
			rowSpans := &pb.TermUpdate_RowSpans{
				Row: int32(row),
			}
			update.Rows = append(update.Rows, rowSpans)
			span := &pb.TermUpdate_Span{}
			var attr vt100.Attr
			for _, cell := range l {
				if cell.Attr != attr {
					attr = cell.Attr
					rowSpans.Spans = append(rowSpans.Spans, span)
					span = &pb.TermUpdate_Span{Attr: int32(attr)}
				}
				// TODO: super inefficient.
				span.Text += fmt.Sprintf("%c", cell.Ch)
			}
			if len(span.Text) > 0 {
				rowSpans.Spans = append(rowSpans.Spans, span)
			}
		}

		err := cmd.send(&pb.Output_TermUpdate{update})
		if err != nil {
			done = err
		}
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

	mu.Lock()

	// done is the error reported by the terminal.
	// We expect EOF in normal execution.
	if done != io.EOF {
		return 0, err
	}

	// Reap the subprocess and report the exit code.
	if err := cmd.cmd.Wait(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			serr := eerr.Sys().(syscall.WaitStatus)
			return serr.ExitStatus(), nil
		} else {
			return 0, err
		}
	}
	return 0, nil
}

// runHandlingErrors calls run() and forwards any subprocess errors
// on to the client.
func (cmd *command) runHandlingErrors() {
	exitCode, err := cmd.run()
	if err != nil {
		cmd.sendError(err.Error())
		exitCode = 1
	}
	cmd.send(&pb.Output_ExitCode{int32(exitCode)})
}

func getEnv() map[string]string {
	env := map[string]string{}
	for _, keyval := range os.Environ() {
		eq := strings.Index(keyval, "=")
		if eq < 0 {
			panic("bad env?")
		}
		env[keyval[:eq]] = keyval[eq+1:]
	}
	return env
}

func serveWS(w http.ResponseWriter, r *http.Request) error {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	conn := &conn{
		ws: wsConn,
	}

	aliases, err := bash.GetAliases()
	if err != nil {
		return err
	}
	err = conn.writeMsg(&pb.ServerMsg{Msg: &pb.ServerMsg_Hello{&pb.Hello{
		Alias: aliases,
		Env:   getEnv(),
	}}})
	if err != nil {
		return err
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
			go cmd.runHandlingErrors()
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
