// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package proto

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type RespOutput struct {
	_tab flatbuffers.Table
}

func GetRootAsRespOutput(buf []byte, offset flatbuffers.UOffsetT) *RespOutput {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &RespOutput{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *RespOutput) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *RespOutput) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *RespOutput) Text() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func RespOutputStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func RespOutputAddText(builder *flatbuffers.Builder, text flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(text), 0)
}
func RespOutputEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
