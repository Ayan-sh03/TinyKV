package main

import (
	"bytes"
	"testing"
)

func TestMarshalString(t *testing.T) {
	v := Value{typ: "string", str: "OK"}
	expected := "+OK\r\n"
	if string(v.Marshal()) != expected {
		t.Errorf("Marshal string = %q, want %q", string(v.Marshal()), expected)
	}
}

func TestMarshalBulk(t *testing.T) {
	v := Value{typ: "bulk", bulk: "hello"}
	expected := "$5\r\nhello\r\n"
	if string(v.Marshal()) != expected {
		t.Errorf("Marshal bulk = %q, want %q", string(v.Marshal()), expected)
	}
}

func TestMarshalError(t *testing.T) {
	v := Value{typ: "error", str: "ERR unknown"}
	expected := "-ERR unknown\r\n"
	if string(v.Marshal()) != expected {
		t.Errorf("Marshal error = %q, want %q", string(v.Marshal()), expected)
	}
}

func TestMarshalNull(t *testing.T) {
	v := Value{typ: "null"}
	expected := "$-1\r\n"
	if string(v.Marshal()) != expected {
		t.Errorf("Marshal null = %q, want %q", string(v.Marshal()), expected)
	}
}

func TestMarshalArray(t *testing.T) {
	v := Value{
		typ: "array",
		array: []Value{
			{typ: "bulk", bulk: "SET"},
			{typ: "bulk", bulk: "key"},
			{typ: "bulk", bulk: "val"},
		},
	}
	expected := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$3\r\nval\r\n"
	if string(v.Marshal()) != expected {
		t.Errorf("Marshal array = %q, want %q", string(v.Marshal()), expected)
	}
}

func TestReadBulk(t *testing.T) {
	input := "$5\r\nhello\r\n"
	resp := NewResp(bytes.NewBufferString(input))

	v, err := resp.Read()
	if err != nil {
		t.Fatalf("Read bulk: %v", err)
	}
	if v.typ != "bulk" || v.bulk != "hello" {
		t.Errorf("Read bulk = %+v, want {typ:bulk, bulk:hello}", v)
	}
}

func TestReadArray(t *testing.T) {
	input := "*2\r\n$3\r\nSET\r\n$3\r\nkey\r\n"
	resp := NewResp(bytes.NewBufferString(input))

	v, err := resp.Read()
	if err != nil {
		t.Fatalf("Read array: %v", err)
	}
	if v.typ != "array" {
		t.Fatalf("Read array type = %v, want array", v.typ)
	}
	if len(v.array) != 2 {
		t.Errorf("Read array length = %d, want 2", len(v.array))
	}
	if v.array[0].bulk != "SET" {
		t.Errorf("Read array[0] = %v, want SET", v.array[0].bulk)
	}
	if v.array[1].bulk != "key" {
		t.Errorf("Read array[1] = %v, want key", v.array[1].bulk)
	}
}

func TestRoundTrip(t *testing.T) {
	original := Value{
		typ: "array",
		array: []Value{
			{typ: "bulk", bulk: "SET"},
			{typ: "bulk", bulk: "mykey"},
			{typ: "bulk", bulk: "myval"},
		},
	}

	marshalled := original.Marshal()
	resp := NewResp(bytes.NewBuffer(marshalled))

	v, err := resp.Read()
	if err != nil {
		t.Fatalf("RoundTrip Read: %v", err)
	}

	if v.typ != "array" {
		t.Fatalf("RoundTrip type = %v, want array", v.typ)
	}
	if len(v.array) != 3 {
		t.Errorf("RoundTrip length = %d, want 3", len(v.array))
	}
	if v.array[0].bulk != "SET" {
		t.Errorf("RoundTrip array[0] = %v, want SET", v.array[0].bulk)
	}
	if v.array[1].bulk != "mykey" {
		t.Errorf("RoundTrip array[1] = %v, want mykey", v.array[1].bulk)
	}
	if v.array[2].bulk != "myval" {
		t.Errorf("RoundTrip array[2] = %v, want myval", v.array[2].bulk)
	}
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	v := Value{typ: "string", str: "OK"}
	err := w.Write(v)
	if err != nil {
		t.Fatalf("Writer.Write: %v", err)
	}

	if buf.String() != "+OK\r\n" {
		t.Errorf("Writer output = %q, want \"+OK\\r\\n\"", buf.String())
	}
}

func TestReadEmptyBulk(t *testing.T) {
	input := "$0\r\n\r\n"
	resp := NewResp(bytes.NewBufferString(input))

	v, err := resp.Read()
	if err != nil {
		t.Fatalf("Read empty bulk: %v", err)
	}
	if v.typ != "bulk" || v.bulk != "" {
		t.Errorf("Read empty bulk = %+v, want {typ:bulk, bulk:''}", v)
	}
}
