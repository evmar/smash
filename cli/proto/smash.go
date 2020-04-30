package proto

import (
	"bufio"
	"fmt"
	"io"
)

type Msg interface {
	Write(w io.Writer) error
	Read(r *bufio.Reader) error
}

func ReadBoolean(r *bufio.Reader) (bool, error) {
	val, err := ReadUint8(r)
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

func ReadUint8(r *bufio.Reader) (byte, error) {
	return r.ReadByte()
}

func ReadInt(r *bufio.Reader) (int, error) {
	shift := 0
	var val int
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		val |= int(b&0b0111_1111) << shift
		if (b & 0b1000_0000) == 0 {
			return val, nil
		}
		shift = shift + 7
	}
}

func ReadString(r *bufio.Reader) (string, error) {
	n, err := ReadInt(r)
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func WriteBoolean(w io.Writer, val bool) error {
	if val {
		return WriteUint8(w, 1)
	} else {
		return WriteUint8(w, 0)
	}
}
func WriteUint8(w io.Writer, val byte) error {
	buf := [1]byte{val}
	_, err := w.Write(buf[:])
	return err
}
func WriteInt(w io.Writer, val int) error {
	if val < 0 {
		panic("negative")
	}
	for {
		b := byte(val & 0b0111_1111)
		val = val >> 7
		if val == 0 {
			return WriteUint8(w, b)
		} else {
			if err := WriteUint8(w, b|0b1000_0000); err != nil {
				return err
			}
		}
	}
}
func WriteString(w io.Writer, str string) error {
	if len(str) >= 1<<16 {
		panic("overlong")
	}
	if err := WriteInt(w, len(str)); err != nil {
		return err
	}
	_, err := w.Write([]byte(str))
	return err
}

type ClientMessage struct {
	// CompleteRequest, RunRequest, KeyEvent
	Alt Msg
}
type CompleteRequest struct {
	Id    int
	Cwd   string
	Input string
	Pos   int
}
type CompleteResponse struct {
	Id          int
	Error       string
	Pos         int
	Completions []string
}
type RunRequest struct {
	Cell int
	Cwd  string
	Argv []string
}
type KeyEvent struct {
	Cell int
	Keys string
}
type RowSpans struct {
	Row   int
	Spans []Span
}
type Span struct {
	Attr int
	Text string
}
type Cursor struct {
	Row    int
	Col    int
	Hidden bool
}
type TermUpdate struct {
	Rows   []RowSpans
	Cursor Cursor
}
type Pair struct {
	Key string
	Val string
}
type Hello struct {
	Alias []Pair
	Env   []Pair
}
type CmdError struct {
	Error string
}
type Exit struct {
	ExitCode int
}
type Output struct {
	// CmdError, TermUpdate, Exit
	Alt Msg
}
type CellOutput struct {
	Cell   int
	Output Output
}
type ServerMsg struct {
	// Hello, CompleteResponse, CellOutput
	Alt Msg
}

func (msg *ClientMessage) Write(w io.Writer) error {
	switch alt := msg.Alt.(type) {
	case *CompleteRequest:
		if err := WriteUint8(w, 1); err != nil {
			return err
		}
		return alt.Write(w)
	case *RunRequest:
		if err := WriteUint8(w, 2); err != nil {
			return err
		}
		return alt.Write(w)
	case *KeyEvent:
		if err := WriteUint8(w, 3); err != nil {
			return err
		}
		return alt.Write(w)
	}
	panic("notimpl")
}
func (msg *CompleteRequest) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Id); err != nil {
		return err
	}
	if err := WriteString(w, msg.Cwd); err != nil {
		return err
	}
	if err := WriteString(w, msg.Input); err != nil {
		return err
	}
	if err := WriteInt(w, msg.Pos); err != nil {
		return err
	}
	return nil
}
func (msg *CompleteResponse) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Id); err != nil {
		return err
	}
	if err := WriteString(w, msg.Error); err != nil {
		return err
	}
	if err := WriteInt(w, msg.Pos); err != nil {
		return err
	}
	if err := WriteInt(w, len(msg.Completions)); err != nil {
		return err
	}
	for _, val := range msg.Completions {
		if err := WriteString(w, val); err != nil {
			return err
		}
	}
	return nil
}
func (msg *RunRequest) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Cell); err != nil {
		return err
	}
	if err := WriteString(w, msg.Cwd); err != nil {
		return err
	}
	if err := WriteInt(w, len(msg.Argv)); err != nil {
		return err
	}
	for _, val := range msg.Argv {
		if err := WriteString(w, val); err != nil {
			return err
		}
	}
	return nil
}
func (msg *KeyEvent) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Cell); err != nil {
		return err
	}
	if err := WriteString(w, msg.Keys); err != nil {
		return err
	}
	return nil
}
func (msg *RowSpans) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Row); err != nil {
		return err
	}
	if err := WriteInt(w, len(msg.Spans)); err != nil {
		return err
	}
	for _, val := range msg.Spans {
		if err := val.Write(w); err != nil {
			return err
		}
	}
	return nil
}
func (msg *Span) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Attr); err != nil {
		return err
	}
	if err := WriteString(w, msg.Text); err != nil {
		return err
	}
	return nil
}
func (msg *Cursor) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Row); err != nil {
		return err
	}
	if err := WriteInt(w, msg.Col); err != nil {
		return err
	}
	if err := WriteBoolean(w, msg.Hidden); err != nil {
		return err
	}
	return nil
}
func (msg *TermUpdate) Write(w io.Writer) error {
	if err := WriteInt(w, len(msg.Rows)); err != nil {
		return err
	}
	for _, val := range msg.Rows {
		if err := val.Write(w); err != nil {
			return err
		}
	}
	if err := msg.Cursor.Write(w); err != nil {
		return err
	}
	return nil
}
func (msg *Pair) Write(w io.Writer) error {
	if err := WriteString(w, msg.Key); err != nil {
		return err
	}
	if err := WriteString(w, msg.Val); err != nil {
		return err
	}
	return nil
}
func (msg *Hello) Write(w io.Writer) error {
	if err := WriteInt(w, len(msg.Alias)); err != nil {
		return err
	}
	for _, val := range msg.Alias {
		if err := val.Write(w); err != nil {
			return err
		}
	}
	if err := WriteInt(w, len(msg.Env)); err != nil {
		return err
	}
	for _, val := range msg.Env {
		if err := val.Write(w); err != nil {
			return err
		}
	}
	return nil
}
func (msg *CmdError) Write(w io.Writer) error {
	if err := WriteString(w, msg.Error); err != nil {
		return err
	}
	return nil
}
func (msg *Exit) Write(w io.Writer) error {
	if err := WriteInt(w, msg.ExitCode); err != nil {
		return err
	}
	return nil
}
func (msg *Output) Write(w io.Writer) error {
	switch alt := msg.Alt.(type) {
	case *CmdError:
		if err := WriteUint8(w, 1); err != nil {
			return err
		}
		return alt.Write(w)
	case *TermUpdate:
		if err := WriteUint8(w, 2); err != nil {
			return err
		}
		return alt.Write(w)
	case *Exit:
		if err := WriteUint8(w, 3); err != nil {
			return err
		}
		return alt.Write(w)
	}
	panic("notimpl")
}
func (msg *CellOutput) Write(w io.Writer) error {
	if err := WriteInt(w, msg.Cell); err != nil {
		return err
	}
	if err := msg.Output.Write(w); err != nil {
		return err
	}
	return nil
}
func (msg *ServerMsg) Write(w io.Writer) error {
	switch alt := msg.Alt.(type) {
	case *Hello:
		if err := WriteUint8(w, 1); err != nil {
			return err
		}
		return alt.Write(w)
	case *CompleteResponse:
		if err := WriteUint8(w, 2); err != nil {
			return err
		}
		return alt.Write(w)
	case *CellOutput:
		if err := WriteUint8(w, 3); err != nil {
			return err
		}
		return alt.Write(w)
	}
	panic("notimpl")
}
func (msg *ClientMessage) Read(r *bufio.Reader) error {
	alt, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch alt {
	case 1:
		var val CompleteRequest
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	case 2:
		var val RunRequest
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	case 3:
		var val KeyEvent
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	default:
		return fmt.Errorf("bad tag %d when reading ClientMessage", alt)
	}
}
func (msg *CompleteRequest) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Id, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Cwd, err = ReadString(r)
	if err != nil {
		return err
	}
	msg.Input, err = ReadString(r)
	if err != nil {
		return err
	}
	msg.Pos, err = ReadInt(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *CompleteResponse) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Id, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Error, err = ReadString(r)
	if err != nil {
		return err
	}
	msg.Pos, err = ReadInt(r)
	if err != nil {
		return err
	}
	{
		n, err := ReadInt(r)
		if err != nil {
			return err
		}
		var val string
		for i := 0; i < n; i++ {
			val, err = ReadString(r)
			if err != nil {
				return err
			}
			msg.Completions = append(msg.Completions, val)
		}
	}
	return nil
}
func (msg *RunRequest) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Cell, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Cwd, err = ReadString(r)
	if err != nil {
		return err
	}
	{
		n, err := ReadInt(r)
		if err != nil {
			return err
		}
		var val string
		for i := 0; i < n; i++ {
			val, err = ReadString(r)
			if err != nil {
				return err
			}
			msg.Argv = append(msg.Argv, val)
		}
	}
	return nil
}
func (msg *KeyEvent) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Cell, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Keys, err = ReadString(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *RowSpans) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Row, err = ReadInt(r)
	if err != nil {
		return err
	}
	{
		n, err := ReadInt(r)
		if err != nil {
			return err
		}
		var val Span
		for i := 0; i < n; i++ {
			if err := val.Read(r); err != nil {
				return err
			}
			msg.Spans = append(msg.Spans, val)
		}
	}
	return nil
}
func (msg *Span) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Attr, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Text, err = ReadString(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *Cursor) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Row, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Col, err = ReadInt(r)
	if err != nil {
		return err
	}
	msg.Hidden, err = ReadBoolean(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *TermUpdate) Read(r *bufio.Reader) error {
	var err error
	err = err
	{
		n, err := ReadInt(r)
		if err != nil {
			return err
		}
		var val RowSpans
		for i := 0; i < n; i++ {
			if err := val.Read(r); err != nil {
				return err
			}
			msg.Rows = append(msg.Rows, val)
		}
	}
	if err := msg.Cursor.Read(r); err != nil {
		return err
	}
	return nil
}
func (msg *Pair) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Key, err = ReadString(r)
	if err != nil {
		return err
	}
	msg.Val, err = ReadString(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *Hello) Read(r *bufio.Reader) error {
	var err error
	err = err
	{
		n, err := ReadInt(r)
		if err != nil {
			return err
		}
		var val Pair
		for i := 0; i < n; i++ {
			if err := val.Read(r); err != nil {
				return err
			}
			msg.Alias = append(msg.Alias, val)
		}
	}
	{
		n, err := ReadInt(r)
		if err != nil {
			return err
		}
		var val Pair
		for i := 0; i < n; i++ {
			if err := val.Read(r); err != nil {
				return err
			}
			msg.Env = append(msg.Env, val)
		}
	}
	return nil
}
func (msg *CmdError) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Error, err = ReadString(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *Exit) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.ExitCode, err = ReadInt(r)
	if err != nil {
		return err
	}
	return nil
}
func (msg *Output) Read(r *bufio.Reader) error {
	alt, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch alt {
	case 1:
		var val CmdError
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	case 2:
		var val TermUpdate
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	case 3:
		var val Exit
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	default:
		return fmt.Errorf("bad tag %d when reading Output", alt)
	}
}
func (msg *CellOutput) Read(r *bufio.Reader) error {
	var err error
	err = err
	msg.Cell, err = ReadInt(r)
	if err != nil {
		return err
	}
	if err := msg.Output.Read(r); err != nil {
		return err
	}
	return nil
}
func (msg *ServerMsg) Read(r *bufio.Reader) error {
	alt, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch alt {
	case 1:
		var val Hello
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	case 2:
		var val CompleteResponse
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	case 3:
		var val CellOutput
		if err := val.Read(r); err != nil {
			return err
		}
		msg.Alt = &val
		return nil
	default:
		return fmt.Errorf("bad tag %d when reading ServerMsg", alt)
	}
}
