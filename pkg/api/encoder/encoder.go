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

func JsonEncodeSlice[T any](value *[]T, writer *jwriter.Writer, childEncoder JsonEncoder[T]) error {
	if value == nil {
		writer.RawString("null")
		return nil
	}
	writer.RawByte('[')
	entryCounter := 0
	for _, child := range *value {
		if entryCounter > 0 {
			writer.RawString(",")
		}
		entryCounter++
		err := childEncoder(&child, writer)
		if err != nil {
			return err
		}
	}
	writer.RawByte(']')

	return nil
}

func JsonEncoderSlice[T any](childEncoder JsonEncoder[T]) JsonEncoder[[]T] {
	return func(value *[]T, writer *jwriter.Writer) error {
		return JsonEncodeSlice[T](value, writer, childEncoder)
	}
}

func JsonEncodeStringMap[T any](value *map[string]T, writer *jwriter.Writer, childEncoder JsonEncoder[T]) error {
	if value == nil {
		writer.RawString("null")
		return nil
	}
	writer.RawByte('{')
	entryCounter := 0
	for key, child := range *value {
		if entryCounter > 0 {
			writer.RawString(",")
		}
		entryCounter++
		writer.String(key)
		writer.RawByte(':')
		err := childEncoder(&child, writer)
		if err != nil {
			return err
		}
	}
	writer.RawByte('}')

	return nil
}

func JsonEncoderStringMap[T any](childEncoder JsonEncoder[T]) JsonEncoder[map[string]T] {
	return func(value *map[string]T, writer *jwriter.Writer) error {
		return JsonEncodeStringMap[T](value, writer, childEncoder)
	}
}
