package main

import (
	"github.com/mailru/easyjson/jwriter"
	"github.com/pschichtel/auto-marshal/pkg/api/encoder"
	"os"
	"reflect"
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

func JsonEncodeLinkedListFields(value *LinkedList, writer *jwriter.Writer, first bool) error {
	if !first {
		writer.RawString(",")
	}
	writer.String("value")
	writer.RawString(":")
	err := JsonEncodeAttribute(value.Value, writer)
	if err != nil {
		return err
	}
	writer.RawString(",")
	writer.String("next")
	writer.RawString(":")
	err = JsonEncodeLinkedList(value.Next, writer)
	if err != nil {
		return err
	}
	return nil
}

func JsonEncodeLinkedList(value *LinkedList, writer *jwriter.Writer) error {
	if value == nil {
		writer.RawString("null")
		return nil
	}

	writer.RawString("{")
	err := JsonEncodeLinkedListFields(value, writer, true)
	if err != nil {
		return err
	}
	writer.RawString("}")
	return nil
}

func JsonEncodeNameAttributeFields(value *NameAttribute, writer *jwriter.Writer, first bool) {
	if !first {
		writer.RawString(",")
	}
	writer.String("name")
	writer.RawString(":")
	writer.String(value.Name)
}

func JsonEncodeNameAttribute(value *NameAttribute, writer *jwriter.Writer) {
	if value == nil {
		writer.RawString("null")
		return
	}

	writer.RawString("{")
	JsonEncodeNameAttributeFields(value, writer, true)
	writer.RawString("}")
}

func JsonEncodeAmountAttributeFields(value *AmountAttribute, writer *jwriter.Writer, first bool) {
	if !first {
		writer.RawString(",")
	}
	writer.String("amount")
	writer.RawString(":")
	writer.Int32(value.Amount)
}

func JsonEncodeAmountAttribute(value *AmountAttribute, writer *jwriter.Writer) {
	if value == nil {
		writer.RawString("null")
		return
	}

	writer.RawString("{")
	JsonEncodeAmountAttributeFields(value, writer, true)
	writer.RawString("}")
}

func JsonEncodeAttribute(value *Attribute, writer *jwriter.Writer) error {
	if value == nil {
		writer.RawString("null")
		return nil
	}

	writer.RawString("{")
	writer.String("type")
	writer.RawString(":")
	switch kind := (*value).(type) {
	case NameAttribute:
		writer.String("NameAttribute")
		JsonEncodeNameAttributeFields(&kind, writer, false)
	case AmountAttribute:
		writer.String("AmountAttribute")
		JsonEncodeAmountAttributeFields(&kind, writer, false)
	default:
		return encoder.JsonEncodingError("Unknown interface type: " + reflect.TypeOf(kind).Name())
	}
	writer.RawString("}")

	return nil
}

func main() {
	var name Attribute = NameAttribute{Name: "a"}
	var amount Attribute = AmountAttribute{Amount: 5}

	list := LinkedList{Value: &name, Next: &LinkedList{Value: &amount}}

	jsonBytes, err := encoder.EncodeJson(&list, JsonEncodeLinkedList)
	if err != nil {
		println("error", err.Error())
		os.Exit(1)
		return
	}

	println("json", string(jsonBytes))
}
