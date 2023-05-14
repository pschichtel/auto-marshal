package encoder

import (
	"bytes"
	"github.com/mailru/easyjson/jwriter"
	"io"
)

type JsonEncoder[T any] func(*T, *JsonWriter) error

type JsonWriter struct {
	writer jwriter.Writer
}

type JsonEncodingError string

func (s JsonEncodingError) Error() string {
	return string(s)
}

func (writer *JsonWriter) BeginObject() {
	writer.writer.RawString("{")
}

func (writer *JsonWriter) EndObject() {
	writer.writer.RawString("}")
}

func (writer *JsonWriter) NullValue() {
	writer.writer.RawString("null")
}

func (writer *JsonWriter) String(value string) {
	writer.writer.String(value)
}

func (writer *JsonWriter) Int32(value int32) {
	writer.writer.Int32(value)
}

func (writer *JsonWriter) ValueSeparator() {
	writer.writer.RawString(":")
}

func (writer *JsonWriter) FieldSeparator() {
	writer.writer.RawString(",")
}

func EncodeJson[T any](value *T, encoder JsonEncoder[T]) ([]byte, error) {
	writer := JsonWriter{}
	err := encoder(value, &writer)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	_, err = writer.writer.Buffer.DumpTo(io.Writer(&out))
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
