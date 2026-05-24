package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type API struct {
	aof *Aof
}

func NewAPI(aof *Aof) *API {
	return &API{aof: aof}
}

var writeCommands = map[string]bool{
	"SET": true, "HSET": true, "HDEL": true, "DEL": true, "INCR": true, "DECR": true,
	"INCRBY": true, "DECRBY": true, "APPEND": true, "LPOP": true,
	"RPOP": true, "LPUSH": true, "RPUSH": true,
}

func (api *API) writeAof(command string, args []Value) {
	if !writeCommands[command] {
		return
	}
	value := Value{
		typ:   "array",
		array: append([]Value{{typ: "bulk", bulk: command}}, args...),
	}
	api.aof.Write(value)
}

func (api *API) exec(w http.ResponseWriter, command string, args []Value) {
	api.writeAof(command, args)

	handler, ok := Handlers[command]
	if !ok {
		http.Error(w, "unknown command", http.StatusBadRequest)
		return
	}

	result := handler(args)
	writeValue(w, result)
}

func writeValue(w http.ResponseWriter, v Value) {
	switch v.typ {
	case "string":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(v.str))
	case "bulk":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(v.bulk))
	case "null":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("null"))
	case "error":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(v.str))
	case "array":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		result := make([]string, len(v.array))
		for i, item := range v.array {
			if item.typ == "bulk" {
				result[i] = item.bulk
			} else {
				result[i] = item.str
			}
		}
		json.NewEncoder(w).Encode(result)
	default:
		http.Error(w, "unknown response type", http.StatusInternalServerError)
	}
}

func readBody(w http.ResponseWriter, r *http.Request) (string, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return "", false
	}
	return string(body), true
}

func (api *API) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PONG"))
}

func (api *API) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, ok := readBody(w, r)
	if !ok {
		return
	}
	api.exec(w, "SET", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: value}})
}

func (api *API) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	api.exec(w, "GET", []Value{{typ: "bulk", bulk: key}})
}

func (api *API) handleDel(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	api.exec(w, "DEL", []Value{{typ: "bulk", bulk: key}})
}

func (api *API) handleIncr(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	api.exec(w, "INCR", []Value{{typ: "bulk", bulk: key}})
}

func (api *API) handleDecr(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	api.exec(w, "DECR", []Value{{typ: "bulk", bulk: key}})
}

func (api *API) handleIncrBy(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	amount, ok := readBody(w, r)
	if !ok {
		return
	}
	api.exec(w, "INCRBY", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: amount}})
}

func (api *API) handleDecrBy(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	amount, ok := readBody(w, r)
	if !ok {
		return
	}
	api.exec(w, "DECRBY", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: amount}})
}

func (api *API) handleAppend(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, ok := readBody(w, r)
	if !ok {
		return
	}
	api.exec(w, "APPEND", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: value}})
}

func (api *API) handleListPush(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	side := r.URL.Query().Get("side")
	if side == "" {
		side = "left"
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var values []string
	if err := json.Unmarshal(body, &values); err != nil {
		http.Error(w, "body must be a JSON array of strings", http.StatusBadRequest)
		return
	}

	args := []Value{{typ: "bulk", bulk: key}}
	for _, v := range values {
		args = append(args, Value{typ: "bulk", bulk: v})
	}

	command := "LPUSH"
	if side == "right" {
		command = "RPUSH"
	}

	api.exec(w, command, args)
}

func (api *API) handleListRange(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" {
		startStr = "0"
	}
	if endStr == "" {
		endStr = "-1"
	}

	api.exec(w, "LRANGE", []Value{
		{typ: "bulk", bulk: key},
		{typ: "bulk", bulk: startStr},
		{typ: "bulk", bulk: endStr},
	})
}

func (api *API) handleListPop(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	side := r.URL.Query().Get("side")
	if side == "" {
		side = "left"
	}

	command := "LPOP"
	if side == "right" {
		command = "RPOP"
	}

	api.exec(w, command, []Value{{typ: "bulk", bulk: key}})
}

func (api *API) handleHashSet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	field := r.PathValue("field")
	value, ok := readBody(w, r)
	if !ok {
		return
	}
	api.exec(w, "HSET", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: field}, {typ: "bulk", bulk: value}})
}

func (api *API) handleHashGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	field := r.PathValue("field")
	api.exec(w, "HGET", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: field}})
}

func (api *API) handleHashGetAll(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	api.exec(w, "HGETALL", []Value{{typ: "bulk", bulk: key}})
}

func (api *API) handleHashDel(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	field := r.PathValue("field")
	api.exec(w, "HDEL", []Value{{typ: "bulk", bulk: key}, {typ: "bulk", bulk: field}})
}

func (api *API) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", api.handlePing)

	mux.HandleFunc("PUT /kv/{key}", api.handleSet)
	mux.HandleFunc("GET /kv/{key}", api.handleGet)
	mux.HandleFunc("DELETE /kv/{key}", api.handleDel)
	mux.HandleFunc("POST /kv/{key}/_incr", api.handleIncr)
	mux.HandleFunc("POST /kv/{key}/_decr", api.handleDecr)
	mux.HandleFunc("POST /kv/{key}/_incrby", api.handleIncrBy)
	mux.HandleFunc("POST /kv/{key}/_decrby", api.handleDecrBy)
	mux.HandleFunc("POST /kv/{key}/_append", api.handleAppend)

	mux.HandleFunc("PUT /list/{key}", api.handleListPush)
	mux.HandleFunc("GET /list/{key}", api.handleListRange)
	mux.HandleFunc("POST /list/{key}/_pop", api.handleListPop)

	mux.HandleFunc("PUT /hash/{key}/{field}", api.handleHashSet)
	mux.HandleFunc("GET /hash/{key}/{field}", api.handleHashGet)
	mux.HandleFunc("GET /hash/{key}", api.handleHashGetAll)
	mux.HandleFunc("DELETE /hash/{key}/{field}", api.handleHashDel)

	fmt.Println("HTTP API listening on :8080")
	http.ListenAndServe(":8080", mux)
}
