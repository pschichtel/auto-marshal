package main

import (
	"github.com/mailru/easyjson/jlexer"
	"github.com/pschichtel/auto-marshal/pkg/api/decoder"
	"os"
)

type LinkedList struct {
	Value *Attribute
	Next  *LinkedList
}

type Attribute interface {
	AttributeMarkerFunction()
}

type AmountAttribute struct {
	Amount int32
}

func (amount AmountAttribute) AttributeMarkerFunction() {

}

type NameAttribute struct {
	Name string
}

func (name NameAttribute) AttributeMarkerFunction() {

}

func JsonDecodeLinkedList(value *LinkedList, lexer *jlexer.Lexer) error {
	lexer.Delim('{')
	err := lexer.Error()
	if err != nil {
		return err
	}
	for {
		fieldName := lexer.String()
		lexer.Delim(':')

		switch fieldName {
		case "value":
			var tmp Attribute = NameAttribute{}
			value.Value = &tmp
		default:

		}

		if lexer.IsDelim('}') {
			break
		}
	}
	return nil
}

func main() {

	list := LinkedList{}
	json := "{\"value\":{\"type\":\"NameAttribute\",\"name\":\"a\"},\"next\":{\"value\":{\"type\":\"AmountAttribute\",\"amount\":5},\"next\":null}}"

	err := decoder.DecodeJson([]byte(json), &list, JsonDecodeLinkedList)
	if err != nil {
		println("error", err.Error())
		os.Exit(1)
		return
	}
}
