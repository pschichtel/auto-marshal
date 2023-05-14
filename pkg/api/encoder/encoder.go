package encoder

import (
	"bytes"
	"github.com/mailru/easyjson/jwriter"
	"io"
)

type JsonEncoder[T any] func(*T, *jwriter.Writer) error

type JsonEncodingError string

func (s JsonEncodingError) Error() string {
	return string(s)
}

func EncodeJson[T any](value *T, encoder JsonEncoder[T]) ([]byte, error) {
	writer := jwriter.Writer{}
	err := encoder(value, &writer)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	_, err = writer.Buffer.DumpTo(io.Writer(&out))
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
