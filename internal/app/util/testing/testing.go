package main

import (
	"github.com/pschichtel/auto-marshal/pkg/api/encoder"
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

func JsonEncodeLinkedListFields(value *LinkedList, writer *encoder.JsonWriter, first bool) error {
	if !first {
		writer.FieldSeparator()
	}
	writer.String("value")
	writer.ValueSeparator()
	err := JsonEncodeAttribute(value.Value, writer)
	if err != nil {
		return err
	}
	writer.FieldSeparator()
	writer.String("next")
	writer.ValueSeparator()
	err = JsonEncodeLinkedList(value.Next, writer)
	if err != nil {
		return err
	}
	return nil
}

func JsonEncodeLinkedList(value *LinkedList, writer *encoder.JsonWriter) error {
	if value == nil {
		writer.NullValue()
		return nil
	}

	writer.BeginObject()
	err := JsonEncodeLinkedListFields(value, writer, true)
	if err != nil {
		return err
	}
	writer.EndObject()
	return nil
}

func JsonEncodeNameAttributeFields(value *NameAttribute, writer *encoder.JsonWriter, first bool) {
	if !first {
		writer.FieldSeparator()
	}
	writer.String("name")
	writer.ValueSeparator()
	writer.String(value.Name)
}

func JsonEncodeNameAttribute(value *NameAttribute, writer *encoder.JsonWriter) {
	if value == nil {
		writer.NullValue()
		return
	}

	writer.BeginObject()
	JsonEncodeNameAttributeFields(value, writer, true)
	writer.EndObject()
}

func JsonEncodeAmountAttributeFields(value *AmountAttribute, writer *encoder.JsonWriter, first bool) {
	if !first {
		writer.FieldSeparator()
	}
	writer.String("amount")
	writer.ValueSeparator()
	writer.Int32(value.Amount)
}

func JsonEncodeAmountAttribute(value *AmountAttribute, writer *encoder.JsonWriter) {
	if value == nil {
		writer.NullValue()
		return
	}

	writer.BeginObject()
	JsonEncodeAmountAttributeFields(value, writer, true)
	writer.EndObject()
}

func JsonEncodeAttribute(value *Attribute, writer *encoder.JsonWriter) error {
	if value == nil {
		writer.NullValue()
		return nil
	}

	writer.BeginObject()
	writer.String("type")
	writer.ValueSeparator()
	switch kind := (*value).(type) {
	case NameAttribute:
		writer.String("NameAttribute")
		JsonEncodeNameAttributeFields(&kind, writer, false)
	case AmountAttribute:
		writer.String("AmountAttribute")
		JsonEncodeAmountAttributeFields(&kind, writer, false)
	default:
		return encoder.JsonEncodingError("Unknown interface type!")
	}
	writer.EndObject()

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
