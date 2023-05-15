package decoder

import (
	"github.com/mailru/easyjson/jlexer"
)

type JsonDecoder[T any] func(*T, *jlexer.Lexer) error

func DecodeJson[T any](json []byte, value *T, decoder JsonDecoder[T]) error {
	lexer := jlexer.Lexer{Data: json}
	err := decoder(value, &lexer)
	if err != nil {
		return err
	}

	return nil
}
