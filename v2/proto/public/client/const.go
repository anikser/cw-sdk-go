package ProtobufClient

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	"github.com/juju/errors"
)

type SerializationFormat int

const (
	ProtobufSerialization SerializationFormat = iota
	JSONSerialization
)

func (sf SerializationFormat) String() string {
	if sf == ProtobufSerialization {
		return "protobuf"
	}
	if sf == JSONSerialization {
		return "json"
	}
	return "invalid"
}

// Shared jsonpb Marshaler/Unmarshaler to be used by other packages
var JSONMarshaler = &jsonpb.Marshaler{}
var JSONUnmarshaler = &jsonpb.Unmarshaler{}

// Convenience function that takes care of the bytes.Buffer internally
func MarshalJSON(msg proto.Message) ([]byte, error) {
	var b bytes.Buffer
	err := JSONMarshaler.Marshal(&b, msg)
	return b.Bytes(), errors.Trace(err)
}
