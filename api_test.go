package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setupTestAPI(t *testing.T) (*API, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.aof")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	aof, err := NewAof(f.Name())
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	api := NewAPI(aof)
	cleanup := func() {
		aof.Close()
		os.Remove(f.Name())
	}

	resetStrings()
	resetHash()
	for k := range SETsL {
		delete(SETsL, k)
	}

	return api, cleanup
}

func TestHandlePing(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	api.handlePing(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ping status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "PONG" {
		t.Errorf("ping body = %q, want PONG", w.Body.String())
	}
}

func TestHandleSetGet(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/kv/mykey", strings.NewReader("myval"))
	req.SetPathValue("key", "mykey")
	w := httptest.NewRecorder()
	api.handleSet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SET status = %d, want %d", w.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/kv/mykey", nil)
	req.SetPathValue("key", "mykey")
	w = httptest.NewRecorder()
	api.handleGet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "myval" {
		t.Errorf("GET body = %q, want myval", w.Body.String())
	}
}

func TestHandleGetMissing(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/kv/nonexistent", nil)
	req.SetPathValue("key", "nonexistent")
	w := httptest.NewRecorder()
	api.handleGet(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET missing status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDel(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "v"}})

	req := httptest.NewRequest(http.MethodDelete, "/kv/k", nil)
	req.SetPathValue("key", "k")
	w := httptest.NewRecorder()
	api.handleDel(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DEL status = %d, want %d", w.Code, http.StatusOK)
	}

	got := get([]Value{{typ: "bulk", bulk: "k"}})
	if got.typ != "null" {
		t.Errorf("after DEL, GET = %+v, want null", got)
	}
}

func TestHandleIncr(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})

	req := httptest.NewRequest(http.MethodPost, "/kv/counter/_incr", nil)
	req.SetPathValue("key", "counter")
	w := httptest.NewRecorder()
	api.handleIncr(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("INCR status = %d, want %d", w.Code, http.StatusOK)
	}

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "11" {
		t.Errorf("after INCR, counter = %v, want 11", got.bulk)
	}
}

func TestHandleDecr(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})

	req := httptest.NewRequest(http.MethodPost, "/kv/counter/_decr", nil)
	req.SetPathValue("key", "counter")
	w := httptest.NewRecorder()
	api.handleDecr(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DECR status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleAppend(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	set([]Value{{typ: "bulk", bulk: "k"}, {typ: "bulk", bulk: "hello"}})

	req := httptest.NewRequest(http.MethodPost, "/kv/k/_append", strings.NewReader(" world"))
	req.SetPathValue("key", "k")
	w := httptest.NewRecorder()
	api.handleAppend(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("APPEND status = %d, want %d", w.Code, http.StatusOK)
	}

	got := get([]Value{{typ: "bulk", bulk: "k"}})
	if got.bulk != "hello world" {
		t.Errorf("after APPEND, GET = %v, want 'hello world'", got.bulk)
	}
}

func TestHandleHasCRUD(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/hash/myhash/f1", strings.NewReader("v1"))
	req.SetPathValue("key", "myhash")
	req.SetPathValue("field", "f1")
	w := httptest.NewRecorder()
	api.handleHashSet(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HSET status = %d, want %d", w.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/hash/myhash/f1", nil)
	req.SetPathValue("key", "myhash")
	req.SetPathValue("field", "f1")
	w = httptest.NewRecorder()
	api.handleHashGet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HGET status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "v1" {
		t.Errorf("HGET body = %q, want v1", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/hash/myhash", nil)
	req.SetPathValue("key", "myhash")
	w = httptest.NewRecorder()
	api.handleHashGetAll(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HGETALL status = %d, want %d", w.Code, http.StatusOK)
	}

	var result []string
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("HGETALL result length = %d, want 2", len(result))
	}

	req = httptest.NewRequest(http.MethodDelete, "/hash/myhash/f1", nil)
	req.SetPathValue("key", "myhash")
	req.SetPathValue("field", "f1")
	w = httptest.NewRecorder()
	api.handleHashDel(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HDEL status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleListPushAndRange(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	body, _ := json.Marshal([]string{"a", "b", "c"})
	req := httptest.NewRequest(http.MethodPut, "/list/mylist", io.NopCloser(strings.NewReader(string(body))))
	req.SetPathValue("key", "mylist")
	w := httptest.NewRecorder()
	api.handleListPush(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("LPUSH status = %d, want %d", w.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/list/mylist?start=0&end=-1", nil)
	req.SetPathValue("key", "mylist")
	w = httptest.NewRecorder()
	api.handleListRange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("LRANGE status = %d, want %d", w.Code, http.StatusOK)
	}

	var result []string
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 3 {
		t.Errorf("LRANGE result length = %d, want 3", len(result))
	}
}

func TestHandleIncrBy(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})

	req := httptest.NewRequest(http.MethodPost, "/kv/counter/_incrby", strings.NewReader("5"))
	req.SetPathValue("key", "counter")
	w := httptest.NewRecorder()
	api.handleIncrBy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("INCRBY status = %d, want %d", w.Code, http.StatusOK)
	}

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "15" {
		t.Errorf("after INCRBY, counter = %v, want 15", got.bulk)
	}
}

func TestHandleDecrBy(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	set([]Value{{typ: "bulk", bulk: "counter"}, {typ: "bulk", bulk: "10"}})

	req := httptest.NewRequest(http.MethodPost, "/kv/counter/_decrby", strings.NewReader("3"))
	req.SetPathValue("key", "counter")
	w := httptest.NewRecorder()
	api.handleDecrBy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DECRBY status = %d, want %d", w.Code, http.StatusOK)
	}

	got := get([]Value{{typ: "bulk", bulk: "counter"}})
	if got.bulk != "7" {
		t.Errorf("after DECRBY, counter = %v, want 7", got.bulk)
	}
}
