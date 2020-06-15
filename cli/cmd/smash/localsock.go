package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// handleLocal handles an incoming local connection,
// by reading a command from the connection and writing
// its output to it.
func handleLocal(conn net.Conn) error {
	defer conn.Close()
	var buf [1 << 10]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		return err
	}
	cmd := strings.TrimSpace(string(buf[:n]))
	cmdFn := localCommands[cmd]
	if cmdFn != nil {
		return cmdFn(conn)
	}
	_, err = io.WriteString(conn, "bad command\n")
	return err
}

// getSockPath gets a (hopefully unique) path for storing the smash socket.
// (Note that the path doesn't need to be predictable across invocations,
// as the socket path is passed to subcommands via the environment.)
func getSockPath() (string, error) {
	path := os.Getenv("XDG_RUNTIME_DIR")
	if path == "" {
		var err error
		path, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	path = filepath.Join(path, "smash")
	if err := os.MkdirAll(path, 0700); err != nil {
		return "", err
	}

	sockName := fmt.Sprintf("sock.%d", os.Getpid())
	path = filepath.Join(path, sockName)
	if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
		if err := os.Remove(path); err != nil {
			return "", err
		}
	}

	return path, nil
}

// deleteOnExit attempts to delete the given path when you ctl-c.
func deleteOnExit(path string) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(path)
		os.Exit(128 + int(syscall.SIGTERM))
	}()
}

// setupLocalCommandSock creates the listening local command socket, and
// returns its path and the socket.
func setupLocalCommandSock() (string, net.Listener, error) {
	path, err := getSockPath()
	if err != nil {
		return "", nil, err
	}
	l, err := net.Listen("unix", path)
	if err != nil {
		deleteOnExit(path)
	}
	return path, l, err
}

// readLocalCommands loops forever, reading commands from the local socket.
func readLocalCommands(sock net.Listener) error {
	defer sock.Close()
	for {
		conn, err := sock.Accept()
		if err != nil {
			return err
		}
		go func() {
			err := handleLocal(conn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "local conn: %s\n", err)
			}
		}()
	}
}
