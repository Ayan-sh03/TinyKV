package main

import (
	"fmt"
	"testing"
)

func resetHashTable() {
	hashTable = NewHashTable(100)
}

func TestHashTableSetAndGet(t *testing.T) {
	resetHashTable()

	hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}, {typ: "bulk", bulk: "v"}})

	got := hgetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})
	if got.typ != "bulk" || got.bulk != "v" {
		t.Errorf("CHGET = %+v, want v", got)
	}
}

func TestHashTableGetMissing(t *testing.T) {
	resetHashTable()

	got := hgetHT([]Value{{typ: "bulk", bulk: "nokey"}, {typ: "bulk", bulk: "nofield"}})
	if got.typ != "null" {
		t.Errorf("CHGET nonexistent = %+v, want null", got)
	}
}

func TestHashTableSetOverwrite(t *testing.T) {
	resetHashTable()

	hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}, {typ: "bulk", bulk: "v1"}})
	hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}, {typ: "bulk", bulk: "v2"}})

	got := hgetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})
	if got.bulk != "v2" {
		t.Errorf("after overwrite, CHGET = %v, want v2", got.bulk)
	}
}

func TestHashTableDelete(t *testing.T) {
	resetHashTable()

	hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}, {typ: "bulk", bulk: "v"}})
	hdelHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})

	got := hgetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})
	if got.typ != "null" {
		t.Errorf("after CHDEL, CHGET = %+v, want null", got)
	}
}

func TestHashTableGetAll(t *testing.T) {
	resetHashTable()

	hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f1"}, {typ: "bulk", bulk: "v1"}})
	hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f2"}, {typ: "bulk", bulk: "v2"}})

	got := hgetallHT([]Value{{typ: "bulk", bulk: "h"}})
	if got.typ != "array" {
		t.Fatalf("CHGETALL = %+v, want array", got)
	}
	if len(got.array) != 4 {
		t.Errorf("CHGETALL array length = %d, want 4", len(got.array))
	}
}

func TestHashTableGetAllMissing(t *testing.T) {
	resetHashTable()

	got := hgetallHT([]Value{{typ: "bulk", bulk: "nohash"}})
	if got.typ != "null" {
		t.Errorf("CHGETALL nonexistent = %+v, want null", got)
	}
}

func TestHashTableMultipleFields(t *testing.T) {
	resetHash()

	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "a"}, {typ: "bulk", bulk: "1"}})
	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "b"}, {typ: "bulk", bulk: "2"}})
	hset([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "c"}, {typ: "bulk", bulk: "3"}})

	got := hgetall([]Value{{typ: "bulk", bulk: "h"}})
	if got.typ != "array" {
		t.Fatalf("HGETALL = %+v, want array", got)
	}
	if len(got.array) != 6 {
		t.Errorf("HGETALL array length = %d, want 6 (3 fields * 2)", len(got.array))
	}
}

func TestHashTableSetWrongArgs(t *testing.T) {
	got := hsetHT([]Value{{typ: "bulk", bulk: "h"}, {typ: "bulk", bulk: "f"}})
	if got.typ != "error" {
		t.Errorf("CHSET with 2 args = %+v, want error", got)
	}
}

func TestHashTableGetWrongArgs(t *testing.T) {
	got := hgetHT([]Value{{typ: "bulk", bulk: "h"}})
	if got.typ != "error" {
		t.Errorf("CHGET with 1 arg = %+v, want error", got)
	}
}

func TestHashTableDeleteWrongArgs(t *testing.T) {
	got := hdelHT([]Value{{typ: "bulk", bulk: "h"}})
	if got.typ != "error" {
		t.Errorf("CHDEL with 1 arg = %+v, want error", got)
	}
}

func TestHashTableGetAllWrongArgs(t *testing.T) {
	got := hgetallHT([]Value{})
	if got.typ != "error" {
		t.Errorf("CHGETALL with no args = %+v, want error", got)
	}
}

func TestHashTableResize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping resize test in short mode")
	}
	resetHashTable()

	for i := 0; i < 20; i++ {
		hsetHT([]Value{
			{typ: "bulk", bulk: fmt.Sprintf("hash%d", i)},
			{typ: "bulk", bulk: "field"},
			{typ: "bulk", bulk: "value"},
		})
	}

	got := hgetHT([]Value{{typ: "bulk", bulk: "hash0"}, {typ: "bulk", bulk: "field"}})
	if got.typ != "bulk" || got.bulk != "value" {
		t.Errorf("after resize, CHGET = %+v, want value", got)
	}
}
