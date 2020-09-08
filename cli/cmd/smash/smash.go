package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/evmar/smash/bash"
	"github.com/evmar/smash/proto"
	"github.com/evmar/smash/vt100"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

var completer *bash.Bash
var globalLastTermForCmd *vt100.Terminal
var globalSockPathForEnv string

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

func (c *conn) writeMsg(msg proto.Msg) error {
	m := &proto.ServerMsg{msg}
	w := &bytes.Buffer{}
	if err := m.Write(w); err != nil {
		return err
	}
	c.Lock()
	defer c.Unlock()
	return c.ws.WriteMessage(websocket.BinaryMessage, w.Bytes())
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
	req *proto.RunRequest
	cmd *exec.Cmd

	// stdin accepts input keys and forwards them to the subprocess.
	stdin chan []byte
}

func newCmd(conn *conn, req *proto.RunRequest) *command {
	cmd := &exec.Cmd{Path: req.Argv[0], Args: req.Argv}
	// TODO: accept environment from the client
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SMASH_SOCK="+globalSockPathForEnv)
	cmd.Dir = req.Cwd
	return &command{
		conn: conn,
		req:  req,
		cmd:  cmd,
	}
}

func (cmd *command) send(msg proto.Msg) error {
	return cmd.conn.writeMsg(&proto.CellOutput{
		Cell:   cmd.req.Cell,
		Output: proto.Output{msg},
	})
}

func (cmd *command) sendError(msg string) error {
	return cmd.send(&proto.CmdError{msg})
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
		update := &proto.TermUpdate{}
		if tr.Dirty.Cursor {
			update.Cursor = proto.Cursor{
				Row:    term.Row,
				Col:    term.Col,
				Hidden: term.HideCursor,
			}
		}
		for row, l := range term.Lines {
			if !(allDirty || tr.Dirty.Lines[row]) {
				continue
			}
			rowSpans := proto.RowSpans{
				Row: row,
			}
			span := proto.Span{}
			var attr vt100.Attr
			for _, cell := range l {
				if cell.Attr != attr {
					attr = cell.Attr
					rowSpans.Spans = append(rowSpans.Spans, span)
					span = proto.Span{Attr: int(attr)}
				}
				// TODO: super inefficient.
				span.Text += fmt.Sprintf("%c", cell.Ch)
			}
			if len(span.Text) > 0 {
				rowSpans.Spans = append(rowSpans.Spans, span)
			}
			update.Rows = append(update.Rows, rowSpans)
		}

		err := cmd.send(update)
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

		if drawPending { // There can be no draw pending if done != nil.
			renderFromDirty()
			tr.Dirty.Reset()
			drawPending = false
		}

		mu.Unlock()

		if done != nil {
			break
		}
	}

	mu.Lock()
	globalLastTermForCmd = term

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
	cmd.send(&proto.Exit{exitCode})
}

var localCommands = map[string]func(w io.Writer) error{
	"that": func(w io.Writer) error {
		if globalLastTermForCmd == nil {
			return nil
		}
		_, err := io.WriteString(w, globalLastTermForCmd.ToString())
		return err
	},
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

func mapPairs(m map[string]string) []proto.Pair {
	pairs := []proto.Pair{}
	for k, v := range m {
		pairs = append(pairs, proto.Pair{k, v})
	}
	return pairs
}

func serveWS(w http.ResponseWriter, r *http.Request) error {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	conn := &conn{
		ws: wsConn,
	}

	smashPath, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return err
	}

	aliases, err := bash.GetAliases()
	if err != nil {
		return err
	}
	env := getEnv()
	env["SMASH"] = smashPath
	env["SMASH_SOCK"] = globalSockPathForEnv
	hello := &proto.Hello{
		Alias: mapPairs(aliases),
		Env:   mapPairs(env),
	}
	if err = conn.writeMsg(hello); err != nil {
		return err
	}

	commands := map[int]*command{}
	for {
		_, buf, err := conn.ws.ReadMessage()
		if err != nil {
			return fmt.Errorf("reading client message: %s", err)
		}
		var msg proto.ClientMessage
		if err := msg.Read(bufio.NewReader(bytes.NewBuffer(buf))); err != nil {
			return fmt.Errorf("parsing client message: %s", err)
		}

		switch msg := msg.Alt.(type) {
		case *proto.RunRequest:
			cmd := newCmd(conn, msg)
			commands[int(msg.Cell)] = cmd
			go cmd.runHandlingErrors()
		case *proto.KeyEvent:
			cmd := commands[int(msg.Cell)]
			if cmd == nil {
				log.Println("got key msg for unknown command", msg.Cell)
				continue
			}
			// TODO: what if cmd failed?
			// TODO: what if pipe is blocked?
			cmd.stdin <- []byte(msg.Keys)
		case *proto.CompleteRequest:
			if msg.Cwd == "" {
				panic("incomplete complete request")
			}
			go func() {
				if err := completer.Chdir(msg.Cwd); err != nil {
					log.Println(err) // TODO
				}
				pos, completions, err := completer.Complete(msg.Input[0:msg.Pos])
				if err != nil {
					log.Println(err) // TODO
				}
				err = conn.writeMsg(&proto.CompleteResponse{
					Id:          msg.Id,
					Pos:         pos,
					Completions: completions,
				})
				if err != nil {
					log.Println(err) // TODO
				}
			}()
		default:
			log.Println("unhandled msg", msg)
		}
	}
}

func serve() error {
	sockPath, localSock, err := setupLocalCommandSock()
	if err != nil {
		return err
	}
	go func() {
		if err := readLocalCommands(localSock); err != nil {
			fmt.Fprintf(os.Stderr, "local sock: %s\n", err)
		}
	}()
	globalSockPathForEnv = sockPath

	b, err := bash.StartBash()
	if err != nil {
		return err
	}
	completer = b

	http.Handle("/", http.FileServer(http.Dir("../web/dist")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := serveWS(w, r); err != nil {
			log.Printf("error: %s", err)
		}
	})
	addr := ":8080"
	fmt.Printf("listening on %q\n", addr)
	return http.ListenAndServe(addr, nil)
}

func localCommand(cmd string) error {
	sockPath := os.Getenv("SMASH_SOCK")
	if sockPath == "" {
		return fmt.Errorf("no $SMASH_SOCK; are you running under smash?")
	}

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = fmt.Fprintf(conn, "%s\n", cmd); err != nil {
		return err
	}
	if _, err = io.Copy(os.Stdout, conn); err != nil {
		return err
	}
	return nil
}

func main() {
	var cmd = "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	if _, isLocal := localCommands[cmd]; isLocal {
		err = localCommand(cmd)
	} else {
		switch cmd {
		case "serve":
			err = serve()
		case "help":
		default:
			fmt.Println("TODO: usage")
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
