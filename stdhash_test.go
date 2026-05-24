package main

import (
	"testing"
)

func resetHash() {
	HSETsMu.Lock()
	for k := range HSETs {
		delete(HSETs, k)
	}
	HSETsMu.Unlock()
}

func TestHsetAndHget(t *testing.T) {
	resetHash()

	result := hset([]Value{{typ: "bulk", bulk: "myhash"}, {typ: "bulk", bulk: "field1"}, {typ: "bulk", bulk: "value1"}})
	if result.typ != "string" || result.str != "OK" {
		t.Fatalf("HSET returned %+v, want OK", result)
	}

	got := hget([]Value{{typ: "bulk", bulk: "myhash"}, {typ: "bulk", bulk: "field1"}})
	if got.typ != "bulk" || got.bulk != "value1" {
		t.Errorf("HGET myhash field1 = %+v, want value1", got)
	}
}

func TestHgetMissing(t *testing.T) {
	resetHash()

	got := hget([]Value{{typ: "bulk", bulk: "nohash"}, {typ: "bulk", bulk: "nofield"}})
	if got.typ != "null" {
		t.Errorf("HGET nonexistent = %+v, want null", got)
	}
}

func TestHsetOverwrite(t *testing.T) {
	resetHash()

	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}, {typ: "bulk", bulk: "v1"}})
	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}, {typ: "bulk", bulk: "v2"}})

	got := hget([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})
	if got.bulk != "v2" {
		t.Errorf("after overwrite, HGET = %v, want v2", got.bulk)
	}
}

func TestHgetall(t *testing.T) {
	resetHash()

	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f1"}, {typ: "bulk", bulk: "v1"}})
	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f2"}, {typ: "bulk", bulk: "v2"}})

	got := hgetall([]Value{{typ: "bulk", bulk: "h"}})
	if got.typ != "array" {
		t.Fatalf("HGETALL = %+v, want array", got)
	}
	if len(got.array) != 4 {
		t.Errorf("HGETALL array length = %d, want 4", len(got.array))
	}
}

func TestHgetallMissing(t *testing.T) {
	resetHash()

	got := hgetall([]Value{{typ: "bulk", bulk: "nohash"}})
	if got.typ != "null" {
		t.Errorf("HGETALL nonexistent = %+v, want null", got)
	}
}

func TestHsetWrongArgs(t *testing.T) {
	got := hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})
	if got.typ != "error" {
		t.Errorf("HSET with 2 args = %+v, want error", got)
	}
}

func TestHgetWrongArgs(t *testing.T) {
	got := hget([]Value{{typ: "bulk", bulk: "h"}})
	if got.typ != "error" {
		t.Errorf("HGET with 1 arg = %+v, want error", got)
	}
}

func TestHgetallWrongArgs(t *testing.T) {
	got := hgetall([]Value{})
	if got.typ != "error" {
		t.Errorf("HGETALL with no args = %+v, want error", got)
	}
}

func TestHdel(t *testing.T) {
	resetHash()

	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f1"}, {typ: "bulk", bulk: "v1"}})
	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f2"}, {typ: "bulk", bulk: "v2"}})

	result := hdel([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f1"}})
	if result.typ != "string" || result.str != "OK" {
		t.Fatalf("HDEL returned %+v, want OK", result)
	}

	got := hget([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f1"}})
	if got.typ != "null" {
		t.Errorf("after HDEL, HGET f1 = %+v, want null", got)
	}

	got = hget([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f2"}})
	if got.typ != "bulk" || got.bulk != "v2" {
		t.Errorf("after HDEL f1, HGET f2 = %+v, want v2", got)
	}
}

func TestHdelMissingHash(t *testing.T) {
	resetHash()

	result := hdel([]Value{{typ: "bulk", bulk: "nohash"}, {typ: "bulk", bulk: "nofield"}})
	if result.typ != "string" || result.str != "OK" {
		t.Errorf("HDEL on nonexistent hash = %+v, want OK", result)
	}
}

func TestHdelWrongArgs(t *testing.T) {
	got := hdel([]Value{})
	if got.typ != "error" {
		t.Errorf("HDEL with no args = %+v, want error", got)
	}
}
