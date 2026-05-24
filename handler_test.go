package main

import (
	"sync"
	"testing"
)

func resetStrings() {
	SETsMu.Lock()
	for k := range SETs {
		delete(SETs, k)
	}
	SETsMu.Unlock()
	INCRsMU.Lock()
	for k := range SETs {
		delete(SETs, k)
	}
	INCRsMU.Unlock()
}

func TestPing(t *testing.T) {
	got := ping([]Value{})
	if got.typ != "string" || got.str != "PONG" {
		t.Errorf("ping() = %+v, want PONG", got)
	}

	got = ping([]Value{{typ: "bulk", bulk: "hello"}})
	if got.str != "hello" {
		t.Errorf("ping(hello) = %v, want hello", got.str)
	}
}

func TestSetAndGet(t *testing.T) {
	resetStrings()

	result := set([]Value{{typ: "bulk", bulk: "mykey"}, {typ: "bulk", bulk: "myval"}})
	if result.typ != "string" || result.str != "OK" {
		t.Fatalf("SET returned %+v, want OK", result)
	}

	got := get([]Value{{typ: "bulk", bulk: "mykey"}})
	if got.typ != "bulk" || got.bulk != "myval" {
		t.Errorf("GET mykey = %+v, want myval", got)
	}
}

func TestGetMissing(t *testing.T) {
	resetStrings()

	got := get([]Value{{typ: "bulk", bulk: "nonexistent"}})
	if got.typ != "null" {
		t.Errorf("GET nonexistent = %+v, want null", got)
	}
}

func TestSetOverwrite(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "v1"}})
	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "v2"}})

	got := get([]Value{{typ: "bulk", bulk: "k"}})
	if got.bulk != "v2" {
		t.Errorf("after overwrite, GET k = %v, want v2", got.bulk)
	}
}

func TestDel(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "v"}})
	del([]Value{{typ: "bulk", bulk: "k"}})

	got := get([]Value{{typ: "bulk", bulk: "k"}})
	if got.typ != "null" {
		t.Errorf("after DEL, GET k = %+v, want null", got)
	}
}

func TestAppend(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "hello"}})
	appendto([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: " world"}})

	got := get([]Value{{typ: "bulk", bulk: "k"}})
	if got.bulk != "hello world" {
		t.Errorf("after APPEND, GET k = %v, want 'hello world'", got.bulk)
	}
}

func TestIncr(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})
	incr([]Value{{typ: "bulk", bulk: "counter"}})

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "11" {
		t.Errorf("after INCR, GET counter = %v, want 11", got.bulk)
	}
}

func TestDecr(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})
	decr([]Value{{typ: "bulk", bulk: "counter"}})

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "9" {
		t.Errorf("after DECR, GET counter = %v, want 9", got.bulk)
	}
}

func TestIncrBy(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})
	incrBy([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "5"}})

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "15" {
		t.Errorf("after INCRBY 5, GET counter = %v, want 15", got.bulk)
	}
}

func TestDecrBy(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})
	decrBy([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "3"}})

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "7" {
		t.Errorf("after DECRBY 3, GET counter = %v, want 7", got.bulk)
	}
}

func TestIncrNonInteger(t *testing.T) {
	resetStrings()

	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "notanumber"}})
	got := incr([]Value{{typ: "bulk", bulk: "k"}})
	if got.typ != "error" {
		t.Errorf("INCR on non-integer = %+v, want error", got)
	}
}

func TestSetWrongArgs(t *testing.T) {
	got := set([]Value{})
	if got.typ != "error" {
		t.Errorf("SET with no args = %+v, want error", got)
	}
}

func TestGetWrongArgs(t *testing.T) {
	got := get([]Value{})
	if got.typ != "error" {
		t.Errorf("GET with no args = %+v, want error", got)
	}
}

func TestDelWrongArgs(t *testing.T) {
	got := del([]Value{})
	if got.typ != "error" {
		t.Errorf("DEL with no args = %+v, want error", got)
	}
}

func TestAppendWrongArgs(t *testing.T) {
	got := appendto([]Value{})
	if got.typ != "error" {
		t.Errorf("APPEND with no args = %+v, want error", got)
	}
}

func TestIncrWrongArgs(t *testing.T) {
	got := incr([]Value{})
	if got.typ != "error" {
		t.Errorf("INCR with no args = %+v, want error", got)
	}
}

func TestDecrWrongArgs(t *testing.T) {
	got := decr([]Value{})
	if got.typ != "error" {
		t.Errorf("DECR with no args = %+v, want error", got)
	}
}

func TestIncrByWrongArgs(t *testing.T) {
	got := incrBy([]Value{})
	if got.typ != "error" {
		t.Errorf("INCRBY with no args = %+v, want error", got)
	}
}

func TestDecrByWrongArgs(t *testing.T) {
	got := decrBy([]Value{})
	if got.typ != "error" {
		t.Errorf("DECRBY with no args = %+v, want error", got)
	}
}

func TestConcurrentSet(t *testing.T) {
	resetStrings()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			set([]Value{{typ: "bulk", bulk: "ckey"}, {typ: "bulk", bulk: "val"}})
		}(i)
	}
	wg.Wait()

	got := get([]Value{{typ: "bulk", bulk: "ckey"}})
	if got.typ != "bulk" {
		t.Errorf("after concurrent SETs, GET ckey = %+v, want bulk", got)
	}
}
